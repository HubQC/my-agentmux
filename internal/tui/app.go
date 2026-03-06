package tui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/monitor"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/tmux"
	"github.com/cqi/my_agentmux/internal/tui/components"
)

// tickMsg triggers periodic refresh.
type tickMsg time.Time

// eventMsg carries a watcher event.
type eventMsg monitor.Event

// resourceMsg carries a watcher resource event.
type resourceMsg monitor.ResourceEvent

// Model is the main bubbletea model for the dashboard.
type Model struct {
	cfg        *config.Config
	sessionMgr *session.Manager
	watcher    *monitor.Watcher
	tmuxClient *tmux.Client

	sessionTree components.SessionTree
	logViewer   components.LogViewer
	statusBar   components.StatusBar

	// Log lines per agent name.
	agentLogs map[string][]string

	// Resources per agent name.
	agentResources map[string]monitor.ResourceEvent

	// Project groups for tree grouping.
	projectGroups map[string][]string

	// Split mode
	splitMode   bool
	rightPaneID string

	width  int
	height int

	quitting bool
}

// NewModel creates the dashboard model.
func NewModel(cfg *config.Config, sessionMgr *session.Manager, tmuxClient *tmux.Client, splitMode bool, rightPaneID string, projectGroups *config.ProjectConfig) Model {
	logger, _ := monitor.NewLogger(cfg.LogsDir(), cfg.Monitor.MaxLogSizeMB)
	watcher := monitor.NewWatcher(tmuxClient, logger, cfg.Monitor.PollIntervalMs)

	var pGroups map[string][]string
	if projectGroups != nil {
		pGroups = projectGroups.Groups
	}

	return Model{
		cfg:            cfg,
		sessionMgr:     sessionMgr,
		watcher:        watcher,
		tmuxClient:     tmuxClient,
		sessionTree:    components.NewSessionTree(),
		logViewer:      components.NewLogViewer(),
		statusBar:      components.NewStatusBar(),
		agentLogs:      make(map[string][]string),
		agentResources: make(map[string]monitor.ResourceEvent),
		splitMode:      splitMode,
		rightPaneID:    rightPaneID,
		projectGroups:  pGroups,
	}
}

// Init returns the initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.listenForEvents(),
		m.listenForResourceEvents(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			m.watcher.Stop()
			return m, tea.Quit

		case "up", "k":
			m.sessionTree.MoveUp()
			m.syncLogViewer()
			return m, nil

		case "down", "j":
			m.sessionTree.MoveDown()
			m.syncLogViewer()
			return m, nil

		case "enter":
			node := m.sessionTree.SelectedNode()
			if node != nil && node.IsGroup {
				m.sessionTree.Toggle()
			} else if agent := m.sessionTree.SelectedAgent(); agent != nil {
				if m.splitMode && m.rightPaneID != "" && agent.Status == "running" {
					// In split mode, instantly attach the right pane to this session
					tmuxSession := m.cfg.SessionPrefix + "-" + agent.Name
					// Use env -u TMUX so nested attach works cleanly
					cmdStr := fmt.Sprintf("env -u TMUX %s attach-session -t %s", m.cfg.TmuxBinary, tmuxSession)
					_ = exec.Command(m.cfg.TmuxBinary, "respawn-pane", "-k", "-t", m.rightPaneID, cmdStr).Run()
				} else {
					m.syncLogViewer()
				}
			}
			return m, nil

		case "left":
			m.sessionTree.Collapse()
			return m, nil

		case "right":
			m.sessionTree.Expand()
			return m, nil

		case "pgup":
			m.logViewer.ScrollUp(10)
			return m, nil

		case "pgdown":
			m.logViewer.ScrollDown(10)
			return m, nil

		case "a", "A":
			// Attach interactively using tea.ExecProcess
			if agent := m.sessionTree.SelectedAgent(); agent != nil && agent.Status == "running" {
				tmuxSession := m.cfg.SessionPrefix + "-" + agent.Name
				c := exec.Command(m.cfg.TmuxBinary, "attach-session", "-t", tmuxSession)
				
				// Suspend the TUI, attach to tmux, and resume when detached
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					// When tmux detaches, force a refresh
					return tickMsg{}
				})
			}
			return m, nil

		case "d":
			// Destroy selected agent
			if agent := m.sessionTree.SelectedAgent(); agent != nil {
				_ = m.sessionMgr.Destroy(context.Background(), agent.Name)
				m.watcher.Unwatch(agent.Name)
				return m, tickCmd()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Calculate which tree row was clicked
			// The sidebar starts at row 2 (title + separator) inside the border (row 1)
			clickedRow := msg.Y - 3 // account for border + title + separator
			if msg.X < m.sessionTree.Width && clickedRow >= 0 && clickedRow < len(m.sessionTree.FlatItems) {
				m.sessionTree.Selected = clickedRow
				node := m.sessionTree.SelectedNode()
				if node != nil && node.IsGroup {
					m.sessionTree.Toggle()
				} else {
					m.syncLogViewer()
				}
			}
		}
		return m, nil

	case tickMsg:
		m.refreshAgents()
		return m, tickCmd()

	case eventMsg:
		event := monitor.Event(msg)
		lines := strings.Split(event.Content, "\n")
		m.agentLogs[event.AgentName] = append(m.agentLogs[event.AgentName], lines...)

		// Cap per-agent log lines at 1000
		if len(m.agentLogs[event.AgentName]) > 1000 {
			m.agentLogs[event.AgentName] = m.agentLogs[event.AgentName][len(m.agentLogs[event.AgentName])-500:]
		}

		m.syncLogViewer()
		return m, m.listenForEvents()

	case resourceMsg:
		event := monitor.ResourceEvent(msg)
		m.agentResources[event.AgentName] = event
		
		// Update agentList immediately for responsive UI
		for i, fi := range m.sessionTree.FlatItems {
			if fi.Node.Agent != nil && fi.Node.Agent.Name == event.AgentName {
				m.sessionTree.FlatItems[i].Node.Agent.CPU = event.CPU
				m.sessionTree.FlatItems[i].Node.Agent.Memory = event.Memory
				break
			}
		}
		return m, m.listenForResourceEvents()
	}

	return m, nil
}

// View renders the full dashboard.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Render components
	sidebar := m.sessionTree.Render()

	var rendered string
	if m.splitMode {
		// In split mode, the right pane is an actual tmux pane, so we only render the sidebar and status
		rendered = lipgloss.JoinVertical(lipgloss.Left, sidebar, m.statusBar.Render())
	} else {
		logPanel := m.logViewer.Render()
		mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, logPanel)
		rendered = lipgloss.JoinVertical(lipgloss.Left, mainArea, m.statusBar.Render())
	}

	return rendered
}

// refreshAgents refreshes the agent list from the session manager.
func (m *Model) refreshAgents() {
	sessions := m.sessionMgr.List(context.Background())

	agents := make([]components.AgentInfo, len(sessions))
	running := 0
	for i, s := range sessions {
		ag := components.AgentInfo{
			Name:      s.Name,
			Type:      s.AgentType,
			Status:    s.Status,
			WorkDir:   s.WorkDir,
			CreatedAt: s.CreatedAt,
			Group:     s.Group,
		}
		if res, ok := m.agentResources[s.Name]; ok {
			ag.CPU = res.CPU
			ag.Memory = res.Memory
		}
		agents[i] = ag

		if s.Status == "running" {
			running++

			// Ensure we're watching running agents
			if !m.watcher.IsWatching(s.Name) {
				_ = m.watcher.Watch(s.Name, s.TmuxName)
			}
		}
	}

	m.sessionTree.BuildTree(agents, m.projectGroups)

	// Fix selection if out of bounds
	if m.sessionTree.Selected >= len(m.sessionTree.FlatItems) && len(m.sessionTree.FlatItems) > 0 {
		m.sessionTree.Selected = len(m.sessionTree.FlatItems) - 1
	}

	// Update status bar
	m.statusBar.TotalAgents = len(agents)
	m.statusBar.RunningAgents = running
	if agent := m.sessionTree.SelectedAgent(); agent != nil {
		m.statusBar.SelectedAgent = agent.Name
	} else {
		m.statusBar.SelectedAgent = ""
	}

	m.syncLogViewer()
}

// syncLogViewer updates the log viewer to show the selected agent's logs.
func (m *Model) syncLogViewer() {
	agent := m.sessionTree.SelectedAgent()
	if agent == nil {
		m.logViewer.AgentName = ""
		m.logViewer.Lines = nil
		return
	}

	m.logViewer.AgentName = agent.Name
	if logs, ok := m.agentLogs[agent.Name]; ok {
		m.logViewer.Lines = logs
	} else {
		m.logViewer.Lines = nil
	}
}

// updateLayout recalculates component sizes for the current terminal size.
func (m *Model) updateLayout() {
	sidebarWidth := 30
	if m.splitMode {
		// Take full width of its pane
		sidebarWidth = m.width
	} else if m.width < 100 {
		sidebarWidth = 25
	}
	
	mainHeight := m.height - 1 // Leave 1 line for status bar
	if mainHeight < 5 {
		mainHeight = 5
	}

	m.sessionTree.Width = sidebarWidth
	m.sessionTree.Height = mainHeight

	m.logViewer.Width = m.width - sidebarWidth
	m.logViewer.Height = mainHeight
	
	m.statusBar.Width = m.width
}

// tickCmd returns a command that sends a tick after 1 second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// listenForEvents returns a command that reads the next event from the watcher.
func (m Model) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.watcher.Events()
		if !ok {
			return nil
		}
		return eventMsg(event)
	}
}

// listenForResourceEvents returns a command that reads the next resource event from the watcher.
func (m Model) listenForResourceEvents() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.watcher.ResourceEvents()
		if !ok {
			return nil
		}
		return resourceMsg(event)
	}
}

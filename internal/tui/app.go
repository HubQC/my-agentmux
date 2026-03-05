package tui

import (
	"context"
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

	agentList components.AgentList
	logViewer components.LogViewer
	statusBar components.StatusBar

	// Log lines per agent name.
	agentLogs map[string][]string

	// Resources per agent name.
	agentResources map[string]monitor.ResourceEvent

	width  int
	height int

	quitting bool
}

// NewModel creates the dashboard model.
func NewModel(cfg *config.Config, sessionMgr *session.Manager, tmuxClient *tmux.Client) Model {
	logger, _ := monitor.NewLogger(cfg.LogsDir(), cfg.Monitor.MaxLogSizeMB)
	watcher := monitor.NewWatcher(tmuxClient, logger, cfg.Monitor.PollIntervalMs)

	return Model{
		cfg:        cfg,
		sessionMgr: sessionMgr,
		watcher:    watcher,
		tmuxClient: tmuxClient,
		agentList:      components.NewAgentList(),
		logViewer:      components.NewLogViewer(),
		statusBar:      components.NewStatusBar(),
		agentLogs:      make(map[string][]string),
		agentResources: make(map[string]monitor.ResourceEvent),
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
			m.agentList.MoveUp()
			m.syncLogViewer()
			return m, nil

		case "down", "j":
			m.agentList.MoveDown()
			m.syncLogViewer()
			return m, nil

		case "pgup":
			m.logViewer.ScrollUp(10)
			return m, nil

		case "pgdown":
			m.logViewer.ScrollDown(10)
			return m, nil

		case "a":
			// Attach — quit TUI first, then attach
			if agent := m.agentList.SelectedAgent(); agent != nil && agent.Status == "running" {
				m.quitting = true
				m.watcher.Stop()
				tmuxSession := m.cfg.SessionPrefix + "-" + agent.Name
				c := exec.Command(m.cfg.TmuxBinary, "attach-session", "-t", tmuxSession)
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return tea.Quit()
				})
			}
			return m, nil

		case "d":
			// Destroy selected agent
			if agent := m.agentList.SelectedAgent(); agent != nil {
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
		for i := range m.agentList.Agents {
			if m.agentList.Agents[i].Name == event.AgentName {
				m.agentList.Agents[i].CPU = event.CPU
				m.agentList.Agents[i].Memory = event.Memory
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
	sidebar := m.agentList.Render()
	logPanel := m.logViewer.Render()
	statusBar := m.statusBar.Render()

	// Layout: sidebar | log panel
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, logPanel)

	// Stack: main | status bar
	return lipgloss.JoinVertical(lipgloss.Left, mainArea, statusBar)
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

	m.agentList.Agents = agents

	// Fix selection if out of bounds
	if m.agentList.Selected >= len(agents) && len(agents) > 0 {
		m.agentList.Selected = len(agents) - 1
	}

	// Update status bar
	m.statusBar.TotalAgents = len(agents)
	m.statusBar.RunningAgents = running
	if agent := m.agentList.SelectedAgent(); agent != nil {
		m.statusBar.SelectedAgent = agent.Name
	} else {
		m.statusBar.SelectedAgent = ""
	}

	m.syncLogViewer()
}

// syncLogViewer updates the log viewer to show the selected agent's logs.
func (m *Model) syncLogViewer() {
	agent := m.agentList.SelectedAgent()
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
	if m.width < 80 {
		sidebarWidth = 24
	}

	mainHeight := m.height - 1 // status bar takes 1 row
	if mainHeight < 5 {
		mainHeight = 5
	}

	m.agentList.Width = sidebarWidth
	m.agentList.Height = mainHeight

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

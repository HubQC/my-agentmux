package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// AgentInfo holds display info for an agent.
type AgentInfo struct {
	Name      string
	Type      string
	Status    string // "running", "stopped"
	WorkDir   string
	CreatedAt time.Time
	CPU       float64
	Memory    uint64
	Group     string

	// Git context
	GitBranch string
	GitRepo   string

	// Codex Integrations
	CodexProfile    string
	CodexReasoning  string
	CodexMCPs       []string
	CodexMultiAgent bool

	// Gemini Integrations
	GeminiMCPs []string
}

// AgentList renders a sidebar list of agents.
type AgentList struct {
	Agents   []AgentInfo
	Selected int
	Width    int
	Height   int
}

// NewAgentList creates a new agent list component.
func NewAgentList() AgentList {
	return AgentList{
		Selected: 0,
		Width:    28,
	}
}

// Render returns the rendered agent list sidebar.
func (al AgentList) Render() string {
	// Colors
	primary := lipgloss.Color("#7C3AED")
	green := lipgloss.Color("#22C55E")
	red := lipgloss.Color("#EF4444")
	text := lipgloss.Color("#F9FAFB")
	dimText := lipgloss.Color("#9CA3AF")
	highlight := lipgloss.Color("#312E81")
	borderColor := lipgloss.Color("#374151")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primary).
		PaddingLeft(1)

	itemStyle := lipgloss.NewStyle().
		Foreground(text).
		PaddingLeft(1).
		Width(al.Width - 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(text).
		Background(highlight).
		Bold(true).
		PaddingLeft(1).
		Width(al.Width - 2)

	statusRunning := lipgloss.NewStyle().Foreground(green)
	statusStopped := lipgloss.NewStyle().Foreground(red)

	typeStyle := lipgloss.NewStyle().Foreground(dimText).Italic(true)
	uptimeStyle := lipgloss.NewStyle().Foreground(dimText)

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("⚡ AGENTS"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", al.Width-2)))
	b.WriteString("\n")

	if len(al.Agents) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(dimText).
			Italic(true).
			PaddingLeft(1).
			Render("No agents running")
		b.WriteString(placeholder)
		b.WriteString("\n")
	}

	for i, agent := range al.Agents {
		// Status icon
		var statusIcon string
		if agent.Status == "running" {
			statusIcon = statusRunning.Render("●")
		} else {
			statusIcon = statusStopped.Render("○")
		}

		// Agent name line
		name := fmt.Sprintf("%s %s", statusIcon, agent.Name)

		// Detail line
		uptime := formatUptime(agent.CreatedAt)
		detail := typeStyle.Render(agent.Type) + " " + uptimeStyle.Render(uptime)

		if agent.Status == "running" && (agent.CPU > 0 || agent.Memory > 0) {
			memMB := float64(agent.Memory) / 1024 / 1024
			metrics := fmt.Sprintf(" [%.1f%% CPU | %.1f MB] ", agent.CPU, memMB)
			detail += lipgloss.NewStyle().Foreground(lipgloss.Color("#6366F1")).Render(metrics)
		}

		var entry string
		if i == al.Selected {
			entry = selectedStyle.Render(name) + "\n" + selectedStyle.Render("  "+detail)
		} else {
			entry = itemStyle.Render(name) + "\n" + itemStyle.Render("  "+detail)
		}

		b.WriteString(entry)
		b.WriteString("\n")
	}

	// Pad remaining height
	rendered := b.String()
	lines := strings.Count(rendered, "\n")
	for lines < al.Height-3 {
		rendered += "\n"
		lines++
	}

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(al.Width).
		Height(al.Height)

	return borderStyle.Render(rendered)
}

// MoveUp moves selection up.
func (al *AgentList) MoveUp() {
	if al.Selected > 0 {
		al.Selected--
	}
}

// MoveDown moves selection down.
func (al *AgentList) MoveDown() {
	if al.Selected < len(al.Agents)-1 {
		al.Selected++
	}
}

// SelectedAgent returns the currently selected agent, or nil.
func (al AgentList) SelectedAgent() *AgentInfo {
	if len(al.Agents) == 0 || al.Selected >= len(al.Agents) {
		return nil
	}
	return &al.Agents[al.Selected]
}

func formatUptime(created time.Time) string {
	d := time.Since(created)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

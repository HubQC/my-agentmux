package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ActionMenuItem represents a single menu item.
type ActionMenuItem struct {
	Key     string // keyboard shortcut
	Label   string
	Action  string // action identifier
	Enabled bool
}

// ActionMenu renders a popup action menu for a selected agent.
type ActionMenu struct {
	Active    bool
	AgentName string
	Items     []ActionMenuItem
	Selected  int
	Width     int
}

// NewActionMenu creates a new action menu component.
func NewActionMenu() ActionMenu {
	return ActionMenu{Width: 32}
}

// Show displays the action menu for the given agent.
func (am *ActionMenu) Show(agent *AgentInfo) {
	if agent == nil {
		return
	}
	am.Active = true
	am.AgentName = agent.Name
	am.Selected = 0

	isRunning := agent.Status == "running"
	am.Items = []ActionMenuItem{
		{Key: "a", Label: "Attach (fullscreen)", Action: "attach", Enabled: isRunning},
		{Key: "l", Label: "View logs", Action: "logs", Enabled: true},
		{Key: "s", Label: "Send command", Action: "send", Enabled: isRunning},
		{Key: "r", Label: "Restart agent", Action: "restart", Enabled: true},
		{Key: "d", Label: "Stop agent", Action: "stop", Enabled: isRunning},
	}
}

// Hide closes the action menu.
func (am *ActionMenu) Hide() {
	am.Active = false
}

// MoveUp moves the selection up.
func (am *ActionMenu) MoveUp() {
	if am.Selected > 0 {
		am.Selected--
	}
}

// MoveDown moves the selection down.
func (am *ActionMenu) MoveDown() {
	if am.Selected < len(am.Items)-1 {
		am.Selected++
	}
}

// SelectedAction returns the action identifier of the selected item.
func (am ActionMenu) SelectedAction() string {
	if am.Selected >= 0 && am.Selected < len(am.Items) {
		item := am.Items[am.Selected]
		if item.Enabled {
			return item.Action
		}
	}
	return ""
}

// HandleKey processes a key press when the menu is active.
// Returns the action to execute, or "" if no action.
func (am *ActionMenu) HandleKey(key string) string {
	switch key {
	case "esc", "q":
		am.Hide()
		return ""
	case "up", "k":
		am.MoveUp()
		return ""
	case "down", "j":
		am.MoveDown()
		return ""
	case "enter":
		action := am.SelectedAction()
		am.Hide()
		return action
	default:
		// Check direct shortcut keys
		for i, item := range am.Items {
			if item.Key == key && item.Enabled {
				am.Selected = i
				am.Hide()
				return item.Action
			}
		}
	}
	return ""
}

// Render returns the rendered action menu popup.
func (am ActionMenu) Render() string {
	if !am.Active {
		return ""
	}

	bgColor := lipgloss.Color("#1E1B4B")
	borderColor := lipgloss.Color("#7C3AED")
	text := lipgloss.Color("#F9FAFB")
	dimText := lipgloss.Color("#9CA3AF")
	disabledColor := lipgloss.Color("#4B5563")
	highlightBg := lipgloss.Color("#312E81")
	cyan := lipgloss.Color("#06B6D4")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cyan).
		PaddingLeft(1)

	var b strings.Builder
	b.WriteString(titleStyle.Render("⚡ "+am.AgentName) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(borderColor).PaddingLeft(1).Render(strings.Repeat("─", am.Width-4)) + "\n")

	for i, item := range am.Items {
		keyStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
		labelStyle := lipgloss.NewStyle().Foreground(text)

		if !item.Enabled {
			keyStyle = keyStyle.Foreground(disabledColor)
			labelStyle = labelStyle.Foreground(disabledColor)
		}

		line := "  " + keyStyle.Render("["+item.Key+"]") + " " + labelStyle.Render(item.Label)

		if i == am.Selected {
			lineStyle := lipgloss.NewStyle().
				Background(highlightBg).
				Width(am.Width - 4)
			b.WriteString(lineStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(dimText).Italic(true).PaddingLeft(1)
	b.WriteString("\n" + helpStyle.Render("↑/↓ navigate • Enter select • Esc close"))

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(bgColor).
		Width(am.Width).
		Padding(0, 0)

	return popupStyle.Render(b.String())
}

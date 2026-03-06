package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders a bottom status bar with key bindings and info.
type StatusBar struct {
	TotalAgents   int
	RunningAgents int
	SelectedAgent string
	Width         int
}

// NewStatusBar creates a new status bar component.
func NewStatusBar() StatusBar {
	return StatusBar{}
}

// Render returns the rendered status bar.
func (sb StatusBar) Render() string {
	primary := lipgloss.Color("#7C3AED")
	text := lipgloss.Color("#F9FAFB")
	dimText := lipgloss.Color("#DDD6FE")

	barStyle := lipgloss.NewStyle().
		Background(primary).
		Foreground(text).
		Width(sb.Width).
		Padding(0, 1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Background(primary).
		Foreground(text)

	descStyle := lipgloss.NewStyle().
		Background(primary).
		Foreground(dimText)

	infoStyle := lipgloss.NewStyle().
		Background(primary).
		Foreground(text).
		Bold(true)

	// Key bindings
	keys := []struct{ key, desc string }{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"Enter", "select"},
		{"←/→", "fold"},
		{"pgup/pgdn", "scroll"},
		{"a", "embed"},
		{"d", "stop"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, keyStyle.Render(k.key)+" "+descStyle.Render(k.desc))
	}
	helpText := strings.Join(parts, descStyle.Render("  │  "))

	// Right side info
	info := infoStyle.Render(fmt.Sprintf(" %d/%d agents", sb.RunningAgents, sb.TotalAgents))

	// Calculate spacing
	helpLen := lipgloss.Width(helpText)
	infoLen := lipgloss.Width(info)
	spacer := sb.Width - helpLen - infoLen - 2
	if spacer < 1 {
		spacer = 1
	}

	return barStyle.Render(helpText + strings.Repeat(" ", spacer) + info)
}

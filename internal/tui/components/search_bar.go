package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SearchBar renders a filter/search input for the session tree.
type SearchBar struct {
	Query  string
	Active bool // Whether the search bar is currently focused
	Width  int
}

// NewSearchBar creates a new search bar component.
func NewSearchBar() SearchBar {
	return SearchBar{}
}

// Render returns the rendered search bar.
func (sb SearchBar) Render() string {
	if !sb.Active && sb.Query == "" {
		return "" // Don't render when inactive and empty
	}

	borderColor := lipgloss.Color("#374151")
	cyan := lipgloss.Color("#06B6D4")
	text := lipgloss.Color("#F9FAFB")
	muted := lipgloss.Color("#6B7280")

	var content string
	if sb.Active {
		prefix := lipgloss.NewStyle().Foreground(cyan).Bold(true).Render("🔍 /")
		input := lipgloss.NewStyle().Foreground(text).Render(sb.Query)
		cursor := lipgloss.NewStyle().Foreground(cyan).Render("▎")
		content = prefix + input + cursor
	} else if sb.Query != "" {
		prefix := lipgloss.NewStyle().Foreground(muted).Render("🔍 ")
		input := lipgloss.NewStyle().Foreground(text).Render(sb.Query)
		esc := lipgloss.NewStyle().Foreground(muted).Italic(true).Render("  (Esc to clear)")
		content = prefix + input + esc
	}

	style := lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottomForeground(borderColor).
		Width(sb.Width - 2).
		PaddingLeft(1)

	return style.Render(content) + "\n"
}

// HandleKey processes a key press when the search bar is active.
// Returns true if the key was consumed.
func (sb *SearchBar) HandleKey(key string) bool {
	if !sb.Active {
		return false
	}

	switch key {
	case "esc":
		sb.Active = false
		sb.Query = ""
		return true
	case "enter":
		sb.Active = false
		return true
	case "backspace":
		if len(sb.Query) > 0 {
			sb.Query = sb.Query[:len(sb.Query)-1]
		}
		return true
	default:
		// Only accept printable single characters
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			sb.Query += key
			return true
		}
	}
	return false
}

// MatchesAgent returns true if the agent matches the current search query.
func (sb SearchBar) MatchesAgent(agent *AgentInfo) bool {
	if sb.Query == "" {
		return true // No filter — show all
	}
	q := strings.ToLower(sb.Query)
	return strings.Contains(strings.ToLower(agent.Name), q) ||
		strings.Contains(strings.ToLower(agent.Type), q) ||
		strings.Contains(strings.ToLower(agent.Group), q) ||
		strings.Contains(strings.ToLower(agent.Status), q) ||
		strings.Contains(strings.ToLower(agent.WorkDir), q)
}

package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LogViewer renders a scrollable log panel for an agent.
type LogViewer struct {
	AgentName string
	Lines     []string
	Width     int
	Height    int
	Offset    int // scroll offset (0 = bottom)
}

// NewLogViewer creates a new log viewer component.
func NewLogViewer() LogViewer {
	return LogViewer{}
}

// Render returns the rendered log viewer panel.
func (lv LogViewer) Render() string {
	cyan := lipgloss.Color("#06B6D4")
	dimText := lipgloss.Color("#9CA3AF")
	muted := lipgloss.Color("#6B7280")
	borderColor := lipgloss.Color("#374151")
	panelBg := lipgloss.Color("#0F172A")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cyan).
		PaddingLeft(1)

	contentStyle := lipgloss.NewStyle().
		Foreground(dimText).
		PaddingLeft(1)

	placeholderStyle := lipgloss.NewStyle().
		Foreground(muted).
		Italic(true).
		PaddingLeft(1)

	var content string
	viewHeight := lv.Height - 4 // account for title + border

	if lv.AgentName == "" {
		// No agent selected
		pad := strings.Repeat("\n", viewHeight/2)
		content = pad + placeholderStyle.Render("Select an agent from the sidebar") + "\n"
	} else if len(lv.Lines) == 0 {
		pad := strings.Repeat("\n", viewHeight/2)
		content = pad + placeholderStyle.Render("No output yet for "+lv.AgentName) + "\n"
	} else {
		// Show lines with scrolling
		startIdx := 0
		if len(lv.Lines) > viewHeight {
			startIdx = len(lv.Lines) - viewHeight - lv.Offset
			if startIdx < 0 {
				startIdx = 0
			}
		}
		endIdx := startIdx + viewHeight
		if endIdx > len(lv.Lines) {
			endIdx = len(lv.Lines)
		}

		var b strings.Builder
		for _, line := range lv.Lines[startIdx:endIdx] {
			// Truncate long lines
			if len(line) > lv.Width-4 {
				line = line[:lv.Width-7] + "..."
			}
			b.WriteString(contentStyle.Render(line))
			b.WriteString("\n")
		}
		content = b.String()
	}

	// Title
	title := "📋 OUTPUT"
	if lv.AgentName != "" {
		title = "📋 " + strings.ToUpper(lv.AgentName)
	}

	header := titleStyle.Render(title) + "\n" +
		lipgloss.NewStyle().Foreground(borderColor).PaddingLeft(1).Render(strings.Repeat("─", lv.Width-4)) + "\n"

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBg).
		Width(lv.Width).
		Height(lv.Height)

	return borderStyle.Render(header + content)
}

// AppendLine adds a line to the log viewer.
func (lv *LogViewer) AppendLine(line string) {
	lv.Lines = append(lv.Lines, line)
	// Auto-scroll to bottom
	lv.Offset = 0
}

// AppendLines adds multiple lines.
func (lv *LogViewer) AppendLines(lines []string) {
	lv.Lines = append(lv.Lines, lines...)
	lv.Offset = 0
}

// Clear removes all lines.
func (lv *LogViewer) Clear() {
	lv.Lines = nil
	lv.Offset = 0
}

// ScrollUp scrolls up by n lines.
func (lv *LogViewer) ScrollUp(n int) {
	lv.Offset += n
	maxOffset := len(lv.Lines) - (lv.Height - 4)
	if maxOffset < 0 {
		maxOffset = 0
	}
	if lv.Offset > maxOffset {
		lv.Offset = maxOffset
	}
}

// ScrollDown scrolls down by n lines.
func (lv *LogViewer) ScrollDown(n int) {
	lv.Offset -= n
	if lv.Offset < 0 {
		lv.Offset = 0
	}
}

package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PipelineGraph visualizes a sequence of agents.
type PipelineGraph struct {
	Name     string
	Sequence []string
	Agents   map[string]*AgentInfo
	Width    int
}

// NewPipelineGraph creates a new instance.
func NewPipelineGraph() PipelineGraph {
	return PipelineGraph{
		Agents: make(map[string]*AgentInfo),
	}
}

// Render returns the rendered DAG.
func (p *PipelineGraph) Render() string {
	if len(p.Sequence) == 0 {
		return ""
	}

	primary := lipgloss.Color("#7C3AED")
	green := lipgloss.Color("#22C55E")
	red := lipgloss.Color("#EF4444")
	text := lipgloss.Color("#F9FAFB")
	dimText := lipgloss.Color("#9CA3AF")

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primary).Padding(1, 0, 1, 2)
	nodeStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2)
	arrowStyle := lipgloss.NewStyle().Foreground(dimText).PaddingLeft(6)

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Pipeline: %s", p.Name)))
	b.WriteString("\n")

	for i, step := range p.Sequence {
		// find agent status
		status := "pending"
		var color lipgloss.Color = dimText

		if ag, ok := p.Agents[step]; ok {
			status = ag.Status
			if status == "running" {
				color = green
			} else {
				color = red
			}
		}

		style := nodeStyle.BorderForeground(color).Foreground(text)
		nodeContent := fmt.Sprintf("%-12s\n[%s]", step, status)
		b.WriteString("  " + style.Render(nodeContent) + "\n")

		if i < len(p.Sequence)-1 {
			b.WriteString(arrowStyle.Render("│"))
			b.WriteString("\n")
			b.WriteString(arrowStyle.Render("▼"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

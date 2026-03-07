package components

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TreeNode represents a node in the session tree (either a group or a session).
type TreeNode struct {
	Label     string
	IsGroup   bool
	Collapsed bool
	Agent     *AgentInfo // nil for group nodes
	Children  []*TreeNode
}

// flatItem is a flattened tree item for cursor navigation.
type flatItem struct {
	Node  *TreeNode
	Depth int
}

// SessionTree renders a collapsible, grouped tree view of sessions.
type SessionTree struct {
	Root      []*TreeNode
	FlatItems []flatItem
	Selected  int
	Width     int
	Height    int
}

// NewSessionTree creates a new session tree component.
func NewSessionTree() SessionTree {
	return SessionTree{
		Selected: 0,
		Width:    30,
	}
}

// BuildTree builds the tree from a list of agents.
// Grouping priority: explicit Group > projectGroups config > fallback by workdir.
func (st *SessionTree) BuildTree(agents []AgentInfo, projectGroups map[string][]string) {
	groupMap := make(map[string][]*AgentInfo)

	for i := range agents {
		agent := &agents[i]
		groupName := resolveGroup(agent, projectGroups)
		groupMap[groupName] = append(groupMap[groupName], agent)
	}

	// Sort group names
	groupNames := make([]string, 0, len(groupMap))
	for name := range groupMap {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	root := make([]*TreeNode, 0, len(groupNames))
	for _, gName := range groupNames {
		members := groupMap[gName]

		// Sort agents within group by creation time
		sort.Slice(members, func(i, j int) bool {
			return members[i].CreatedAt.Before(members[j].CreatedAt)
		})

		children := make([]*TreeNode, len(members))
		for i, m := range members {
			children[i] = &TreeNode{
				Label: m.Name,
				Agent: m,
			}
		}

		groupNode := &TreeNode{
			Label:    gName,
			IsGroup:  true,
			Children: children,
		}
		root = append(root, groupNode)
	}

	st.Root = root
	st.flatten()
}

// resolveGroup determines the group for an agent.
func resolveGroup(agent *AgentInfo, projectGroups map[string][]string) string {
	// 1. Explicit Group field
	if agent.Group != "" {
		return agent.Group
	}

	// 2. Project config groups
	for groupName, members := range projectGroups {
		for _, name := range members {
			if name == agent.Name {
				return groupName
			}
		}
	}

	// 3. Fallback: group by workdir basename
	if agent.WorkDir != "" {
		return filepath.Base(agent.WorkDir)
	}

	return "ungrouped"
}

// flatten creates the flat list of visible items for cursor navigation.
func (st *SessionTree) flatten() {
	st.FlatItems = nil
	for _, node := range st.Root {
		st.FlatItems = append(st.FlatItems, flatItem{Node: node, Depth: 0})
		if !node.Collapsed {
			for _, child := range node.Children {
				st.FlatItems = append(st.FlatItems, flatItem{Node: child, Depth: 1})
			}
		}
	}
}

// MoveUp moves the cursor up.
func (st *SessionTree) MoveUp() {
	if st.Selected > 0 {
		st.Selected--
	}
}

// MoveDown moves the cursor down.
func (st *SessionTree) MoveDown() {
	if st.Selected < len(st.FlatItems)-1 {
		st.Selected++
	}
}

// Toggle collapses or expands the group under the cursor.
func (st *SessionTree) Toggle() {
	if st.Selected >= len(st.FlatItems) {
		return
	}
	item := st.FlatItems[st.Selected]
	if item.Node.IsGroup {
		item.Node.Collapsed = !item.Node.Collapsed
		st.flatten()
		// Clamp selection
		if st.Selected >= len(st.FlatItems) {
			st.Selected = len(st.FlatItems) - 1
		}
	}
}

// Collapse collapses the group under the cursor (or the parent group).
func (st *SessionTree) Collapse() {
	if st.Selected >= len(st.FlatItems) {
		return
	}
	item := st.FlatItems[st.Selected]
	if item.Node.IsGroup && !item.Node.Collapsed {
		item.Node.Collapsed = true
		st.flatten()
	} else if !item.Node.IsGroup {
		// Find parent group and collapse it
		for i := st.Selected - 1; i >= 0; i-- {
			if st.FlatItems[i].Node.IsGroup {
				st.FlatItems[i].Node.Collapsed = true
				st.Selected = i
				st.flatten()
				break
			}
		}
	}
}

// Expand expands the group under the cursor.
func (st *SessionTree) Expand() {
	if st.Selected >= len(st.FlatItems) {
		return
	}
	item := st.FlatItems[st.Selected]
	if item.Node.IsGroup && item.Node.Collapsed {
		item.Node.Collapsed = false
		st.flatten()
	}
}

// SelectedAgent returns the AgentInfo if a session node is selected, nil if a group node.
func (st *SessionTree) SelectedAgent() *AgentInfo {
	if len(st.FlatItems) == 0 || st.Selected >= len(st.FlatItems) {
		return nil
	}
	return st.FlatItems[st.Selected].Node.Agent
}

// SelectedNode returns the currently selected tree node.
func (st *SessionTree) SelectedNode() *TreeNode {
	if len(st.FlatItems) == 0 || st.Selected >= len(st.FlatItems) {
		return nil
	}
	return st.FlatItems[st.Selected].Node
}

// Render returns the rendered session tree sidebar.
func (st SessionTree) Render() string {
	primary := lipgloss.Color("#7C3AED")
	green := lipgloss.Color("#22C55E")
	red := lipgloss.Color("#EF4444")
	text := lipgloss.Color("#F9FAFB")
	dimText := lipgloss.Color("#9CA3AF")
	highlight := lipgloss.Color("#312E81")
	borderColor := lipgloss.Color("#374151")
	groupColor := lipgloss.Color("#F59E0B")
	codexColor := lipgloss.Color("#0EA5E9") // Nice bright blue for Codex elements

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primary).
		PaddingLeft(1)

	itemStyle := lipgloss.NewStyle().
		Foreground(text).
		PaddingLeft(1).
		Width(st.Width - 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(text).
		Background(highlight).
		Bold(true).
		PaddingLeft(1).
		Width(st.Width - 2)

	groupStyle := lipgloss.NewStyle().
		Foreground(groupColor).
		Bold(true)

	statusRunning := lipgloss.NewStyle().Foreground(green)
	statusStopped := lipgloss.NewStyle().Foreground(red)
	uptimeStyle := lipgloss.NewStyle().Foreground(dimText)

	codexStyle := lipgloss.NewStyle().Foreground(codexColor)
	codexDimStyle := lipgloss.NewStyle().Foreground(codexColor).Faint(true)

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🌳 SESSIONS"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", st.Width-2)))
	b.WriteString("\n")

	if len(st.FlatItems) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(dimText).
			Italic(true).
			PaddingLeft(1).
			Render("No sessions")
		b.WriteString(placeholder)
		b.WriteString("\n")
	}

	for i, fi := range st.FlatItems {
		node := fi.Node
		style := itemStyle
		if i == st.Selected {
			style = selectedStyle
		}

		if node.IsGroup {
			// Group node
			arrow := "▼"
			if node.Collapsed {
				arrow = "▸"
			}
			runCount := 0
			for _, child := range node.Children {
				if child.Agent != nil && child.Agent.Status == "running" {
					runCount++
				}
			}
			label := fmt.Sprintf("%s %s (%d)", arrow, groupStyle.Render(node.Label), len(node.Children))
			b.WriteString(style.Render(label))
			b.WriteString("\n")
		} else {
			// Session node
			var statusIcon string
			if node.Agent != nil && node.Agent.Status == "running" {
				statusIcon = statusRunning.Render("●")
			} else {
				statusIcon = statusStopped.Render("○")
			}

			uptime := ""
			if node.Agent != nil {
				uptime = formatUptime(node.Agent.CreatedAt)
			}

			line := fmt.Sprintf("  %s %s", statusIcon, node.Label)
			if uptime != "" {
				line += " " + uptimeStyle.Render(uptime)
			}
			b.WriteString(style.Render(line))
			b.WriteString("\n")

			// Codex Integration Display
			if node.Agent != nil && node.Agent.CodexProfile != "" {
				// Line 1: Profile & Reasoning
				profileLine := fmt.Sprintf("    ↳ %s", codexStyle.Render("["+node.Agent.CodexProfile+"]"))
				if node.Agent.CodexReasoning != "" {
					profileLine += " " + codexDimStyle.Render("(Reasoning: "+node.Agent.CodexReasoning+")")
				}
				if node.Agent.CodexMultiAgent {
					profileLine += " " + codexStyle.Render("🤖 Multi-Agent")
				}
				b.WriteString(style.Render(profileLine))
				b.WriteString("\n")

				// Line 2: MCP Servers
				if len(node.Agent.CodexMCPs) > 0 {
					mcpStr := strings.Join(node.Agent.CodexMCPs, ", ")
					// truncate if too long
					if len(mcpStr) > 40 {
						mcpStr = mcpStr[:37] + "..."
					}
					mcpLine := fmt.Sprintf("      🔌 MCP: %s", codexDimStyle.Render(mcpStr))
					b.WriteString(style.Render(mcpLine))
					b.WriteString("\n")
				}
			}
		}
	}

	// Pad remaining height
	rendered := b.String()
	lines := strings.Count(rendered, "\n")
	for lines < st.Height-3 {
		rendered += "\n"
		lines++
	}

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(st.Width).
		Height(st.Height)

	return borderStyle.Render(rendered)
}

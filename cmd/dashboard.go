package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/tmux"
	tuipkg "github.com/cqi/my_agentmux/internal/tui"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the real-time TUI dashboard",
	Long: `Launch an interactive terminal dashboard that shows all running
agents, their live output, and lets you manage them with keyboard shortcuts.

Key bindings:
  ↑/k, ↓/j — navigate agents
  pgup/pgdn — scroll logs
  a — attach to selected agent
  d — stop selected agent
  q — quit dashboard`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create session manager
		mgr, err := session.NewManager(cfg)
		if err != nil {
			return fmt.Errorf("initializing session manager: %w", err)
		}

		// Create tmux client
		tmuxClient, err := tmux.NewClient(cfg.TmuxBinary)
		if err != nil {
			return fmt.Errorf("initializing tmux: %w", err)
		}

		// Create TUI model
		model := tuipkg.NewModel(cfg, mgr, tmuxClient)

		// Run bubbletea program
		p := tea.NewProgram(model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("dashboard error: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

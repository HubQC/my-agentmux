package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cqi/my_agentmux/internal/config"
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

		// Check split mode
		splitMode, _ := cmd.Flags().GetBool("split")
		var rightPaneID string

		if splitMode {
			// Ensure we are inside tmux!
			if os.Getenv("TMUX") == "" {
				// Re-launch this command inside a new tmux session
				args := []string{"new-session", "-A", "-s", cfg.SessionPrefix + "-dashboard", os.Args[0] + " dashboard --split"}
				syscallCmd := exec.Command(cfg.TmuxBinary, args...)
				syscallCmd.Stdin = os.Stdin
				syscallCmd.Stdout = os.Stdout
				syscallCmd.Stderr = os.Stderr
				return syscallCmd.Run()
			}

			// Create the split pane: 70% width on the right, return pane ID
			splitCmd := exec.Command(cfg.TmuxBinary, "split-window", "-h", "-p", "70", "-P", "-F", "#{pane_id}", "echo 'Waiting for selection...'; cat")
			var out bytes.Buffer
			splitCmd.Stdout = &out
			if err := splitCmd.Run(); err != nil {
				return fmt.Errorf("creating split window: %w", err)
			}
			rightPaneID = strings.TrimSpace(out.String())

			// Cleanup the right pane when dashboard exits
			defer func() {
				if rightPaneID != "" {
					_ = exec.Command(cfg.TmuxBinary, "kill-pane", "-t", rightPaneID).Run()
				}
			}()
		}

		// Try to load project config to resolve groups
		var projectCfg *config.ProjectConfig
		workDir, _ := os.Getwd()
		if pCfg, err := config.LoadProjectConfig(workDir); err == nil {
			projectCfg = pCfg
		}

		// Create TUI model
		model := tuipkg.NewModel(cfg, mgr, tmuxClient, splitMode, rightPaneID, projectCfg)

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
	dashboardCmd.Flags().Bool("split", false, "open dashboard with an interactive side-by-side terminal split")
	rootCmd.AddCommand(dashboardCmd)
}

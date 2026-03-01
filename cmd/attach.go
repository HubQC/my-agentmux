package cmd

import (
	"fmt"

	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach <agent-name>",
	Short: "Attach to an agent session",
	Long: `Attach the current terminal to a running agent's tmux session.

This replaces the current process with tmux attach-session.
Detach with Ctrl+B, D (default tmux prefix).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := args[0]

		mgr, err := session.NewManager(cfg)
		if err != nil {
			return err
		}

		fmt.Printf("Attaching to agent %q... (detach with Ctrl+B, D)\n", agentName)

		// This replaces the current process — won't return on success
		if err := mgr.Attach(cmd.Context(), agentName); err != nil {
			return fmt.Errorf("attaching to %q: %w", agentName, err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

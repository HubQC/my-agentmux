package cmd

import (
	"fmt"

	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <agent-name>",
	Short: "Stop an agent session",
	Long: `Stop and remove an agent session. The tmux session will be destroyed
and the agent state will be cleaned up.

Use --all to stop all running agents.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := session.NewManager(cfg)
		if err != nil {
			return err
		}

		stopAll, _ := cmd.Flags().GetBool("all")

		if stopAll {
			count, err := mgr.DestroyAll(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Printf("✓ Stopped %d agent(s)\n", count)
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("specify an agent name or use --all")
		}

		agentName := args[0]
		if err := mgr.Destroy(cmd.Context(), agentName); err != nil {
			return err
		}

		fmt.Printf("✓ Agent %q stopped\n", agentName)
		return nil
	},
}

func init() {
	stopCmd.Flags().BoolP("all", "a", false, "stop all agent sessions")

	rootCmd.AddCommand(stopCmd)
}

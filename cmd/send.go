package cmd

import (
	"fmt"

	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send <agent-name> <message>",
	Short: "Send a message to an agent session",
	Long: `Send input text to an agent's tmux session.

The message is typed into the agent's terminal pane and Enter is
pressed automatically. Use --no-enter to send without pressing Enter.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := args[0]
		message := args[1]

		// Join remaining args as part of the message
		for _, a := range args[2:] {
			message += " " + a
		}

		noEnter, _ := cmd.Flags().GetBool("no-enter")

		mgr, err := session.NewManager(cfg)
		if err != nil {
			return err
		}

		if err := mgr.SendKeys(cmd.Context(), agentName, message, !noEnter); err != nil {
			return err
		}

		fmt.Printf("✓ Sent to %q: %s\n", agentName, message)
		return nil
	},
}

func init() {
	sendCmd.Flags().Bool("no-enter", false, "send without pressing Enter")

	rootCmd.AddCommand(sendCmd)
}

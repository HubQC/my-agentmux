package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all agent sessions",
	Long:    "Display all tracked agent sessions with their status, type, and uptime.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := session.NewManager(cfg)
		if err != nil {
			return err
		}

		sessions := mgr.List(cmd.Context())

		if len(sessions) == 0 {
			fmt.Println("No agent sessions found. Start one with: agentmux start <name>")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tWORKDIR\tUPTIME")
		fmt.Fprintln(w, "----\t----\t------\t-------\t------")

		for _, s := range sessions {
			uptime := "-"
			if s.Status == "running" {
				uptime = formatDuration(time.Since(s.CreatedAt))
			}

			statusIcon := statusSymbol(s.Status)
			fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\t%s\n",
				s.Name, s.AgentType, statusIcon, s.Status, s.WorkDir, uptime)
		}

		return w.Flush()
	},
}

func statusSymbol(status string) string {
	switch status {
	case "running":
		return "●"
	case "stopped":
		return "○"
	case "error":
		return "✗"
	default:
		return "?"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func init() {
	rootCmd.AddCommand(listCmd)
}

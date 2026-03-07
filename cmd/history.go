package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cqi/my_agentmux/internal/history"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View past agent session history",
	Long: `View a log of all past agent sessions with duration, status, and type.

Use flags to filter by agent type, status, or date range.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := history.NewStore(cfg.DataDir)
		if err != nil {
			return err
		}

		statsMode, _ := cmd.Flags().GetBool("stats")
		if statsMode {
			return showStats(store)
		}

		limit, _ := cmd.Flags().GetInt("limit")
		agentType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")

		opts := history.FilterOptions{
			Limit:     limit,
			AgentType: agentType,
			Status:    status,
		}

		entries := store.ListFiltered(opts)
		if len(entries) == 0 {
			fmt.Println("No session history found.")
			fmt.Println("History is recorded automatically when agents start and stop.")
			return nil
		}

		fmt.Println("📜 Session History")
		fmt.Println(strings.Repeat("─", 75))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tDURATION\tSTARTED")
		for _, e := range entries {
			durStr := historyFormatDuration(e.Duration)
			startStr := e.StartedAt.Format("Jan 02 15:04")
			statusIcon := historyStatusSymbol(e.Status)
			fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\t%s\n",
				e.Name, e.AgentType, statusIcon, e.Status, durStr, startStr)
		}
		w.Flush()

		fmt.Printf("\nShowing %d session(s). Use --stats for aggregate view.\n", len(entries))
		return nil
	},
}

func showStats(store *history.Store) error {
	stats := store.Stats()

	fmt.Println("📊 Session Statistics")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("Total sessions:  %d\n", stats.TotalSessions)
	fmt.Printf("Completed:       %d\n", stats.Completed)
	fmt.Printf("Failed:          %d\n", stats.Failed)
	fmt.Printf("Total time:      %s\n", historyFormatDuration(stats.TotalDuration))

	if len(stats.AgentTypeCounts) > 0 {
		fmt.Println("\nAgent Type Breakdown:")
		for t, count := range stats.AgentTypeCounts {
			fmt.Printf("  %-15s %d sessions\n", t, count)
		}
	}
	return nil
}

func historyStatusSymbol(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "stopped":
		return "⏹"
	case "running":
		return "🔄"
	default:
		return "•"
	}
}

func historyFormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func init() {
	historyCmd.Flags().Int("limit", 20, "max number of entries to show")
	historyCmd.Flags().String("type", "", "filter by agent type")
	historyCmd.Flags().String("status", "", "filter by status (completed, failed, stopped)")
	historyCmd.Flags().Bool("stats", false, "show aggregate statistics")
	rootCmd.AddCommand(historyCmd)
}

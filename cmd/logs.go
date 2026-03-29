package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/cqi/my_agentmux/internal/monitor"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/tmux"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <agent-name>",
	Short: "View agent output logs",
	Long: `View or follow the output log for an agent session.

By default, prints the existing log content. Use --follow to
continuously stream new output as it appears.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := args[0]
		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")

		// Verify the agent exists
		mgr, err := session.NewManager(cfg)
		if err != nil {
			return err
		}

		agentSession, err := mgr.Get(cmd.Context(), agentName)
		if err != nil {
			return err
		}

		logger, err := monitor.NewLogger(cfg.LogsDir(), cfg.Monitor.MaxLogSizeMB)
		if err != nil {
			return err
		}
		defer logger.Close()

		if !follow {
			// Static mode: print existing log content
			content, err := logger.ReadAll(agentName)
			if err != nil {
				return fmt.Errorf("reading logs: %w", err)
			}

			if content == "" {
				// No log file yet — do a one-time capture
				tmuxClient, err := tmux.NewClient(cfg.TmuxBinary)
				if err != nil {
					return err
				}
				content, err = tmuxClient.CapturePane(cmd.Context(), agentSession.TmuxName, 0, 0)
				if err != nil {
					return fmt.Errorf("capturing output: %w", err)
				}
			}

			if tail > 0 {
				content = tailLines(content, tail)
			}

			fmt.Print(content)
			return nil
		}

		// Follow mode: poll and stream output
		if agentSession.Status != session.StatusRunning {
			return fmt.Errorf("agent %q is not running (status: %s)", agentName, agentSession.Status)
		}

		tmuxClient, err := tmux.NewClient(cfg.TmuxBinary)
		if err != nil {
			return err
		}

		watcher := monitor.NewWatcher(tmuxClient, logger, cfg.Monitor.PollIntervalMs)
		defer watcher.Stop()

		if err := watcher.Watch(agentName, agentSession.TmuxName); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Following logs for %q (Ctrl-C to stop)...\n", agentName)

		// Stream events until interrupted
		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case event, ok := <-watcher.Events():
				if !ok {
					return nil
				}
				_, err := io.WriteString(os.Stdout, event.Content+"\n")
				if err != nil {
					return nil
				}
			}
		}
	},
}

// tailLines returns the last n lines of content.
func tailLines(content string, n int) string {
	lines := splitLines(content)
	if len(lines) <= n {
		return content
	}

	result := ""
	for _, line := range lines[len(lines)-n:] {
		result += line + "\n"
	}
	return result
}

// splitLines splits content into lines, handling \r\n and \n.
func splitLines(content string) []string {
	var lines []string
	current := ""
	for _, r := range content {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else if r == '\r' {
			continue
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "follow log output (like tail -f)")
	logsCmd.Flags().IntP("tail", "n", 0, "show last N lines only")

	rootCmd.AddCommand(logsCmd)
}

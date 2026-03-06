package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/config"
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

		// Try to load project config to resolve groups
		var projectCfg *config.ProjectConfig
		workDir, _ := os.Getwd()
		if pCfg, err := config.LoadProjectConfig(workDir); err == nil {
			projectCfg = pCfg
		}

		useTree, _ := cmd.Flags().GetBool("tree")
		if useTree {
			printTree(sessions, projectCfg)
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
	listCmd.Flags().Bool("tree", false, "display sessions grouped in a tree view")
	rootCmd.AddCommand(listCmd)
}

// printTree prints sessions grouped by their resolved group name.
func printTree(sessions []*session.AgentSession, c *config.ProjectConfig) {
	// 1. Group sessions
	groups := make(map[string][]*session.AgentSession)
	
	for _, s := range sessions {
		groupName := "ungrouped"
		if s.Group != "" {
			groupName = s.Group
		} else if c != nil && c.Groups != nil {
			found := false
			for gName, members := range c.Groups {
				for _, m := range members {
					if m == s.Name {
						groupName = gName
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found && s.WorkDir != "" {
				groupName = filepath.Base(s.WorkDir)
			}
		} else if s.WorkDir != "" {
			groupName = filepath.Base(s.WorkDir)
		}
		
		groups[groupName] = append(groups[groupName], s)
	}

	// 2. Sort groups
	var groupNames []string
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	// 3. Print tree
	for _, gName := range groupNames {
		members := groups[gName]
		
		// Sort within group
		sort.Slice(members, func(i, j int) bool {
			return members[i].CreatedAt.Before(members[j].CreatedAt)
		})

		fmt.Printf("▼ %s (%d)\n", gName, len(members))
		
		for _, s := range members {
			uptime := "-"
			if s.Status == "running" {
				uptime = formatDuration(time.Since(s.CreatedAt))
			}
			
			fmt.Printf("  %s %s (%s, %s)\n", statusSymbol(s.Status), s.Name, s.AgentType, uptime)
		}
	}
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/tmux"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose your AgentMux environment",
	Long: `Runs a comprehensive health check of your AgentMux environment.

Checks include:
  • tmux installation and version
  • AgentMux configuration and data directories
  • Installed agent CLI tools
  • Orphaned tmux sessions
  • System information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		issues := 0
		warnings := 0

		fmt.Println("🩺 AgentMux Doctor")
		fmt.Println("==================")

		// --- System Info ---
		fmt.Println("\n📋 System Information")
		fmt.Printf("   OS:       %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("   Go:       %s\n", runtime.Version())
		fmt.Printf("   Version:  %s\n", Version)

		// --- tmux Check ---
		fmt.Println("\n🔍 Checking tmux...")
		tmuxPath, err := exec.LookPath(cfg.TmuxBinary)
		if err != nil {
			fmt.Println("   ❌ tmux not found on PATH")
			printInstallHint("tmux")
			issues++
		} else {
			fmt.Printf("   ✅ tmux found: %s\n", tmuxPath)

			tmuxClient, err := tmux.NewClient(cfg.TmuxBinary)
			if err == nil {
				ver, err := tmuxClient.Version(ctx)
				if err == nil {
					fmt.Printf("   ✅ tmux version: %s\n", strings.TrimSpace(ver))
				}
			}
		}

		// --- Config Check ---
		fmt.Println("\n🔍 Checking configuration...")
		configFile := filepath.Join(cfg.DataDir, "config.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Println("   ⚠️  No config file found (using defaults)")
			fmt.Println("      Run 'agentmux init' to create one.")
			warnings++
		} else {
			fmt.Printf("   ✅ Config: %s\n", configFile)
		}

		// --- Data Directories ---
		fmt.Println("\n🔍 Checking data directories...")
		dirs := map[string]string{
			"Data":     cfg.DataDir,
			"Sessions": cfg.SessionsDir(),
			"Logs":     cfg.LogsDir(),
			"Agents":   cfg.AgentsDir(),
			"Plans":    cfg.PlansDir(),
		}
		for label, dir := range dirs {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				fmt.Printf("   ✅ %-8s %s\n", label+":", dir)
			} else {
				fmt.Printf("   ❌ %-8s %s (missing)\n", label+":", dir)
				issues++
			}
		}

		// --- Agent Definitions ---
		fmt.Println("\n🔍 Checking agent definitions...")
		agentsDir := cfg.AgentsDir()
		entries, err := os.ReadDir(agentsDir)
		if err == nil {
			mdCount := 0
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					mdCount++
				}
			}
			if mdCount > 0 {
				fmt.Printf("   ✅ %d custom agent definition(s) found\n", mdCount)
			} else {
				fmt.Println("   ℹ️  No custom agent definitions (create .md files in ~/.agentmux/agents/)")
			}
		}

		// --- Agent CLI Tools ---
		fmt.Println("\n🔍 Checking agent CLI tools...")
		presets := agent.ListPresets()
		installed := 0
		for _, p := range presets {
			if p.Binary == "" {
				continue // skip shell preset
			}
			if p.IsInstalled() {
				fmt.Printf("   ✅ %-12s %s\n", p.Name, p.Description)
				installed++
			} else {
				fmt.Printf("   ⬚  %-12s not installed (%s)\n", p.Name, p.Binary)
			}
		}
		if installed == 0 {
			fmt.Println("   ⚠️  No agent CLI tools detected. Install at least one agent to get started.")
			warnings++
		}

		// --- Orphaned Sessions ---
		fmt.Println("\n🔍 Checking for orphaned sessions...")
		mgr, mgrErr := session.NewManager(cfg)
		if mgrErr == nil {
			sessions := mgr.List(ctx)
			orphaned := 0
			for _, s := range sessions {
				if s.Status == session.StatusStopped {
					orphaned++
				}
			}
			running := len(sessions) - orphaned
			fmt.Printf("   ✅ %d running session(s)\n", running)
			if orphaned > 0 {
				fmt.Printf("   ⚠️  %d stale session(s) in state (run 'agentmux cleanup' to remove)\n", orphaned)
				warnings++
			}
		}

		// --- tmux Orphans ---
		tmuxClient, tmuxErr := tmux.NewClient(cfg.TmuxBinary)
		if tmuxErr == nil && tmuxClient.ServerRunning(ctx) {
			tmuxSessions, err := tmuxClient.ListSessions(ctx)
			if err == nil {
				orphanCount := 0
				for _, ts := range tmuxSessions {
					if strings.HasPrefix(ts.Name, cfg.SessionPrefix+"-") {
						// Check if we're tracking it
						agentName := strings.TrimPrefix(ts.Name, cfg.SessionPrefix+"-")
						if mgrErr == nil {
							if _, err := mgr.Get(ctx, agentName); err != nil {
								orphanCount++
							}
						}
					}
				}
				if orphanCount > 0 {
					fmt.Printf("   ⚠️  %d orphaned tmux session(s) with prefix %q\n", orphanCount, cfg.SessionPrefix)
					fmt.Println("      Run 'agentmux cleanup' to remove them.")
					warnings++
				}
			}
		}

		// --- Summary ---
		fmt.Println("\n" + strings.Repeat("─", 40))
		if issues == 0 && warnings == 0 {
			fmt.Println("🎉 All checks passed! Your environment is healthy.")
		} else if issues == 0 {
			fmt.Printf("✅ No critical issues. %d warning(s) found.\n", warnings)
		} else {
			fmt.Printf("❌ %d issue(s) and %d warning(s) found.\n", issues, warnings)
		}

		return nil
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove orphaned tmux sessions and stale state",
	Long: `Finds and removes:
  • Stale entries in the session state file (where tmux session no longer exists)
  • Orphaned tmux sessions with the AgentMux prefix that are not tracked`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		cleaned := 0

		mgr, err := session.NewManager(cfg)
		if err != nil {
			return fmt.Errorf("initializing session manager: %w", err)
		}

		tmuxClient, err := tmux.NewClient(cfg.TmuxBinary)
		if err != nil {
			return fmt.Errorf("initializing tmux: %w", err)
		}

		// 1. Clean stale state entries
		sessions := mgr.List(ctx)
		for _, s := range sessions {
			if s.Status == session.StatusStopped {
				if dryRun {
					fmt.Printf("  Would remove stale state: %s\n", s.Name)
				} else {
					_ = mgr.Destroy(ctx, s.Name)
					fmt.Printf("  ✓ Removed stale state: %s\n", s.Name)
				}
				cleaned++
			}
		}

		// 2. Clean orphaned tmux sessions
		if tmuxClient.ServerRunning(ctx) {
			tmuxSessions, err := tmuxClient.ListSessions(ctx)
			if err == nil {
				for _, ts := range tmuxSessions {
					if strings.HasPrefix(ts.Name, cfg.SessionPrefix+"-") {
						agentName := strings.TrimPrefix(ts.Name, cfg.SessionPrefix+"-")
						if _, err := mgr.Get(ctx, agentName); err != nil {
							// This tmux session is not tracked — it's an orphan
							if dryRun {
								fmt.Printf("  Would kill orphaned tmux session: %s\n", ts.Name)
							} else {
								_ = tmuxClient.KillSession(ctx, ts.Name)
								fmt.Printf("  ✓ Killed orphaned tmux session: %s\n", ts.Name)
							}
							cleaned++
						}
					}
				}
			}
		}

		if cleaned == 0 {
			fmt.Println("✨ No orphaned sessions found. Everything is clean!")
		} else if dryRun {
			fmt.Printf("\n%d item(s) would be cleaned. Run without --dry-run to apply.\n", cleaned)
		} else {
			fmt.Printf("\n✓ Cleaned %d item(s).\n", cleaned)
		}

		return nil
	},
}

func printInstallHint(tool string) {
	switch tool {
	case "tmux":
		switch runtime.GOOS {
		case "darwin":
			fmt.Println("      Install: brew install tmux")
		case "linux":
			fmt.Println("      Install: sudo apt install tmux  (or your distro's package manager)")
		default:
			fmt.Println("      Install: see https://github.com/tmux/tmux/wiki/Installing")
		}
	}
}

func init() {
	cleanupCmd.Flags().Bool("dry-run", false, "show what would be cleaned without making changes")
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(cleanupCmd)
}

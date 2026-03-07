package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

// savedSession holds the config needed to re-launch a stopped agent.
type savedSession struct {
	Name      string            `json:"name"`
	AgentType string            `json:"agent_type"`
	WorkDir   string            `json:"work_dir"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Group     string            `json:"group,omitempty"`
}

var resumeCmd = &cobra.Command{
	Use:   "resume <agent-name>",
	Short: "Re-launch a previously stopped agent with its saved configuration",
	Long: `Resume a stopped agent session using its previously saved configuration.

The agent will be re-created with the same agent type, working directory,
arguments, and environment variables as the original session.

Use 'agentmux resume --list' to see available saved sessions.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		listMode, _ := cmd.Flags().GetBool("list")

		sessionsFile := filepath.Join(cfg.DataDir, "saved_sessions.json")

		if listMode {
			return listSavedSessions(sessionsFile)
		}

		if len(args) == 0 {
			return fmt.Errorf("agent name is required (or use --list)")
		}

		agentName := args[0]
		return resumeSession(sessionsFile, agentName)
	},
}

var saveCmd = &cobra.Command{
	Use:   "save <agent-name>",
	Short: "Save a running agent's configuration for later resuming",
	Long: `Save the configuration of a running agent so it can be resumed later
with 'agentmux resume <name>'.

This is useful for preserving session configurations across reboots.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := args[0]
		sessionsFile := filepath.Join(cfg.DataDir, "saved_sessions.json")

		mgr, err := session.NewManager(cfg)
		if err != nil {
			return err
		}

		sess, err := mgr.Get(cmd.Context(), agentName)
		if err != nil {
			return fmt.Errorf("agent %q not found: %w", agentName, err)
		}

		saved := savedSession{
			Name:      sess.Name,
			AgentType: sess.AgentType,
			WorkDir:   sess.WorkDir,
			Group:     sess.Group,
		}

		if err := addSavedSession(sessionsFile, saved); err != nil {
			return err
		}

		fmt.Printf("✓ Saved configuration for agent %q\n", agentName)
		fmt.Printf("  Resume with: agentmux resume %s\n", agentName)
		return nil
	},
}

func listSavedSessions(filePath string) error {
	sessions, err := loadSavedSessions(filePath)
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No saved sessions. Use 'agentmux save <name>' to save a running agent's config.")
		return nil
	}

	fmt.Println("📋 Saved Sessions")
	fmt.Println(strings.Repeat("─", 50))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tWORKDIR\tGROUP")
	for _, s := range sessions {
		group := s.Group
		if group == "" {
			group = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.AgentType, s.WorkDir, group)
	}
	w.Flush()

	fmt.Printf("\nResume with: agentmux resume <name>\n")
	return nil
}

func resumeSession(filePath string, name string) error {
	sessions, err := loadSavedSessions(filePath)
	if err != nil {
		return err
	}

	var saved *savedSession
	for i, s := range sessions {
		if s.Name == name {
			saved = &sessions[i]
			break
		}
	}

	if saved == nil {
		return fmt.Errorf("no saved session %q found (run 'agentmux resume --list')", name)
	}

	mgr, err := session.NewManager(cfg)
	if err != nil {
		return err
	}

	runner := agent.NewRunner(cfg, mgr)
	opts := agent.LaunchOptions{
		Name:      saved.Name,
		AgentType: saved.AgentType,
		WorkDir:   saved.WorkDir,
		ExtraArgs: saved.Args,
		Command:   saved.Command,
		Env:       saved.Env,
		Group:     saved.Group,
	}

	agentSession, err := runner.Launch(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("resuming agent %q: %w", name, err)
	}

	fmt.Printf("✓ Resumed agent %q (tmux: %s, workdir: %s)\n",
		agentSession.Name, agentSession.TmuxName, agentSession.WorkDir)
	fmt.Printf("  Attach:  agentmux attach %s\n", agentSession.Name)
	fmt.Printf("  Stop:    agentmux stop %s\n", agentSession.Name)
	return nil
}

func loadSavedSessions(filePath string) ([]savedSession, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sessions []savedSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("parsing saved sessions: %w", err)
	}
	return sessions, nil
}

func addSavedSession(filePath string, saved savedSession) error {
	sessions, err := loadSavedSessions(filePath)
	if err != nil {
		return err
	}

	// Replace existing or append
	found := false
	for i, s := range sessions {
		if s.Name == saved.Name {
			sessions[i] = saved
			found = true
			break
		}
	}
	if !found {
		sessions = append(sessions, saved)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0o644)
}

func init() {
	resumeCmd.Flags().Bool("list", false, "list all saved sessions")
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(saveCmd)
}

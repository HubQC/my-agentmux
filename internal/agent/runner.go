package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/session"
)

// Runner handles launching agent CLI processes inside tmux sessions.
type Runner struct {
	cfg        *config.Config
	sessionMgr *session.Manager
}

// NewRunner creates a new agent runner.
func NewRunner(cfg *config.Config, sessionMgr *session.Manager) *Runner {
	return &Runner{
		cfg:        cfg,
		sessionMgr: sessionMgr,
	}
}

// LaunchOptions configures an agent launch.
type LaunchOptions struct {
	Name      string
	AgentType string
	WorkDir   string
	ExtraArgs []string
	Env       map[string]string
	Command   string // override: use this exact command instead of preset
}

// Launch starts an agent CLI in a new tmux session.
func (r *Runner) Launch(ctx context.Context, opts LaunchOptions) (*session.AgentSession, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	// Resolve agent type
	agentType := opts.AgentType
	if agentType == "" {
		agentType = r.cfg.DefaultAgentType
	}

	// Build the command to run
	command := opts.Command
	if command == "" && agentType != "shell" {
		preset, err := GetPreset(agentType)
		if err != nil {
			return nil, err
		}

		if preset.Binary != "" && !preset.IsInstalled() {
			return nil, fmt.Errorf("agent %q binary %q not found on PATH (install it first)", agentType, preset.Binary)
		}

		if preset.Binary != "" {
			command = preset.Command(opts.ExtraArgs)
		}

		// Merge preset env with user env
		if preset.Env != nil {
			if opts.Env == nil {
				opts.Env = make(map[string]string)
			}
			for k, v := range preset.Env {
				if _, exists := opts.Env[k]; !exists {
					opts.Env[k] = v
				}
			}
		}
	} else if command == "" && agentType == "shell" {
		// Plain shell — no command, tmux starts user's default shell
		command = ""
	}

	// Always inject basic AgentMux env variables
	if opts.Env == nil {
		opts.Env = make(map[string]string)
	}
	opts.Env["AGENTMUX_AGENT_NAME"] = opts.Name
	opts.Env["AGENTMUX_SESSION_PREFIX"] = r.cfg.SessionPrefix

	// Create the session via session manager
	createOpts := session.CreateOptions{
		Name:      opts.Name,
		AgentType: agentType,
		WorkDir:   opts.WorkDir,
		Command:   command,
		Env:       opts.Env,
	}

	agentSession, err := r.sessionMgr.Create(ctx, createOpts)
	if err != nil {
		return nil, fmt.Errorf("creating agent session: %w", err)
	}

	return agentSession, nil
}

// LaunchCustom starts a custom command (not a preset) in a new tmux session.
func (r *Runner) LaunchCustom(ctx context.Context, name string, command string, workDir string) (*session.AgentSession, error) {
	return r.Launch(ctx, LaunchOptions{
		Name:      name,
		AgentType: "custom",
		WorkDir:   workDir,
		Command:   command,
	})
}

// FormatAgentCommand returns a human-readable description of what command
// would be run for a given agent type.
func FormatAgentCommand(agentType string, extraArgs []string) string {
	if agentType == "shell" {
		return "(default shell)"
	}

	preset, err := GetPreset(agentType)
	if err != nil {
		return fmt.Sprintf("(unknown: %s)", agentType)
	}

	cmd := preset.Command(extraArgs)
	if cmd == "" {
		return "(default shell)"
	}

	// Truncate long commands
	if len(cmd) > 60 {
		cmd = cmd[:57] + "..."
	}

	return cmd
}

// ValidateAgentType checks if an agent type is valid (preset or special).
func ValidateAgentType(agentType string) error {
	specialTypes := []string{"shell", "custom"}
	for _, st := range specialTypes {
		if agentType == st {
			return nil
		}
	}

	_, err := GetPreset(agentType)
	if err != nil {
		available := append([]string{"shell", "custom"}, strings.Split(AvailablePresets(), ", ")...)
		return fmt.Errorf("unknown agent type %q (available: %s)", agentType, strings.Join(available, ", "))
	}
	return nil
}

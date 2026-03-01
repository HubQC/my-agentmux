package cmd

import (
	"fmt"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <agent-name>",
	Short: "Start a new agent session",
	Long: `Create and start a new agent session in an isolated tmux session.

The agent will run in a detached tmux session that you can attach to,
monitor, or send commands to.

By default, the configured agent type (e.g. claude) is launched.
Use --agent-type to pick a different preset, --command to run an
arbitrary command, or --args to pass extra arguments to the agent CLI.

If a custom agent definition exists with the given name (in
~/.agentmux/agents/), its settings are used as defaults.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := args[0]

		workDir, _ := cmd.Flags().GetString("workdir")
		agentType, _ := cmd.Flags().GetString("agent-type")
		extraArgs, _ := cmd.Flags().GetStringSlice("args")
		customCmd, _ := cmd.Flags().GetString("command")
		projectConfig, _ := cmd.Flags().GetString("config")

		// Apply project-level config if specified
		activeCfg := cfg
		if projectConfig != "" {
			projCfg, err := config.LoadProjectConfigFile(projectConfig)
			if err != nil {
				return fmt.Errorf("loading project config: %w", err)
			}
			if projCfg != nil {
				activeCfg = config.MergeProjectConfig(cfg, projCfg)
			}
		}

		// Check for custom agent definition matching the name
		agentDef, _ := config.GetAgentDefinition(activeCfg.AgentsDir(), agentName)
		if agentDef != nil {
			// Apply definition defaults (flags override)
			if agentType == "" {
				agentType = agentDef.AgentType
			}
			if workDir == "" {
				workDir = agentDef.WorkDir
			}
			if len(extraArgs) == 0 && len(agentDef.Args) > 0 {
				extraArgs = agentDef.Args
			}
		}

		// When --command is specified without --agent-type, use "custom"
		if customCmd != "" && agentType == "" {
			agentType = "custom"
		}

		// Validate agent type early (unless using --command)
		if customCmd == "" {
			resolvedType := agentType
			if resolvedType == "" {
				resolvedType = activeCfg.DefaultAgentType
			}
			if err := agent.ValidateAgentType(resolvedType); err != nil {
				return err
			}
		}

		// Create session manager and runner
		mgr, err := session.NewManager(activeCfg)
		if err != nil {
			return err
		}

		runner := agent.NewRunner(activeCfg, mgr)

		// Merge environment from definition
		var env map[string]string
		if agentDef != nil && len(agentDef.Env) > 0 {
			env = agentDef.Env
		}

		opts := agent.LaunchOptions{
			Name:      agentName,
			AgentType: agentType,
			WorkDir:   workDir,
			ExtraArgs: extraArgs,
			Command:   customCmd,
			Env:       env,
		}

		agentSession, err := runner.Launch(cmd.Context(), opts)
		if err != nil {
			return err
		}

		// Show what command is running
		cmdDesc := agent.FormatAgentCommand(agentSession.AgentType, extraArgs)
		if customCmd != "" {
			cmdDesc = customCmd
		}

		fmt.Printf("✓ Agent %q started (tmux: %s, workdir: %s)\n",
			agentSession.Name, agentSession.TmuxName, agentSession.WorkDir)
		fmt.Printf("  Command: %s\n", cmdDesc)
		if agentDef != nil {
			fmt.Printf("  Definition: %s\n", agentDef.SourceFile)
		}
		fmt.Printf("  Attach:  agentmux attach %s\n", agentSession.Name)
		fmt.Printf("  Stop:    agentmux stop %s\n", agentSession.Name)

		return nil
	},
}

func init() {
	startCmd.Flags().StringP("workdir", "w", "", "working directory for the agent (default: current dir)")
	startCmd.Flags().StringP("agent-type", "t", "", "agent type preset (default: from config)")
	startCmd.Flags().StringSliceP("args", "a", nil, "extra arguments to pass to the agent CLI")
	startCmd.Flags().StringP("command", "c", "", "custom command to run (overrides agent type preset)")
	startCmd.Flags().String("config", "", "project-level config file path")

	rootCmd.AddCommand(startCmd)
}

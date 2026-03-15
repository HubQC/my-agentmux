package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/cqi/my_agentmux/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	cfg       *config.Config
	appAssets embed.FS
)

var rootCmd = &cobra.Command{
	Use:   "agentmux",
	Short: "A blazing fast multi-agent orchestrator",
	Long: `agentmux — Personal Multi-Agent Orchestrator

A tmux-based orchestrator for managing multiple AI coding agents
in parallel. Works in any terminal emulator or IDE.

Run multiple agents simultaneously, monitor their output in real-time,
and intervene at any time. Supports Claude Code, Aider, Codex,
Gemini CLI, and any other CLI-based coding agent.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute(assets embed.FS) error {
	appAssets = assets
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.agentmux/config.yaml)")

	// Register subcommands
	rootCmd.AddCommand(versionCmd)
}

// GetConfig returns the loaded configuration. Available after PersistentPreRunE.
func GetConfig() *config.Config {
	return cfg
}

// PrintError prints a formatted error message to stderr.
func PrintError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
}

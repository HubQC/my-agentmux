//go:build desktop

package cmd

import (
	"github.com/cqi/my_agentmux/internal/desktop"
	"github.com/spf13/cobra"
)

var desktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Launch the AgentMux desktop application",
	Long: `Start the Wails-based desktop GUI for AgentMux.
The desktop app provides a graphical interface for managing and
interacting with multiple agent sessions simultaneously.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return desktop.Run(appAssets, cfg)
	},
}

func init() {
	rootCmd.AddCommand(desktopCmd)
}

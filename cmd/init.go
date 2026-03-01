package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cqi/my_agentmux/internal/devcontainer"
	"github.com/cqi/my_agentmux/internal/platform"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize agentmux configuration",
	Long: `Create default agentmux configuration and optionally generate
a devcontainer configuration for the current project.

This creates:
  - ~/.agentmux/config.yaml (global config)
  - ~/.agentmux/agents/ (agent definitions directory)
  - A sample agent definition

With --devcontainer, also generates:
  - .devcontainer/devcontainer.json (in current directory)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		genDevcontainer, _ := cmd.Flags().GetBool("devcontainer")
		force, _ := cmd.Flags().GetBool("force")

		created := 0

		// 1. Ensure config file exists
		configFile := filepath.Join(cfg.DataDir, "config.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) || force {
			if err := cfg.Save(configFile); err != nil {
				return fmt.Errorf("creating config: %w", err)
			}
			fmt.Printf("✓ Created %s\n", configFile)
			created++
		} else {
			fmt.Printf("  Exists: %s\n", configFile)
		}

		// 2. Create agents directory with sample
		agentsDir := cfg.AgentsDir()
		sampleFile := filepath.Join(agentsDir, "example.md")
		if _, err := os.Stat(sampleFile); os.IsNotExist(err) || force {
			sampleContent := `---
description: Example agent definition
agent_type: claude
args: []
---

You are a helpful coding assistant. Focus on writing clean,
well-tested code with clear documentation.
`
			if err := os.MkdirAll(agentsDir, 0o755); err != nil {
				return fmt.Errorf("creating agents directory: %w", err)
			}
			if err := os.WriteFile(sampleFile, []byte(sampleContent), 0o644); err != nil {
				return fmt.Errorf("creating sample agent: %w", err)
			}
			fmt.Printf("✓ Created %s\n", sampleFile)
			created++
		} else {
			fmt.Printf("  Exists: %s\n", sampleFile)
		}

		// 3. Generate devcontainer config
		if genDevcontainer {
			cwd, _ := os.Getwd()
			dcConfig := devcontainer.DefaultConfig()
			filePath, err := devcontainer.Generate(cwd, dcConfig)
			if err != nil {
				return fmt.Errorf("generating devcontainer: %w", err)
			}
			fmt.Printf("✓ Created %s\n", filePath)
			created++
		}

		// 4. Show platform info
		info := platform.Detect()
		fmt.Printf("\n  Platform: %s/%s", info.OS, info.Arch)
		if info.IsWSL {
			fmt.Print(" (WSL)")
		}
		if info.IsDocker {
			fmt.Print(" (Docker)")
		}
		fmt.Println()

		if created == 0 {
			fmt.Println("\n  Everything already initialized. Use --force to overwrite.")
		} else {
			fmt.Printf("\n✓ Initialized %d file(s)\n", created)
		}

		return nil
	},
}

func init() {
	initCmd.Flags().Bool("devcontainer", false, "generate .devcontainer/devcontainer.json")
	initCmd.Flags().Bool("force", false, "overwrite existing files")

	rootCmd.AddCommand(initCmd)
}

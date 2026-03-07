package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cqi/my_agentmux/internal/codex"
	"github.com/spf13/cobra"
)

var codexCmd = &cobra.Command{
	Use:   "codex",
	Short: "Show available Codex configurations and usage advice",
	Long: `Provides a deep-dive view into your available Codex profiles
and sub-agents found in ~/.codex/config.toml (or local project overrides).

This command is designed to help you choose the best configuration
for your next session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := codex.LoadConfig()
		if err != nil || cfg == nil {
			fmt.Println("❌ No Codex configuration found at ~/.codex/config.toml or .codex/config.toml.")
			fmt.Println("Please ensure Codex CLI is installed and configured.")
			return nil
		}

		fmt.Println("🤖 CODEX CONFIGURATION ADVISOR")
		fmt.Println("================================")

		if len(cfg.Profiles) > 0 {
			fmt.Println("\n📌 AVAILABLE PROFILES")
			fmt.Println("To launch a profile, run: agentmux start <session-name> -t codex -a \"--profile <profile-name>\"")

			var keys []string
			for k := range cfg.Profiles {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				p := cfg.Profiles[k]
				fmt.Printf("\n🔹 Profile: %s\n", k)

				model := p.Model
				if model == "" {
					model = "default"
				}
				fmt.Printf("   Model:       %s (%s)\n", model, p.ModelProvider)

				reasoning := p.ModelReasoningEffort
				if reasoning == "" {
					reasoning = "default"
				} else if reasoning == "high" {
					reasoning = "high (🔥 Recommended for complex logic & heavy refactoring)"
				} else if reasoning == "medium" {
					reasoning = "medium (⚡ Balanced speed and reasoning)"
				} else if reasoning == "low" {
					reasoning = "low (🚀 Best for fast, simple code completions)"
				}
				fmt.Printf("   Reasoning:   %s\n", reasoning)

				if p.ReviewModel != "" {
					fmt.Printf("   Reviewer:    %s\n", p.ReviewModel)
				}
				if p.Personality != "" {
					fmt.Printf("   Personality: Custom configuration applied\n")
				}
			}
		} else {
			fmt.Println("\nNo Codex profiles found in your config.")
		}

		if len(cfg.Agents) > 0 {
			fmt.Println("\n\n🛠️  SUB-AGENTS (Multi-Agent Roles)")
			fmt.Println("To launch a specific sub-agent, use its config file as an argument.")

			var keys []string
			for k := range cfg.Agents {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				a := cfg.Agents[k]
				fmt.Printf("\n🔸 Agent: %s\n", k)
				fmt.Printf("   Config:      %s\n", a.ConfigFile)
				desc := a.Description
				if desc != "" {
					// Wrap text for better readability
					desc = strings.ReplaceAll(desc, "\n", " ")
					if len(desc) > 80 {
						desc = desc[:77] + "..."
					}
					fmt.Printf("   Description: %s\n", desc)
				}
			}
		}

		fmt.Println("\n💡 Quick Tip: Create custom agents in ~/.agentmux/agents/ to save your favorite profiles! See docs/CODEX.md")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(codexCmd)
}

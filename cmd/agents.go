package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/codex"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "List available agent definitions",
	Long: `List all available agent types, including built-in presets
and custom agent definitions from the agents directory.

Custom agents are defined as Markdown files with YAML frontmatter
in ~/.agentmux/agents/ (or the configured agents directory).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showAll, _ := cmd.Flags().GetBool("all")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Built-in presets
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tDESCRIPTION")
		fmt.Fprintln(w, "----\t----\t------\t-----------")

		presets := agent.ListPresets()
		sort.Slice(presets, func(i, j int) bool {
			return presets[i].Name < presets[j].Name
		})

		for _, p := range presets {
			status := "○ not found"
			if p.Binary == "" {
				status = "● built-in"
			} else if p.IsInstalled() {
				status = "● installed"
			}

			if !showAll && status == "○ not found" {
				continue
			}

			fmt.Fprintf(w, "%s\tpreset\t%s\t%s\n", p.Name, status, p.Description)
		}

		// Custom agent definitions
		defs, err := config.LoadAgentDefinitions(cfg.AgentsDir())
		if err == nil {
			sort.Slice(defs, func(i, j int) bool {
				return defs[i].Name < defs[j].Name
			})

			for _, d := range defs {
				desc := d.Description
				if desc == "" {
					desc = fmt.Sprintf("Custom (%s)", d.AgentType)
				}
				// Truncate long descriptions
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				fmt.Fprintf(w, "%s\tcustom\t● defined\t%s\n", d.Name, desc)
			}
		}

		w.Flush()

		// Codex specific configurations
		codexCfg, err := codex.LoadConfig()
		if err == nil && codexCfg != nil {
			// Print Codex Profiles
			hasProfiles := len(codexCfg.Profiles) > 0
			if hasProfiles {
				fmt.Fprintln(w, "\nCODEX PROFILES\tMODEL/PROVIDER\tREASONING\tDESCRIPTION")
				fmt.Fprintln(w, "--------------\t--------------\t---------\t-----------")

				// Get sorted profile keys for consistent output
				var profileKeys []string
				for key := range codexCfg.Profiles {
					profileKeys = append(profileKeys, key)
				}
				sort.Strings(profileKeys)

				for _, key := range profileKeys {
					profile := codexCfg.Profiles[key]
					provider := profile.ModelProvider
					if profile.Model != "" {
						provider = fmt.Sprintf("%s (%s)", profile.Model, profile.ModelProvider)
					}
					reasoning := profile.ModelReasoningEffort
					if reasoning == "" {
						reasoning = "-"
					}

					// Basic description string parsing
					desc := fmt.Sprintf("Codex configuration profile for %s", key)
					if profile.Personality != "" {
						desc += " [Custom Personality]"
					}

					if profile.ReviewModel != "" {
						desc += fmt.Sprintf(" (Reviewer: %s)", profile.ReviewModel)
					}

					if len(desc) > 50 {
						desc = desc[:47] + "..."
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", key, provider, reasoning, desc)
				}
				w.Flush()
			}

			// Print Codex Agents (Task Planner, Supervisor, etc)
			hasAgents := len(codexCfg.Agents) > 0
			if hasAgents {
				fmt.Fprintln(w, "\nCODEX AGENTS\tCONFIG FILE\tDESCRIPTION")
				fmt.Fprintln(w, "------------\t-----------\t-----------")

				var agentKeys []string
				for key := range codexCfg.Agents {
					agentKeys = append(agentKeys, key)
				}
				sort.Strings(agentKeys)

				for _, key := range agentKeys {
					a := codexCfg.Agents[key]

					desc := a.Description
					if desc == "" {
						desc = fmt.Sprintf("Codex multi-agent role: %s", key)
					}

					if len(desc) > 50 {
						desc = desc[:47] + "..."
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", key, a.ConfigFile, desc)
				}
				w.Flush()
			}
		}

		return nil
	},
}

func init() {
	agentsCmd.Flags().BoolP("all", "a", false, "show all presets including uninstalled ones")

	rootCmd.AddCommand(agentsCmd)
}

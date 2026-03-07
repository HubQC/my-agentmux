package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cqi/my_agentmux/internal/templates"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"tmpl"},
	Short:   "List and install built-in agent templates",
	Long: `Browse and install curated agent definition templates.

Templates are pre-configured agent definitions for common tasks like
code review, test writing, documentation, and security auditing.

Install a template to create an agent definition file that you can
customize and launch.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tag, _ := cmd.Flags().GetString("tag")

		all := templates.BuiltinTemplates()
		if tag != "" {
			all = templates.FindByTag(tag)
			if len(all) == 0 {
				fmt.Printf("No templates found with tag %q\n", tag)
				return nil
			}
		}

		fmt.Println("📦 Agent Templates")
		fmt.Println(strings.Repeat("─", 60))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION\tTAGS")
		for _, t := range all {
			tags := strings.Join(t.Tags, ", ")
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Name, t.AgentType, t.Description, tags)
		}
		w.Flush()

		fmt.Printf("\nInstall: agentmux templates install <name>\n")
		return nil
	},
}

var templatesInstallCmd = &cobra.Command{
	Use:   "install <template-name>",
	Short: "Install a template as an agent definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		tmpl := templates.FindByName(name)
		if tmpl == nil {
			fmt.Printf("❌ Template %q not found.\n\nAvailable templates:\n", name)
			for _, t := range templates.BuiltinTemplates() {
				fmt.Printf("  • %s — %s\n", t.Name, t.Description)
			}
			return nil
		}

		agentsDir := cfg.AgentsDir()
		path, err := templates.InstallTemplate(*tmpl, agentsDir)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Installed template %q\n", name)
		fmt.Printf("   File: %s\n", path)
		fmt.Printf("   Launch: agentmux start my-agent -t %s\n", name)
		fmt.Printf("   Customize the file to adjust the prompt and settings.\n")
		return nil
	},
}

func init() {
	templatesCmd.Flags().String("tag", "", "filter templates by tag (e.g., quality, security)")
	templatesCmd.AddCommand(templatesInstallCmd)
	rootCmd.AddCommand(templatesCmd)
}

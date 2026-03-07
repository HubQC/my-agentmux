package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cqi/my_agentmux/internal/gemini"
	"github.com/spf13/cobra"
)

// geminiCmd represents the gemini command
var geminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Inspect the active Gemini CLI configuration and MCP servers",
	Long: `The 'gemini' command interrogates the local ~/.gemini/settings.json 
configuration. This is a read-only advisory view that shows your active
MCP servers.

Usage Examples:
  agentmux gemini
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load Gemini Configuration
		cfg, err := gemini.LoadConfig()
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No Gemini configuration found. Have you installed and run the Gemini CLI yet?")
				fmt.Println("Looking for configuration at: ~/.gemini/settings.json")
				return nil
			}
			return fmt.Errorf("failed to load gemini configuration: %w", err)
		}

		fmt.Println("🚀 Gemini Environment Overview")
		fmt.Println("========================================")

		printGeminiMCPs(cfg)

		fmt.Println("\n💡 To launch an agent session with Gemini:")
		fmt.Println("   agentmux launch my-gemini --agent gemini")
		fmt.Println("\n   Your specific MCP settings are passed implicitly.")
		return nil
	},
}

func printGeminiMCPs(cfg *gemini.Config) {
	if len(cfg.MCPServers) == 0 {
		fmt.Println("\n🔌 MCP Servers:")
		fmt.Println("  No MCP servers configured.")
		fmt.Println("  Get started by adding servers via the gemini cli commands or editing ~/.gemini/settings.json manually.")
		return
	}

	fmt.Printf("\n🔌 MCP Servers (%d configured):\n", len(cfg.MCPServers))
	fmt.Println("  --------------------------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  SERVER NAME\tCOMMAND\tARGS")

	// Sort server names to ensure deterministic output
	var names []string
	for name := range cfg.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		mcp := cfg.MCPServers[name]
		argsStr := strings.Join(mcp.Args, " ")
		if len(argsStr) > 40 {
			argsStr = argsStr[:37] + "..."
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\n", name, mcp.Command, argsStr)
	}
	w.Flush()
}

func init() {
	rootCmd.AddCommand(geminiCmd)
}

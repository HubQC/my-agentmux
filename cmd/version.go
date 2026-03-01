package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information — set at build time via ldflags
var (
	Version   = "0.1.0-dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of agentmux",
	Long:  "Display the version, git commit, build date, and platform information.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("agentmux %s\n", Version)
		fmt.Printf("  commit:   %s\n", GitCommit)
		fmt.Printf("  built:    %s\n", BuildDate)
		fmt.Printf("  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  go:       %s\n", runtime.Version())
	},
}

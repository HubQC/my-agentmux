package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/cqi/my_agentmux/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := cmd.Execute(assets); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

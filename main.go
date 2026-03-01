package main

import (
	"os"

	"github.com/cqi/my_agentmux/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package devcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents a devcontainer.json configuration.
type Config struct {
	Name           string            `json:"name"`
	Image          string            `json:"image,omitempty"`
	Features       map[string]any    `json:"features,omitempty"`
	ForwardPorts   []int             `json:"forwardPorts,omitempty"`
	PostCreateCmd  string            `json:"postCreateCommand,omitempty"`
	Customizations map[string]any    `json:"customizations,omitempty"`
	RemoteEnv      map[string]string `json:"remoteEnv,omitempty"`
}

// DefaultConfig returns a devcontainer config suitable for agentmux development.
func DefaultConfig() Config {
	return Config{
		Name:  "agentmux-dev",
		Image: "mcr.microsoft.com/devcontainers/base:ubuntu",
		Features: map[string]any{
			"ghcr.io/devcontainers/features/go:1":         map[string]string{"version": "latest"},
			"ghcr.io/devcontainers-extra/features/tmux:1": map[string]string{},
		},
		ForwardPorts:  []int{},
		PostCreateCmd: "go mod download && go build -o agentmux .",
		Customizations: map[string]any{
			"vscode": map[string]any{
				"extensions": []string{
					"golang.go",
					"ms-vscode.makefile-tools",
				},
				"settings": map[string]any{
					"go.useLanguageServer":                     true,
					"terminal.integrated.defaultProfile.linux": "zsh",
				},
			},
		},
		RemoteEnv: map[string]string{
			"AGENTMUX_DATA_DIR": "/home/vscode/.agentmux",
		},
	}
}

// ConfigWithAgents returns a config that includes install commands for agent CLIs.
func ConfigWithAgents(agents []string) Config {
	cfg := DefaultConfig()

	if len(agents) > 0 {
		var installCmds string
		for _, agent := range agents {
			switch agent {
			case "claude":
				installCmds += " && npm install -g @anthropic-ai/claude-code"
			case "aider":
				installCmds += " && pip install aider-chat"
			case "codex":
				installCmds += " && npm install -g @openai/codex"
			}
		}
		if installCmds != "" {
			cfg.PostCreateCmd += installCmds
		}
	}

	return cfg
}

// Generate writes a devcontainer.json file to the given output directory.
func Generate(outputDir string, cfg Config) (string, error) {
	devcontainerDir := filepath.Join(outputDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0o755); err != nil {
		return "", fmt.Errorf("creating .devcontainer directory: %w", err)
	}

	filePath := filepath.Join(devcontainerDir, "devcontainer.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshalling devcontainer config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", fmt.Errorf("writing devcontainer.json: %w", err)
	}

	return filePath, nil
}

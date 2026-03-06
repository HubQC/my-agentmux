package agent

import (
	"fmt"
	"os/exec"
	"strings"
)

// Preset defines a built-in agent CLI configuration.
type Preset struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Binary      string            `json:"binary"`
	DefaultArgs []string          `json:"default_args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	CheckCmd    string            `json:"check_cmd,omitempty"` // command to verify installation
}

// Command returns the full command string for this preset with the given extra args.
func (p Preset) Command(extraArgs []string) string {
	parts := []string{p.Binary}
	parts = append(parts, p.DefaultArgs...)
	parts = append(parts, extraArgs...)
	return strings.Join(parts, " ")
}

// IsInstalled checks if the agent CLI binary is available on PATH.
func (p Preset) IsInstalled() bool {
	_, err := exec.LookPath(p.Binary)
	return err == nil
}

// builtinPresets contains all built-in agent presets.
var builtinPresets = map[string]Preset{
	"claude": {
		Name:        "claude",
		Description: "Anthropic Claude Code — AI coding assistant",
		Binary:      "claude",
		DefaultArgs: []string{},
		CheckCmd:    "claude --version",
	},
	"aider": {
		Name:        "aider",
		Description: "Aider — AI pair programming in your terminal",
		Binary:      "aider",
		DefaultArgs: []string{},
		CheckCmd:    "aider --version",
	},
	"codex": {
		Name:        "codex",
		Description: "OpenAI Codex CLI — AI coding agent",
		Binary:      "codex",
		DefaultArgs: []string{},
		CheckCmd:    "codex --version",
	},
	"gemini": {
		Name:        "gemini",
		Description: "Google Gemini CLI — AI coding assistant",
		Binary:      "gemini",
		DefaultArgs: []string{},
		CheckCmd:    "gemini --version",
	},
	"copilot": {
		Name:        "copilot",
		Description: "GitHub Copilot CLI",
		Binary:      "gh",
		DefaultArgs: []string{"copilot"},
		CheckCmd:    "gh copilot --version",
	},
	"cline": {
		Name:        "cline",
		Description: "Cline — CLI coding agent",
		Binary:      "cline",
		DefaultArgs: []string{},
		CheckCmd:    "cline --version",
	},
	"openhands": {
		Name:        "openhands",
		Description: "OpenHands — CLI wrapper for OpenHands",
		Binary:      "openhands",
		DefaultArgs: []string{},
		CheckCmd:    "openhands --version",
	},
	"ollama": {
		Name:        "ollama",
		Description: "Ollama — Run large language models locally",
		Binary:      "ollama",
		DefaultArgs: []string{"run"},
		CheckCmd:    "ollama --version",
	},
	"shell": {
		Name:        "shell",
		Description: "Plain shell session (bash/zsh)",
		Binary:      "",
		DefaultArgs: []string{},
	},
}

// GetPreset returns a built-in agent preset by name.
func GetPreset(name string) (Preset, error) {
	preset, ok := builtinPresets[name]
	if !ok {
		return Preset{}, fmt.Errorf("unknown agent type %q (available: %s)", name, AvailablePresets())
	}
	return preset, nil
}

// ListPresets returns all built-in presets.
func ListPresets() []Preset {
	presets := make([]Preset, 0, len(builtinPresets))
	for _, p := range builtinPresets {
		presets = append(presets, p)
	}
	return presets
}

// AvailablePresets returns a comma-separated list of preset names.
func AvailablePresets() string {
	names := make([]string, 0, len(builtinPresets))
	for name := range builtinPresets {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

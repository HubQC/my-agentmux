package codex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// MCPConfig represents a Model Context Protocol server configuration
type MCPConfig struct {
	Type             string            `toml:"type"`
	Command          string            `toml:"command"`
	Args             []string          `toml:"args"`
	Env              map[string]string `toml:"env"`
	StartupTimeoutMs int               `toml:"startup_timeout_ms"`
}

// ProfileConfig represents a specialized agent profile (e.g., gpt-5.3-codex)
type ProfileConfig struct {
	Model                string `toml:"model"`
	ModelProvider        string `toml:"model_provider"`
	ReviewModel          string `toml:"review_model"`
	Personality          string `toml:"personality"`
	ModelReasoningEffort string `toml:"model_reasoning_effort"`
}

// AgentConfig represents a sub-agent configuration (supervisor, task_planner)
type AgentConfig struct {
	ConfigFile  string `toml:"config_file"`
	Description string `toml:"description"`
}

// Config represents the top-level Codex configuration structure
type Config struct {
	Model                string `toml:"model"`
	Profile              string `toml:"profile"`
	ModelReasoningEffort string `toml:"model_reasoning_effort"`
	Features             struct {
		MultiAgent bool `toml:"multi_agent"`
	} `toml:"features"`
	Profiles   map[string]ProfileConfig `toml:"profiles"`
	MCPServers map[string]MCPConfig     `toml:"mcp_servers"`
	Agents     map[string]AgentConfig   `toml:"agents"`
}

// LoadConfig reads and parses the ~/.codex/config.toml file
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not an error if the user doesn't have codex installed
		}
		return nil, fmt.Errorf("reading codex config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing codex config: %w", err)
	}

	return &cfg, nil
}

package gemini

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the parsed contents of a Gemini settings.json
type Config struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single MCP server definition
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// LoadConfig attempts to parse the global Gemini settings from ~/.gemini/settings.json.
// It returns a generic config struct or an error if it fails to parse.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".gemini", "settings.json")
	return loadFromFile(configPath)
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

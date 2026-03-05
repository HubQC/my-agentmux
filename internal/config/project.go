package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents a project-level configuration file
// found at `.agentmux/config.yaml` in a project directory.
type ProjectConfig struct {
	// Agents maps agent names to their launch overrides.
	Agents map[string]ProjectAgent `yaml:"agents,omitempty"`

	// DefaultWorkDir is the default working directory for agents in this project.
	DefaultWorkDir string `yaml:"default_workdir,omitempty"`

	// DefaultAgentType overrides the global default agent type for this project.
	DefaultAgentType string `yaml:"default_agent_type,omitempty"`

	// Env are environment variables set for all agents in this project.
	Env map[string]string `yaml:"env,omitempty"`

	// Pipelines defines orchestrated sequences of agents (e.g., ["architect", "coder"]).
	Pipelines map[string][]string `yaml:"pipelines,omitempty"`

	// SourceFile is the path to the config file (not serialized).
	SourceFile string `yaml:"-"`
}

// ProjectAgent defines per-agent overrides in a project config.
type ProjectAgent struct {
	AgentType string            `yaml:"agent_type,omitempty"`
	Args      []string          `yaml:"args,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	WorkDir   string            `yaml:"workdir,omitempty"`
}

// LoadProjectConfig loads a project-level config from the given directory.
// It looks for `.agentmux/config.yaml` in the directory.
// Returns nil (no error) if the file doesn't exist.
func LoadProjectConfig(projectDir string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectDir, ".agentmux", "config.yaml")
	return LoadProjectConfigFile(configPath)
}

// LoadProjectConfigFile loads a project config from a specific file path.
func LoadProjectConfigFile(configPath string) (*ProjectConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading project config: %w", err)
	}

	cfg := &ProjectConfig{
		SourceFile: configPath,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing project config %s: %w", configPath, err)
	}

	return cfg, nil
}

// SaveProjectConfig writes the project config to `.agentmux/config.yaml`.
func SaveProjectConfig(projectDir string, cfg *ProjectConfig) error {
	configDir := filepath.Join(projectDir, ".agentmux")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("creating .agentmux directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling project config: %w", err)
	}

	return os.WriteFile(configPath, data, 0o644)
}

// MergeProjectConfig merges project-level overrides into a global config.
func MergeProjectConfig(global *Config, project *ProjectConfig) *Config {
	if project == nil {
		return global
	}

	// Create a copy
	merged := *global

	if project.DefaultAgentType != "" {
		merged.DefaultAgentType = project.DefaultAgentType
	}

	return &merged
}

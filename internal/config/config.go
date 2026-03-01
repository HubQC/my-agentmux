package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds global agentmux configuration.
type Config struct {
	// DataDir is the root directory for agentmux data (sessions, logs, etc.).
	DataDir string `yaml:"data_dir"`

	// DefaultAgentType is the default agent CLI to use when starting agents.
	DefaultAgentType string `yaml:"default_agent_type"`

	// TmuxBinary is the path to the tmux binary.
	TmuxBinary string `yaml:"tmux_binary"`

	// SessionPrefix is the prefix for all agentmux tmux sessions.
	SessionPrefix string `yaml:"session_prefix"`

	// LogLevel controls logging verbosity (debug, info, warn, error).
	LogLevel string `yaml:"log_level"`

	// Monitor configuration.
	Monitor MonitorConfig `yaml:"monitor"`
}

// MonitorConfig holds configuration for agent output monitoring.
type MonitorConfig struct {
	// PollIntervalMs is how often to poll tmux panes for output (milliseconds).
	PollIntervalMs int `yaml:"poll_interval_ms"`

	// MaxLogSizeMB is the max log file size before rotation (MB).
	MaxLogSizeMB int `yaml:"max_log_size_mb"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DataDir:          filepath.Join(homeDir, ".agentmux"),
		DefaultAgentType: "claude",
		TmuxBinary:       "tmux",
		SessionPrefix:    "amux",
		LogLevel:         "info",
		Monitor: MonitorConfig{
			PollIntervalMs: 500,
			MaxLogSizeMB:   50,
		},
	}
}

// Load reads the configuration from the given file path.
// If cfgFile is empty, it looks for ~/.agentmux/config.yaml.
// If the file doesn't exist, it returns defaults.
func Load(cfgFile string) (*Config, error) {
	cfg := DefaultConfig()

	if cfgFile == "" {
		cfgFile = filepath.Join(cfg.DataDir, "config.yaml")
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file — use defaults, ensure data dir exists
			return cfg, ensureDataDir(cfg.DataDir)
		}
		return nil, fmt.Errorf("reading config file %s: %w", cfgFile, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", cfgFile, err)
	}

	// Expand ~ in DataDir if present
	if cfg.DataDir == "" {
		cfg.DataDir = DefaultConfig().DataDir
	}

	return cfg, ensureDataDir(cfg.DataDir)
}

// Save writes the current config to the given file path.
func (c *Config) Save(cfgFile string) error {
	if cfgFile == "" {
		cfgFile = filepath.Join(c.DataDir, "config.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(cfgFile), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	return os.WriteFile(cfgFile, data, 0o644)
}

// SessionsDir returns the path to the sessions state directory.
func (c *Config) SessionsDir() string {
	return filepath.Join(c.DataDir, "sessions")
}

// LogsDir returns the path to the logs directory.
func (c *Config) LogsDir() string {
	return filepath.Join(c.DataDir, "logs")
}

// AgentsDir returns the path to agent definitions directory.
func (c *Config) AgentsDir() string {
	return filepath.Join(c.DataDir, "agents")
}

// PlansDir returns the path to the plans directory.
func (c *Config) PlansDir() string {
	return filepath.Join(c.DataDir, "plans")
}

// ensureDataDir creates the data directory and its subdirectories.
func ensureDataDir(dataDir string) error {
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "sessions"),
		filepath.Join(dataDir, "logs"),
		filepath.Join(dataDir, "agents"),
		filepath.Join(dataDir, "plans"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}

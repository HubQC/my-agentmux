package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TmuxBinary != "tmux" {
		t.Errorf("expected TmuxBinary=tmux, got %s", cfg.TmuxBinary)
	}
	if cfg.SessionPrefix != "amux" {
		t.Errorf("expected SessionPrefix=amux, got %s", cfg.SessionPrefix)
	}
	if cfg.DefaultAgentType != "codex" {
		t.Errorf("expected DefaultAgentType=codex, got %s", cfg.DefaultAgentType)
	}
	if cfg.Monitor.PollIntervalMs != 500 {
		t.Errorf("expected PollIntervalMs=500, got %d", cfg.Monitor.PollIntervalMs)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/tmp/agentmux-test-nonexistent-config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}
	if cfg.TmuxBinary != "tmux" {
		t.Errorf("expected defaults when config missing, got TmuxBinary=%s", cfg.TmuxBinary)
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := []byte(`
data_dir: /tmp/agentmux-test-data
default_agent_type: aider
tmux_binary: /usr/local/bin/tmux
session_prefix: test
log_level: debug
monitor:
  poll_interval_ms: 1000
  max_log_size_mb: 100
`)
	if err := os.WriteFile(cfgFile, content, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DefaultAgentType != "aider" {
		t.Errorf("expected DefaultAgentType=aider, got %s", cfg.DefaultAgentType)
	}
	if cfg.TmuxBinary != "/usr/local/bin/tmux" {
		t.Errorf("expected TmuxBinary=/usr/local/bin/tmux, got %s", cfg.TmuxBinary)
	}
	if cfg.SessionPrefix != "test" {
		t.Errorf("expected SessionPrefix=test, got %s", cfg.SessionPrefix)
	}
	if cfg.Monitor.PollIntervalMs != 1000 {
		t.Errorf("expected PollIntervalMs=1000, got %d", cfg.Monitor.PollIntervalMs)
	}
}

func TestSaveAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.DataDir = filepath.Join(tmpDir, "data")
	cfg.DefaultAgentType = "codex"
	cfg.LogLevel = "debug"

	if err := cfg.Save(cfgFile); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if loaded.DefaultAgentType != "codex" {
		t.Errorf("expected DefaultAgentType=codex after reload, got %s", loaded.DefaultAgentType)
	}
	if loaded.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug after reload, got %s", loaded.LogLevel)
	}
}

func TestConfigDirHelpers(t *testing.T) {
	cfg := &Config{DataDir: "/home/test/.agentmux"}

	if cfg.SessionsDir() != "/home/test/.agentmux/sessions" {
		t.Errorf("unexpected SessionsDir: %s", cfg.SessionsDir())
	}
	if cfg.LogsDir() != "/home/test/.agentmux/logs" {
		t.Errorf("unexpected LogsDir: %s", cfg.LogsDir())
	}
	if cfg.AgentsDir() != "/home/test/.agentmux/agents" {
		t.Errorf("unexpected AgentsDir: %s", cfg.AgentsDir())
	}
	if cfg.PlansDir() != "/home/test/.agentmux/plans" {
		t.Errorf("unexpected PlansDir: %s", cfg.PlansDir())
	}
}

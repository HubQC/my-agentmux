package devcontainer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Name == "" {
		t.Error("expected non-empty name")
	}
	if cfg.Image == "" {
		t.Error("expected non-empty image")
	}
	if cfg.PostCreateCmd == "" {
		t.Error("expected non-empty post create command")
	}
	if len(cfg.Features) == 0 {
		t.Error("expected at least one feature")
	}
}

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()

	filePath, err := Generate(tmpDir, cfg)
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected devcontainer.json to exist")
	}

	// Verify it's in the right path
	expected := filepath.Join(tmpDir, ".devcontainer", "devcontainer.json")
	if filePath != expected {
		t.Errorf("expected path %q, got %q", expected, filePath)
	}

	// Verify it's valid JSON
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["name"] != "agentmux-dev" {
		t.Errorf("expected name 'agentmux-dev', got %v", parsed["name"])
	}
}

func TestConfigWithAgents(t *testing.T) {
	cfg := ConfigWithAgents([]string{"claude", "aider"})

	if cfg.PostCreateCmd == DefaultConfig().PostCreateCmd {
		t.Error("expected post create command to include agent installs")
	}

	// Should contain claude install
	if !containsStr(cfg.PostCreateCmd, "claude-code") {
		t.Error("expected claude install command")
	}

	// Should contain aider install
	if !containsStr(cfg.PostCreateCmd, "aider-chat") {
		t.Error("expected aider install command")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Agent definition tests ---

func TestLoadAgentDefinitionMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	defFile := filepath.Join(tmpDir, "reviewer.md")

	content := `---
description: Code reviewer agent
agent_type: claude
args:
  - "--model"
  - "sonnet"
env:
  REVIEW_MODE: "strict"
workdir: /tmp/reviews
---

You are a code reviewer. Focus on correctness, performance, and style.
Always provide constructive feedback.
`
	if err := os.WriteFile(defFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := LoadAgentDefinition(defFile)
	if err != nil {
		t.Fatalf("failed to load definition: %v", err)
	}

	if def.Name != "reviewer" {
		t.Errorf("expected name 'reviewer', got %q", def.Name)
	}
	if def.Description != "Code reviewer agent" {
		t.Errorf("expected description 'Code reviewer agent', got %q", def.Description)
	}
	if def.AgentType != "claude" {
		t.Errorf("expected agent_type 'claude', got %q", def.AgentType)
	}
	if len(def.Args) != 2 || def.Args[0] != "--model" {
		t.Errorf("expected args [--model sonnet], got %v", def.Args)
	}
	if def.Env["REVIEW_MODE"] != "strict" {
		t.Errorf("expected env REVIEW_MODE=strict, got %v", def.Env)
	}
	if def.WorkDir != "/tmp/reviews" {
		t.Errorf("expected workdir '/tmp/reviews', got %q", def.WorkDir)
	}
	if def.SystemPrompt == "" {
		t.Error("expected non-empty system prompt")
	}
	if def.SourceFile != defFile {
		t.Errorf("expected source file %q, got %q", defFile, def.SourceFile)
	}
}

func TestLoadAgentDefinitionYAML(t *testing.T) {
	tmpDir := t.TempDir()
	defFile := filepath.Join(tmpDir, "coder.yaml")

	content := `description: Coding agent
agent_type: aider
args:
  - "--auto-commits"
`
	if err := os.WriteFile(defFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := LoadAgentDefinition(defFile)
	if err != nil {
		t.Fatalf("failed to load definition: %v", err)
	}

	if def.Name != "coder" {
		t.Errorf("expected name 'coder', got %q", def.Name)
	}
	if def.AgentType != "aider" {
		t.Errorf("expected agent_type 'aider', got %q", def.AgentType)
	}
}

func TestLoadAgentDefinitionPlainMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	defFile := filepath.Join(tmpDir, "helper.md")

	content := `You are a helpful assistant.
Focus on clear explanations.
`
	if err := os.WriteFile(defFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := LoadAgentDefinition(defFile)
	if err != nil {
		t.Fatalf("failed to load definition: %v", err)
	}

	if def.Name != "helper" {
		t.Errorf("expected name 'helper', got %q", def.Name)
	}
	if def.AgentType != "claude" {
		t.Errorf("expected default agent_type 'claude', got %q", def.AgentType)
	}
	if def.SystemPrompt == "" {
		t.Error("expected non-empty system prompt")
	}
}

func TestLoadAgentDefinitions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create several definitions
	files := map[string]string{
		"agent-a.md":       "---\ndescription: Agent A\nagent_type: claude\n---\nPrompt A",
		"agent-b.yaml":     "description: Agent B\nagent_type: aider\n",
		"not-an-agent.txt": "This should be ignored",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	defs, err := LoadAgentDefinitions(tmpDir)
	if err != nil {
		t.Fatalf("failed to load definitions: %v", err)
	}

	if len(defs) != 2 {
		t.Errorf("expected 2 definitions, got %d", len(defs))
	}
}

func TestLoadAgentDefinitionsNonExistent(t *testing.T) {
	defs, err := LoadAgentDefinitions("/nonexistent/path")
	if err != nil {
		t.Fatalf("expected nil error for nonexistent dir, got: %v", err)
	}
	if len(defs) != 0 {
		t.Errorf("expected 0 definitions, got %d", len(defs))
	}
}

func TestGetAgentDefinition(t *testing.T) {
	tmpDir := t.TempDir()
	defFile := filepath.Join(tmpDir, "myagent.md")
	if err := os.WriteFile(defFile, []byte("---\nagent_type: codex\n---\nDo stuff"), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := GetAgentDefinition(tmpDir, "myagent")
	if err != nil {
		t.Fatalf("failed to get definition: %v", err)
	}
	if def.AgentType != "codex" {
		t.Errorf("expected codex, got %q", def.AgentType)
	}

	_, err = GetAgentDefinition(tmpDir, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent definition")
	}
}

func TestParseFrontmatterUnclosed(t *testing.T) {
	content := "---\nkey: value\nno closing marker"
	_, err := LoadAgentDefinition(createTempFile(t, "bad.md", content))
	if err == nil {
		t.Error("expected error for unclosed frontmatter")
	}
}

func TestLoadAgentDefinitionStrictSchema(t *testing.T) {
	tmpDir := t.TempDir()
	defFile := filepath.Join(tmpDir, "strict.yaml")

	content := `description: Strict agent
agent_type: aider
unknown_field: should_fail
`
	if err := os.WriteFile(defFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAgentDefinition(defFile)
	if err == nil {
		t.Error("expected error for unknown field in YAML")
	}
}

// --- Project config tests ---

func TestLoadProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".agentmux")
	os.MkdirAll(configDir, 0o755)

	configContent := `default_agent_type: aider
default_workdir: /tmp/myproject
env:
  PROJECT_NAME: myproject
agents:
  reviewer:
    agent_type: claude
    args:
      - "--verbose"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to load project config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil project config")
	}

	if cfg.DefaultAgentType != "aider" {
		t.Errorf("expected default_agent_type 'aider', got %q", cfg.DefaultAgentType)
	}
	if cfg.DefaultWorkDir != "/tmp/myproject" {
		t.Errorf("expected default_workdir '/tmp/myproject', got %q", cfg.DefaultWorkDir)
	}
	if cfg.Env["PROJECT_NAME"] != "myproject" {
		t.Errorf("expected env PROJECT_NAME=myproject, got %v", cfg.Env)
	}
	if len(cfg.Agents) != 1 {
		t.Errorf("expected 1 agent override, got %d", len(cfg.Agents))
	}
}

func TestLoadProjectConfigNonExistent(t *testing.T) {
	cfg, err := LoadProjectConfig("/nonexistent/dir")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config for nonexistent dir")
	}
}

func TestSaveAndLoadProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &ProjectConfig{
		DefaultAgentType: "gemini",
		Env:              map[string]string{"FOO": "bar"},
	}

	if err := SaveProjectConfig(tmpDir, cfg); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if loaded.DefaultAgentType != "gemini" {
		t.Errorf("expected 'gemini', got %q", loaded.DefaultAgentType)
	}
	if loaded.Env["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got %v", loaded.Env)
	}
}

func TestMergeProjectConfig(t *testing.T) {
	global := DefaultConfig()
	project := &ProjectConfig{
		DefaultAgentType: "aider",
	}

	merged := MergeProjectConfig(global, project)
	if merged.DefaultAgentType != "aider" {
		t.Errorf("expected merged default_agent_type 'aider', got %q", merged.DefaultAgentType)
	}
	// Global should be unchanged
	if global.DefaultAgentType == "aider" {
		t.Error("global config should not be modified")
	}
}

func TestMergeProjectConfigNil(t *testing.T) {
	global := DefaultConfig()
	merged := MergeProjectConfig(global, nil)
	if merged != global {
		t.Error("nil project config should return global as-is")
	}
}

// --- Helpers ---

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

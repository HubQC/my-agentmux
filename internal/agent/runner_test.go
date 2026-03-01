package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/session"
)

// --- Unit tests (no tmux needed) ---

func TestPresetCommand(t *testing.T) {
	preset := Preset{
		Binary:      "claude",
		DefaultArgs: []string{"--model", "sonnet"},
	}

	got := preset.Command(nil)
	if got != "claude --model sonnet" {
		t.Errorf("expected 'claude --model sonnet', got %q", got)
	}

	got = preset.Command([]string{"--verbose"})
	if got != "claude --model sonnet --verbose" {
		t.Errorf("expected extra args appended, got %q", got)
	}
}

func TestPresetIsInstalled(t *testing.T) {
	// "echo" should always be found on PATH
	p := Preset{Binary: "echo"}
	if !p.IsInstalled() {
		t.Error("expected 'echo' to be found on PATH")
	}

	// Non-existent binary
	p2 := Preset{Binary: "nonexistent_binary_xyz_12345"}
	if p2.IsInstalled() {
		t.Error("expected nonexistent binary to NOT be found")
	}
}

func TestValidateAgentType(t *testing.T) {
	// Valid presets
	for _, name := range []string{"claude", "aider", "codex", "gemini", "copilot", "shell", "custom"} {
		if err := ValidateAgentType(name); err != nil {
			t.Errorf("expected %q to be valid, got: %v", name, err)
		}
	}

	// Invalid
	if err := ValidateAgentType("nonexistent"); err == nil {
		t.Error("expected error for invalid agent type")
	}
}

func TestFormatAgentCommand(t *testing.T) {
	// Shell type
	if got := FormatAgentCommand("shell", nil); got != "(default shell)" {
		t.Errorf("expected '(default shell)', got %q", got)
	}

	// Unknown type
	got := FormatAgentCommand("unknown_type", nil)
	if !strings.Contains(got, "unknown") {
		t.Errorf("expected unknown format, got %q", got)
	}

	// Known preset (claude)
	got = FormatAgentCommand("claude", []string{"--verbose"})
	if !strings.Contains(got, "claude") {
		t.Errorf("expected 'claude' in output, got %q", got)
	}
	if !strings.Contains(got, "--verbose") {
		t.Errorf("expected '--verbose' in output, got %q", got)
	}

	// Long command truncation
	longArgs := make([]string, 20)
	for i := range longArgs {
		longArgs[i] = "--some-very-long-argument-name"
	}
	got = FormatAgentCommand("claude", longArgs)
	if len(got) > 63 { // 60 + "..."
		t.Errorf("expected truncated output (max ~63 chars), got %d chars: %q", len(got), got)
	}
}

func TestAvailablePresets(t *testing.T) {
	presets := AvailablePresets()
	for _, expected := range []string{"claude", "aider", "codex", "gemini"} {
		if !strings.Contains(presets, expected) {
			t.Errorf("expected %q in available presets: %q", expected, presets)
		}
	}
}

func TestGetPreset(t *testing.T) {
	p, err := GetPreset("claude")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.Binary != "claude" {
		t.Errorf("expected binary 'claude', got %q", p.Binary)
	}

	_, err = GetPreset("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent preset")
	}
}

func TestListPresets(t *testing.T) {
	presets := ListPresets()
	if len(presets) < 5 {
		t.Errorf("expected at least 5 presets, got %d", len(presets))
	}
}

// --- Integration tests (require tmux) ---

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	tmpDir := t.TempDir()
	return &config.Config{
		DataDir:          tmpDir,
		DefaultAgentType: "shell",
		TmuxBinary:       "tmux",
		SessionPrefix:    "amux-test",
		LogLevel:         "debug",
	}
}

func skipIfNoTmux(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed")
	}
}

func TestRunnerLaunchShell(t *testing.T) {
	skipIfNoTmux(t)

	cfg := testConfig(t)
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := session.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	runner := NewRunner(cfg, mgr)
	ctx := context.Background()

	sess, err := runner.Launch(ctx, LaunchOptions{
		Name:      "test-shell",
		AgentType: "shell",
		WorkDir:   "/tmp",
	})
	if err != nil {
		t.Fatalf("failed to launch shell: %v", err)
	}
	t.Cleanup(func() {
		_ = mgr.Destroy(ctx, "test-shell")
	})

	if sess.Name != "test-shell" {
		t.Errorf("expected name=test-shell, got %s", sess.Name)
	}
	if sess.AgentType != "shell" {
		t.Errorf("expected type=shell, got %s", sess.AgentType)
	}
	if sess.Status != "running" {
		t.Errorf("expected status=running, got %s", sess.Status)
	}
	if sess.WorkDir != "/tmp" {
		t.Errorf("expected workdir=/tmp, got %s", sess.WorkDir)
	}

	// Verify session shows up in list
	sessions := mgr.List(ctx)
	found := false
	for _, s := range sessions {
		if s.Name == "test-shell" {
			found = true
			break
		}
	}
	if !found {
		t.Error("launched session not found in list")
	}
}

func TestRunnerLaunchCustomCommand(t *testing.T) {
	skipIfNoTmux(t)

	cfg := testConfig(t)
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := session.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	runner := NewRunner(cfg, mgr)
	ctx := context.Background()

	sess, err := runner.LaunchCustom(ctx, "test-custom", "echo hello", "/tmp")
	if err != nil {
		t.Fatalf("failed to launch custom command: %v", err)
	}
	t.Cleanup(func() {
		_ = mgr.Destroy(ctx, "test-custom")
	})

	if sess.Name != "test-custom" {
		t.Errorf("expected name=test-custom, got %s", sess.Name)
	}
	if sess.AgentType != "custom" {
		t.Errorf("expected type=custom, got %s", sess.AgentType)
	}
}

func TestRunnerLaunchDuplicate(t *testing.T) {
	skipIfNoTmux(t)

	cfg := testConfig(t)
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := session.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	runner := NewRunner(cfg, mgr)
	ctx := context.Background()

	_, err = runner.Launch(ctx, LaunchOptions{
		Name:      "test-dup",
		AgentType: "shell",
	})
	if err != nil {
		t.Fatalf("first launch failed: %v", err)
	}
	t.Cleanup(func() {
		_ = mgr.Destroy(ctx, "test-dup")
	})

	// Second launch with same name should fail
	_, err = runner.Launch(ctx, LaunchOptions{
		Name:      "test-dup",
		AgentType: "shell",
	})
	if err == nil {
		t.Error("expected error when launching duplicate agent")
	}
}

func TestRunnerLaunchEmptyName(t *testing.T) {
	skipIfNoTmux(t)

	cfg := testConfig(t)
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := session.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	runner := NewRunner(cfg, mgr)

	_, err = runner.Launch(context.Background(), LaunchOptions{
		AgentType: "shell",
	})
	if err == nil {
		t.Error("expected error when launching with empty name")
	}
}

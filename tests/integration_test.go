package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/devcontainer"
	"github.com/cqi/my_agentmux/internal/monitor"
	"github.com/cqi/my_agentmux/internal/platform"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/workflow"
)

// TestEndToEndAgentLifecycle tests the full agent lifecycle:
// create config → start agent → send keys → capture output → stop agent.
func TestEndToEndAgentLifecycle(t *testing.T) {
	skipIfNoTmux(t)
	tmpDir := t.TempDir()
	cfg := testConfig(t, tmpDir)

	// Create session manager
	mgr, err := session.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	runner := agent.NewRunner(cfg, mgr)

	// Start a shell agent
	sess, err := runner.Launch(ctx, agent.LaunchOptions{
		Name:      "e2e-test",
		AgentType: "shell",
		WorkDir:   tmpDir,
	})
	if err != nil {
		t.Fatalf("failed to start agent: %v", err)
	}
	defer mgr.Destroy(ctx, "e2e-test")

	if sess.Status != "running" {
		t.Errorf("expected status 'running', got %q", sess.Status)
	}

	// List — should see our agent
	sessions := mgr.List(ctx)
	found := false
	for _, s := range sessions {
		if s.Name == "e2e-test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("agent 'e2e-test' not found in list")
	}

	// Send keys
	err = mgr.SendKeys(ctx, "e2e-test", "echo E2E_INTEGRATION_TEST", true)
	if err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}

	// Wait for output
	time.Sleep(500 * time.Millisecond)

	// Capture output
	output, err := mgr.CaptureOutput(ctx, "e2e-test")
	if err != nil {
		t.Fatalf("failed to capture output: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty captured output")
	}

	// Stop
	err = mgr.Destroy(ctx, "e2e-test")
	if err != nil {
		t.Fatalf("failed to destroy agent: %v", err)
	}

	// Verify it's gone
	_, err = mgr.Get(ctx, "e2e-test")
	if err == nil {
		t.Error("expected error after destroy")
	}
}

// TestEndToEndMonitorFlow tests the monitor subsystem.
func TestEndToEndMonitorFlow(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := testConfig(t, tmpDir)

	// Logger
	logger, err := monitor.NewLogger(cfg.LogsDir(), cfg.Monitor.MaxLogSizeMB)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Write and read
	err = logger.Write("test-agent", "line 1\nline 2\n")
	if err != nil {
		t.Fatalf("failed to write log: %v", err)
	}

	content, err := logger.ReadAll("test-agent")
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty log content")
	}
}

// TestEndToEndWorkflowFlow tests plan lifecycle.
func TestEndToEndWorkflowFlow(t *testing.T) {
	tmpDir := t.TempDir()
	plansDir := filepath.Join(tmpDir, "plans")

	store, err := workflow.NewPlanStore(plansDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create → approve
	plan, err := store.Create("E2E Test Plan", "Integration test plan", "")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	if err := store.Approve(plan.ID); err != nil {
		t.Fatalf("failed to approve: %v", err)
	}

	// Create → reject
	plan2, _ := store.Create("Rejected Plan", "", "")
	if err := store.Reject(plan2.ID, "E2E test rejection"); err != nil {
		t.Fatalf("failed to reject: %v", err)
	}

	// List — should have 2
	plans, err := store.List()
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(plans))
	}

	// Delete
	if err := store.Delete(plan.ID); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
}

// TestEndToEndDevcontainer tests devcontainer generation.
func TestEndToEndDevcontainer(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := devcontainer.ConfigWithAgents([]string{"claude", "aider"})
	filePath, err := devcontainer.Generate(tmpDir, cfg)
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("devcontainer.json should exist")
	}
}

// TestEndToEndPlatformDetect tests platform detection.
func TestEndToEndPlatformDetect(t *testing.T) {
	info := platform.Detect()

	if info.OS == "" {
		t.Error("expected non-empty OS")
	}
	if info.Arch == "" {
		t.Error("expected non-empty Arch")
	}
}

// TestEndToEndConfigAndAgentDefs tests config + agent definitions.
func TestEndToEndConfigAndAgentDefs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config
	cfg := config.DefaultConfig()
	cfg.DataDir = tmpDir
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := cfg.Save(configFile); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Create agent definition
	agentsDir := cfg.AgentsDir()
	os.MkdirAll(agentsDir, 0o755)

	defContent := "---\ndescription: E2E test agent\nagent_type: shell\n---\nYou are a test agent.\n"
	defFile := filepath.Join(agentsDir, "e2e-agent.md")
	os.WriteFile(defFile, []byte(defContent), 0o644)

	// Load definitions
	defs, err := config.LoadAgentDefinitions(agentsDir)
	if err != nil {
		t.Fatalf("failed to load definitions: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("expected 1 definition, got %d", len(defs))
	}
	if defs[0].Name != "e2e-agent" {
		t.Errorf("expected name 'e2e-agent', got %q", defs[0].Name)
	}
}

// TestEndToEndPipeline tests running an orchestrated sequence of agents.
func TestEndToEndPipeline(t *testing.T) {
	skipIfNoTmux(t)
	tmpDir := t.TempDir()

	// Global config
	cfg := config.DefaultConfig()
	cfg.DataDir = tmpDir
	cfg.SessionPrefix = "amux-pipe"
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	// Project config
	projCfg := &config.ProjectConfig{
		DefaultAgentType: "shell", // Use fast shell agents for testing
		Pipelines: map[string][]string{
			"test-pipeline": {"agent-1", "agent-2"},
		},
	}
	if err := config.SaveProjectConfig(tmpDir, projCfg); err != nil {
		t.Fatalf("failed to save project config: %v", err)
	}

	activeCfg := config.MergeProjectConfig(cfg, projCfg)
	mgr, err := session.NewManager(activeCfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	runner := agent.NewRunner(activeCfg, mgr)
	ctx := context.Background()

	pipeline := projCfg.Pipelines["test-pipeline"]

	// Emulate pipeline command behaviour
	for _, agentName := range pipeline {
		opts := agent.LaunchOptions{
			Name:    agentName,
			WorkDir: tmpDir,
			// Since we want this to exit immediately so the pipeline moves on,
			// we override the command to just be `true` or `echo done`.
			Command: "echo " + agentName + " done",
		}

		sess, err := runner.Launch(ctx, opts)
		if err != nil {
			t.Fatalf("failed to launch agent %q: %v", agentName, err)
		}

		// Verify session was created
		if sess.Status != "running" {
			t.Errorf("expected status 'running' for %s", agentName)
		}

		// Wait briefly
		time.Sleep(100 * time.Millisecond)

		// Terminate to allow pipeline to continue
		_ = mgr.Destroy(ctx, agentName)
	}
}

// --- Helpers ---

func testConfig(t *testing.T, tmpDir string) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.DataDir = tmpDir
	cfg.SessionPrefix = "amux-e2e"

	// Ensure directories exist
	os.MkdirAll(cfg.SessionsDir(), 0o755)
	os.MkdirAll(cfg.LogsDir(), 0o755)
	os.MkdirAll(cfg.AgentsDir(), 0o755)
	os.MkdirAll(cfg.PlansDir(), 0o755)

	return cfg
}

func skipIfNoTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("skipping: tmux not found in PATH")
	}
}

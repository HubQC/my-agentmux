package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cqi/my_agentmux/internal/config"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	tmpDir := t.TempDir()
	return &config.Config{
		DataDir:          tmpDir,
		DefaultAgentType: "test",
		TmuxBinary:       "tmux",
		SessionPrefix:    "amux-test",
		LogLevel:         "debug",
	}
}

func TestStateLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "sessions.json")

	state, err := NewState(stateFile)
	if err != nil {
		t.Fatalf("failed to create state: %v", err)
	}

	// Put a session
	sess := &AgentSession{
		Name:      "test-agent",
		TmuxName:  "amux-test-agent",
		AgentType: "claude",
		WorkDir:   "/tmp",
		Status:    "running",
	}
	if err := state.Put(sess); err != nil {
		t.Fatalf("failed to put session: %v", err)
	}

	// Get it back
	got := state.Get("test-agent")
	if got == nil {
		t.Fatal("expected to get session back")
	}
	if got.AgentType != "claude" {
		t.Errorf("expected type=claude, got %s", got.AgentType)
	}

	// List sessions
	all := state.List()
	if len(all) != 1 {
		t.Errorf("expected 1 session, got %d", len(all))
	}

	// Persistence: reload from disk
	state2, err := NewState(stateFile)
	if err != nil {
		t.Fatalf("failed to reload state: %v", err)
	}
	got2 := state2.Get("test-agent")
	if got2 == nil {
		t.Fatal("expected session to persist across reload")
	}

	// Remove
	if err := state.Remove("test-agent"); err != nil {
		t.Fatalf("failed to remove: %v", err)
	}
	if state.Get("test-agent") != nil {
		t.Error("expected nil after remove")
	}

	// Clear
	_ = state.Put(sess)
	if err := state.Clear(); err != nil {
		t.Fatalf("failed to clear: %v", err)
	}
	if len(state.List()) != 0 {
		t.Error("expected 0 sessions after clear")
	}
}

func TestManagerCreateAndDestroy(t *testing.T) {
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed")
	}

	cfg := testConfig(t)
	// Ensure sessions directory exists
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Create a session
	sess, err := mgr.Create(ctx, CreateOptions{
		Name: "test-agent",
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	t.Cleanup(func() {
		_ = mgr.Destroy(ctx, "test-agent")
	})

	if sess.Name != "test-agent" {
		t.Errorf("expected name=test-agent, got %s", sess.Name)
	}
	if sess.Status != "running" {
		t.Errorf("expected status=running, got %s", sess.Status)
	}

	// List — should show 1 running
	sessions := mgr.List(ctx)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Status != "running" {
		t.Errorf("expected running status, got %s", sessions[0].Status)
	}

	// Get
	got, err := mgr.Get(ctx, "test-agent")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if got.TmuxName != sess.TmuxName {
		t.Errorf("expected tmux name %s, got %s", sess.TmuxName, got.TmuxName)
	}

	// Duplicate create should fail
	_, err = mgr.Create(ctx, CreateOptions{Name: "test-agent"})
	if err == nil {
		t.Error("expected error when creating duplicate session")
	}

	// Destroy
	if err := mgr.Destroy(ctx, "test-agent"); err != nil {
		t.Fatalf("failed to destroy: %v", err)
	}

	// Should be gone
	sessions = mgr.List(ctx)
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after destroy, got %d", len(sessions))
	}
}

func TestManagerDestroyAll(t *testing.T) {
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed")
	}

	cfg := testConfig(t)
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Create 3 sessions
	for _, name := range []string{"agent-a", "agent-b", "agent-c"} {
		_, err := mgr.Create(ctx, CreateOptions{Name: name})
		if err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}
	t.Cleanup(func() {
		mgr.DestroyAll(ctx)
	})

	// Destroy all
	count, err := mgr.DestroyAll(ctx)
	if err != nil {
		t.Fatalf("failed to destroy all: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 destroyed, got %d", count)
	}

	sessions := mgr.List(ctx)
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestManagerGetNotFound(t *testing.T) {
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed")
	}

	cfg := testConfig(t)
	os.MkdirAll(cfg.SessionsDir(), 0o755)

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = mgr.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

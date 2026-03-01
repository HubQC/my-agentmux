package tmux

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// skipIfNoTmux skips the test if tmux is not installed.
func skipIfNoTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed, skipping integration test")
	}
}

// testSessionName returns a unique session name for testing.
func testSessionName(t *testing.T) string {
	t.Helper()
	// Use a name unlikely to collide with real sessions
	return fmt.Sprintf("amux-test-%d", time.Now().UnixNano()%100000)
}

func TestNewClient(t *testing.T) {
	skipIfNoTmux(t)

	client, err := NewClient("")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	if client.Binary == "" {
		t.Error("expected Binary to be set")
	}
}

func TestNewClientBadBinary(t *testing.T) {
	_, err := NewClient("/nonexistent/tmux")
	if err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestVersion(t *testing.T) {
	skipIfNoTmux(t)

	client, _ := NewClient("")
	ctx := context.Background()

	version, err := client.Version(ctx)
	if err != nil {
		t.Fatalf("failed to get version: %v", err)
	}
	if !strings.HasPrefix(version, "tmux") {
		t.Errorf("unexpected version format: %q", version)
	}
	t.Logf("tmux version: %s", version)
}

func TestSessionLifecycle(t *testing.T) {
	skipIfNoTmux(t)

	client, _ := NewClient("")
	ctx := context.Background()
	name := testSessionName(t)

	// Cleanup in case of test failure
	t.Cleanup(func() {
		_ = client.KillSession(ctx, name)
	})

	// Create session
	session, err := client.NewSession(ctx, SessionOptions{
		Name:       name,
		WindowName: "main",
		Detached:   true,
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if session.Name != name {
		t.Errorf("expected session name %q, got %q", name, session.Name)
	}
	t.Logf("created session: %+v", session)

	// Has session
	if !client.HasSession(ctx, name) {
		t.Error("expected HasSession to return true")
	}

	// Get session
	got, err := client.GetSession(ctx, name)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if got.Name != name {
		t.Errorf("expected name %q, got %q", name, got.Name)
	}

	// List sessions (should contain ours)
	sessions, err := client.ListSessions(ctx)
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("session %q not found in list of %d sessions", name, len(sessions))
	}

	// Kill session
	if err := client.KillSession(ctx, name); err != nil {
		t.Fatalf("failed to kill session: %v", err)
	}

	// Verify gone
	if client.HasSession(ctx, name) {
		t.Error("session should not exist after kill")
	}
}

func TestWindowAndPaneOperations(t *testing.T) {
	skipIfNoTmux(t)

	client, _ := NewClient("")
	ctx := context.Background()
	name := testSessionName(t)

	t.Cleanup(func() {
		_ = client.KillSession(ctx, name)
	})

	// Create session
	_, err := client.NewSession(ctx, SessionOptions{
		Name:       name,
		WindowName: "win1",
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// List windows
	windows, err := client.ListWindows(ctx, name)
	if err != nil {
		t.Fatalf("failed to list windows: %v", err)
	}
	if len(windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(windows))
	}
	if windows[0].Name != "win1" {
		t.Errorf("expected window name %q, got %q", "win1", windows[0].Name)
	}

	// List panes (should have 1)
	panes, err := client.ListPanes(ctx, name)
	if err != nil {
		t.Fatalf("failed to list panes: %v", err)
	}
	if len(panes) != 1 {
		t.Fatalf("expected 1 pane, got %d", len(panes))
	}

	// Split window (creates pane 2)
	newPane, err := client.SplitWindow(ctx, SplitOptions{
		Target:     name,
		Horizontal: true,
		Percent:    50,
	})
	if err != nil {
		t.Fatalf("failed to split window: %v", err)
	}
	if newPane == nil {
		t.Fatal("expected non-nil pane from split")
	}
	t.Logf("new pane from split: %+v", newPane)

	// Now should have 2 panes
	panes, err = client.ListPanes(ctx, name)
	if err != nil {
		t.Fatalf("failed to list panes after split: %v", err)
	}
	if len(panes) != 2 {
		t.Errorf("expected 2 panes after split, got %d", len(panes))
	}
}

func TestSendKeysAndCapture(t *testing.T) {
	skipIfNoTmux(t)

	client, _ := NewClient("")
	ctx := context.Background()
	name := testSessionName(t)

	t.Cleanup(func() {
		_ = client.KillSession(ctx, name)
	})

	// Create session
	_, err := client.NewSession(ctx, SessionOptions{
		Name: name,
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Send a command
	err = client.SendKeys(ctx, name, "echo AGENTMUX_TEST_OUTPUT", true)
	if err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}

	// Wait a moment for the command to execute
	time.Sleep(500 * time.Millisecond)

	// Capture pane output
	output, err := client.CapturePane(ctx, name, 0, 0)
	if err != nil {
		t.Fatalf("failed to capture pane: %v", err)
	}

	if !strings.Contains(output, "AGENTMUX_TEST_OUTPUT") {
		t.Errorf("expected captured output to contain 'AGENTMUX_TEST_OUTPUT', got:\n%s", output)
	}
	t.Logf("captured output:\n%s", output)
}

func TestParseSession(t *testing.T) {
	line := "$1|mysession|2|200|50|1709123456|0"
	s, err := parseSession(line)
	if err != nil {
		t.Fatalf("failed to parse session: %v", err)
	}
	if s.ID != "$1" {
		t.Errorf("expected ID=$1, got %s", s.ID)
	}
	if s.Name != "mysession" {
		t.Errorf("expected Name=mysession, got %s", s.Name)
	}
	if s.Windows != 2 {
		t.Errorf("expected Windows=2, got %d", s.Windows)
	}
	if s.Attached {
		t.Error("expected Attached=false")
	}
}

func TestParseWindow(t *testing.T) {
	line := "@0|$1|0|main|200|50|1|1"
	w, err := parseWindow(line)
	if err != nil {
		t.Fatalf("failed to parse window: %v", err)
	}
	if w.ID != "@0" {
		t.Errorf("expected ID=@0, got %s", w.ID)
	}
	if w.Name != "main" {
		t.Errorf("expected Name=main, got %s", w.Name)
	}
	if !w.Active {
		t.Error("expected Active=true")
	}
}

func TestParsePane(t *testing.T) {
	line := "%0|$1|@0|0|100|25|1|12345|bash"
	p, err := parsePane(line)
	if err != nil {
		t.Fatalf("failed to parse pane: %v", err)
	}
	if p.ID != "%0" {
		t.Errorf("expected ID=%%0, got %s", p.ID)
	}
	if p.PID != 12345 {
		t.Errorf("expected PID=12345, got %d", p.PID)
	}
	if p.Command != "bash" {
		t.Errorf("expected Command=bash, got %s", p.Command)
	}
}

func TestParseInvalidLines(t *testing.T) {
	if _, err := parseSession("too|few"); err == nil {
		t.Error("expected error for invalid session line")
	}
	if _, err := parseWindow("too|few"); err == nil {
		t.Error("expected error for invalid window line")
	}
	if _, err := parsePane("too|few"); err == nil {
		t.Error("expected error for invalid pane line")
	}
}

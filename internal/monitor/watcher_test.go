package monitor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cqi/my_agentmux/internal/tmux"
)

// --- Logger unit tests ---

func TestLoggerWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write some content
	if err := logger.Write("test-agent", "line1\n"); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if err := logger.Write("test-agent", "line2\n"); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read it back
	content, err := logger.ReadAll("test-agent")
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if !strings.Contains(content, "line1") || !strings.Contains(content, "line2") {
		t.Errorf("expected both lines in content, got: %q", content)
	}
}

func TestLoggerWriteTimestamped(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	if err := logger.WriteTimestamped("test-agent", "hello"); err != nil {
		t.Fatalf("failed to write timestamped: %v", err)
	}

	content, err := logger.ReadAll("test-agent")
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	// Should contain timestamp format and message
	if !strings.Contains(content, "hello") {
		t.Errorf("expected 'hello' in content, got: %q", content)
	}
	if !strings.Contains(content, "[") || !strings.Contains(content, "]") {
		t.Errorf("expected timestamp brackets in content, got: %q", content)
	}
}

func TestLoggerEmptyWrite(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	// Writing empty content should be no-op
	if err := logger.Write("test-agent", ""); err != nil {
		t.Fatalf("empty write should succeed: %v", err)
	}

	// File shouldn't exist since nothing was written
	content, err := logger.ReadAll("test-agent")
	if err != nil {
		t.Fatalf("read should succeed: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty content, got: %q", content)
	}
}

func TestLoggerReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	content, err := logger.ReadAll("nonexistent")
	if err != nil {
		t.Fatalf("read nonexistent should succeed: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty content, got: %q", content)
	}
}

func TestLoggerLogPath(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	expected := filepath.Join(tmpDir, "myagent.log")
	if got := logger.LogPath("myagent"); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestLoggerRotation(t *testing.T) {
	tmpDir := t.TempDir()
	// Set max size to 1 byte to force rotation
	logger, err := NewLogger(tmpDir, 0) // 0 defaults to 50MB, override below
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	// Override max size directly for testing
	logger.maxSizeMB = 0 // will be < any file size, so first write creates, second rotates
	defer logger.Close()

	// Write enough to exceed 0 MB limit
	if err := logger.Write("test-agent", "first content\n"); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Close and reopen to force file stat
	logger.CloseAgent("test-agent")

	if err := logger.Write("test-agent", "second content\n"); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	// Check that rotation happened — .1 file should exist
	rotatedPath := logger.LogPath("test-agent") + ".1"
	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Error("expected rotated log file to exist")
	}
}

func TestLoggerMultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Write("agent-a", "content-a\n")
	logger.Write("agent-b", "content-b\n")

	contentA, _ := logger.ReadAll("agent-a")
	contentB, _ := logger.ReadAll("agent-b")

	if !strings.Contains(contentA, "content-a") {
		t.Errorf("agent-a: expected 'content-a', got %q", contentA)
	}
	if !strings.Contains(contentB, "content-b") {
		t.Errorf("agent-b: expected 'content-b', got %q", contentB)
	}
	if strings.Contains(contentA, "content-b") {
		t.Error("agent-a should not contain agent-b's content")
	}
}

// --- Watcher unit tests ---

func TestWatcherCreateAndStop(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := NewLogger(tmpDir, 50)
	defer logger.Close()

	// Watcher can be created with nil tmux client for unit tests
	watcher := NewWatcher(nil, logger, 100)
	if watcher == nil {
		t.Fatal("expected non-nil watcher")
	}
	watcher.Stop()

	// Double stop should be safe
	watcher.Stop()
}

// --- Integration tests (require tmux) ---

func skipIfNoTmux(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/usr/bin/tmux"); os.IsNotExist(err) {
		t.Skip("tmux not installed")
	}
}

func TestWatcherIntegration(t *testing.T) {
	skipIfNoTmux(t)

	tmpDir := t.TempDir()
	logger, err := NewLogger(tmpDir, 50)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	tmuxClient, err := tmux.NewClient("")
	if err != nil {
		t.Fatalf("failed to create tmux client: %v", err)
	}

	ctx := context.Background()
	sessionName := "amux-test-watcher"

	// Create a tmux session
	_, err = tmuxClient.NewSession(ctx, tmux.SessionOptions{
		Name:     sessionName,
		Detached: true,
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	t.Cleanup(func() {
		_ = tmuxClient.KillSession(ctx, sessionName)
	})

	// Create watcher with fast polling
	watcher := NewWatcher(tmuxClient, logger, 100)
	defer watcher.Stop()

	// Start watching
	if err := watcher.Watch("test-agent", sessionName); err != nil {
		t.Fatalf("failed to start watching: %v", err)
	}

	if !watcher.IsWatching("test-agent") {
		t.Error("expected IsWatching to return true")
	}

	// Duplicate watch should fail
	if err := watcher.Watch("test-agent", sessionName); err == nil {
		t.Error("expected error for duplicate watch")
	}

	// Send a command to the session
	err = tmuxClient.SendKeys(ctx, sessionName, "echo WATCHER_TEST_OUTPUT", true)
	if err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}

	// Wait for the watcher to capture the output
	var capturedEvent *Event
	timeout := time.After(5 * time.Second)
	for capturedEvent == nil {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for watcher event")
		case event, ok := <-watcher.Events():
			if !ok {
				t.Fatal("events channel closed unexpectedly")
			}
			if strings.Contains(event.Content, "WATCHER_TEST_OUTPUT") {
				capturedEvent = &event
			}
		}
	}

	if capturedEvent.AgentName != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %q", capturedEvent.AgentName)
	}

	// Check that content was logged
	logContent, err := logger.ReadAll("test-agent")
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	if !strings.Contains(logContent, "WATCHER_TEST_OUTPUT") {
		t.Errorf("expected log to contain 'WATCHER_TEST_OUTPUT', got: %q", logContent)
	}

	// Unwatch
	watcher.Unwatch("test-agent")
	if watcher.IsWatching("test-agent") {
		t.Error("expected IsWatching to return false after unwatch")
	}
}

package tui

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/cqi/my_agentmux/internal/tmux"
)

func TestAppModel(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("skipping: tmux not found in PATH")
	}

	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.DataDir = tmpDir
	os.MkdirAll(cfg.SessionsDir(), 0o755)
	os.MkdirAll(cfg.LogsDir(), 0o755)

	sessionMgr, err := session.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tmuxClient, err := tmux.NewClient("tmux")
	if err != nil {
		t.Fatalf("failed to create tmux client: %v", err)
	}

	model := NewModel(cfg, sessionMgr, tmuxClient, false, "", nil)
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for the dashboard to render (status bar shows "agents")
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("agents"))
	}, teatest.WithDuration(3*time.Second))

	// Send navigation keys to test basic input handling
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

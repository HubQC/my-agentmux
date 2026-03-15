//go:build desktop

package desktop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cqi/my_agentmux/internal/session"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TerminalService handles terminal emulation events for the Wails frontend.
type TerminalService struct {
	ctx      context.Context
	mgr      *session.Manager
	mu       sync.RWMutex
	streams  map[string]context.CancelFunc
}

// NewTerminalService creates a new TerminalService.
func NewTerminalService(mgr *session.Manager) *TerminalService {
	return &TerminalService{
		mgr:     mgr,
		streams: make(map[string]context.CancelFunc),
	}
}

// startup is called by Wails.
func (t *TerminalService) startup(ctx context.Context) {
	t.ctx = ctx
}

// AttachTerminal starts streaming output for a session.
func (t *TerminalService) AttachTerminal(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If already streaming, do nothing
	if _, ok := t.streams[name]; ok {
		return nil
	}

	// Create a cancellable context for this stream
	streamCtx, cancel := context.WithCancel(t.ctx)
	t.streams[name] = cancel

	// Start streaming goroutine
	go t.streamOutput(streamCtx, name)

	return nil
}

// DetachTerminal stops streaming output for a session.
func (t *TerminalService) DetachTerminal(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if cancel, ok := t.streams[name]; ok {
		cancel()
		delete(t.streams, name)
	}
	return nil
}

// SendInput sends raw input (keystrokes) to the session's tmux pane.
func (t *TerminalService) SendInput(name, input string) error {
	// For raw terminal input, we use SendKeys without appending Enter
	return t.mgr.SendKeys(t.ctx, name, input, false)
}

// streamOutput polls tmux capture-pane and emits events to the frontend.
func (t *TerminalService) streamOutput(ctx context.Context, name string) {
	var lastOutput string
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	eventName := fmt.Sprintf("terminal:output:%s", name)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			output, err := t.mgr.CaptureOutput(ctx, name)
			if err != nil {
				// If session is gone, stop streaming
				t.DetachTerminal(name)
				return
			}

			// Simple diff: only emit if content changed
			// In a more advanced version, we'd use tmux's escape sequences
			// and compute a minimal delta or use a PTY proxy.
			if output != lastOutput {
				runtime.EventsEmit(t.ctx, eventName, output)
				lastOutput = output
			}
		}
	}
}

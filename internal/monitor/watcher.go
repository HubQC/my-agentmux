package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cqi/my_agentmux/internal/tmux"
	"github.com/shirou/gopsutil/v3/process"
)

// Event represents a captured output event from an agent.
type Event struct {
	AgentName string
	Content   string
	Timestamp time.Time
}

// ResourceEvent represents CPU and Memory usage of an agent's pane process.
type ResourceEvent struct {
	AgentName string
	CPU       float64
	Memory    uint64 // in bytes
	Timestamp time.Time
}

// Watcher polls tmux panes for output and streams events.
type Watcher struct {
	mu           sync.Mutex
	tmux         *tmux.Client
	logger       *Logger
	pollInterval time.Duration
	agents       map[string]*watchedAgent
	eventCh      chan Event
	resourceCh   chan ResourceEvent
	stopCh       chan struct{}
	stopped      bool
}

type watchedAgent struct {
	tmuxSession string
	lastOutput  string
	pid         int
	cancel      context.CancelFunc
}

// NewWatcher creates a new output watcher.
func NewWatcher(tmuxClient *tmux.Client, logger *Logger, pollIntervalMs int) *Watcher {
	if pollIntervalMs <= 0 {
		pollIntervalMs = 500
	}

	return &Watcher{
		tmux:         tmuxClient,
		logger:       logger,
		pollInterval: time.Duration(pollIntervalMs) * time.Millisecond,
		agents:       make(map[string]*watchedAgent),
		eventCh:      make(chan Event, 100),
		resourceCh:   make(chan ResourceEvent, 100),
		stopCh:       make(chan struct{}),
	}
}

// Events returns the channel of output events.
func (w *Watcher) Events() <-chan Event {
	return w.eventCh
}

// ResourceEvents returns the channel of resource events.
func (w *Watcher) ResourceEvents() <-chan ResourceEvent {
	return w.resourceCh
}

// Watch starts polling a tmux session for output changes.
func (w *Watcher) Watch(agentName string, tmuxSession string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.agents[agentName]; exists {
		return fmt.Errorf("already watching agent %q", agentName)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wa := &watchedAgent{
		tmuxSession: tmuxSession,
		cancel:      cancel,
	}
	w.agents[agentName] = wa

	go w.pollLoop(ctx, agentName, wa)
	return nil
}

// Unwatch stops polling a specific agent.
func (w *Watcher) Unwatch(agentName string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if wa, exists := w.agents[agentName]; exists {
		wa.cancel()
		delete(w.agents, agentName)
	}
}

// Stop stops all watchers and closes the event channel.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped {
		return
	}
	w.stopped = true

	for name, wa := range w.agents {
		wa.cancel()
		delete(w.agents, name)
	}

	close(w.stopCh)
	close(w.eventCh)
	close(w.resourceCh)
}

// IsWatching returns true if the agent is being watched.
func (w *Watcher) IsWatching(agentName string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, exists := w.agents[agentName]
	return exists
}

// pollLoop continuously captures pane output and emits events on changes.
func (w *Watcher) pollLoop(ctx context.Context, agentName string, wa *watchedAgent) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.captureAndEmit(ctx, agentName, wa)
			w.captureResources(ctx, agentName, wa)
		}
	}
}

// captureAndEmit captures the current pane output and emits an event if changed.
func (w *Watcher) captureAndEmit(ctx context.Context, agentName string, wa *watchedAgent) {
	output, err := w.tmux.CapturePane(ctx, wa.tmuxSession, 0, 0)
	if err != nil {
		// Session may have ended — silently stop watching
		return
	}

	// Trim trailing whitespace for comparison
	trimmed := strings.TrimRight(output, " \t\n\r")

	if trimmed == wa.lastOutput {
		return // no change
	}

	// Find new content (diff from last capture)
	newContent := trimmed
	if wa.lastOutput != "" {
		// Simple diff: if new output starts with old output, extract the new part
		if strings.HasPrefix(trimmed, wa.lastOutput) {
			newContent = strings.TrimSpace(trimmed[len(wa.lastOutput):])
		}
	}

	wa.lastOutput = trimmed

	if newContent == "" {
		return
	}

	// Log the new content
	_ = w.logger.Write(agentName, newContent+"\n")

	// Emit event (non-blocking)
	event := Event{
		AgentName: agentName,
		Content:   newContent,
		Timestamp: time.Now(),
	}

	select {
	case w.eventCh <- event:
	default:
		// Channel full — drop event to avoid blocking
	}
}

// captureResources captures CPU and memory usage and emits a ResourceEvent.
func (w *Watcher) captureResources(ctx context.Context, agentName string, wa *watchedAgent) {
	// Discover PID if we haven't yet
	if wa.pid <= 0 {
		pid, err := w.tmux.GetPanePID(ctx, wa.tmuxSession)
		if err != nil || pid <= 0 {
			return
		}
		wa.pid = pid
	}

	p, err := process.NewProcess(int32(wa.pid))
	if err != nil {
		// Process died or invalid
		return
	}

	// For true agent usage, we should sum children, but to keep it simple and performant,
	// checking the main pane process (often a shell) and its direct children is good enough.
	// For deeper accuracy, we could recursively sum `p.Children()`.
	
	// Default to main process metrics
	cpu, _ := p.CPUPercent()
	memInfo, _ := p.MemoryInfo()
	var mem uint64
	if memInfo != nil {
		mem = memInfo.RSS
	}

	// Try to sum direct children to capture CLI tools (like python/node) running under the shell
	children, err := p.Children()
	if err == nil {
		for _, child := range children {
			childCPU, _ := child.CPUPercent()
			cpu += childCPU
			
			childMem, _ := child.MemoryInfo()
			if childMem != nil {
				mem += childMem.RSS
			}
		}
	}

	event := ResourceEvent{
		AgentName: agentName,
		CPU:       cpu,
		Memory:    mem,
		Timestamp: time.Now(),
	}

	select {
	case w.resourceCh <- event:
	default:
	}
}

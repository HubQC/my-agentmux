package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/tmux"
)

// Manager handles agent session lifecycle.
type Manager struct {
	cfg   *config.Config
	tmux  *tmux.Client
	state *State
}

// NewManager creates a new session manager.
func NewManager(cfg *config.Config) (*Manager, error) {
	tmuxClient, err := tmux.NewClient(cfg.TmuxBinary)
	if err != nil {
		return nil, fmt.Errorf("initializing tmux: %w", err)
	}

	stateFile := filepath.Join(cfg.SessionsDir(), "sessions.json")
	state, err := NewState(stateFile)
	if err != nil {
		return nil, fmt.Errorf("initializing state: %w", err)
	}

	return &Manager{
		cfg:   cfg,
		tmux:  tmuxClient,
		state: state,
	}, nil
}

// CreateOptions configures a new agent session.
type CreateOptions struct {
	Name      string
	AgentType string
	WorkDir   string
	Command   string
	Args      []string
	Env       map[string]string
	Group     string

	// Codex Integration
	CodexProfile    string
	CodexReasoning  string
	CodexMCPs       []string
	CodexMultiAgent bool

	// Gemini Integration
	GeminiMCPs []string
}

// Create creates a new agent session.
func (m *Manager) Create(ctx context.Context, opts CreateOptions) (*AgentSession, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	// Check if session already exists
	if existing := m.state.Get(opts.Name); existing != nil {
		tmuxName := m.tmuxSessionName(opts.Name)
		if m.tmux.HasSession(ctx, tmuxName) {
			return nil, fmt.Errorf("agent %q: %w", opts.Name, ErrAgentAlreadyRunning)
		}
		// Stale state — clean up
		_ = m.state.Remove(opts.Name)
	}

	// Default values
	if opts.AgentType == "" {
		opts.AgentType = m.cfg.DefaultAgentType
	}
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}

	tmuxName := m.tmuxSessionName(opts.Name)

	// Build the command to run inside the session
	sessionCmd := ""
	if opts.Command != "" {
		sessionCmd = opts.Command
		for _, arg := range opts.Args {
			sessionCmd += " " + shellQuote(arg)
		}
	}

	// Create tmux session
	tmuxOpts := tmux.SessionOptions{
		Name:       tmuxName,
		StartDir:   opts.WorkDir,
		WindowName: opts.Name,
		Detached:   true,
		Command:    sessionCmd,
		Env:        opts.Env,
	}

	_, err := m.tmux.NewSession(ctx, tmuxOpts)
	if err != nil {
		return nil, fmt.Errorf("creating tmux session: %w", err)
	}

	// Track the session
	agentSession := &AgentSession{
		Name:      opts.Name,
		TmuxName:  tmuxName,
		AgentType: opts.AgentType,
		WorkDir:   opts.WorkDir,
		CreatedAt: time.Now(),
		Status:    StatusRunning,
		Group:     opts.Group,

		CodexProfile:    opts.CodexProfile,
		CodexReasoning:  opts.CodexReasoning,
		CodexMCPs:       opts.CodexMCPs,
		CodexMultiAgent: opts.CodexMultiAgent,

		GeminiMCPs: opts.GeminiMCPs,
	}

	if err := m.state.Put(agentSession); err != nil {
		// Try to clean up the tmux session
		_ = m.tmux.KillSession(ctx, tmuxName)
		return nil, fmt.Errorf("saving session state: %w", err)
	}

	return agentSession, nil
}

// List returns all tracked agent sessions, sorted by creation time.
// It also reconciles state with actual tmux sessions.
func (m *Manager) List(ctx context.Context) []*AgentSession {
	sessions := m.state.List()

	// Reconcile: mark sessions as stopped if tmux session is gone
	for _, s := range sessions {
		if !m.tmux.HasSession(ctx, s.TmuxName) && s.Status == StatusRunning {
			s.Status = StatusStopped
			_ = m.state.Put(s)
		}
	}

	// Sort by creation time
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.Before(sessions[j].CreatedAt)
	})

	return sessions
}

// Get returns a specific agent session by name.
func (m *Manager) Get(ctx context.Context, name string) (*AgentSession, error) {
	session := m.state.Get(name)
	if session == nil {
		return nil, fmt.Errorf("agent %q: %w", name, ErrAgentNotFound)
	}

	// Reconcile status
	if !m.tmux.HasSession(ctx, session.TmuxName) && session.Status == StatusRunning {
		session.Status = StatusStopped
		_ = m.state.Put(session)
	}

	return session, nil
}

// Attach attaches the current terminal to an agent's tmux session.
func (m *Manager) Attach(ctx context.Context, name string) error {
	session, err := m.Get(ctx, name)
	if err != nil {
		return err
	}

	if session.Status != StatusRunning {
		return fmt.Errorf("agent %q: %w (status: %s)", name, ErrAgentNotRunning, session.Status)
	}

	// Use syscall.Exec to replace the current process with tmux attach
	tmuxPath, err := exec.LookPath(m.cfg.TmuxBinary)
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	args := []string{"tmux", "attach-session", "-t", session.TmuxName}
	return syscall.Exec(tmuxPath, args, os.Environ())
}

// Destroy stops and removes an agent session.
func (m *Manager) Destroy(ctx context.Context, name string) error {
	session := m.state.Get(name)
	if session == nil {
		return fmt.Errorf("agent %q: %w", name, ErrAgentNotFound)
	}

	// Kill the tmux session if it's still running
	if m.tmux.HasSession(ctx, session.TmuxName) {
		if err := m.tmux.KillSession(ctx, session.TmuxName); err != nil {
			return fmt.Errorf("killing tmux session: %w", err)
		}
	}

	// Remove from state
	if err := m.state.Remove(name); err != nil {
		return fmt.Errorf("removing session state: %w", err)
	}

	return nil
}

// DestroyAll stops and removes all agent sessions.
func (m *Manager) DestroyAll(ctx context.Context) (int, error) {
	sessions := m.state.List()
	count := 0

	for _, s := range sessions {
		if m.tmux.HasSession(ctx, s.TmuxName) {
			_ = m.tmux.KillSession(ctx, s.TmuxName)
		}
		count++
	}

	if err := m.state.Clear(); err != nil {
		return count, fmt.Errorf("clearing state: %w", err)
	}

	return count, nil
}

// SendKeys sends input to an agent's tmux session.
func (m *Manager) SendKeys(ctx context.Context, name string, keys string, pressEnter bool) error {
	session, err := m.Get(ctx, name)
	if err != nil {
		return err
	}

	if session.Status != StatusRunning {
		return fmt.Errorf("agent %q: %w", name, ErrAgentNotRunning)
	}

	return m.tmux.SendKeys(ctx, session.TmuxName, keys, pressEnter)
}

// CaptureOutput captures the visible output of an agent's tmux pane.
func (m *Manager) CaptureOutput(ctx context.Context, name string) (string, error) {
	session, err := m.Get(ctx, name)
	if err != nil {
		return "", err
	}

	return m.tmux.CapturePane(ctx, session.TmuxName, 0, 0)
}

// tmuxSessionName generates the tmux session name for an agent.
func (m *Manager) tmuxSessionName(agentName string) string {
	return fmt.Sprintf("%s-%s", m.cfg.SessionPrefix, agentName)
}

// shellQuote wraps an argument in single quotes for safe shell interpolation.
// Single quotes inside the value are escaped as '\'' (end quote, escaped quote, start quote).
func shellQuote(s string) string {
	// If the string is simple (alphanumeric, dash, underscore, dot, slash), no quoting needed
	safe := true
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '/' || c == ':' || c == '=') {
			safe = false
			break
		}
	}
	if safe && s != "" {
		return s
	}

	// Wrap in single quotes, escaping any embedded single quotes
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return "'" + escaped + "'"
}

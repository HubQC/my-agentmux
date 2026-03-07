package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AgentSession represents a tracked agent session.
type AgentSession struct {
	Name      string    `json:"name"`
	TmuxName  string    `json:"tmux_name"`
	AgentType string    `json:"agent_type"`
	WorkDir   string    `json:"work_dir"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "running", "stopped", "error"
	PID       int       `json:"pid,omitempty"`
	Group     string    `json:"group,omitempty"`

	// Codex Integrations
	CodexProfile    string   `json:"codex_profile,omitempty"`
	CodexReasoning  string   `json:"codex_reasoning,omitempty"`
	CodexMCPs       []string `json:"codex_mcps,omitempty"`
	CodexMultiAgent bool     `json:"codex_multi_agent,omitempty"`

	// Gemini Integrations
	GeminiMCPs []string `json:"gemini_mcps,omitempty"`
}

// State manages persistent session state stored as JSON.
type State struct {
	mu       sync.RWMutex
	filePath string
	Sessions map[string]*AgentSession `json:"sessions"`
}

// NewState creates a new State that persists to the given file path.
func NewState(filePath string) (*State, error) {
	s := &State{
		filePath: filePath,
		Sessions: make(map[string]*AgentSession),
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, fmt.Errorf("creating state directory: %w", err)
	}

	// Load existing state if available
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	return s, nil
}

// Get returns a session by name, or nil if not found.
func (s *State) Get(name string) *AgentSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Sessions[name]
}

// Put adds or updates a session in state and persists.
func (s *State) Put(session *AgentSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Sessions[session.Name] = session
	return s.save()
}

// Remove deletes a session from state and persists.
func (s *State) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Sessions, name)
	return s.save()
}

// List returns all tracked sessions.
func (s *State) List() []*AgentSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AgentSession, 0, len(s.Sessions))
	for _, sess := range s.Sessions {
		result = append(result, sess)
	}
	return result
}

// Clear removes all sessions from state and persists.
func (s *State) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Sessions = make(map[string]*AgentSession)
	return s.save()
}

// load reads state from disk.
func (s *State) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	// Handle empty file
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.Sessions)
}

// save writes state to disk atomically.
func (s *State) save() error {
	data, err := json.MarshalIndent(s.Sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		return fmt.Errorf("writing temp state file: %w", err)
	}

	if err := os.Rename(tmpFile, s.filePath); err != nil {
		return fmt.Errorf("renaming temp state file: %w", err)
	}

	return nil
}

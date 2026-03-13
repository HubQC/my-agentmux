//go:build desktop

package desktop

import (
	"context"
	"fmt"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/session"
)

// SessionService provides session management for the Wails frontend.
type SessionService struct {
	ctx    context.Context
	mgr    *session.Manager
	runner *agent.Runner
}

// NewSessionService creates a new SessionService.
func NewSessionService(mgr *session.Manager, runner *agent.Runner) *SessionService {
	return &SessionService{
		mgr:    mgr,
		runner: runner,
	}
}

// startup is called by Wails.
func (s *SessionService) startup(ctx context.Context) {
	s.ctx = ctx
}

// ListSessions returns all tracked agent sessions.
func (s *SessionService) ListSessions() []SessionInfo {
	sessions := s.mgr.List(s.ctx)
	result := make([]SessionInfo, len(sessions))
	for i, sess := range sessions {
		result[i] = SessionInfo{
			Name:      sess.Name,
			AgentType: sess.AgentType,
			Status:    sess.Status,
			WorkDir:   sess.WorkDir,
			TmuxName:  sess.TmuxName,
			CreatedAt: sess.CreatedAt,
			Group:     sess.Group,
		}
	}
	return result
}

// CreateSession starts a new agent session.
func (s *SessionService) CreateSession(opts CreateSessionOpts) error {
	launchOpts := agent.LaunchOptions{
		Name:      opts.Name,
		AgentType: opts.AgentType,
		WorkDir:   opts.WorkDir,
		ExtraArgs: opts.Args,
		Env:       opts.Env,
		Command:   opts.Command,
		Group:     opts.Group,
	}

	_, err := s.runner.Launch(s.ctx, launchOpts)
	if err != nil {
		return fmt.Errorf("launching agent: %w", err)
	}
	return nil
}

// StopSession kills the agent session and its tmux session.
func (s *SessionService) StopSession(name string) error {
	return s.mgr.Destroy(s.ctx, name)
}

// GetSession returns details for a single session.
func (s *SessionService) GetSession(name string) (*SessionInfo, error) {
	sess, err := s.mgr.Get(s.ctx, name)
	if err != nil {
		return nil, err
	}
	return &SessionInfo{
		Name:      sess.Name,
		AgentType: sess.AgentType,
		Status:    sess.Status,
		WorkDir:   sess.WorkDir,
		TmuxName:  sess.TmuxName,
		CreatedAt: sess.CreatedAt,
		Group:     sess.Group,
	}, nil
}

// SendKeys sends input strings to the session's tmux pane.
func (s *SessionService) SendKeys(name, keys string) error {
	return s.mgr.SendKeys(s.ctx, name, keys, true)
}

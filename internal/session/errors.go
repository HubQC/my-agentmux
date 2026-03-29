package session

import "errors"

// Sentinel errors for session operations.
var (
	// ErrAgentNotFound is returned when an agent name doesn't match any tracked session.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrAgentAlreadyRunning is returned when trying to create a session that already exists.
	ErrAgentAlreadyRunning = errors.New("agent is already running")

	// ErrAgentNotRunning is returned when an operation requires a running agent.
	ErrAgentNotRunning = errors.New("agent is not running")
)

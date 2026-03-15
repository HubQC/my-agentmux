//go:build desktop

package desktop

import "time"

// SessionInfo represents a summary of an agent session for the UI.
type SessionInfo struct {
	Name      string    `json:"name"`
	AgentType string    `json:"agent_type"`
	Status    string    `json:"status"` // running, stopped, error
	WorkDir   string    `json:"work_dir"`
	TmuxName  string    `json:"tmux_name"`
	CreatedAt time.Time `json:"created_at"`
	Group     string    `json:"group,omitempty"`
}

// HealthInfo represents the health status of an agent.
type HealthInfo struct {
	AgentName    string    `json:"agent_name"`
	Healthy      bool      `json:"healthy"`
	Status       string    `json:"status"`
	LastOutput   time.Time `json:"last_output"`
	RestartCount int       `json:"restart_count"`
}

// CreateSessionOpts defines options for creating a new session.
type CreateSessionOpts struct {
	Name      string            `json:"name"`
	AgentType string            `json:"agent_type"`
	WorkDir   string            `json:"work_dir,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Command   string            `json:"command,omitempty"`
	Group     string            `json:"group,omitempty"`
}

// ResourceInfo represents resource usage of a session.
type ResourceInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage uint64  `json:"memory_usage"`
	PID         int     `json:"pid"`
}

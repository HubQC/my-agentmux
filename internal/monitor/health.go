package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RestartPolicy defines how agents should be restarted on failure.
type RestartPolicy string

const (
	RestartNever     RestartPolicy = "never"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartAlways    RestartPolicy = "always"
)

// HealthConfig defines health monitoring settings for an agent.
type HealthConfig struct {
	// RestartPolicy determines restart behavior.
	RestartPolicy RestartPolicy `yaml:"restart_policy,omitempty"`

	// MaxRestarts is the maximum number of restart attempts (0 = unlimited).
	MaxRestarts int `yaml:"max_restarts,omitempty"`

	// IdleTimeout is the max time without output before considering unhealthy.
	// Zero means no idle timeout.
	IdleTimeout time.Duration `yaml:"idle_timeout,omitempty"`

	// OnUnhealthy is a callback when the agent is deemed unhealthy.
	OnUnhealthy func(agentName string, reason string) `yaml:"-"`

	// OnRestart is a callback when the agent is restarted.
	OnRestart func(agentName string, attempt int) `yaml:"-"`
}

// HealthStatus represents the health of an agent.
type HealthStatus struct {
	AgentName    string
	Healthy      bool
	LastOutput   time.Time
	RestartCount int
	Status       string // "healthy", "idle", "stopped", "restarting"
}

// HealthMonitor tracks agent health and applies restart policies.
type HealthMonitor struct {
	mu       sync.RWMutex
	agents   map[string]*agentHealth
	configs  map[string]HealthConfig
	watcher  *Watcher
	interval time.Duration

	stopCh chan struct{}
}

// agentHealth holds internal health tracking state.
type agentHealth struct {
	name         string
	lastOutput   time.Time
	restartCount int
	status       string
}

// NewHealthMonitor creates a health monitor.
func NewHealthMonitor(watcher *Watcher, checkInterval time.Duration) *HealthMonitor {
	if checkInterval <= 0 {
		checkInterval = 10 * time.Second
	}
	return &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		watcher:  watcher,
		interval: checkInterval,
		stopCh:   make(chan struct{}),
	}
}

// Register adds an agent to health monitoring with the given config.
func (hm *HealthMonitor) Register(agentName string, cfg HealthConfig) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.agents[agentName] = &agentHealth{
		name:       agentName,
		lastOutput: time.Now(),
		status:     "healthy",
	}
	hm.configs[agentName] = cfg
}

// Unregister stops monitoring an agent.
func (hm *HealthMonitor) Unregister(agentName string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.agents, agentName)
	delete(hm.configs, agentName)
}

// RecordOutput marks that the agent produced output (updating its health timestamp).
func (hm *HealthMonitor) RecordOutput(agentName string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	if a, ok := hm.agents[agentName]; ok {
		a.lastOutput = time.Now()
		a.status = "healthy"
	}
}

// GetStatus returns the health status of an agent.
func (hm *HealthMonitor) GetStatus(agentName string) *HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	a, ok := hm.agents[agentName]
	if !ok {
		return nil
	}

	cfg := hm.configs[agentName]
	healthy := true

	// Check idle timeout
	if cfg.IdleTimeout > 0 && time.Since(a.lastOutput) > cfg.IdleTimeout {
		healthy = false
		a.status = "idle"
	}

	return &HealthStatus{
		AgentName:    agentName,
		Healthy:      healthy,
		LastOutput:   a.lastOutput,
		RestartCount: a.restartCount,
		Status:       a.status,
	}
}

// AllStatuses returns health status for all monitored agents.
func (hm *HealthMonitor) AllStatuses() []HealthStatus {
	hm.mu.RLock()
	names := make([]string, 0, len(hm.agents))
	for name := range hm.agents {
		names = append(names, name)
	}
	hm.mu.RUnlock()

	statuses := make([]HealthStatus, 0, len(names))
	for _, name := range names {
		if s := hm.GetStatus(name); s != nil {
			statuses = append(statuses, *s)
		}
	}
	return statuses
}

// Start begins the health monitoring loop.
func (hm *HealthMonitor) Start(ctx context.Context) {
	go hm.monitorLoop(ctx)
}

// Stop halts the health monitor.
func (hm *HealthMonitor) Stop() {
	close(hm.stopCh)
}

func (hm *HealthMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hm.stopCh:
			return
		case <-ticker.C:
			hm.checkAll()
		}
	}
}

func (hm *HealthMonitor) checkAll() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	for name, agent := range hm.agents {
		cfg := hm.configs[name]

		// Check idle timeout
		if cfg.IdleTimeout > 0 && time.Since(agent.lastOutput) > cfg.IdleTimeout {
			agent.status = "idle"
			if cfg.OnUnhealthy != nil {
				idleDur := time.Since(agent.lastOutput).Round(time.Second)
				cfg.OnUnhealthy(name, fmt.Sprintf("no output for %s", idleDur))
			}
		}

		// Check if agent has stopped (via watcher)
		if !hm.watcher.IsWatching(name) {
			if agent.status != "stopped" && agent.status != "restarting" {
				agent.status = "stopped"
				hm.handleStopped(name, cfg, agent)
			}
		}
	}
}

func (hm *HealthMonitor) handleStopped(name string, cfg HealthConfig, agent *agentHealth) {
	switch cfg.RestartPolicy {
	case RestartAlways:
		hm.attemptRestart(name, cfg, agent)
	case RestartOnFailure:
		// Only restart if agent stopped unexpectedly
		hm.attemptRestart(name, cfg, agent)
	case RestartNever, "":
		// Do nothing
	}
}

func (hm *HealthMonitor) attemptRestart(name string, cfg HealthConfig, agent *agentHealth) {
	if cfg.MaxRestarts > 0 && agent.restartCount >= cfg.MaxRestarts {
		agent.status = "stopped"
		if cfg.OnUnhealthy != nil {
			cfg.OnUnhealthy(name, fmt.Sprintf("max restarts (%d) exceeded", cfg.MaxRestarts))
		}
		return
	}

	agent.restartCount++
	agent.status = "restarting"
	agent.lastOutput = time.Now()

	if cfg.OnRestart != nil {
		cfg.OnRestart(name, agent.restartCount)
	}
}

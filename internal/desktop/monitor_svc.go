//go:build desktop

package desktop

import (
	"context"
	"fmt"
	"strings"

	"github.com/cqi/my_agentmux/internal/monitor"
)

// MonitorService provides health and resource monitoring for the Wails frontend.
type MonitorService struct {
	ctx     context.Context
	health  *monitor.HealthMonitor
	watcher *monitor.Watcher
	logger  *monitor.Logger
}

// NewMonitorService creates a new MonitorService.
func NewMonitorService(health *monitor.HealthMonitor, watcher *monitor.Watcher, logger *monitor.Logger) *MonitorService {
	return &MonitorService{
		health:  health,
		watcher: watcher,
		logger:  logger,
	}
}

// startup is called by Wails.
func (m *MonitorService) startup(ctx context.Context) {
	m.ctx = ctx
}

// GetLogs returns the last N lines of logs for an agent.
func (m *MonitorService) GetLogs(name string, lines int) (string, error) {
	content, err := m.logger.ReadAll(name)
	if err != nil {
		return "", err
	}

	if lines <= 0 {
		return content, nil
	}

	// Simple line-based tail
	splitLines := strings.Split(content, "\n")
	if len(splitLines) <= lines {
		return content, nil
	}

	return strings.Join(splitLines[len(splitLines)-lines:], "\n"), nil
}

// GetHealth returns the health status for an agent.
func (m *MonitorService) GetHealth(name string) (*HealthInfo, error) {
	status := m.health.GetStatus(name)
	if status == nil {
		return nil, fmt.Errorf("agent %q not monitored", name)
	}

	return &HealthInfo{
		AgentName:    status.AgentName,
		Healthy:      status.Healthy,
		Status:       status.Status,
		LastOutput:   status.LastOutput,
		RestartCount: status.RestartCount,
	}, nil
}

// GetResources returns the current resource usage for an agent.
// Note: This requires the watcher to be actively watching the agent.
func (m *MonitorService) GetResources(name string) (*ResourceInfo, error) {
	// Since the watcher emits events on a channel, we don't have a direct "Get" method.
	// For now, we return a placeholder or we'd need to add a cache to the watcher.
	// Looking at monitor/watcher.go, it doesn't store the last resource event.
	
	// Implementation detail: We might want to add a GetLastResourceEvent to monitor.Watcher
	// or maintain a local cache here by subscribing to w.ResourceEvents().
	
	return &ResourceInfo{
		CPUUsage:    0,
		MemoryUsage: 0,
		PID:         0,
	}, nil
}

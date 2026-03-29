package monitor

import (
	"sync"
	"testing"
	"time"
)

func TestHealthMonitorRegisterAndGetStatus(t *testing.T) {
	// Create a minimal watcher for testing
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	hm.Register("test-agent", HealthConfig{
		RestartPolicy: RestartNever,
		IdleTimeout:   30 * time.Second,
	})

	status := hm.GetStatus("test-agent")
	if status == nil {
		t.Fatal("expected non-nil status")
	}

	if status.AgentName != "test-agent" {
		t.Errorf("AgentName = %q, want %q", status.AgentName, "test-agent")
	}
	if !status.Healthy {
		t.Error("expected agent to be healthy after registration")
	}
	if status.Status != "healthy" {
		t.Errorf("Status = %q, want %q", status.Status, "healthy")
	}
	if status.RestartCount != 0 {
		t.Errorf("RestartCount = %d, want 0", status.RestartCount)
	}
}

func TestHealthMonitorUnregister(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	hm.Register("test-agent", HealthConfig{})
	hm.Unregister("test-agent")

	status := hm.GetStatus("test-agent")
	if status != nil {
		t.Error("expected nil status after unregister")
	}
}

func TestHealthMonitorRecordOutput(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	hm.Register("test-agent", HealthConfig{})

	// Manually set status to something else to verify RecordOutput resets it
	hm.agents["test-agent"].status = "idle"

	hm.RecordOutput("test-agent")

	status := hm.GetStatus("test-agent")
	if status.Status != "healthy" {
		t.Errorf("Status = %q, want %q after RecordOutput", status.Status, "healthy")
	}
}

func TestHealthMonitorIdleTimeout(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	hm.Register("test-agent", HealthConfig{
		IdleTimeout: 10 * time.Millisecond,
	})

	// Set lastOutput in the past
	hm.agents["test-agent"].lastOutput = time.Now().Add(-1 * time.Second)

	status := hm.GetStatus("test-agent")
	if status.Healthy {
		t.Error("expected agent to be unhealthy after idle timeout")
	}
	if status.Status != "idle" {
		t.Errorf("Status = %q, want %q", status.Status, "idle")
	}
}

func TestHealthMonitorAllStatuses(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	hm.Register("agent-1", HealthConfig{})
	hm.Register("agent-2", HealthConfig{})

	statuses := hm.AllStatuses()
	if len(statuses) != 2 {
		t.Errorf("AllStatuses returned %d, want 2", len(statuses))
	}
}

func TestHealthMonitorConcurrentAccess(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	hm.Register("agent-1", HealthConfig{})

	var wg sync.WaitGroup
	const goroutines = 20

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = hm.GetStatus("agent-1")
		}()
	}

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hm.RecordOutput("agent-1")
		}()
	}

	// Concurrent AllStatuses
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = hm.AllStatuses()
		}()
	}

	wg.Wait()
}

func TestRestartPolicyConstants(t *testing.T) {
	tests := []struct {
		policy RestartPolicy
		want   string
	}{
		{RestartNever, "never"},
		{RestartOnFailure, "on-failure"},
		{RestartAlways, "always"},
	}
	for _, tt := range tests {
		if string(tt.policy) != tt.want {
			t.Errorf("RestartPolicy %v = %q, want %q", tt.policy, string(tt.policy), tt.want)
		}
	}
}

func TestHealthMonitorRecordOutputNonexistent(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	// Should not panic when recording output for non-existent agent
	hm.RecordOutput("nonexistent")
}

func TestHealthMonitorGetStatusNonexistent(t *testing.T) {
	hm := &HealthMonitor{
		agents:   make(map[string]*agentHealth),
		configs:  make(map[string]HealthConfig),
		interval: 1 * time.Second,
		stopCh:   make(chan struct{}),
	}

	status := hm.GetStatus("nonexistent")
	if status != nil {
		t.Error("expected nil status for nonexistent agent")
	}
}

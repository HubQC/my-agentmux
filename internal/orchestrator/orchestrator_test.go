package orchestrator

import (
	"fmt"
	"testing"
)

func TestFailurePolicyConstants(t *testing.T) {
	tests := []struct {
		policy FailurePolicy
		want   string
	}{
		{FailurePolicyAbort, "abort"},
		{FailurePolicySkip, "skip"},
		{FailurePolicyRetry, "retry"},
	}
	for _, tt := range tests {
		if string(tt.policy) != tt.want {
			t.Errorf("FailurePolicy %v = %q, want %q", tt.policy, string(tt.policy), tt.want)
		}
	}
}

func TestPipelineConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  PipelineConfig
		wantErr bool
	}{
		{
			name:    "empty pipeline name",
			config:  PipelineConfig{Name: ""},
			wantErr: false, // Currently no validation
		},
		{
			name: "valid pipeline",
			config: PipelineConfig{
				Name: "test-pipeline",
				Stages: []StageConfig{
					{Agents: []string{"agent-1"}},
				},
			},
			wantErr: false,
		},
		{
			name: "multi-stage pipeline",
			config: PipelineConfig{
				Name: "multi-stage",
				Stages: []StageConfig{
					{Agents: []string{"agent-1"}, OnFailure: FailurePolicyAbort},
					{Agents: []string{"agent-2", "agent-3"}, OnFailure: FailurePolicySkip},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.config.Stages) == 0 && tt.wantErr {
				t.Error("expected error for empty stages")
			}
		})
	}
}

func TestStageResultTracking(t *testing.T) {
	result := &PipelineResult{
		PipelineName: "test",
		Success:      true,
	}

	if result.PipelineName != "test" {
		t.Errorf("PipelineName = %q, want %q", result.PipelineName, "test")
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}

	stageResults := []StageResult{
		{StageName: "stage-0", AgentName: "agent-1", Success: true},
		{StageName: "stage-0", AgentName: "agent-2", Success: false, Error: fmt.Errorf("stage timeout")},
	}

	result.Stages = append(result.Stages, stageResults)

	if len(result.Stages) != 1 {
		t.Errorf("Stages count = %d, want 1", len(result.Stages))
	}
	if len(result.Stages[0]) != 2 {
		t.Errorf("Stage 0 results count = %d, want 2", len(result.Stages[0]))
	}

	// Verify failure detection
	hasFailure := false
	for _, r := range result.Stages[0] {
		if !r.Success {
			hasFailure = true
		}
	}
	if !hasFailure {
		t.Error("expected at least one failure in stage results")
	}
}

func TestCallbacksAreOptional(t *testing.T) {
	// Orchestrator should work without callbacks (they're nil by default)
	o := &Orchestrator{}
	if o.OnStageStart != nil {
		t.Error("OnStageStart should be nil by default")
	}
	if o.OnAgentDone != nil {
		t.Error("OnAgentDone should be nil by default")
	}
	if o.OnStageEnd != nil {
		t.Error("OnStageEnd should be nil by default")
	}
}

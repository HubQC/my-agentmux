package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/session"
)

// FailurePolicy defines how the pipeline handles agent failures.
type FailurePolicy string

const (
	FailurePolicyAbort FailurePolicy = "abort" // Stop the entire pipeline
	FailurePolicySkip  FailurePolicy = "skip"  // Skip failed agent and continue
	FailurePolicyRetry FailurePolicy = "retry" // Retry the failed agent
)

// StageConfig defines a pipeline stage containing one or more parallel agents.
type StageConfig struct {
	// Agents to run in this stage (run in parallel if > 1).
	Agents []string `yaml:"agents"`

	// OnFailure determines behavior when an agent in this stage fails.
	OnFailure FailurePolicy `yaml:"on_failure,omitempty"`

	// Timeout is the max duration for this stage. Zero means no timeout.
	Timeout time.Duration `yaml:"timeout,omitempty"`

	// MaxRetries is the number of retry attempts (only used with retry policy).
	MaxRetries int `yaml:"max_retries,omitempty"`
}

// PipelineConfig defines an enhanced pipeline with stages.
type PipelineConfig struct {
	Name   string        `yaml:"name"`
	Stages []StageConfig `yaml:"stages"`

	// GlobalTimeout is the max duration for the entire pipeline.
	GlobalTimeout time.Duration `yaml:"global_timeout,omitempty"`
}

// StageResult holds the result of a completed stage.
type StageResult struct {
	StageName string
	AgentName string
	Success   bool
	Error     error
	Duration  time.Duration
}

// PipelineResult holds all results from a pipeline run.
type PipelineResult struct {
	PipelineName string
	Stages       [][]StageResult // Results per stage
	Success      bool
	Duration     time.Duration
}

// Orchestrator manages the execution of enhanced pipelines.
type Orchestrator struct {
	cfg    *config.Config
	mgr    *session.Manager
	runner *agent.Runner

	// Callbacks
	OnStageStart func(stageIdx int, agents []string)
	OnAgentDone  func(stageIdx int, agentName string, success bool, err error)
	OnStageEnd   func(stageIdx int, results []StageResult)
}

// NewOrchestrator creates a new pipeline orchestrator.
func NewOrchestrator(cfg *config.Config, mgr *session.Manager, runner *agent.Runner) *Orchestrator {
	return &Orchestrator{
		cfg:    cfg,
		mgr:    mgr,
		runner: runner,
	}
}

// RunPipeline executes a pipeline configuration.
func (o *Orchestrator) RunPipeline(ctx context.Context, pipeline PipelineConfig) (*PipelineResult, error) {
	startTime := time.Now()

	// Apply global timeout if set
	if pipeline.GlobalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pipeline.GlobalTimeout)
		defer cancel()
	}

	result := &PipelineResult{
		PipelineName: pipeline.Name,
		Success:      true,
	}

	for i, stage := range pipeline.Stages {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf("pipeline timed out or cancelled at stage %d", i+1)
		default:
		}

		if o.OnStageStart != nil {
			o.OnStageStart(i, stage.Agents)
		}

		stageResults, err := o.runStage(ctx, i, stage)
		result.Stages = append(result.Stages, stageResults)

		if o.OnStageEnd != nil {
			o.OnStageEnd(i, stageResults)
		}

		if err != nil {
			result.Success = false
			if stage.OnFailure == FailurePolicyAbort || stage.OnFailure == "" {
				result.Duration = time.Since(startTime)
				return result, fmt.Errorf("stage %d failed: %w", i+1, err)
			}
			// FailurePolicySkip — continue to next stage
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// runStage executes all agents in a stage (in parallel if >1).
func (o *Orchestrator) runStage(ctx context.Context, stageIdx int, stage StageConfig) ([]StageResult, error) {
	// Apply stage timeout if set
	if stage.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, stage.Timeout)
		defer cancel()
	}

	if len(stage.Agents) == 1 {
		// Single agent — run directly
		result := o.runAgent(ctx, stageIdx, stage.Agents[0], stage)
		return []StageResult{result}, resultError(result)
	}

	// Multiple agents — run in parallel
	var (
		mu      sync.Mutex
		results = make([]StageResult, len(stage.Agents))
		wg      sync.WaitGroup
	)

	for i, agentName := range stage.Agents {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			result := o.runAgent(ctx, stageIdx, name, stage)
			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, agentName)
	}

	wg.Wait()

	// Check for any failures
	var firstErr error
	for _, r := range results {
		if !r.Success && firstErr == nil {
			firstErr = r.Error
		}
	}

	return results, firstErr
}

// runAgent launches and waits for a single agent, with optional retries.
func (o *Orchestrator) runAgent(ctx context.Context, stageIdx int, agentName string, stage StageConfig) StageResult {
	maxAttempts := 1
	if stage.OnFailure == FailurePolicyRetry && stage.MaxRetries > 0 {
		maxAttempts = stage.MaxRetries + 1
	}

	startTime := time.Now()

	for attempt := 0; attempt < maxAttempts; attempt++ {
		opts := agent.LaunchOptions{
			Name: agentName,
		}

		_, err := o.runner.Launch(ctx, opts)
		if err != nil {
			if attempt == maxAttempts-1 {
				result := StageResult{
					AgentName: agentName,
					Success:   false,
					Error:     err,
					Duration:  time.Since(startTime),
				}
				if o.OnAgentDone != nil {
					o.OnAgentDone(stageIdx, agentName, false, err)
				}
				return result
			}
			continue // Retry
		}

		// Wait for agent completion
		err = o.waitForCompletion(ctx, agentName)

		if err == nil {
			result := StageResult{
				AgentName: agentName,
				Success:   true,
				Duration:  time.Since(startTime),
			}
			if o.OnAgentDone != nil {
				o.OnAgentDone(stageIdx, agentName, true, nil)
			}
			return result
		}

		// Clean up before retry
		_ = o.mgr.Destroy(ctx, agentName)

		if attempt == maxAttempts-1 {
			result := StageResult{
				AgentName: agentName,
				Success:   false,
				Error:     err,
				Duration:  time.Since(startTime),
			}
			if o.OnAgentDone != nil {
				o.OnAgentDone(stageIdx, agentName, false, err)
			}
			return result
		}
	}

	// Should not reach here
	return StageResult{AgentName: agentName, Success: false}
}

// waitForCompletion polls until an agent finishes or the context is cancelled.
func (o *Orchestrator) waitForCompletion(ctx context.Context, agentName string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sess, err := o.mgr.Get(ctx, agentName)
			if err != nil {
				return nil // Session gone = agent finished
			}
			if sess.Status != "running" {
				_ = o.mgr.Destroy(ctx, agentName)
				return nil
			}
		}
	}
}

func resultError(r StageResult) error {
	if r.Success {
		return nil
	}
	return r.Error
}

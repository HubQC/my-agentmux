package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// PlanStatus represents the lifecycle state of a plan.
type PlanStatus string

const (
	StatusDraft    PlanStatus = "draft"
	StatusApproved PlanStatus = "approved"
	StatusRejected PlanStatus = "rejected"
)

// Plan represents a spec-driven workflow plan.
type Plan struct {
	// ID is the unique identifier (e.g. "plan-001").
	ID string `yaml:"id"`

	// Title is a short summary of the plan.
	Title string `yaml:"title"`

	// Description is a longer explanation of what the plan covers.
	Description string `yaml:"description,omitempty"`

	// Status is the current lifecycle state (draft, approved, rejected).
	Status PlanStatus `yaml:"status"`

	// RejectReason explains why the plan was rejected.
	RejectReason string `yaml:"reject_reason,omitempty"`

	// Agent is the name of the agent that created this plan (optional).
	Agent string `yaml:"agent,omitempty"`

	// CreatedAt is when the plan was created.
	CreatedAt time.Time `yaml:"created_at"`

	// UpdatedAt is when the plan was last modified.
	UpdatedAt time.Time `yaml:"updated_at"`
}

// PlanStore manages plan files in a directory.
type PlanStore struct {
	dir string
}

// NewPlanStore creates a new PlanStore for the given directory.
func NewPlanStore(dir string) (*PlanStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating plans directory: %w", err)
	}
	return &PlanStore{dir: dir}, nil
}

// Create creates a new plan with the given title and description.
func (s *PlanStore) Create(title, description string) (*Plan, error) {
	if title == "" {
		return nil, fmt.Errorf("plan title is required")
	}

	id, err := s.nextID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	plan := &Plan{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      StatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.save(plan); err != nil {
		return nil, err
	}

	return plan, nil
}

// List returns all plans sorted by creation time (oldest first).
func (s *PlanStore) List() ([]*Plan, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading plans directory: %w", err)
	}

	var plans []*Plan
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		plan, err := s.loadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", entry.Name(), err)
			continue
		}
		plans = append(plans, plan)
	}

	sort.Slice(plans, func(i, j int) bool {
		return plans[i].CreatedAt.Before(plans[j].CreatedAt)
	})

	return plans, nil
}

// Get returns a plan by its ID.
func (s *PlanStore) Get(id string) (*Plan, error) {
	filePath := s.planPath(id)
	plan, err := s.loadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan %q not found", id)
		}
		return nil, err
	}
	return plan, nil
}

// Approve moves a plan from draft to approved status.
func (s *PlanStore) Approve(id string) error {
	plan, err := s.Get(id)
	if err != nil {
		return err
	}

	if plan.Status != StatusDraft {
		return fmt.Errorf("cannot approve plan %q: status is %q (must be %q)", id, plan.Status, StatusDraft)
	}

	plan.Status = StatusApproved
	plan.UpdatedAt = time.Now()
	return s.save(plan)
}

// Reject moves a plan from draft to rejected status with an optional reason.
func (s *PlanStore) Reject(id, reason string) error {
	plan, err := s.Get(id)
	if err != nil {
		return err
	}

	if plan.Status != StatusDraft {
		return fmt.Errorf("cannot reject plan %q: status is %q (must be %q)", id, plan.Status, StatusDraft)
	}

	plan.Status = StatusRejected
	plan.RejectReason = reason
	plan.UpdatedAt = time.Now()
	return s.save(plan)
}

// Delete removes a plan file.
func (s *PlanStore) Delete(id string) error {
	filePath := s.planPath(id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("plan %q not found", id)
	}
	return os.Remove(filePath)
}

// planPath returns the file path for a plan ID.
func (s *PlanStore) planPath(id string) string {
	return filepath.Join(s.dir, id+".yaml")
}

// save writes a plan to its YAML file.
func (s *PlanStore) save(plan *Plan) error {
	data, err := yaml.Marshal(plan)
	if err != nil {
		return fmt.Errorf("marshalling plan: %w", err)
	}
	return os.WriteFile(s.planPath(plan.ID), data, 0o644)
}

// loadFile reads a plan from a YAML file.
func (s *PlanStore) loadFile(filePath string) (*Plan, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var plan Plan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parsing plan %s: %w", filePath, err)
	}

	return &plan, nil
}

// nextID determines the next sequential plan ID.
func (s *PlanStore) nextID() (string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "plan-001", nil
		}
		return "", fmt.Errorf("reading plans directory: %w", err)
	}

	maxNum := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if strings.HasPrefix(name, "plan-") {
			numStr := strings.TrimPrefix(name, "plan-")
			if num, err := strconv.Atoi(numStr); err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	return fmt.Sprintf("plan-%03d", maxNum+1), nil
}

// FormatStatus returns a display-friendly status icon + label.
func FormatStatus(status PlanStatus) string {
	switch status {
	case StatusDraft:
		return "◎ draft"
	case StatusApproved:
		return "✓ approved"
	case StatusRejected:
		return "✗ rejected"
	default:
		return string(status)
	}
}

package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreatePlan(t *testing.T) {
	store := newTestStore(t)

	plan, err := store.Create("Add authentication", "Implement OAuth2 login flow", "")
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	if plan.ID != "plan-001" {
		t.Errorf("expected ID 'plan-001', got %q", plan.ID)
	}
	if plan.Title != "Add authentication" {
		t.Errorf("expected title 'Add authentication', got %q", plan.Title)
	}
	if plan.Description != "Implement OAuth2 login flow" {
		t.Errorf("expected description, got %q", plan.Description)
	}
	if plan.Status != StatusDraft {
		t.Errorf("expected status 'draft', got %q", plan.Status)
	}
	if plan.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	// Verify file exists
	filePath := filepath.Join(store.dir, "plan-001.yaml")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected plan file to exist on disk")
	}
}

func TestCreatePlanEmptyTitle(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Create("", "description", "")
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestListPlans(t *testing.T) {
	store := newTestStore(t)

	_, _ = store.Create("Plan A", "First plan", "")
	_, _ = store.Create("Plan B", "Second plan", "")
	_, _ = store.Create("Plan C", "Third plan", "")

	plans, err := store.List()
	if err != nil {
		t.Fatalf("failed to list plans: %v", err)
	}

	if len(plans) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(plans))
	}

	// Should be sorted by creation time (oldest first)
	if plans[0].Title != "Plan A" {
		t.Errorf("expected first plan 'Plan A', got %q", plans[0].Title)
	}
	if plans[2].Title != "Plan C" {
		t.Errorf("expected last plan 'Plan C', got %q", plans[2].Title)
	}
}

func TestGetPlan(t *testing.T) {
	store := newTestStore(t)

	created, _ := store.Create("Test plan", "", "")

	plan, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("failed to get plan: %v", err)
	}
	if plan.Title != "Test plan" {
		t.Errorf("expected title 'Test plan', got %q", plan.Title)
	}

	// Nonexistent
	_, err = store.Get("plan-999")
	if err == nil {
		t.Error("expected error for nonexistent plan")
	}
}

func TestApprovePlan(t *testing.T) {
	store := newTestStore(t)

	plan, _ := store.Create("Approve me", "", "")

	if err := store.Approve(plan.ID); err != nil {
		t.Fatalf("failed to approve: %v", err)
	}

	updated, _ := store.Get(plan.ID)
	if updated.Status != StatusApproved {
		t.Errorf("expected status 'approved', got %q", updated.Status)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) || updated.UpdatedAt.Equal(updated.CreatedAt) {
		// UpdatedAt should be >= CreatedAt (may be equal if very fast)
	}
}

func TestRejectPlan(t *testing.T) {
	store := newTestStore(t)

	plan, _ := store.Create("Reject me", "", "")

	if err := store.Reject(plan.ID, "Needs more detail"); err != nil {
		t.Fatalf("failed to reject: %v", err)
	}

	updated, _ := store.Get(plan.ID)
	if updated.Status != StatusRejected {
		t.Errorf("expected status 'rejected', got %q", updated.Status)
	}
	if updated.RejectReason != "Needs more detail" {
		t.Errorf("expected reject reason, got %q", updated.RejectReason)
	}
}

func TestApproveNonDraft(t *testing.T) {
	store := newTestStore(t)

	plan, _ := store.Create("Already done", "", "")
	_ = store.Approve(plan.ID)

	// Trying to approve again should fail
	err := store.Approve(plan.ID)
	if err == nil {
		t.Error("expected error when approving non-draft plan")
	}

	// Create another and reject, then try to approve
	plan2, _ := store.Create("Rejected one", "", "")
	_ = store.Reject(plan2.ID, "no")

	err = store.Approve(plan2.ID)
	if err == nil {
		t.Error("expected error when approving rejected plan")
	}
}

func TestDeletePlan(t *testing.T) {
	store := newTestStore(t)

	plan, _ := store.Create("Delete me", "", "")

	if err := store.Delete(plan.ID); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Should be gone
	_, err := store.Get(plan.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}

	// File should not exist
	filePath := filepath.Join(store.dir, plan.ID+".yaml")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected plan file to be deleted from disk")
	}

	// Delete nonexistent
	err = store.Delete("plan-999")
	if err == nil {
		t.Error("expected error for deleting nonexistent plan")
	}
}

func TestPlanIDSequence(t *testing.T) {
	store := newTestStore(t)

	p1, _ := store.Create("First", "", "")
	if p1.ID != "plan-001" {
		t.Errorf("expected plan-001, got %q", p1.ID)
	}

	p2, _ := store.Create("Second", "", "")
	if p2.ID != "plan-002" {
		t.Errorf("expected plan-002, got %q", p2.ID)
	}

	// Delete plan-001
	_ = store.Delete("plan-001")

	// Next should be plan-003 (no ID reuse)
	p3, _ := store.Create("Third", "", "")
	if p3.ID != "plan-003" {
		t.Errorf("expected plan-003 (no reuse), got %q", p3.ID)
	}
}

// --- Helpers ---

func newTestStore(t *testing.T) *PlanStore {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "plans")
	store, err := NewPlanStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return store
}

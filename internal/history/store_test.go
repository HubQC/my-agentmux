package history

import (
	"testing"
	"time"
)

func TestStoreNewAndList(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	entries := store.List()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestStoreRecord(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	entry := Entry{
		Name:      "test-agent",
		AgentType: "claude",
		WorkDir:   "/tmp/test",
		StartedAt: time.Now(),
		Status:    "completed",
		Duration:  5 * time.Minute,
	}

	if err := store.Record(entry); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	entries := store.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "test-agent" {
		t.Errorf("Name = %q, want %q", entries[0].Name, "test-agent")
	}
}

func TestStoreListFilteredByType(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	now := time.Now()
	store.Record(Entry{Name: "a1", AgentType: "claude", StartedAt: now, Status: "completed"})
	store.Record(Entry{Name: "a2", AgentType: "aider", StartedAt: now.Add(1 * time.Second), Status: "completed"})
	store.Record(Entry{Name: "a3", AgentType: "claude", StartedAt: now.Add(2 * time.Second), Status: "failed"})

	filtered := store.ListFiltered(FilterOptions{AgentType: "claude"})
	if len(filtered) != 2 {
		t.Errorf("expected 2 claude entries, got %d", len(filtered))
	}
}

func TestStoreListFilteredByStatus(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	now := time.Now()
	store.Record(Entry{Name: "a1", AgentType: "claude", StartedAt: now, Status: "completed"})
	store.Record(Entry{Name: "a2", AgentType: "claude", StartedAt: now.Add(1 * time.Second), Status: "failed"})

	filtered := store.ListFiltered(FilterOptions{Status: "failed"})
	if len(filtered) != 1 {
		t.Errorf("expected 1 failed entry, got %d", len(filtered))
	}
}

func TestStoreListFilteredWithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	for i := 0; i < 10; i++ {
		store.Record(Entry{
			Name:      "agent",
			AgentType: "shell",
			StartedAt: time.Now().Add(time.Duration(i) * time.Second),
			Status:    "completed",
		})
	}

	filtered := store.ListFiltered(FilterOptions{Limit: 3})
	if len(filtered) != 3 {
		t.Errorf("expected 3 entries with limit, got %d", len(filtered))
	}
}

func TestStoreStats(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	now := time.Now()
	store.Record(Entry{Name: "a1", AgentType: "claude", StartedAt: now, Status: "completed", Duration: 5 * time.Minute})
	store.Record(Entry{Name: "a2", AgentType: "aider", StartedAt: now, Status: "failed", Duration: 2 * time.Minute})
	store.Record(Entry{Name: "a3", AgentType: "claude", StartedAt: now, Status: "completed", Duration: 10 * time.Minute})

	stats := store.Stats()
	if stats.TotalSessions != 3 {
		t.Errorf("TotalSessions = %d, want 3", stats.TotalSessions)
	}
	if stats.Completed != 2 {
		t.Errorf("Completed = %d, want 2", stats.Completed)
	}
	if stats.Failed != 1 {
		t.Errorf("Failed = %d, want 1", stats.Failed)
	}
	if stats.AgentTypeCounts["claude"] != 2 {
		t.Errorf("claude count = %d, want 2", stats.AgentTypeCounts["claude"])
	}
	if stats.TotalDuration != 17*time.Minute {
		t.Errorf("TotalDuration = %v, want %v", stats.TotalDuration, 17*time.Minute)
	}
}

func TestStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create and populate store
	store1, _ := NewStore(tmpDir)
	store1.Record(Entry{Name: "test", AgentType: "shell", StartedAt: time.Now(), Status: "completed"})

	// Load from same directory — should see the entry
	store2, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore (re-open) failed: %v", err)
	}

	entries := store2.List()
	if len(entries) != 1 {
		t.Errorf("persisted entries: expected 1, got %d", len(entries))
	}
}

func TestStoreCapsAt1000(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	for i := 0; i < 1005; i++ {
		store.Record(Entry{
			Name:      "agent",
			AgentType: "shell",
			StartedAt: time.Now(),
			Status:    "completed",
		})
	}

	entries := store.List()
	if len(entries) > 1000 {
		t.Errorf("expected max 1000 entries, got %d", len(entries))
	}
}

package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a historical agent session record.
type Entry struct {
	Name      string        `json:"name"`
	AgentType string        `json:"agent_type"`
	WorkDir   string        `json:"work_dir"`
	Group     string        `json:"group,omitempty"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   *time.Time    `json:"ended_at,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	ExitCode  int           `json:"exit_code,omitempty"`
	Status    string        `json:"status"` // "completed", "failed", "stopped", "running"

	// Resource metrics (peak values)
	PeakCPU    float64 `json:"peak_cpu,omitempty"`
	PeakMemory uint64  `json:"peak_memory,omitempty"`

	// Git context
	GitBranch string `json:"git_branch,omitempty"`
	GitRepo   string `json:"git_repo,omitempty"`
}

// Store persists session history to a JSON file.
type Store struct {
	filePath string
	entries  []Entry
}

// NewStore creates a history store backed by a JSON file.
func NewStore(dataDir string) (*Store, error) {
	filePath := filepath.Join(dataDir, "history.json")
	s := &Store{filePath: filePath}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("loading history: %w", err)
	}

	return s, nil
}

// Record adds a new history entry.
func (s *Store) Record(entry Entry) error {
	s.entries = append(s.entries, entry)

	// Cap at 1000 entries, removing oldest
	if len(s.entries) > 1000 {
		s.entries = s.entries[len(s.entries)-1000:]
	}

	return s.save()
}

// List returns all history entries, newest first.
func (s *Store) List() []Entry {
	result := make([]Entry, len(s.entries))
	copy(result, s.entries)

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartedAt.After(result[j].StartedAt)
	})

	return result
}

// ListFiltered returns entries matching the given filters.
func (s *Store) ListFiltered(opts FilterOptions) []Entry {
	all := s.List()
	var filtered []Entry

	for _, e := range all {
		if opts.AgentType != "" && e.AgentType != opts.AgentType {
			continue
		}
		if opts.Status != "" && e.Status != opts.Status {
			continue
		}
		if !opts.Since.IsZero() && e.StartedAt.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && e.StartedAt.After(opts.Until) {
			continue
		}
		if opts.Limit > 0 && len(filtered) >= opts.Limit {
			break
		}
		filtered = append(filtered, e)
	}

	return filtered
}

// FilterOptions defines filters for listing history.
type FilterOptions struct {
	AgentType string
	Status    string
	Since     time.Time
	Until     time.Time
	Limit     int
}

// Stats returns aggregate statistics from the history.
func (s *Store) Stats() HistoryStats {
	stats := HistoryStats{
		AgentTypeCounts: make(map[string]int),
	}

	for _, e := range s.entries {
		stats.TotalSessions++
		stats.AgentTypeCounts[e.AgentType]++

		if e.Status == "completed" {
			stats.Completed++
		} else if e.Status == "failed" {
			stats.Failed++
		}

		stats.TotalDuration += e.Duration
	}

	return stats
}

// HistoryStats holds aggregate history statistics.
type HistoryStats struct {
	TotalSessions   int
	Completed       int
	Failed          int
	TotalDuration   time.Duration
	AgentTypeCounts map[string]int
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = nil
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.entries)
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0o644)
}

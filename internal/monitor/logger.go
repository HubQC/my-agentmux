package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger manages per-agent log files.
type Logger struct {
	mu        sync.Mutex
	logsDir   string
	maxSizeMB int
	files     map[string]*os.File
}

// NewLogger creates a new logger that writes to the given directory.
func NewLogger(logsDir string, maxSizeMB int) (*Logger, error) {
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating logs directory: %w", err)
	}

	if maxSizeMB <= 0 {
		maxSizeMB = 50
	}

	return &Logger{
		logsDir:   logsDir,
		maxSizeMB: maxSizeMB,
		files:     make(map[string]*os.File),
	}, nil
}

// LogPath returns the log file path for a given agent.
func (l *Logger) LogPath(agentName string) string {
	return filepath.Join(l.logsDir, agentName+".log")
}

// Write appends output to the agent's log file.
func (l *Logger) Write(agentName string, content string) error {
	if content == "" {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := l.getOrOpenFile(agentName)
	if err != nil {
		return err
	}

	// Check if rotation is needed
	if err := l.rotateIfNeeded(agentName, f); err != nil {
		return err
	}

	// Re-get file handle after potential rotation
	f, err = l.getOrOpenFile(agentName)
	if err != nil {
		return err
	}

	_, err = f.WriteString(content)
	return err
}

// WriteTimestamped appends timestamped output to the agent's log file.
func (l *Logger) WriteTimestamped(agentName string, content string) error {
	if content == "" {
		return nil
	}
	ts := time.Now().Format("2006-01-02T15:04:05")
	return l.Write(agentName, fmt.Sprintf("[%s] %s\n", ts, content))
}

// ReadAll reads the entire log file for an agent.
func (l *Logger) ReadAll(agentName string) (string, error) {
	logPath := l.LogPath(agentName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading log file: %w", err)
	}
	return string(data), nil
}

// Close closes all open log file handles.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var lastErr error
	for name, f := range l.files {
		if err := f.Close(); err != nil {
			lastErr = err
		}
		delete(l.files, name)
	}
	return lastErr
}

// CloseAgent closes the log file handle for a specific agent.
func (l *Logger) CloseAgent(agentName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if f, ok := l.files[agentName]; ok {
		delete(l.files, agentName)
		return f.Close()
	}
	return nil
}

// getOrOpenFile returns the open file handle for an agent, opening it if needed.
// Must be called with l.mu held.
func (l *Logger) getOrOpenFile(agentName string) (*os.File, error) {
	if f, ok := l.files[agentName]; ok {
		return f, nil
	}

	logPath := l.LogPath(agentName)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file %s: %w", logPath, err)
	}

	l.files[agentName] = f
	return f, nil
}

// rotateIfNeeded rotates the log file if it exceeds maxSizeMB.
// Must be called with l.mu held.
func (l *Logger) rotateIfNeeded(agentName string, f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return nil // can't stat, skip rotation
	}

	maxBytes := int64(l.maxSizeMB) * 1024 * 1024
	if info.Size() < maxBytes {
		return nil
	}

	// Close current file
	f.Close()
	delete(l.files, agentName)

	// Rotate: rename current to .1, discard any existing .1
	logPath := l.LogPath(agentName)
	rotatedPath := logPath + ".1"
	_ = os.Remove(rotatedPath)
	if err := os.Rename(logPath, rotatedPath); err != nil {
		return fmt.Errorf("rotating log file: %w", err)
	}

	return nil
}

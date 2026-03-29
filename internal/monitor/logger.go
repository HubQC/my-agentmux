package monitor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// logFile wraps an os.File with a buffered writer for reduced syscall overhead.
type logFile struct {
	file   *os.File
	writer *bufio.Writer
}

// Logger manages per-agent log files with buffered writes.
type Logger struct {
	mu        sync.Mutex
	logsDir   string
	maxSizeMB int
	files     map[string]*logFile

	stopCh chan struct{}
}

// NewLogger creates a new logger that writes to the given directory.
// Writes are buffered and flushed periodically to reduce I/O overhead.
func NewLogger(logsDir string, maxSizeMB int) (*Logger, error) {
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating logs directory: %w", err)
	}

	if maxSizeMB <= 0 {
		maxSizeMB = 50
	}

	l := &Logger{
		logsDir:   logsDir,
		maxSizeMB: maxSizeMB,
		files:     make(map[string]*logFile),
		stopCh:    make(chan struct{}),
	}

	// Start periodic flush goroutine
	go l.flushLoop()

	return l, nil
}

// flushLoop periodically flushes all buffered writers.
func (l *Logger) flushLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.flushAll()
		case <-l.stopCh:
			return
		}
	}
}

// flushAll flushes all open buffered writers.
func (l *Logger) flushAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, lf := range l.files {
		_ = lf.writer.Flush()
	}
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

	lf, err := l.getOrOpenFile(agentName)
	if err != nil {
		return err
	}

	// Check if rotation is needed
	if err := l.rotateIfNeeded(agentName, lf); err != nil {
		return err
	}

	// Re-get file handle after potential rotation
	lf, err = l.getOrOpenFile(agentName)
	if err != nil {
		return err
	}

	_, err = lf.writer.WriteString(content)
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
	// Flush before reading to ensure all buffered data is on disk
	l.mu.Lock()
	if lf, ok := l.files[agentName]; ok {
		_ = lf.writer.Flush()
	}
	l.mu.Unlock()

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

// Close flushes and closes all open log file handles and stops the flush loop.
func (l *Logger) Close() error {
	// Stop the flush goroutine
	select {
	case <-l.stopCh:
		// Already stopped
	default:
		close(l.stopCh)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var lastErr error
	for name, lf := range l.files {
		if err := lf.writer.Flush(); err != nil {
			lastErr = err
		}
		if err := lf.file.Close(); err != nil {
			lastErr = err
		}
		delete(l.files, name)
	}
	return lastErr
}

// CloseAgent flushes and closes the log file handle for a specific agent.
func (l *Logger) CloseAgent(agentName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if lf, ok := l.files[agentName]; ok {
		delete(l.files, agentName)
		_ = lf.writer.Flush()
		return lf.file.Close()
	}
	return nil
}

// getOrOpenFile returns the open logFile for an agent, opening it if needed.
// Must be called with l.mu held.
func (l *Logger) getOrOpenFile(agentName string) (*logFile, error) {
	if lf, ok := l.files[agentName]; ok {
		return lf, nil
	}

	logPath := l.LogPath(agentName)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file %s: %w", logPath, err)
	}

	lf := &logFile{
		file:   f,
		writer: bufio.NewWriterSize(f, 8192), // 8KB buffer
	}
	l.files[agentName] = lf
	return lf, nil
}

// rotateIfNeeded rotates the log file if it exceeds maxSizeMB.
// Must be called with l.mu held.
func (l *Logger) rotateIfNeeded(agentName string, lf *logFile) error {
	info, err := lf.file.Stat()
	if err != nil {
		return nil // can't stat, skip rotation
	}

	maxBytes := int64(l.maxSizeMB) * 1024 * 1024
	if info.Size() < maxBytes {
		return nil
	}

	// Flush and close current file
	_ = lf.writer.Flush()
	lf.file.Close()
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

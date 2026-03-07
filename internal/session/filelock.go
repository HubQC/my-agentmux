package session

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// FileLock provides cross-process file locking using flock.
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a new file lock at the given path.
func NewFileLock(path string) *FileLock {
	return &FileLock{path: path}
}

// Lock acquires an exclusive file lock, blocking until available or timeout.
func (fl *FileLock) Lock(timeout time.Duration) error {
	f, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("opening lock file: %w", err)
	}
	fl.file = f

	// Try non-blocking first
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == nil {
		return nil // Got the lock immediately
	}

	// Fall back to timed blocking
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	done := make(chan error, 1)
	go func() {
		done <- syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	}()

	select {
	case err := <-done:
		if err != nil {
			_ = f.Close()
			fl.file = nil
			return fmt.Errorf("acquiring lock: %w", err)
		}
		return nil
	case <-time.After(timeout):
		_ = f.Close()
		fl.file = nil
		return fmt.Errorf("timeout waiting for state file lock (another agentmux process may be running)")
	}
}

// Unlock releases the file lock.
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return nil
	}
	err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	closeErr := fl.file.Close()
	fl.file = nil
	if err != nil {
		return err
	}
	return closeErr
}

package tmux

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Session represents a tmux session.
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Windows   int       `json:"windows"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Created   time.Time `json:"created"`
	Attached  bool      `json:"attached"`
}

// Window represents a tmux window within a session.
type Window struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Active    bool   `json:"active"`
	Panes     int    `json:"panes"`
}

// Pane represents a tmux pane within a window.
type Pane struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	WindowID  string `json:"window_id"`
	Index     int    `json:"index"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Active    bool   `json:"active"`
	PID       int    `json:"pid"`
	Command   string `json:"command"`
}

// SessionOptions configures a new tmux session.
type SessionOptions struct {
	Name       string
	StartDir   string
	WindowName string
	Width      int
	Height     int
	Detached   bool
	Command    string
	Env        map[string]string
}

// SplitOptions configures a pane split.
type SplitOptions struct {
	Target     string // target pane identifier
	Horizontal bool   // split horizontally (top/bottom) if true, vertically (left/right) if false
	Percent    int    // percentage of the split
	StartDir   string
	Command    string
}

// parseSession parses a tmux format string into a Session.
// Expected format: id|name|windows|width|height|created_epoch|attached
func parseSession(line string) (Session, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 7 {
		return Session{}, fmt.Errorf("invalid session format: %q (expected 7 fields, got %d)", line, len(parts))
	}

	windows, _ := strconv.Atoi(parts[2])
	width, _ := strconv.Atoi(parts[3])
	height, _ := strconv.Atoi(parts[4])
	epoch, _ := strconv.ParseInt(parts[5], 10, 64)

	return Session{
		ID:       parts[0],
		Name:     parts[1],
		Windows:  windows,
		Width:    width,
		Height:   height,
		Created:  time.Unix(epoch, 0),
		Attached: parts[6] == "1",
	}, nil
}

// parseWindow parses a tmux format string into a Window.
// Expected format: id|session_id|index|name|width|height|active|panes
func parseWindow(line string) (Window, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 8 {
		return Window{}, fmt.Errorf("invalid window format: %q (expected 8 fields, got %d)", line, len(parts))
	}

	index, _ := strconv.Atoi(parts[2])
	width, _ := strconv.Atoi(parts[4])
	height, _ := strconv.Atoi(parts[5])
	panes, _ := strconv.Atoi(parts[7])

	return Window{
		ID:        parts[0],
		SessionID: parts[1],
		Index:     index,
		Name:      parts[3],
		Width:     width,
		Height:    height,
		Active:    parts[6] == "1",
		Panes:     panes,
	}, nil
}

// parsePane parses a tmux format string into a Pane.
// Expected format: id|session_id|window_id|index|width|height|active|pid|command
func parsePane(line string) (Pane, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 9 {
		return Pane{}, fmt.Errorf("invalid pane format: %q (expected 9 fields, got %d)", line, len(parts))
	}

	index, _ := strconv.Atoi(parts[3])
	width, _ := strconv.Atoi(parts[4])
	height, _ := strconv.Atoi(parts[5])
	pid, _ := strconv.Atoi(parts[7])

	return Pane{
		ID:        parts[0],
		SessionID: parts[1],
		WindowID:  parts[2],
		Index:     index,
		Width:     width,
		Height:    height,
		Active:    parts[6] == "1",
		PID:       pid,
		Command:   parts[8],
	}, nil
}

// sessionFormat is the tmux format string for listing sessions.
const sessionFormat = "#{session_id}|#{session_name}|#{session_windows}|#{session_width}|#{session_height}|#{session_created}|#{session_attached}"

// windowFormat is the tmux format string for listing windows.
const windowFormat = "#{window_id}|#{session_id}|#{window_index}|#{window_name}|#{window_width}|#{window_height}|#{window_active}|#{window_panes}"

// paneFormat is the tmux format string for listing panes.
const paneFormat = "#{pane_id}|#{session_id}|#{window_id}|#{pane_index}|#{pane_width}|#{pane_height}|#{pane_active}|#{pane_pid}|#{pane_current_command}"

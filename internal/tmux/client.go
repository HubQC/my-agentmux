package tmux

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Client wraps tmux CLI commands for session/window/pane management.
type Client struct {
	// Binary is the path to the tmux executable.
	Binary string

	// DefaultTimeout for tmux commands.
	DefaultTimeout time.Duration
}

// NewClient creates a new tmux client.
// If binary is empty, it defaults to "tmux".
func NewClient(binary string) (*Client, error) {
	if binary == "" {
		binary = "tmux"
	}

	// Verify tmux is installed and reachable
	path, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("tmux not found at %q: %w (install tmux first)", binary, err)
	}

	return &Client{
		Binary:         path,
		DefaultTimeout: 10 * time.Second,
	}, nil
}

// run executes a tmux command and returns its stdout.
func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.DefaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, c.Binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tmux %s failed: %w (stderr: %s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ServerRunning checks if the tmux server is running.
func (c *Client) ServerRunning(ctx context.Context) bool {
	_, err := c.run(ctx, "list-sessions")
	return err == nil
}

// ---- Session operations ----

// NewSession creates a new tmux session with the given options.
func (c *Client) NewSession(ctx context.Context, opts SessionOptions) (*Session, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("session name is required")
	}

	args := []string{"new-session", "-d", "-s", opts.Name, "-P", "-F", sessionFormat}

	if opts.StartDir != "" {
		args = append(args, "-c", opts.StartDir)
	}
	if opts.WindowName != "" {
		args = append(args, "-n", opts.WindowName)
	}
	if opts.Width > 0 && opts.Height > 0 {
		args = append(args, "-x", fmt.Sprintf("%d", opts.Width), "-y", fmt.Sprintf("%d", opts.Height))
	} else {
		// Default size for detached sessions to avoid "size missing" errors
		// when performing operations like split-window
		args = append(args, "-x", "200", "-y", "50")
	}
	if opts.Command != "" {
		args = append(args, opts.Command)
	}


	output, err := c.run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("creating session %q: %w", opts.Name, err)
	}

	session, err := parseSession(output)
	if err != nil {
		return nil, err
	}

	// Set environment variables after session creation
	for k, v := range opts.Env {
		_, _ = c.run(ctx, "set-environment", "-t", opts.Name, k, v)
	}

	return &session, nil
}

// KillSession destroys a tmux session by name.
func (c *Client) KillSession(ctx context.Context, name string) error {
	_, err := c.run(ctx, "kill-session", "-t", name)
	if err != nil {
		return fmt.Errorf("killing session %q: %w", name, err)
	}
	return nil
}

// HasSession checks if a tmux session with the given name exists.
func (c *Client) HasSession(ctx context.Context, name string) bool {
	_, err := c.run(ctx, "has-session", "-t", name)
	return err == nil
}

// ListSessions returns all active tmux sessions.
func (c *Client) ListSessions(ctx context.Context) ([]Session, error) {
	output, err := c.run(ctx, "list-sessions", "-F", sessionFormat)
	if err != nil {
		// If no server is running, return empty list
		if strings.Contains(err.Error(), "no server running") || strings.Contains(err.Error(), "no sessions") {
			return nil, nil
		}
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	if output == "" {
		return nil, nil
	}

	var sessions []Session
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		s, err := parseSession(line)
		if err != nil {
			continue // skip malformed lines
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// GetSession returns info about a specific session by name.
func (c *Client) GetSession(ctx context.Context, name string) (*Session, error) {
	output, err := c.run(ctx, "list-sessions", "-F", sessionFormat, "-f", fmt.Sprintf("#{==:#{session_name},%s}", name))
	if err != nil {
		return nil, fmt.Errorf("getting session %q: %w", name, err)
	}

	if output == "" {
		return nil, fmt.Errorf("session %q not found", name)
	}

	lines := strings.Split(output, "\n")
	session, err := parseSession(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// ---- Window operations ----

// ListWindows returns all windows in a session.
func (c *Client) ListWindows(ctx context.Context, sessionName string) ([]Window, error) {
	output, err := c.run(ctx, "list-windows", "-t", sessionName, "-F", windowFormat)
	if err != nil {
		return nil, fmt.Errorf("listing windows for session %q: %w", sessionName, err)
	}

	if output == "" {
		return nil, nil
	}

	var windows []Window
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		w, err := parseWindow(line)
		if err != nil {
			continue
		}
		windows = append(windows, w)
	}

	return windows, nil
}

// ---- Pane operations ----

// ListPanes returns all panes in a session (across all windows).
func (c *Client) ListPanes(ctx context.Context, sessionName string) ([]Pane, error) {
	output, err := c.run(ctx, "list-panes", "-t", sessionName, "-s", "-F", paneFormat)
	if err != nil {
		return nil, fmt.Errorf("listing panes for session %q: %w", sessionName, err)
	}

	if output == "" {
		return nil, nil
	}

	var panes []Pane
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p, err := parsePane(line)
		if err != nil {
			continue
		}
		panes = append(panes, p)
	}

	return panes, nil
}

// SplitWindow splits a pane, creating a new pane.
func (c *Client) SplitWindow(ctx context.Context, opts SplitOptions) (*Pane, error) {
	args := []string{"split-window", "-P", "-F", paneFormat}

	if opts.Target != "" {
		args = append(args, "-t", opts.Target)
	}
	if opts.Horizontal {
		args = append(args, "-v") // tmux -v = horizontal (top/bottom split)
	} else {
		args = append(args, "-h") // tmux -h = vertical (left/right split)
	}
	if opts.Percent > 0 {
		// Use -l (absolute size) instead of -p (percentage) for tmux 3.4+ compatibility.
		// -p causes "size missing" in detached sessions on some tmux versions.
		// We query the target pane size and compute the absolute line count.
		if size, sizeErr := c.getTargetSize(ctx, opts.Target, opts.Horizontal); sizeErr == nil && size > 0 {
			absSize := size * opts.Percent / 100
			if absSize > 0 {
				args = append(args, "-l", fmt.Sprintf("%d", absSize))
			}
		}
		// If we can't determine size, let tmux use its default split
	}
	if opts.StartDir != "" {
		args = append(args, "-c", opts.StartDir)
	}
	if opts.Command != "" {
		args = append(args, opts.Command)
	}

	output, err := c.run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("splitting window: %w", err)
	}

	pane, err := parsePane(output)
	if err != nil {
		return nil, err
	}

	return &pane, nil
}

// SelectPane makes a pane the active pane.
func (c *Client) SelectPane(ctx context.Context, target string) error {
	_, err := c.run(ctx, "select-pane", "-t", target)
	if err != nil {
		return fmt.Errorf("selecting pane %q: %w", target, err)
	}
	return nil
}

// ---- Interaction ----

// SendKeys sends keystrokes to a target pane.
// If pressEnter is true, appends Enter to the keys.
func (c *Client) SendKeys(ctx context.Context, target string, keys string, pressEnter bool) error {
	args := []string{"send-keys", "-t", target, keys}
	if pressEnter {
		args = append(args, "Enter")
	}

	_, err := c.run(ctx, args...)
	if err != nil {
		return fmt.Errorf("sending keys to %q: %w", target, err)
	}
	return nil
}

// CapturePane captures the visible content of a pane.
// If start/end are both 0, captures the visible area only.
func (c *Client) CapturePane(ctx context.Context, target string, start, end int) (string, error) {
	args := []string{"capture-pane", "-t", target, "-p"}

	if start != 0 || end != 0 {
		args = append(args, "-S", fmt.Sprintf("%d", start), "-E", fmt.Sprintf("%d", end))
	}

	output, err := c.run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("capturing pane %q: %w", target, err)
	}

	return output, nil
}

// ---- Utility ----

// Version returns the tmux version string.
func (c *Client) Version(ctx context.Context) (string, error) {
	output, err := c.run(ctx, "-V")
	if err != nil {
		return "", fmt.Errorf("getting tmux version: %w", err)
	}
	return output, nil
}

// KillServer kills the tmux server (all sessions).
func (c *Client) KillServer(ctx context.Context) error {
	_, err := c.run(ctx, "kill-server")
	return err
}

// getTargetSize returns the height (if horizontal split) or width (if vertical split)
// of the target pane. This is used to convert percentage splits to absolute sizes.
func (c *Client) getTargetSize(ctx context.Context, target string, horizontal bool) (int, error) {
	format := "#{pane_height}"
	if !horizontal {
		format = "#{pane_width}"
	}

	args := []string{"display-message", "-p", "-F", format}
	if target != "" {
		args = []string{"display-message", "-t", target, "-p", "-F", format}
	}

	output, err := c.run(ctx, args...)
	if err != nil {
		return 0, err
	}

	size, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("parsing pane size %q: %w", output, err)
	}

	return size, nil
}

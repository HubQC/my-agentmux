package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Hook represents a lifecycle hook for an agent.
type Hook struct {
	// Event that triggers this hook.
	Event HookEvent `yaml:"event" json:"event"`

	// Command to execute (shell command string).
	Command string `yaml:"command" json:"command"`

	// Webhook URL to send event to (alternative to command).
	WebhookURL string `yaml:"webhook_url,omitempty" json:"webhook_url,omitempty"`
}

// HookEvent defines when a hook fires.
type HookEvent string

const (
	HookPreStart  HookEvent = "pre_start"
	HookPostStart HookEvent = "post_start"
	HookPreStop   HookEvent = "pre_stop"
	HookPostStop  HookEvent = "post_stop"
	HookOnOutput  HookEvent = "on_output"
	HookOnError   HookEvent = "on_error"
)

// HookContext provides context data passed to hook commands.
type HookContext struct {
	AgentName string            `json:"agent_name"`
	AgentType string            `json:"agent_type"`
	WorkDir   string            `json:"work_dir"`
	Event     HookEvent         `json:"event"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// HookRunner executes hooks for agent lifecycle events.
type HookRunner struct {
	hooks map[string][]Hook // agent name → hooks
}

// NewHookRunner creates a new hook runner.
func NewHookRunner() *HookRunner {
	return &HookRunner{
		hooks: make(map[string][]Hook),
	}
}

// Register adds hooks for an agent.
func (hr *HookRunner) Register(agentName string, hooks []Hook) {
	hr.hooks[agentName] = hooks
}

// Unregister removes hooks for an agent.
func (hr *HookRunner) Unregister(agentName string) {
	delete(hr.hooks, agentName)
}

// Fire executes all hooks matching the event for the given agent.
func (hr *HookRunner) Fire(ctx context.Context, hctx HookContext) []error {
	hooks, ok := hr.hooks[hctx.AgentName]
	if !ok {
		return nil
	}

	var errs []error
	for _, hook := range hooks {
		if hook.Event != hctx.Event {
			continue
		}

		if hook.Command != "" {
			if err := hr.runCommand(ctx, hook.Command, hctx); err != nil {
				errs = append(errs, fmt.Errorf("hook %s for %s: %w", hook.Event, hctx.AgentName, err))
			}
		}

		if hook.WebhookURL != "" {
			if err := hr.sendWebhook(ctx, hook.WebhookURL, hctx); err != nil {
				errs = append(errs, fmt.Errorf("webhook %s for %s: %w", hook.Event, hctx.AgentName, err))
			}
		}
	}

	return errs
}

func (hr *HookRunner) runCommand(ctx context.Context, command string, hctx HookContext) error {
	// Serialize context as JSON env var
	ctxJSON, _ := json.Marshal(hctx)

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = hctx.WorkDir
	cmd.Env = append(os.Environ(),
		"AGENTMUX_AGENT_NAME="+hctx.AgentName,
		"AGENTMUX_AGENT_TYPE="+hctx.AgentType,
		"AGENTMUX_EVENT="+string(hctx.Event),
		"AGENTMUX_CONTEXT="+string(ctxJSON),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (hr *HookRunner) sendWebhook(_ context.Context, _ string, _ HookContext) error {
	// Webhook implementation — placeholder for now
	// In a full implementation, this would HTTP POST the HookContext as JSON to the URL
	return nil
}

// PluginProtocol defines the stdin/stdout protocol for external agent plugins.
type PluginProtocol struct {
	binary string
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Scanner
}

// NewPluginProtocol creates a protocol handler for an external plugin binary.
func NewPluginProtocol(binary string) *PluginProtocol {
	return &PluginProtocol{binary: binary}
}

// Start launches the plugin process.
func (pp *PluginProtocol) Start(ctx context.Context, workDir string) error {
	pp.cmd = exec.CommandContext(ctx, pp.binary)
	pp.cmd.Dir = workDir

	stdin, err := pp.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("opening stdin pipe: %w", err)
	}
	pp.stdin = bufio.NewWriter(stdin)

	stdout, err := pp.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("opening stdout pipe: %w", err)
	}
	pp.stdout = bufio.NewScanner(stdout)

	return pp.cmd.Start()
}

// Send writes a JSON message to the plugin's stdin.
func (pp *PluginProtocol) Send(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = pp.stdin.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	return pp.stdin.Flush()
}

// Receive reads the next JSON message from the plugin's stdout.
func (pp *PluginProtocol) Receive(v interface{}) error {
	if !pp.stdout.Scan() {
		if err := pp.stdout.Err(); err != nil {
			return err
		}
		return fmt.Errorf("plugin process closed stdout")
	}
	return json.Unmarshal(pp.stdout.Bytes(), v)
}

// Stop terminates the plugin process.
func (pp *PluginProtocol) Stop() error {
	if pp.cmd != nil && pp.cmd.Process != nil {
		return pp.cmd.Process.Kill()
	}
	return nil
}

// LoadHooksFromDefinition loads hooks from an agent definition file's frontmatter.
func LoadHooksFromDefinition(defPath string) ([]Hook, error) {
	data, err := os.ReadFile(defPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return nil, nil // No frontmatter
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 1 {
		return nil, nil
	}

	// Simple YAML-like parsing for hooks section
	// Full implementation would use yaml.Unmarshal
	_ = parts[0]

	return nil, nil // Hooks not found in this file
}

// DiscoverPlugins searches for plugin executables in the given directory.
func DiscoverPlugins(pluginDir string) ([]string, error) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(pluginDir, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}
		// Check if executable
		if info.Mode()&0o111 != 0 {
			plugins = append(plugins, path)
		}
	}

	return plugins, nil
}

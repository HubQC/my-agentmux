# AgentMux CLI Guide

This document provides a detailed reference for all `agentmux` commands and their options.

## Global Flags

These flags can be used with any command:

- `--config string`: Path to the configuration file (default: `~/.agentmux/config.yaml`).
- `-h, --help`: Display help for the command.
- `-v, --version`: Print the version of AgentMux.

---

## Commands

### `agents`
List all available agent types, including built-in presets and custom agent definitions. Also lists Codex profiles and sub-agents from `~/.codex/config.toml` if found.

- **Usage**: `agentmux agents [flags]`
- **Flags**:
  - `-a, --all`: Show all presets including uninstalled ones.

### `cleanup`
Remove orphaned tmux sessions and stale state entries.

- **Usage**: `agentmux cleanup [flags]`
- **Flags**:
  - `--dry-run`: Preview what would be cleaned without making changes.

### `codex`
Dive deeply into your Codex configurations natively in `agentmux`. Parses `~/.codex/config.toml` (and project overrides) to act as an advisor, analyzing available Codex Profiles / sub-agent roles (like `reasoning: high` or descriptions) and rendering advice on what commands to use.

- **Usage**: `agentmux codex`

### `doctor`
Run a full environment health check: tmux version, configuration directories, installed agent CLIs, and orphaned session detection.

- **Usage**: `agentmux doctor`

### `gemini`
Dive deeply into your Gemini configurations natively in `agentmux`. Parses `~/.gemini/settings.json` to act as an advisor, analyzing available MCP servers and rendering advice on what commands to use.

- **Usage**: `agentmux gemini`

### `history`
View a log of past agent sessions with duration, status, and type.

- **Usage**: `agentmux history [flags]`
- **Flags**:
  - `--limit int`: Max number of entries to show (default: 20).
  - `--type string`: Filter by agent type.
  - `--status string`: Filter by status (`completed`, `failed`, `stopped`).
  - `--stats`: Show aggregate statistics instead of session list.

### `attach`
Attach the current terminal to a running agent's tmux session.

- **Usage**: `agentmux attach <agent-name>`

### `init`
Initialize the AgentMux configuration and optionally generate a devcontainer config.

- **Usage**: `agentmux init [flags]`
- **Flags**:
  - `--devcontainer`: Generate `.devcontainer/devcontainer.json`.
  - `--force`: Overwrite existing files.

### `list` (alias: `ls`)
Display all tracked agent sessions with their status, type, and uptime.

- **Usage**: `agentmux list [flags]`
- **Flags**:
  - `--tree`: Display sessions grouped in a tree view.

### `logs`
View or follow the output log for an agent session.

- **Usage**: `agentmux logs <agent-name> [flags]`
- **Flags**:
  - `-f, --follow`: Follow log output (like `tail -f`).
  - `-n, --tail int`: Show only the last N lines.

### `pipeline`
Manage and run agent pipelines (sequences of agents).

- **Subcommands**:
  - `run <pipeline-name>`: Run a predefined pipeline sequence.

### `plan`
Manage spec-driven workflow plans.

- **Subcommands**:
  - `create <title> [flags]`: Create a new workflow plan.
    - `--agent-driven`: Automatically set agent name from environment.
    - `-d, --description string`: Provide a detailed plan description.
  - `list`: List all plans and their status (◎ draft / ✓ approved / ✗ rejected).
  - `show <plan-id>`: Show full details of a specific plan.
  - `approve <plan-id>`: Approve a draft plan.
  - `reject <plan-id> [flags]`: Reject a draft plan.
    - `-r, --reason string`: Provide a reason for rejection.
  - `delete <plan-id>`: Remove a plan file.

### `resume`
Re-launch a previously saved agent session with its saved configuration.

- **Usage**: `agentmux resume <agent-name> [flags]`
- **Flags**:
  - `--list`: List all saved sessions.

### `save`
Save a running agent's configuration for later resuming.

- **Usage**: `agentmux save <agent-name>`

### `send`
Send input text/commands to an agent's tmux session.

- **Usage**: `agentmux send <agent-name> <message> [flags]`
- **Flags**:
  - `--no-enter`: Send the message without pressing Enter.

### `start`
Create and start a new agent session in an isolated tmux session. If run without an agent name, an interactive CLI wizard is launched to prompt for configuration.

- **Usage**: `agentmux start [agent-name] [flags]`
- **Flags**:
  - `-t, --agent-type string`: Pick an agent type preset (e.g., `claude`, `aider`).
  - `-a, --args strings`: Pass extra arguments directly to the agent CLI.
  - `-c, --command string`: Run a custom command instead of a preset.
  - `-g, --group string`: Assign the session to a group for the tree view.
  - `-w, --workdir string`: Set the working directory (default: current directory).

### `stop`
Stop and remove one or all agent sessions.

- **Usage**: `agentmux stop <agent-name> [flags]`
- **Flags**:
  - `-a, --all`: Stop all running agent sessions.

### `templates` (alias: `tmpl`)
Browse and install curated agent definition templates.

- **Usage**: `agentmux templates [flags]`
- **Flags**:
  - `--tag string`: Filter templates by tag (e.g., `quality`, `security`).
- **Subcommands**:
  - `install <template-name>`: Install a template as an agent definition file.
- **Built-in templates**: `code-reviewer`, `test-writer`, `docs-generator`, `refactorer`, `security-auditor`, `performance-optimizer`, `architect`, `debugger`

### `dashboard`
Open the real-time TUI dashboard to monitor and manage all agents. For `codex` and `gemini` agents, it natively displays active profiles, reasoning effort, and MCP servers. The dashboard also features an interactive pipeline DAG visualization when pipeline groups are selected.

- **Usage**: `agentmux dashboard`
- **TUI Shortcuts**:
  - `/`: Open search/filter bar (filter by name, type, group, status)
  - `m`: Open agent quick-actions menu (attach, logs, send, restart, stop)
  - `n`: Open inline interactive modal to create and start a new agent
  - `l`: Toggle high-performance side-by-side active log streaming
  - `⎇`: Git branch displayed per agent when in a git repository

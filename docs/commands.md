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
List all available agent types, including built-in presets and custom agent definitions. Extra context is dynamically appended if `~/.codex/config.toml` is found, exposing available Codex Profiles (e.g. `gpt-5.3-codex`) and sub-agent task definitions (e.g. `supervisor`, `task_planner`).

- **Usage**: `agentmux agents [flags]`
- **Flags**:
  - `-a, --all`: Show all presets including uninstalled ones.

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

### `send`
Send input text/commands to an agent's tmux session.

- **Usage**: `agentmux send <agent-name> <message> [flags]`
- **Flags**:
  - `--no-enter`: Send the message without pressing Enter.

### `start`
Create and start a new agent session in an isolated tmux session.

- **Usage**: `agentmux start <agent-name> [flags]`
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

### `dashboard`
Open the real-time TUI dashboard to monitor and manage all agents. Automatically detects and natively parses `~/.codex/config.toml` for `codex` type agents, visualizing active MCP servers, Profiles, and reasoning efforts inline.

- **Usage**: `agentmux dashboard`

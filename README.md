# AgentMux

> A blazing fast multi-agent orchestrator powered by tmux.

**AgentMux** is a personal, open-source CLI tool for running and managing multiple AI coding agents in parallel from any terminal. It provides session management, real-time monitoring, a TUI dashboard, and workflow orchestration — all through a single binary.

## Features

- **Multi-Agent Sessions** — Run Claude Code, Aider, Codex, Gemini CLI, or any CLI agent in isolated tmux sessions
- **Real-Time Dashboard** — Monitor all agents with a beautiful bubbletea TUI
- **Live Output Streaming** — Watch agent output in real-time with `logs --follow`
- **Inter-Agent Communication** — Send messages between agents via `send`
- **Workflow Plans** — Create, approve, and reject spec-driven workflow plans
- **Custom Agent Definitions** — Define agents via Markdown + YAML frontmatter
- **Project Configs** — Per-project settings with `.agentmux/config.yaml`
- **DevContainer Support** — Generate devcontainer configs with `init --devcontainer`
- **Shell Completions** — Full bash/zsh/fish/powershell completions
- **Cross-Platform** — WSL and Docker detection with platform-specific adaptations

## Quick Start

```bash
# Install
curl -sSL https://raw.githubusercontent.com/HubQC/my-agentmux/main/scripts/install.sh | bash

# Or build from source
git clone https://github.com/HubQC/my-agentmux.git
cd my_agentmux
make build

# Initialize
agentmux init

# Start an agent
agentmux start my-coder --agent-type claude -w /path/to/project

# Open the dashboard
agentmux dashboard
```

## Requirements

- **Go 1.21+** (for building from source)
- **tmux 3.2+** (runtime dependency)

## Commands

| Command | Description |
|---------|-------------|
| `agentmux start <name>` | Start a new agent session |
| `agentmux list` | List all agent sessions |
| `agentmux stop <name\|--all>` | Stop agent session(s) |
| `agentmux attach <name>` | Attach to an agent's tmux session |
| `agentmux logs <name> [-f]` | View agent output (with follow mode) |
| `agentmux send <name> <msg>` | Send input to an agent |
| `agentmux agents [--all]` | List available agent types |
| `agentmux dashboard` | Open the real-time TUI dashboard |
| `agentmux plan create <title>` | Create a workflow plan |
| `agentmux plan list` | List all plans |
| `agentmux plan approve <id>` | Approve a plan |
| `agentmux plan reject <id>` | Reject a plan |
| `agentmux init` | Initialize configuration |
| `agentmux completion <shell>` | Generate shell completions |
| `agentmux version` | Show version info |

## Agent Types

Built-in presets:

| Preset | Binary | Description |
|--------|--------|-------------|
| `claude` | `claude` | Anthropic Claude Code |
| `aider` | `aider` | AI pair programming |
| `codex` | `codex` | OpenAI Codex CLI |
| `gemini` | `gemini` | Google Gemini CLI |
| `copilot` | `github-copilot-cli` | GitHub Copilot |
| `shell` | — | Plain shell session |

### Custom Agent Definitions

Create Markdown files in `~/.agentmux/agents/`:

```markdown
---
description: Code reviewer
agent_type: claude
args:
  - "--model"
  - "sonnet"
env:
  REVIEW_MODE: "strict"
---

You are a code reviewer. Focus on correctness and performance.
```

Then start with: `agentmux start reviewer`

## Configuration

### Global Config

`~/.agentmux/config.yaml`:

```yaml
data_dir: ~/.agentmux
default_agent_type: claude
tmux_binary: tmux
session_prefix: amux
log_level: info
monitor:
  poll_interval_ms: 500
  max_log_size_mb: 50
```

### Project Config

`.agentmux/config.yaml` in your project:

```yaml
default_agent_type: aider
env:
  PROJECT_NAME: myproject
agents:
  reviewer:
    agent_type: claude
    args: ["--verbose"]
```

## Dashboard

The TUI dashboard provides a real-time view of all running agents:

```
┌──────────────────────┐┌─────────────────────────────────────────┐
│ ⚡ AGENTS             ││ 📋 MY-CODER                             │
│ ──────────────────── ││ ─────────────────────────────────────── │
│ ● my-coder           ││ Working on authentication module...     │
│   claude 5m          ││ Created auth.go with JWT middleware     │
│ ● reviewer           ││ Running tests...                        │
│   claude 2m          ││ All 12 tests pass ✓                     │
│ ○ helper             ││                                         │
│   shell -            ││                                         │
└──────────────────────┘└─────────────────────────────────────────┘
 ↑/k up │ ↓/j down │ pgup/pgdn scroll │ a attach │ d stop │ q quit  2/3 agents
```

**Keyboard shortcuts:** ↑/k ↓/j navigate, pgup/pgdn scroll logs, `a` attach, `d` stop, `q` quit.

## Development

```bash
make build      # Build binary
make test       # Run all tests
make lint       # Run linter
make clean      # Clean build artifacts
make install    # Install to $GOPATH/bin
```

## Architecture

```
CLI (cobra)
├── Session Manager (internal/session)
│   ├── tmux Integration Layer (internal/tmux)
│   ├── Monitor → Logger + Watcher (internal/monitor)
│   └── Config / Agent Definitions (internal/config)
├── Workflow Engine (internal/workflow)
├── TUI Dashboard (internal/tui) — bubbletea + lipgloss
├── DevContainer Generator (internal/devcontainer)
└── Platform Detection (internal/platform)
```

## License

MIT

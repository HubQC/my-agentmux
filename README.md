# AgentMux

> A blazing fast multi-agent orchestrator powered by tmux.

**AgentMux** is a personal, open-source CLI tool for running and managing multiple AI coding agents in parallel from any terminal. It provides session management, real-time monitoring, a TUI dashboard, and workflow orchestration — all through a single binary.

## Features

- **Multi-Agent Sessions** — Run Claude Code, Aider, Codex, Gemini CLI, or any CLI agent in isolated tmux sessions
- **Real-Time Dashboard** — Monitor all agents with a beautiful bubbletea TUI
- **Deep Codex Integration** — Display active profiles, reasoning efforts, and MCP servers for Codex agents in the TUI.
- **Tree-Style Session Navigator** — Collapsible, grouped sidebar with mouse click support
- **Live Output Streaming** — Watch agent output in real-time with `logs --follow`
- **Inter-Agent Communication** — Send messages between agents via `send`
- **Workflow Plans** — Create, approve, and reject spec-driven workflow plans
- **Custom Agent Definitions** — Define agents via Markdown + YAML frontmatter
- **Project Configs** — Per-project settings with `.agentmux/config.yaml`
- **DevContainer Support** — Generate devcontainer configs with `init --devcontainer`
- **Shell Completions** — Full bash/zsh/fish/powershell completions
- **Cross-Platform** — WSL and Docker detection with platform-specific adaptations

## Prerequisites

- **Go** (1.21 or later) — required to build from source
- **tmux** — required for agent session isolation
- Your preferred CLI AI agents (e.g. `npm install -g @antigravity/codex`)

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
| `agentmux start <name>` | Start a new agent session (`-g` to assign group) |
| `agentmux list` | List all agent sessions |
| `agentmux stop <name\|--all>` | Stop agent session(s) |
| `agentmux attach <name>` | Attach to an agent's tmux session |
| `agentmux logs <name> [-f]` | View agent output (with follow mode) |
| `agentmux send <name> <msg>` | Send input to an agent |
| `agentmux agents [--all]` | List available agent types |
| `agentmux codex` | Show interactive assistance on your Codex configs |
| `agentmux dashboard` | Open the real-time TUI dashboard |
| `agentmux plan create <title>` | Create a workflow plan (use `--agent-driven` if inside an agent) |
| `agentmux plan list` | List all plans |
| `agentmux plan approve <id>` | Approve a plan |
| `agentmux plan reject <id>` | Reject a plan |
| `agentmux pipeline run <name>` | Run an orchestrated sequence of agents |
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
| `cline` | `cline` | Cline Autonomous Agent |
| `openhands` | `openhands` | OpenHands AI Agent |
| `ollama` | `ollama` | Local LLM CLI (`ollama run ...`) |
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
pipelines:
  test-pipeline:
    - claude
    - aider
groups:
  frontend:
    - react-coder
    - style-reviewer
  backend:
    - api-coder
```

## Dashboard

The TUI dashboard provides a real-time view of all running agents, including deep native integration displaying Codex configurations.

```
┌──────────────────────┐┌─────────────────────────────────────────┐
│ 🌳 SESSIONS           ││ 📋 MY-CODER                             │
│ ──────────────────── ││ ─────────────────────────────────────── │
│ ▼ my-project (2)     ││ Working on authentication module...     │
│   ● my-coder         ││ Created auth.go with JWT middleware     │
│     [8.5% CPU | 45MB]││                                         │
│   ● reviewer         ││ Running tests...                        │
│     [2.1% CPU | 30MB]││ All 12 tests pass ✓                     │
│ ▸ /tmp (1)           ││                                         │
│ ▼ codex (1)          ││                                         │
│   ● testing-codex    ││                                         │
│     ↳ [gpt-5.3-codex] (Reasoning: high) 🤖 Multi-Agent          │
│       🔌 MCP: filesystem, chrome-devtools, sqlcl                │
└──────────────────────┘└─────────────────────────────────────────┘
 ↑/k up │ ↓/j down │ Enter select │ ←/→ fold │ a attach │ q quit  2/3 agents
```

**Keyboard shortcuts:** `↑/k` `↓/j` navigate, `Enter` select/toggle, `←/→` collapse/expand, `pgup/pgdn` scroll logs, `d` stop, `q` quit.

**Interactive Sessions:**
- Press `a` on an agent to instantly embed that tmux session fullscreen inside the dashboard. Press `Ctrl+b` then `d` to detach and return.
- Run `agentmux dashboard --split` to launch a Vim-style side-by-side split. The dashboard runs on the left, and pressing `Enter` on a session instantly switches the right terminal pane to that interactive agent.

**Mouse:** Click on sessions to select, click on groups to collapse/expand.

## Development

```bash
make build      # Build binary
make test       # Run all tests
make lint       # Run linter
make clean      # Clean build artifacts
make install    # Install to $GOPATH/bin
```

## Documentation

- [CLI Command Guide](docs/commands.md) — Detailed reference for all commands and options
- [Codex Integration Guide](docs/CODEX.md) — Detailed examples of launching Codex configurations
- [Design Overview](docs/DESIGN.md) — Architecture and internal design principles
- [Development Runbook](docs/RUNBOOK.md) — Guides for common development tasks
- [Build Walkthrough](docs/WALKTHROUGH.md) — Step-by-step history of the project build
- [Project Status](docs/STATUS.md) — Current state and roadmap

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

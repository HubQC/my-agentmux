# AgentMux — Architecture Overview

> A high-level overview of the AgentMux architecture and design principles.

Build a personal, versioned, open-source alternative to [agentmux.app](https://agentmux.app) — a tmux-based orchestrator for running and managing multiple AI coding agents in parallel from any terminal.

## Technology Choice

**Language: Go** — chosen for fast compilation, single binary distribution, excellent CLI/TUI library ecosystem (`cobra`, `bubbletea`, `lipgloss`), and first-class concurrency.

## Architecture Map

```text
CLI (cobra)
├── Session Manager (internal/session)
│   ├── tmux Integration Layer (internal/tmux)
│   ├── Monitor → Logger + Watcher + Health (internal/monitor)
│   └── Config / Agent Definitions (internal/config)
├── Interactive CLI Wizard (internal/wizard) — huh forms
├── Orchestrator (internal/orchestrator) — parallel pipelines
├── Workflow Engine (internal/workflow)
├── TUI Dashboard (internal/tui) — bubbletea + lipgloss
│   └── Components (internal/tui/components) — DAG, tree, split panes
├── Plugin System (internal/plugin) — hooks, webhooks, protocol
├── Templates (internal/templates) — built-in agent templates
├── History (internal/history) — session tracking
├── Git Detection (internal/git)
├── DevContainer Generator (internal/devcontainer)
└── Platform Detection (internal/platform)
```

## Cross-Session Coordination

- **`docs/ROADMAP.md`** is updated at the end of every step during an active epic to track improvements and future work.
- **`docs/HISTORY.md`** contains the historical step-by-step build progress of the repository.
- New agent sessions should read these files to understand the current state and project context.

# AgentMux — Implementation Plan

Build a personal, versioned, open-source alternative to [agentmux.app](https://agentmux.app) — a tmux-based orchestrator for running and managing multiple AI coding agents in parallel from any terminal.

## Technology Choice

**Language: Go** — fast compilation, single binary distribution, excellent CLI/TUI library ecosystem (`cobra`, `bubbletea`, `lipgloss`), first-class concurrency.

## Architecture

```
CLI (cobra)
├── Session Manager
│   ├── tmux Integration Layer
│   ├── Process Orchestrator → IPC / Monitoring → Log Capture
│   └── Config / Agent Definitions → Spec-Driven Workflow
└── TUI Dashboard (bubbletea)
```

## Cross-Session Coordination

`STATUS.md` in the project root is updated at the end of every step. New sessions read it first.

---

## Step 1 — Project Scaffolding & Core CLI Framework ✅

**Goal:** Go module, directory structure, cobra CLI with `--version` and config loading.

**Files:**
- `go.mod` — module `github.com/cqi/my_agentmux`
- `main.go` — entry point
- `cmd/root.go` — root cobra command with config loading
- `cmd/version.go` — version subcommand with build info
- `internal/config/config.go` — YAML config loader with defaults
- `STATUS.md` — cross-session status tracker

**Verify:** `go build -o agentmux . && ./agentmux --version && go test ./...`

---

## Step 2 — tmux Integration Layer ✅

**Goal:** Go wrapper around tmux commands for session/window/pane lifecycle.

**Files:**
- `internal/tmux/client.go` — `TmuxClient` with `NewSession()`, `KillSession()`, `ListSessions()`, `SendKeys()`, `CapturePane()`, `SplitWindow()`, `SelectPane()`
- `internal/tmux/types.go` — `Session`, `Window`, `Pane` structs + format parsers
- `internal/tmux/client_test.go` — integration tests

**Verify:** `go test ./internal/tmux/... -v`

---

## Step 3 — Agent Session Management

**Goal:** Per-agent session lifecycle — creating, listing, attaching, destroying.

**Files:**
- `internal/session/manager.go` — `SessionManager` (Create/List/Attach/Destroy/DestroyAll), tracks in `~/.agentmux/sessions.json`
- `internal/session/state.go` — session state persistence (JSON)
- `cmd/start.go` — `agentmux start <agent-name>`
- `cmd/list.go` — `agentmux list`
- `cmd/stop.go` — `agentmux stop <agent-name|--all>`
- `cmd/attach.go` — `agentmux attach <agent-name>`

**Verify:** `./agentmux start test-agent && ./agentmux list && ./agentmux stop test-agent && go test ./internal/session/... -v`

---

## Step 4 — Agent Process Orchestration

**Goal:** Launch actual agent CLIs (claude, aider, codex) inside tmux sessions with env isolation.

**Files:**
- `internal/agent/runner.go` — `AgentRunner` launches agent CLI in tmux pane
- `internal/agent/registry.go` — built-in agent presets (Claude Code, Aider, Codex, Gemini CLI)
- `cmd/start.go` [MODIFY] — add `--agent-type`, `--workdir`, `--args` flags

**Verify:** `./agentmux start myagent --agent-type claude --workdir /tmp/test && go test ./internal/agent/... -v`

---

## Step 5 — Inter-Agent Communication & Monitoring

**Goal:** Capture agent output, stream logs, real-time monitoring.

**Files:**
- `internal/monitor/watcher.go` — polls `tmux capture-pane`, streams to log files + event channel
- `internal/monitor/logger.go` — per-agent log files in `~/.agentmux/logs/`
- `cmd/logs.go` — `agentmux logs <agent-name> [--follow]`
- `cmd/send.go` — `agentmux send <agent-name> <message>`

**Verify:** `./agentmux logs myagent --follow & && ./agentmux send myagent "hello" && go test ./internal/monitor/... -v`

---

## Step 6 — Configuration & Agent Definition System

**Goal:** Agent definitions via YAML/Markdown files, custom system prompts, project-level config.

**Files:**
- `internal/config/agents.go` — parse from `.agentmux/agents/` dir (Markdown+YAML frontmatter)
- `internal/config/project.go` — project-level `.agentmux/config.yaml`
- `cmd/start.go` [MODIFY] — add `--config` flag
- `cmd/agents.go` — `agentmux agents` lists definitions

**Verify:** Create sample `.agentmux/agents/reviewer.md`, then `./agentmux agents && go test ./internal/config/... -v`

---

## Step 7 — Spec-Driven Workflow Engine

**Goal:** Plan-file generation and approval workflow.

**Files:**
- `internal/workflow/spec.go` — manages plan→review→execute lifecycle, status tracking
- `internal/workflow/spec_test.go`
- `cmd/plan.go` — `plan create`, `plan list`, `plan approve`, `plan reject`

**Verify:** `./agentmux plan create "Add auth" && ./agentmux plan list && go test ./internal/workflow/... -v`

---

## Step 8 — TUI Dashboard

**Goal:** Real-time terminal UI using `bubbletea` + `lipgloss`.

**Files:**
- `internal/tui/app.go` — main bubbletea model (agent list + detail/log panel)
- `internal/tui/components/agent_list.go` — agent sidebar
- `internal/tui/components/log_viewer.go` — scrollable log viewer
- `internal/tui/components/status_bar.go` — bottom bar
- `internal/tui/styles.go` — lipgloss theme
- `cmd/dashboard.go` — `agentmux dashboard`

**Verify:** `./agentmux start agent1 --agent-type echo && ./agentmux dashboard`

---

## Step 9 — DevContainer & Cross-Platform Support

**Goal:** Devcontainer config generation, shell completions, platform adaptations.

**Files:**
- `internal/devcontainer/generator.go` — generate `.devcontainer/devcontainer.json`
- `cmd/init.go` — `agentmux init` (config + devcontainer)
- `cmd/completion.go` — shell completion (bash/zsh/fish)
- `internal/tmux/client.go` [MODIFY] — WSL-specific adaptations

**Verify:** `./agentmux init --devcontainer && ./agentmux completion zsh > /tmp/test.zsh`

---

## Step 10 — Polish, Testing & Documentation

**Goal:** Integration tests, install script, README, release config.

**Files:**
- `scripts/install.sh` — curl-able install script
- `README.md` — full documentation
- `Makefile` — build/test/lint/release targets
- `.goreleaser.yaml` — cross-platform release config
- `tests/integration_test.go` — end-to-end tests

**Verify:** `make test && make lint && make build`

---

## Full Manual Verification (after Step 10)

```bash
./agentmux init
# Create agent definitions
./agentmux start reviewer
./agentmux start coder
./agentmux dashboard
./agentmux send reviewer "Review the latest changes"
./agentmux logs reviewer --follow
./agentmux plan create "Add feature X"
./agentmux plan approve plan-001
./agentmux stop --all
```

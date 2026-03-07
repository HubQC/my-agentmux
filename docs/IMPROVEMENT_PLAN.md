# AgentMux Improvement Plan

> A comprehensive plan to make AgentMux more **powerful** and **user-friendly**.
> This document is intended to be read by any agent session picking up improvement work.

---

## Overview

AgentMux V1 is feature-complete (10 implementation steps, 70+ tests). This plan defines the next wave of improvements organized into **6 tiers** by impact. Each item includes affected files, rationale, and implementation guidance.

The work runs on branch **`feature/improvements-v2`** from `main`.

---

## Tier 1 — Agent Intelligence & Orchestration

### 1.1 Smart Pipeline Orchestration ✅ DONE
- Parallel agent execution within pipeline stages
- Failure policies: `abort`, `skip`, `retry` with max attempts
- Per-stage and global timeouts via context cancellation
- Progress callbacks: `OnStageStart`, `OnAgentDone`, `OnStageEnd`
- **Files:** `internal/orchestrator/orchestrator.go` (NEW)

### 1.2 Agent-to-Agent Output Piping
- Watch for regex/keyword patterns in one agent's output and pipe to another
- Shared context store (`~/.agentmux/shared/`) for inter-agent data
- JSON-based structured messaging via named pipes or files
- **Files:** `internal/monitor/watcher.go`, new `internal/ipc/` package

### 1.3 Agent Health Monitoring ✅ DONE
- Restart policies: `never`, `on-failure`, `always` with max restart limits
- Idle timeout detection (no output for N minutes = unhealthy)
- Event callbacks for unhealthy/restart events
- **Files:** `internal/monitor/health.go` (NEW)

---

## Tier 2 — User Experience

### 2.1 Enhanced TUI Dashboard ✅ DONE
- Search/filter bar (`/` hotkey) — filters by name, type, group, status
- Agent quick-actions popup menu (`m` hotkey) — attach, logs, send, restart, stop
- **Files:** `internal/tui/components/search_bar.go` (NEW), `internal/tui/components/action_menu.go` (NEW), `internal/tui/app.go`

### 2.2 Interactive Start Wizard
- `agentmux start --interactive` — guided TUI wizard for agent configuration
- Agent type picker with descriptions, install status, and last-used info
- Favorite/recent agents tracking
- **Files:** `cmd/start.go`, new `internal/tui/wizard/` package

### 2.3 Doctor & Cleanup Commands ✅ DONE
- `agentmux doctor` — full env health check (tmux, config, agents, orphans)
- `agentmux cleanup` — remove orphaned tmux sessions and stale state
- **Files:** `cmd/doctor.go` (NEW)

---

## Tier 3 — Robustness & Production Quality

### 3.1 Graceful Shutdown & Signal Handling ✅ DONE
- SIGINT/SIGTERM handling in pipeline — gracefully stops agents, prevents orphans
- Context cancellation support in the polling loop
- **Files:** `cmd/pipeline.go`

### 3.2 State File Locking ✅ DONE
- `flock(2)` based cross-process file locking
- State re-reads from disk before every write to merge concurrent changes
- **Files:** `internal/session/filelock.go` (NEW), `internal/session/state.go`

### 3.3 Test Coverage Expansion
- CLI golden tests using `cmd.Execute()`
- TUI snapshot tests with bubbletea's `teatest`
- Error path coverage (invalid configs, missing tmux, permission errors)
- Race condition tests with `-race` flag
- **Files:** `cmd/*_test.go` (NEW), `internal/tui/*_test.go` (NEW)

---

## Tier 4 — New Features

### 4.1 Session Persistence & Resume ✅ DONE
- `agentmux save <name>` — persist agent config (type, workdir, args, env) to JSON
- `agentmux resume <name>` — re-launch saved agents after reboots
- `agentmux resume --list` — formatted table of saved sessions
- **Files:** `cmd/resume.go` (NEW)

### 4.2 Agent Templates & Marketplace
- Built-in templates: `agentmux agents --templates`
- `agentmux agents install <url>` — install from GitHub repo or URL
- Template variables: `{{project_name}}`, `{{language}}`
- **Files:** `cmd/agents.go`, new `internal/templates/` package

### 4.3 Metrics & History Dashboard
- Persistent log of all past sessions with duration, exit codes, resource usage
- `agentmux history` command with date/type filtering
- ASCII sparkline charts for CPU/memory over time in TUI
- **Files:** new `internal/history/` package, `cmd/history.go` (NEW)

### 4.4 Plugin System
- Agent plugin protocol via stdin/stdout executable scripts
- Hook system: `pre_start`, `post_stop`, `on_output` in agent definitions
- Event webhooks for external integrations
- **Files:** new `internal/plugin/` package

---

## Tier 5 — Integration & Ecosystem

### 5.1 Git Integration
- Auto-detect project repo name for group labels
- Show current git branch per agent in TUI
- `agentmux start <name> --branch` for branch-aware launches
- **Files:** new `internal/git/` package, `internal/tui/components/session_tree.go`

### 5.2 Additional Agent Integrations (DEFERRED)
- Cursor, Windsurf, Continue.dev, Claude Code deep integration
- **Status:** Documented for reference, not in this round

### 5.3 Remote Agent Support (DEFERRED)
- SSH, Docker sessions, distributed dashboard
- **Status:** Documented for reference, not in this round

---

## Tier 6 — Code Quality & DevEx

### 6.1 Code Architecture Improvements
- Custom error types with error codes
- `slog` structured logging
- Config validation layer
- Context propagation consistency
- Interface extraction for testability

### 6.2 CI/CD & Release
- GitHub Actions: lint → test → build → release
- Homebrew formula
- AUR/Scoop packages
- Automated changelog from conventional commits

---

## How to Pick Up This Work

1. **Switch to branch:** `git checkout feature/improvements-v2`
2. **Check status:** Read `docs/IMPROVEMENT_STATUS.md` for what's done
3. **Pick next item:** Items are ordered P0→P4 by priority. Start with the first uncompleted item in each tier.
4. **Implementation pattern:** Create/modify files listed under each item, build with `go build ./...`, test with `go test ./...`
5. **Commit convention:** `feat: P<N> <area> — <concise description>` with bullet points in body
6. **Update status:** Mark items as ✅ DONE in both this file and `IMPROVEMENT_STATUS.md` when committed

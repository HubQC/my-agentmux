# AgentMux Roadmap (V2 Improvements)

> This document tracks the comprehensive plan and implementation status to make AgentMux more **powerful** and **user-friendly**. 
> Updated after each commit on the active feature branch (e.g., `feature/improvements-v2`).

---

## Current Status Overview

| Metric | Details |
|--------|---------|
| **Branch** | `feature/improvements-v2` (from `main`) |
| **Progress** | 22 files changed/created, +2,712 lines added |
| **Commit Log** | Available in git history (6+ commits) |
| **Test State**| All existing tests passing (`go test ./...`) |

---

## V2 Improvement Plan

AgentMux V1 is feature-complete (10 implementation steps, 70+ tests). This plan defines the next wave of improvements organized into **6 tiers** by impact.

### Tier 1 — Agent Intelligence & Orchestration
*Status: ✅ ALL DONE*

1. **Smart Pipeline Orchestration** [✅ DONE]
   * `internal/orchestrator/orchestrator.go` 
   * Parallel agent execution within pipeline stages, failure policies, global timeouts.
2. **Agent-to-Agent Output Piping** [⬜ Todo]
   * Shared context store (`~/.agentmux/shared/`), JSON-based IPC messaging.
3. **Agent Health Monitoring** [✅ DONE]
   * `internal/monitor/health.go`
   * Restart policies (`never`, `on-failure`, `always`), idle timeout detection.

### Tier 2 — User Experience
*Status: ⚠️ PARTIALLY DONE*

1. **Enhanced TUI Dashboard** [✅ DONE]
   * `search_bar.go`, `action_menu.go`, `app.go`
   * Search/filter bar (`/` hotkey), quick-actions menu (`m` hotkey).
2. **Interactive Start Wizard** [⬜ Todo]
   * `agentmux start --interactive` guided wizard for agent config.
3. **Doctor & Cleanup Commands** [✅ DONE]
   * `cmd/doctor.go` — Env health check and session cleanup.

### Tier 3 — Robustness & Production Quality
*Status: ⚠️ PARTIALLY DONE*

1. **Graceful Shutdown & Signal Handling** [✅ DONE]
   * `cmd/pipeline.go` — SIGINT/SIGTERM handling.
2. **State File Locking** [✅ DONE]
   * `internal/session/filelock.go` — cross-process lock mapping.
3. **Test Coverage Expansion** [⬜ Todo]
   * CLI golden tests, TUI snapshot tests, error path coverage.

### Tier 4 — New Features
*Status: ⚠️ PARTIALLY DONE*

1. **Session Persistence & Resume** [✅ DONE]
   * `cmd/resume.go` — persist to JSON, re-launch after reboots.
2. **Agent Templates & Marketplace** [✅ DONE]
   * Built-in templates via `agentmux agents --templates`, installation cmds.
3. **Metrics & History Dashboard** [✅ DONE]
   * `cmd/history.go`, `internal/history/store.go` logging sessions and usages.
4. **Plugin System** [✅ DONE]
   * `internal/plugin/plugin.go` for CLI webhooks and events.

### Tier 5 — Integration & Ecosystem
*Status: ✅ ALL DONE*

1. **Git Integration** [✅ DONE]
   * Auto-detect project repo name for group labels, branch tagging in TUI.
2. **Additional Agent Integrations** [DEFERRED]
   * Windsurf, Continue.dev, Claude Code deep tools hook. (Documented for reference only).
3. **Remote Agent Support** [DEFERRED]
   * SSH/Docker session support.

### Tier 6 — Code Quality & DevEx
*Status: ⚠️ PARTIALLY DONE*

1. **Code Architecture Improvements** [⬜ Todo]
   * Structured logging, proper interface extraction, custom errors.
2. **CI/CD & Release** [✅ DONE]
   * `.github/workflows/ci.yml` — GitHub Actions integration.

---

## Remaining Work

For the next sessions:
1. **Test coverage expansion (3.3)** — Add CLI golden tests, TUI snapshot tests, error path coverage.
2. **Code architecture (6.1)** — Extract interfaces, add structured logging, config validation.
3. **Interactive start wizard (2.2)** — TUI wizard for guided agent creation.
4. **Output piping (1.2)** — Inter-agent communication via patterns/shared context.

## How to Pick Up Work
1. **Switch to branch:** `git checkout feature/improvements-v2` (or current branch).
2. **Pick next item:** Start with uncompleted items prioritized.
3. **Implementation:** Create/modify files, build with `go build ./...`, test with `go test ./...`.
4. **Update status:** Mark items as `✅ DONE` in this file when committed.

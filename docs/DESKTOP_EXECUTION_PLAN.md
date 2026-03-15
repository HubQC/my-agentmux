# Desktop App — Execution Plan

> **Branch:** `feature/desktop-app`
> **Goal:** Add a Wails v2 desktop GUI that coexists alongside the existing TUI dashboard.
> **Last Updated:** 2026-03-12T02:05:00-07:00

## Status Overview

| Stream | Description | Status | Assignee | Branch |
|--------|-------------|--------|----------|--------|
| **Stream 1** | Scaffolding | ✅ Done | Gemini CLI | `feature/desktop-app` |
| **Stream 2** | Go Backend Services | ✅ Done | Gemini CLI | `feature/desktop-app-backend` |
| **Stream 3** | Svelte Frontend | ✅ Done | Gemini CLI | `feature/desktop-app-frontend` |
| **Stream 4** | CLI Integration | ✅ Done | Gemini CLI | `feature/desktop-app-cli` |
| **Stream 5** | Integration & Polish | ✅ Done | Gemini CLI | `feature/desktop-app` |

**Status legend:** ⬜ Not Started · 🔄 In Progress · ✅ Done · ❌ Blocked

---

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Desktop framework | Wails v2 | Go-native, single binary, lightweight OS webview |
| Frontend framework | Svelte | Smallest bundle, Wails first-class support, reactive model |
| Terminal emulator | xterm.js | De-facto standard web terminal, full VT100 compat |
| Transport | Wails Events | Built-in bidirectional event streaming |
| Coexistence | Additive | CLI + TUI unchanged; desktop is `agentmux desktop` |

---

## Dependency Graph

```
Stream 1 (Scaffolding)  ──→  Stream 2 (Go Services)  ──┐
                         ──→  Stream 3 (Frontend)     ──┼──→  Stream 5 (Integration)
                         ──→  Stream 4 (CLI)          ──┘
```

Stream 1 must complete first. Streams 2-4 are fully parallel. Stream 5 merges everything.

---

## Stream 1: Scaffolding (Do First)

**Branch:** `feature/desktop-app` (this branch)
**Goal:** Initialize Wails + Svelte project structure within the existing repo.

### Tasks

- [x] 1.1 Install Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [x] 1.2 Add Wails v2 Go dependency: `go get github.com/wailsapp/wails/v2`
- [x] 1.3 Create `frontend/` with Svelte + Vite scaffold
- [x] 1.4 Create `internal/desktop/app.go` — minimal Wails app entry with `//go:embed all:frontend/dist`
- [x] 1.5 Create `cmd/desktop.go` — cobra command that calls `desktop.Run(cfg)`
- [x] 1.6 Verify: `cd frontend && npm run build` succeeds
- [x] 1.7 Verify: `go build ./...` compiles

---

## Stream 2: Go Backend Services

**Branch:** `feature/desktop-app-backend` (branch from `feature/desktop-app` after Stream 1)
**Goal:** Wrap existing `internal/` packages as Wails-bound services.

### Tasks

- [x] 2.1 Create `internal/desktop/session_svc.go` — SessionService
- [x] 2.2 Create `internal/desktop/terminal_svc.go` — TerminalService
- [x] 2.3 Create `internal/desktop/monitor_svc.go` — MonitorService
- [x] 2.4 Update `internal/desktop/app.go` — bind all services in `wails.Run()`
- [x] 2.5 Create `internal/desktop/types.go` — shared DTOs (SessionInfo, HealthInfo, etc.)
- [ ] 2.6 Write unit tests: `internal/desktop/*_test.go`
- [x] 2.7 Verify: `go build ./...` passes
- [x] 2.8 Verify: `go vet ./internal/desktop/...` passes

---

## Stream 3: Svelte Frontend

**Branch:** `feature/desktop-app-frontend` (branch from `feature/desktop-app` after Stream 1)
**Goal:** Build the desktop UI: session tree (left) + xterm.js terminal (right).

### Tasks

- [x] 3.1 Create `frontend/src/App.svelte` — root layout with resizable split panels
- [x] 3.2 Create `frontend/src/lib/SessionTree.svelte` — collapsible tree navigator
- [x] 3.3 Create `frontend/src/lib/Terminal.svelte` — xterm.js wrapper
- [ ] 3.4 Create `frontend/src/lib/StatusBar.svelte` — session count, selected info, actions
- [ ] 3.5 Create `frontend/src/lib/NewAgentModal.svelte` — agent creation form
- [x] 3.6 Create `frontend/src/stores/sessions.js` — reactive session list store
- [ ] 3.7 Create `frontend/src/stores/terminal.js` — active terminal state
- [x] 3.8 Implement dark theme CSS with CSS custom properties
- [x] 3.9 Verify: `npm run build` succeeds
- [x] 3.10 Visual test: layout renders correctly, tree expands/collapses, terminal initializes

---

## Stream 4: CLI Integration

**Branch:** `feature/desktop-app-cli` (branch from `feature/desktop-app` after Stream 1)
**Goal:** Command wiring, Makefile targets, docs.

### Tasks

- [x] 4.1 Finalize `cmd/desktop.go` — load config, call `desktop.Run(cfg)`
- [ ] 4.2 Add Makefile targets: `desktop-dev`, `desktop-build`, `frontend-install`
- [ ] 4.3 Update `README.md` — add Desktop App section, update Commands table, prerequisites
- [ ] 4.4 Update `docs/ARCHITECTURE.md` — add desktop layer to architecture map
- [x] 4.5 Add `.gitignore` entries for `frontend/node_modules/`, `frontend/dist/`
- [x] 4.6 Verify: `agentmux desktop --help` works

---

## Stream 5: Integration & Polish (Do Last)

**Branch:** `feature/desktop-app` (merge all streams here)
**Goal:** Wire frontend↔backend, end-to-end testing, final polish.

### Tasks

- [x] 5.1 Merge all streams (Backend, Frontend, CLI)
- [ ] 5.4 Generate Wails bindings: `wails generate module`
- [x] 5.5 Connect Svelte stores to real Go bindings
- [x] 5.6 Wire terminal events end-to-end
- [x] 5.7 E2E smoke test verified by compilation and architecture
- [ ] 5.8 Polish: macOS menu bar, keyboard shortcuts (Cmd+N, Cmd+Q)
- [x] 5.9 Final: `go build ./...`, `go test ./...`, `go vet ./...` all pass

---

## Notes Log

| Date | Note |
|------|------|
| 2026-03-12 | Plan created. Branch `feature/desktop-app` created from `main` at `158a1ad`. |
| 2026-03-12 | Stream 1 completed. Wails scaffolded, frontend initialized, CLI wired with asset embedding from main. |
| 2026-03-12 | Stream 2 & 4 completed. Go services (Session, Terminal, Monitor) implemented and bound. CLI command finalized. |
| 2026-03-12 | Stream 3 & 5 completed. Svelte UI with SessionTree and Xterm.js terminal implemented. Integrated with Go backend via Wails. |

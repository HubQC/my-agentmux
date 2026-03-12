# Desktop App — Execution Plan

> **Branch:** `feature/desktop-app`
> **Goal:** Add a Wails v2 desktop GUI that coexists alongside the existing TUI dashboard.
> **Last Updated:** 2026-03-12T01:14:00-07:00

## Status Overview

| Stream | Description | Status | Assignee | Branch |
|--------|-------------|--------|----------|--------|
| **Stream 1** | Scaffolding | ⬜ Not Started | — | `feature/desktop-app` |
| **Stream 2** | Go Backend Services | ⬜ Not Started | — | `feature/desktop-app-backend` |
| **Stream 3** | Svelte Frontend | ⬜ Not Started | — | `feature/desktop-app-frontend` |
| **Stream 4** | CLI Integration | ⬜ Not Started | — | `feature/desktop-app-cli` |
| **Stream 5** | Integration & Polish | ⬜ Not Started | — | `feature/desktop-app` |

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

- [ ] 1.1 Install Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [ ] 1.2 Add Wails v2 Go dependency: `go get github.com/wailsapp/wails/v2`
- [ ] 1.3 Create `frontend/` with Svelte + Vite scaffold
  - `npm create vite@latest ./frontend -- --template svelte`
  - `cd frontend && npm install`
  - `cd frontend && npm install xterm @xterm/addon-fit @xterm/addon-web-links`
- [ ] 1.4 Create `internal/desktop/app.go` — minimal Wails app entry with `//go:embed all:frontend/dist`
- [ ] 1.5 Create `cmd/desktop.go` — cobra command that calls `desktop.Run(cfg)`
- [ ] 1.6 Verify: `cd frontend && npm run build` succeeds
- [ ] 1.7 Verify: `go build ./...` compiles

### Output Files
```
frontend/                    — Svelte scaffold
  ├── src/App.svelte         — placeholder "Hello AgentMux"
  ├── package.json
  ├── vite.config.js
  └── index.html
internal/desktop/app.go      — Wails entry point (embeds frontend/dist)
cmd/desktop.go               — cobra command wiring
```

### Completion Criteria
- `go build ./...` passes
- `cd frontend && npm run build` passes
- `./agentmux desktop --help` prints usage

---

## Stream 2: Go Backend Services

**Branch:** `feature/desktop-app-backend` (branch from `feature/desktop-app` after Stream 1)
**Goal:** Wrap existing `internal/` packages as Wails-bound services.

### Tasks

- [ ] 2.1 Create `internal/desktop/session_svc.go` — SessionService
  - `ListSessions() []SessionInfo` — wraps `session.Manager.List()`
  - `CreateSession(opts) error` — wraps runner.Launch()
  - `StopSession(name) error` — wraps session.Manager.Destroy()
  - `GetSession(name) SessionInfo` — wraps session.Manager.Get()
  - `SendKeys(name, keys) error` — wraps session.Manager.SendKeys()
- [ ] 2.2 Create `internal/desktop/terminal_svc.go` — TerminalService
  - `AttachTerminal(name) error` — start polling tmux capture-pane, emit events
  - `DetachTerminal(name) error` — stop polling
  - `SendInput(name, input) error` — tmux send-keys
  - Uses Wails `runtime.EventsEmit()` for output streaming
- [ ] 2.3 Create `internal/desktop/monitor_svc.go` — MonitorService
  - `GetLogs(name, lines) string`
  - `GetHealth(name) HealthInfo`
  - `GetResources(name) ResourceInfo`
- [ ] 2.4 Update `internal/desktop/app.go` — bind all services in `wails.Run()`
- [ ] 2.5 Create `internal/desktop/types.go` — shared DTOs (SessionInfo, HealthInfo, etc.)
- [ ] 2.6 Write unit tests: `internal/desktop/*_test.go`
- [ ] 2.7 Verify: `go test ./internal/desktop/... -v` passes
- [ ] 2.8 Verify: `go vet ./internal/desktop/...` passes

### Key Design Notes
- **Terminal streaming (MVP):** Poll `tmux capture-pane -p -t <session>` at 100ms, diff against last capture, emit delta via Wails events. For input: receive via Wails `EventsOn`, forward via `tmux send-keys`.
- **Thread safety:** Each `ptyStream` runs in its own goroutine with `context.Context` cancellation.
- All services receive `context.Context` via Wails `OnStartup` hook.

### Completion Criteria
- All service methods implemented with proper error handling
- Unit tests pass
- Services integrate correctly with existing `internal/session`, `internal/tmux`, `internal/monitor`

---

## Stream 3: Svelte Frontend

**Branch:** `feature/desktop-app-frontend` (branch from `feature/desktop-app` after Stream 1)
**Goal:** Build the desktop UI: session tree (left) + xterm.js terminal (right).

### Tasks

- [ ] 3.1 Create `frontend/src/App.svelte` — root layout with resizable split panels
- [ ] 3.2 Create `frontend/src/lib/SessionTree.svelte` — collapsible tree navigator
  - Group nodes with expand/collapse on click
  - Session nodes with status badges (● running, ○ stopped)
  - Agent type label, workdir tooltip
  - Click to select → emit event for terminal panel
- [ ] 3.3 Create `frontend/src/lib/Terminal.svelte` — xterm.js wrapper
  - Init xterm.js with fit addon
  - Listen for `terminal:output:<name>` Wails events → `term.write(data)`
  - On keypress → emit `terminal:input:<name>` Wails event
  - Handles attach/detach on session selection change
- [ ] 3.4 Create `frontend/src/lib/StatusBar.svelte` — session count, selected info, actions
- [ ] 3.5 Create `frontend/src/lib/NewAgentModal.svelte` — agent creation form
- [ ] 3.6 Create `frontend/src/stores/sessions.js` — reactive session list store
  - Polls `SessionService.ListSessions()` every 2s
  - Derived store for grouped tree structure
- [ ] 3.7 Create `frontend/src/stores/terminal.js` — active terminal state
- [ ] 3.8 Implement dark theme CSS with CSS custom properties
  - Colors matching existing TUI aesthetic
  - Smooth transitions, hover effects, premium feel
- [ ] 3.9 Verify: `npm run build` succeeds
- [ ] 3.10 Visual test: layout renders correctly, tree expands/collapses, terminal initializes

### Key Design Notes
- Use Wails-generated bindings from `frontend/wailsjs/go/desktop/` to call Go services
- Use `window.runtime.EventsOn()` / `EventsEmit()` for terminal streaming
- xterm.js theme should match the app's dark palette

### Completion Criteria
- All components render correctly
- Frontend builds without errors
- Mock data shows tree with groups, terminal initializes empty

---

## Stream 4: CLI Integration

**Branch:** `feature/desktop-app-cli` (branch from `feature/desktop-app` after Stream 1)
**Goal:** Command wiring, Makefile targets, docs.

### Tasks

- [ ] 4.1 Finalize `cmd/desktop.go` — load config, call `desktop.Run(cfg)`
- [ ] 4.2 Add Makefile targets: `desktop-dev`, `desktop-build`, `frontend-install`
- [ ] 4.3 Update `README.md` — add Desktop App section, update Commands table, prerequisites
- [ ] 4.4 Update `docs/ARCHITECTURE.md` — add desktop layer to architecture map
- [ ] 4.5 Add `.gitignore` entries for `frontend/node_modules/`, `frontend/dist/`
- [ ] 4.6 Verify: `agentmux desktop --help` works

### Completion Criteria
- CLI command registered and displays help
- Makefile targets work
- Docs updated

---

## Stream 5: Integration & Polish (Do Last)

**Branch:** `feature/desktop-app` (merge all streams here)
**Goal:** Wire frontend↔backend, end-to-end testing, final polish.

### Tasks

- [ ] 5.1 Merge `feature/desktop-app-backend` → `feature/desktop-app`
- [ ] 5.2 Merge `feature/desktop-app-frontend` → `feature/desktop-app`
- [ ] 5.3 Merge `feature/desktop-app-cli` → `feature/desktop-app`
- [ ] 5.4 Generate Wails bindings: `wails generate module`
- [ ] 5.5 Connect Svelte stores to real Go bindings (replace mock data)
- [ ] 5.6 Wire terminal events end-to-end
- [ ] 5.7 E2E smoke test:
  - Launch `agentmux desktop` → window opens
  - Create shell agent from CLI → appears in desktop tree
  - Click session → terminal shows live output
  - Type command → executes in tmux
  - Stop agent → removed from tree
  - Verify same sessions visible in `agentmux dashboard` (TUI)
- [ ] 5.8 Polish: macOS menu bar, keyboard shortcuts (Cmd+N, Cmd+Q)
- [ ] 5.9 Final: `go build ./...`, `go test ./...`, `go vet ./...` all pass

### Completion Criteria
- Desktop app fully functional with session tree + interactive terminal
- TUI and desktop coexist, both showing same tmux sessions
- All tests pass, code compiles cleanly

---

## How to Pick Up This Work

1. **Read this document** to understand current status
2. **Check the Status Overview table** at the top for which streams are available
3. **Branch from the correct base:**
   - Stream 1: work directly on `feature/desktop-app`
   - Streams 2-4: branch from `feature/desktop-app` after Stream 1 completes
   - Stream 5: work on `feature/desktop-app` after merging all sub-branches
4. **Update this document** after completing tasks:
   - Mark tasks with `[x]`
   - Update the Status Overview table
   - Update `Last Updated` timestamp
5. **Commit frequently** with descriptive messages prefixed by stream: e.g. `feat(desktop/stream-2): add SessionService bindings`

## Notes Log

| Date | Note |
|------|------|
| 2026-03-12 | Plan created. Branch `feature/desktop-app` created from `main` at `158a1ad`. |

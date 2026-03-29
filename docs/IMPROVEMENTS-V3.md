# AgentMux V3 Improvements — Implementation Tracker

> **Branch:** `feature/improvements-v3` (from `main` @ `158a1ad`)
> **Created:** 2026-03-29
> **Status:** 🟡 In Progress

This document tracks all V3 improvement tasks. It is designed for **any agent to pick up any task independently** — each item includes full context, exact file locations, and verification steps.

---

## How to Use This Document

1. **Pick an uncompleted task** — start with Phase 1 (all blockers), then move to higher phases.
2. **Mark `[/]` when starting**, `[x]` when done.
3. **Run verification** after each task: `go vet ./... && go test -race -count=1 ./...`
4. **Commit after each phase** with message format: `fix: V3 phase N — short description`

---

## Progress Summary

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Unbreak Build | `[x]` Done ✅ |
| Phase 2 | Safety & Concurrency | `[x]` Done ✅ |
| Phase 3 | Code Quality | `[x]` Done ✅ |
| Phase 4 | Test Coverage | `[x]` Done ✅ |
| Phase 5 | Stubs → Real Code | `[x]` Partial (5.1, 5.2 done; 5.3, 5.4 deferred) |
| Phase 6 | Polish & DevEx | `[x]` Partial (6.2, 6.3 done; 6.1 deferred) |

---

## Phase 1: Unbreak Build (~30 min)

> **Goal:** Make `go vet ./...` and `go test ./...` pass on this branch.

### 1.1 Fix `go.mod` — Stale Dependencies

- `[ ]` **Status:** Not Started
- **Problem:** `go vet ./...` fails with `go: updates to go.mod needed`. The TUI test file imports `github.com/charmbracelet/x/exp/teatest` which is not in `go.mod`.
- **File:** `go.mod`
- **Fix:**
  ```bash
  go get github.com/charmbracelet/x/exp/teatest
  go mod tidy
  ```
- **Verify:** `go vet ./...` exits 0

### 1.2 Remove Committed Binary Artifacts

- `[ ]` **Status:** Not Started
- **Problem:** Two compiled binaries (`agentmux` 8.3MB, `my_agentmux` 6.9MB) are tracked by git in the repo root.
- **Files:** `agentmux`, `my_agentmux`
- **Fix:**
  ```bash
  # Ensure they are in .gitignore
  echo "agentmux" >> .gitignore
  echo "my_agentmux" >> .gitignore
  # Remove from git tracking (keeps local files)
  git rm --cached agentmux my_agentmux
  ```
- **Verify:** `git status` no longer shows the binaries as tracked; `.gitignore` contains both names

### 1.3 Fix `release.yml` Go Version Mismatch

- `[ ]` **Status:** Not Started
- **Problem:** `.github/workflows/release.yml` line 23 hardcodes `go-version: '1.21'` while `go.mod` specifies `go 1.25.7`. The CI workflow (`ci.yml`) correctly uses `go-version-file: go.mod`.
- **File:** `.github/workflows/release.yml`
- **Fix:** Change `go-version: '1.21'` to `go-version-file: go.mod`
- **Verify:** Manual review of the YAML

### 1.4 Fix TUI Test Compilation Error

- `[ ]` **Status:** Not Started
- **Problem:** `internal/tui/app_test.go` line 35 calls `NewModel(cfg, sessionMgr, tmuxClient)` with 3 args, but `NewModel` requires 6 parameters: `(cfg, sessionMgr, tmuxClient, splitMode bool, rightPaneID string, projectGroups *config.ProjectConfig)`.
- **File:** `internal/tui/app_test.go`, line 35
- **Fix:** Change to `NewModel(cfg, sessionMgr, tmuxClient, false, "", nil)`
- **Verify:** `go test ./internal/tui/...` compiles (may skip if tmux unavailable)

---

## Phase 2: Safety & Concurrency (~2 hours)

> **Goal:** Fix command injection, data races, and goroutine leaks.

### 2.1 Fix Command Injection via Unsanitized Args

- `[ ]` **Status:** Not Started
- **Problem:** `internal/session/manager.go` lines 91-97 builds the tmux session command via string concatenation. Shell metacharacters in user-provided args (`;`, `&&`, `$()`, backticks) are interpreted.
  ```go
  // CURRENT (vulnerable):
  sessionCmd = opts.Command
  for _, arg := range opts.Args {
      sessionCmd += " " + arg
  }
  ```
- **File:** `internal/session/manager.go`, lines 91-97
- **Fix:** Quote each argument with `fmt.Sprintf("%q", arg)` or use `strings.Join` with proper shell escaping. Consider a helper function `shellQuote(args []string) string`.
- **Verify:** Unit test with args containing `; rm -rf /` should not execute the injection

### 2.2 Fix HealthMonitor Data Race

- `[ ]` **Status:** Not Started
- **Problem:** `internal/monitor/health.go` — `HealthMonitor` has no mutex. `monitorLoop` runs in a goroutine calling `checkAll()` which reads/writes `hm.agents` and `hm.configs`. Meanwhile, `RecordOutput`, `GetStatus`, `Register`, `Unregister`, `AllStatuses` can be called from other goroutines.
- **File:** `internal/monitor/health.go`
- **Fix:** Add `mu sync.RWMutex` to `HealthMonitor` struct. Use `mu.Lock()` in `Register`, `Unregister`, `RecordOutput`, `checkAll`. Use `mu.RLock()` in `GetStatus`, `AllStatuses`.
- **Verify:** `go test -race ./internal/monitor/...` passes

### 2.3 Fix Goroutine Leak in FileLock Timeout

- `[ ]` **Status:** Not Started
- **Problem:** `internal/session/filelock.go` lines 40-57 — when `flock()` blocks and the timeout fires, the goroutine calling blocking `flock()` is leaked forever (it never gets cancelled).
- **File:** `internal/session/filelock.go`, lines 36-57
- **Fix:** Replace the blocking goroutine approach with a polling loop using `LOCK_NB` and `time.NewTicker`:
  ```go
  deadline := time.Now().Add(timeout)
  ticker := time.NewTicker(50 * time.Millisecond)
  defer ticker.Stop()
  for time.Now().Before(deadline) {
      err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
      if err == nil {
          return nil
      }
      <-ticker.C
  }
  ```
- **Verify:** No goroutine leak under contention; test with concurrent lock attempts

---

## Phase 3: Code Quality (~2 hours)

> **Goal:** Introduce sentinel errors, typed statuses, deterministic ordering, and complete merge logic.

### 3.1 Define Sentinel Errors

- `[ ]` **Status:** Not Started
- **Problem:** Error strings like `"agent %q not found"` and `"agent %q is already running"` are created with `fmt.Errorf` throughout `internal/session/manager.go`. Callers cannot use `errors.Is` to distinguish error types.
- **Files:** `internal/session/manager.go` (define errors), update callers in `cmd/stop.go`, `cmd/resume.go`, etc.
- **Fix:** Create `internal/session/errors.go`:
  ```go
  package session
  
  import "errors"
  
  var (
      ErrAgentNotFound       = errors.New("agent not found")
      ErrAgentAlreadyRunning = errors.New("agent is already running")
      ErrAgentNotRunning     = errors.New("agent is not running")
  )
  ```
  Then update `manager.go` to use: `fmt.Errorf("agent %q: %w", name, ErrAgentNotFound)`
- **Verify:** `go build ./...` passes; existing tests still pass

### 3.2 Define Typed Session Status

- `[ ]` **Status:** Not Started
- **Problem:** Session status is a bare `string` (`"running"`, `"stopped"`, `"error"`). Typos would silently break comparison logic across `manager.go`, `app.go`, `doctor.go`.
- **Files:** `internal/session/state.go` (define type), update all comparisons in `manager.go`, `cmd/doctor.go`, `internal/tui/app.go`
- **Fix:** Add to `internal/session/state.go`:
  ```go
  type SessionStatus string
  const (
      StatusRunning SessionStatus = "running"
      StatusStopped SessionStatus = "stopped"
      StatusError   SessionStatus = "error"
  )
  ```
  Change `Status string` to `Status SessionStatus` in `AgentSession`.
- **Verify:** `go build ./...` passes; grep for any remaining `== "running"` string literals and replace with `== StatusRunning`

### 3.3 Fix Non-Deterministic Preset Ordering

- `[ ]` **Status:** Not Started
- **Problem:** `internal/agent/registry.go` lines 118-124 — `AvailablePresets()` iterates a map, producing random order in error messages and help text.
- **File:** `internal/agent/registry.go`, `AvailablePresets()` function
- **Fix:** Add `sort.Strings(names)` before the `return strings.Join(names, ", ")` line.
- **Verify:** Run `AvailablePresets()` multiple times — output is always alphabetically sorted

### 3.4 Complete `MergeProjectConfig`

- `[ ]` **Status:** Not Started
- **Problem:** `internal/config/project.go` lines 90-103 — `MergeProjectConfig` only merges `DefaultAgentType`. Fields `DefaultWorkDir`, `Env`, and `Monitor` overrides are silently ignored.
- **File:** `internal/config/project.go`, `MergeProjectConfig()` function
- **Fix:** Merge additional fields:
  ```go
  if project.DefaultWorkDir != "" {
      // Store for use by commands, though Config doesn't have this field yet
  }
  // Merge project env into a DefaultEnv field, etc.
  ```
  At minimum, document which fields ARE merged vs. which are handled at the command level.
- **Verify:** Unit test in `config_test.go` verifying merge behavior

---

## Phase 4: Test Coverage (~3-4 hours)

> **Goal:** Get 0%-coverage packages to at least 50%.

### 4.1 Add Orchestrator Tests

- `[ ]` **Status:** Not Started
- **Coverage:** 0% → target 60%+
- **File to create:** `internal/orchestrator/orchestrator_test.go`
- **Test cases:**
  - Single-stage pipeline with 1 agent
  - Multi-stage pipeline with parallel agents
  - Stage failure with `FailurePolicyAbort` vs `FailurePolicySkip`
  - Stage timeout
  - Retry policy
- **Notes:** Will need mock `session.Manager` and `agent.Runner`, or use shell agents with `echo` commands

### 4.2 Add Plugin Tests

- `[ ]` **Status:** Not Started
- **Coverage:** 0% → target 50%+
- **File to create:** `internal/plugin/plugin_test.go`
- **Test cases:**
  - `HookRunner.Fire` with matching event
  - `HookRunner.Fire` with non-matching event (no-op)
  - `runCommand` with simple echo
  - `PluginProtocol.Send`/`Receive` with a subprocess
  - `DiscoverPlugins` with executable vs non-executable files

### 4.3 Add Health Monitor Tests

- `[ ]` **Status:** Not Started
- **Coverage:** (within monitor's 52.7%) → target standalone 70%+
- **File to create:** `internal/monitor/health_test.go`
- **Test cases:**
  - `Register` + `GetStatus` returns healthy
  - `RecordOutput` updates `lastOutput` timestamp
  - Idle timeout triggers unhealthy status
  - `MaxRestarts` exceeded stops restarting
  - Concurrent `RecordOutput` + `checkAll` (race detector)

### 4.4 Add History Package Tests

- `[ ]` **Status:** Not Started
- **Coverage:** 0% → target 60%+
- **Existing file:** `internal/history/store.go`
- **File to create:** `internal/history/store_test.go`
- **Test cases:** Record session, list sessions, filter by date, truncate old entries

### 4.5 Add Git Package Tests

- `[ ]` **Status:** Not Started
- **Coverage:** 0% → target 50%+
- **Existing file:** `internal/git/git.go`
- **File to create:** `internal/git/git_test.go`
- **Test cases:** Detect git repo, detect non-git directory, branch name extraction

### 4.6 Add CLI Smoke Tests

- `[ ]` **Status:** Not Started
- **Coverage:** `cmd/` at ~0%
- **File:** `cmd/cmd_test.go` (exists but minimal)
- **Test cases:**
  - `agentmux version` returns version string
  - `agentmux --help` exits 0
  - `agentmux start` without args triggers wizard or error
  - `agentmux doctor` runs without crash

---

## Phase 5: Stubs → Real Code (~2-3 hours)

### 5.1 Implement Buffered Logger

- `[ ]` **Status:** Not Started
- **File:** `internal/monitor/logger.go`
- **Problem:** `f.WriteString(content)` writes directly to OS file handle every 500ms per agent.
- **Fix:** Wrap `os.File` in `bufio.Writer` in `getOrOpenFile`. Store `*bufio.Writer` in the `files` map (or a wrapper struct). Flush in `Close()`, `CloseAgent()`, and `rotateIfNeeded()`. Add a periodic flush goroutine (every 500ms).
- **Verify:** Benchmark before/after with high-frequency writes

### 5.2 Implement Webhook Support (or Remove Claim)

- `[ ]` **Status:** Not Started
- **File:** `internal/plugin/plugin.go`, `sendWebhook()` function (line 116)
- **Decision needed:** Implement real HTTP POST, or remove webhook references from docs
- **If implementing:**
  ```go
  func (hr *HookRunner) sendWebhook(ctx context.Context, url string, hctx HookContext) error {
      data, _ := json.Marshal(hctx)
      req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
      req.Header.Set("Content-Type", "application/json")
      resp, err := http.DefaultClient.Do(req)
      if err != nil { return err }
      defer resp.Body.Close()
      if resp.StatusCode >= 400 { return fmt.Errorf("webhook returned %d", resp.StatusCode) }
      return nil
  }
  ```

### 5.3 Implement `LoadHooksFromDefinition`

- `[ ]` **Status:** Not Started
- **File:** `internal/plugin/plugin.go`, `LoadHooksFromDefinition()` function (line 188)
- **Problem:** Parses frontmatter but always returns `nil, nil`
- **Fix:** Use `gopkg.in/yaml.v3` to unmarshal the frontmatter into a struct that includes a `hooks` field, then return parsed hooks
- **Verify:** Unit test with a sample agent definition containing hooks

### 5.4 Add Structured Logging with `log/slog`

- `[ ]` **Status:** Not Started
- **Problem:** All output uses `fmt.Printf`. No log levels, no structured output.
- **Fix:** Add a global `slog.Logger` initialized in `cmd/root.go` based on `cfg.LogLevel`. Replace internal debug/warn/error prints with `slog.Debug`, `slog.Warn`, `slog.Error`. Keep user-facing output as `fmt.Printf`.
- **Scope:** Start with `internal/session/manager.go` and `internal/monitor/watcher.go` as pilot files. Don't convert `cmd/` user-facing output.

---

## Phase 6: Polish & DevEx (~1 hour)

### 6.1 Add `CONTRIBUTING.md`

- `[ ]` **Status:** Not Started
- **File to create:** `CONTRIBUTING.md`
- **Content:** Dev setup instructions, `make test` workflow, PR process, link to `docs/ARCHITECTURE.md`

### 6.2 Add Config Validation

- `[ ]` **Status:** Not Started
- **File:** `internal/config/config.go`
- **Fix:** Add `func (c *Config) Validate() error` that checks:
  - `PollIntervalMs > 0`
  - `MaxLogSizeMB > 0`
  - `LogLevel` is one of `debug`, `info`, `warn`, `error`
  - `SessionPrefix` is non-empty and contains only `[a-zA-Z0-9-]`
- **Call from:** `config.Load()` after unmarshaling

### 6.3 Remove Duplicate Release Job from `ci.yml`

- `[ ]` **Status:** Not Started
- **File:** `.github/workflows/ci.yml`, lines 75-94
- **Problem:** Both `ci.yml` and `release.yml` define release-on-tag logic. They will conflict.
- **Fix:** Remove the `release:` job from `ci.yml`. The dedicated `release.yml` (goreleaser) handles this.

---

## Reference: File Map

Key files by package for quick navigation:

```
cmd/
  root.go          — CLI entrypoint, config loading
  start.go         — `agentmux start` command
  doctor.go        — `agentmux doctor` + `cleanup` commands
  pipeline.go      — `agentmux pipeline run` command
  resume.go        — `agentmux save/resume` commands

internal/
  session/
    manager.go     — Session lifecycle (Create, List, Get, Destroy)
    state.go       — JSON state persistence + AgentSession struct
    filelock.go    — Cross-process file locking (flock)
  tmux/
    client.go      — tmux CLI wrapper (386 lines)
    types.go       — Session/Window/Pane structs + parsers
  monitor/
    logger.go      — Per-agent log files
    watcher.go     — Pane output polling + resource monitoring
    health.go      — Restart policies + idle detection
  agent/
    runner.go      — Agent launch logic
    registry.go    — Built-in agent presets (claude, aider, codex, etc.)
  config/
    config.go      — Global config loading/saving
    agents.go      — Agent definition parsing (YAML frontmatter)
    project.go     — Project-level config
  orchestrator/
    orchestrator.go — Pipeline execution with stages
  plugin/
    plugin.go      — Hooks, webhooks, plugin protocol
  tui/
    app.go         — Bubbletea dashboard model (576 lines)
    components/    — SessionTree, LogViewer, SearchBar, ActionMenu, etc.
  workflow/
    spec.go        — Plan CRUD (create/approve/reject)
```

---

## Verification Commands

```bash
# Full build check
go vet ./... && go build ./...

# All tests with race detector
go test -race -count=1 ./...

# Coverage report
go test -cover ./internal/...

# Build binary
make build

# Integration tests (requires tmux)
go test -v -count=1 ./tests/...
```

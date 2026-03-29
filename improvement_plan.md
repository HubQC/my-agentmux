# AgentMux Improvement Plan

Based on a review of the `my_agentmux` repository, the project has a very solid architecture, well-defined boundaries (CLI -> Session/Agent -> tmux/monitor), and excellent integration testing. However, there are several areas where the project can be elevated from a personal tool to a production-grade OSS project.

This document outlines an improvement plan across 5 core areas.

## 1. CI/CD & Automation (High Priority)
The project currently relies on a `Makefile` for local operations, but lacks automated CI pipelines.
- **Action**: Add GitHub Actions workflows (`.github/workflows/ci.yml`).
- **Details**:
  - Run `golangci-lint`, `go test -race`, and `go build` on every PR and push to `main`.
  - Add a `.github/workflows/release.yml` workflow that triggers `goreleaser` automatically when a new Git tag is pushed, offloading the release process from the developer's local machine.

## 2. Test Coverage for CLI & TUI (Medium Priority)
While `internal/*` packages have ~70-80% coverage (from `go test -cover ./...`), the `cmd/` and `internal/tui/` packages show 0% statement coverage.
- **Action**: Introduce UI and CLI testing.
- **Details**: 
  - For the **CLI** (`cmd/` package), integrate a tool like `testscript` (by rogpeppe) or use `os/exec` to run high-level CLI smoke tests against a compiled binary to verify flags and exit codes.
  - For the **TUI** (`internal/tui/`), use `charmbracelet/x/exp/teatest` to write unit tests for `bubbletea` models, verifying state transitions and view output without needing a real terminal.

## 3. Performance: Buffered Logging (Low Priority but easy win)
Agents (especially LLMs streaming fast responses) can generate a lot of disk I/O. Currently, `internal/monitor/logger.go` writes directly to the `os.File` handle.
- **Action**: Wrap log files with `bufio.Writer`.
- **Details**:
  - Modify `getOrOpenFile` to wrap the `os.File` in a `bufio.Writer`.
  - Add a periodic background flush goroutine (e.g., every 500ms) or explicitly flush in `Write` and `Close` to reduce syscalls and protect disk performance during high-throughput streams.

## 4. Resilient Error Handling (Medium Priority)
Currently, error handling primarily uses `fmt.Errorf("... %w", err)` for wrapping.
- **Action**: Implement typed sentinel errors.
- **Details**:
  - Define custom errors in `internal/session` (e.g., `ErrAgentNotFound`, `ErrSessionAlreadyRunning`) and check them using `errors.Is`/`errors.As`.
  - This prevents upstream handlers (like the CLI or TUI) from relying on fragile string matching if they need to conditionally handle specific failures (e.g., silently ignoring "agent not found" on a `stop --all` command).

## 5. Community & Developer Experience (Low Priority)
The `README.md` is excellent, but open-source projects benefit from standard community files.
- **Action**: Add `CONTRIBUTING.md` and Issue Templates.
- **Details**:
  - Create `CONTRIBUTING.md` pointing to `DESIGN.md` and the `Makefile` testing requirements.
  - Add a troubleshooting section in the `README.md` (e.g., "What to do if tmux fails to attach").

# AgentMux — Build Walkthrough

This document tracks what has been built and verified at each step.

---

## Step 1: Project Scaffolding & Core CLI ✅ (2026-02-28)

### What was built
- Go module `github.com/cqi/my_agentmux` with cobra CLI
- Config system loading from `~/.agentmux/config.yaml`
- `agentmux version`, `agentmux --version`, `agentmux --help`

### Tests: 5/5 pass

---

## Step 2: tmux Integration Layer ✅ (2026-02-28)

### What was built
- `internal/tmux/types.go` — `Session`, `Window`, `Pane` structs with format parsers
- `internal/tmux/client.go` — Full tmux client: NewSession, KillSession, ListSessions, GetSession, HasSession, ListWindows, ListPanes, SplitWindow, SelectPane, SendKeys, CapturePane

### Bug fix
tmux 3.4 `-p` (percentage split) fails in detached sessions. Fixed with `-l` (absolute) via `getTargetSize()`.

### Tests: 10/10 pass

---

## Step 3: Agent Session Management ✅ (2026-02-28)

### What was built
- `internal/session/state.go` — Thread-safe JSON state persistence with atomic writes
- `internal/session/manager.go` — Session lifecycle with tmux/state reconciliation:
  - `Create()` — creates tmux session + tracks in JSON state
  - `List()` — lists sessions, reconciles dead tmux sessions
  - `Get()` — get single session with status check
  - `Attach()` — exec into tmux attach-session
  - `Destroy()` / `DestroyAll()` — kill tmux + remove state
  - `SendKeys()` / `CaptureOutput()` — interact with agent panes
- `cmd/start.go` — `agentmux start <name> [-w workdir] [-t agent-type]`
- `cmd/list.go` — `agentmux list` with `●/○/✗` status, tabwriter, uptime
- `cmd/stop.go` — `agentmux stop <name>` / `agentmux stop --all`
- `cmd/attach.go` — `agentmux attach <name>` (replaces process)

### CLI smoke test
```
$ ./agentmux start test-smoke -w /tmp
✓ Agent "test-smoke" started (tmux: amux-test-smoke, workdir: /tmp)

$ ./agentmux list
NAME        TYPE    STATUS     WORKDIR  UPTIME
test-smoke  claude  ● running  /tmp     0s

$ ./agentmux stop test-smoke
✓ Agent "test-smoke" stopped
```

### Tests: 4 new (19/19 total pass)

---

## Step 4: Agent Process Orchestration ✅ (2026-02-28)

### What was built
- `internal/agent/runner.go` — `AgentRunner` launches agent CLIs in tmux sessions:
  - `Launch()` — resolves preset → validates binary → builds command → creates session
  - `LaunchCustom()` — runs arbitrary command in tmux
  - `FormatAgentCommand()` — human-readable command display with truncation
  - `ValidateAgentType()` — validates preset or special types (shell, custom)
- `internal/agent/registry.go` — 6 built-in presets: claude, aider, codex, gemini, copilot, shell
- `cmd/start.go` [MODIFIED] — now uses `agent.Runner` instead of `session.Manager` directly:
  - `--args` / `-a` — pass extra arguments to the agent CLI
  - `--command` / `-c` — run a custom command (overrides preset)
  - Early agent type validation
  - Shows resolved command in output

### CLI smoke test
```
$ ./agentmux start test-shell --agent-type shell -w /tmp
✓ Agent "test-shell" started (tmux: amux-test-shell, workdir: /tmp)
  Command: (default shell)
  Attach:  agentmux attach test-shell
  Stop:    agentmux stop test-shell

$ ./agentmux start test-custom --command "echo hello" -w /tmp
✓ Agent "test-custom" started (tmux: amux-test-custom, workdir: /tmp)
  Command: echo hello
  Attach:  agentmux attach test-custom
  Stop:    agentmux stop test-custom

$ ./agentmux list
NAME         TYPE    STATUS     WORKDIR  UPTIME
test-shell   shell   ● running  /tmp     0s
test-custom  custom  ○ stopped  /tmp     -

$ ./agentmux stop --all
✓ Stopped 2 agent(s)
```

### Tests: 11 new (30/30 total pass)
- 7 unit tests: preset command, install check, type validation, formatting, truncation, listing, get preset
- 4 integration tests: shell launch, custom command, duplicate detection, empty name validation

---

## Step 5: Inter-Agent Communication & Monitoring ✅ (2026-02-28)

### What was built
- `internal/monitor/logger.go` — Per-agent log file manager:
  - Append writes, timestamped entries, file handle pooling
  - Log rotation when exceeding configurable max size
  - Multi-agent isolation (separate log files per agent)
- `internal/monitor/watcher.go` — Tmux pane output watcher:
  - Polls `tmux capture-pane` at configurable intervals
  - Detects output changes via diff, emits `Event` structs to channel
  - Non-blocking event emission (drops when channel full)
  - Per-agent watch/unwatch with context cancellation
- `cmd/logs.go` — `agentmux logs <name>`:
  - Default: prints existing log or one-time pane capture
  - `--follow` / `-f`: real-time streaming via Watcher
  - `--tail N` / `-n N`: show last N lines
- `cmd/send.go` — `agentmux send <name> <message>`:
  - Sends input to agent's tmux pane + Enter
  - `--no-enter`: send without pressing Enter

### CLI smoke test
```
$ ./agentmux start test-mon --agent-type shell -w /tmp
✓ Agent "test-mon" started (tmux: amux-test-mon, workdir: /tmp)

$ ./agentmux send test-mon "echo HELLO_FROM_SEND"
✓ Sent to "test-mon": echo HELLO_FROM_SEND

$ ./agentmux logs test-mon
echo HELLO_FROM_SEND
❯ echo HELLO_FROM_SEND
HELLO_FROM_SEND

$ ./agentmux stop test-mon
✓ Agent "test-mon" stopped
```

### Tests: 9 new (39/39 total pass)
- 7 logger unit tests: write/read, timestamped, empty, nonexistent, path, rotation, multi-agent
- 1 watcher unit test: create/stop lifecycle
- 1 watcher integration test: tmux polling → event capture → log verification

---

## Step 6: Configuration & Agent Definition System ✅ (2026-03-01)

### What was built
- `internal/config/agents.go` — Agent definition loader:
  - Parse `.md` files with YAML frontmatter (description, agent_type, args, env, workdir)
  - Parse `.yaml` / `.yml` files as pure YAML definitions
  - Plain Markdown files treated as system prompt with default agent type
  - `LoadAgentDefinitions()` — batch load from agents directory
  - `GetAgentDefinition()` — look up by name
- `internal/config/project.go` — Project-level config:
  - `LoadProjectConfig()` — reads `.agentmux/config.yaml` from a project dir
  - `SaveProjectConfig()` — writes config with `os.MkdirAll`
  - `MergeProjectConfig()` — merges project overrides into global config
  - Per-agent overrides (agent_type, args, env, workdir)
- `cmd/agents.go` — `agentmux agents [--all]`:
  - Lists built-in presets with install status (● installed / ○ not found / ● built-in)
  - Lists custom agent definitions from agents directory
  - Sorted output with tabwriter
- `cmd/start.go` [MODIFIED] — Agent definition integration:
  - `--config` flag for project-level config file
  - Auto-lookup agent definitions matching the session name
  - Merge env vars from agent definition into launch options
  - Definition defaults (agent_type, workdir, args) with flag overrides

### CLI smoke test
```
$ ./agentmux agents --all
NAME     TYPE    STATUS       DESCRIPTION
----     ----    ------       -----------
aider    preset  ○ not found  Aider — AI pair programming in your terminal
claude   preset  ● installed  Anthropic Claude Code — AI coding assistant
codex    preset  ● installed  OpenAI Codex CLI — AI coding agent
copilot  preset  ○ not found  GitHub Copilot CLI
gemini   preset  ● installed  Google Gemini CLI — AI coding assistant
shell    preset  ● built-in   Plain shell session (bash/zsh)

$ ./agentmux start test-step6 --agent-type shell -w /tmp
✓ Agent "test-step6" started (tmux: amux-test-step6, workdir: /tmp)
  Command: (default shell)

$ ./agentmux list
NAME        TYPE   STATUS     WORKDIR  UPTIME
test-step6  shell  ● running  /tmp     0s

$ ./agentmux stop test-step6
✓ Agent "test-step6" stopped
```

### Tests: 12 new (51/51 total pass)
- 7 agent definition tests: markdown, YAML, plain MD, batch load, nonexistent dir, get by name, unclosed frontmatter
- 5 project config tests: load, nonexistent, save/reload, merge, nil merge

---

## Step 7: Spec-Driven Workflow Engine ✅ (2026-03-01)

### What was built
- `internal/workflow/spec.go` — Plan lifecycle manager:
  - `Plan` struct — ID, Title, Description, Status (draft/approved/rejected), RejectReason, Agent, timestamps
  - `PlanStore` — manages YAML plan files in `~/.agentmux/plans/`:
    - `Create()` — sequential plan IDs (`plan-001`, `plan-002`, ...), writes YAML
    - `List()` — loads all plans, sorted by creation time
    - `Get()` — load single plan by ID
    - `Approve()` — draft → approved transition with validation
    - `Reject()` — draft → rejected with reason, validation
    - `Delete()` — remove plan file
  - `FormatStatus()` — display-friendly status icons (◎ draft / ✓ approved / ✗ rejected)
- `cmd/plan.go` — `agentmux plan` with 6 subcommands:
  - `plan create <title> [-d description]`
  - `plan list` — tabwriter with status icons
  - `plan show <id>` — full plan details
  - `plan approve <id>` — approve a draft plan
  - `plan reject <id> [-r reason]` — reject with reason
  - `plan delete <id>` — remove a plan

### CLI smoke test
```
$ ./agentmux plan create "Add authentication" -d "Implement OAuth2 login flow"
✓ Plan plan-001 created: Add authentication
  Description: Implement OAuth2 login flow
  Status: ◎ draft
  Approve: agentmux plan approve plan-001

$ ./agentmux plan create "Refactor database layer"
✓ Plan plan-002 created: Refactor database layer

$ ./agentmux plan list
ID        TITLE                    STATUS   CREATED
plan-001  Add authentication       ◎ draft  2026-03-01 03:15
plan-002  Refactor database layer  ◎ draft  2026-03-01 03:15

$ ./agentmux plan approve plan-001
✓ Plan plan-001 approved

$ ./agentmux plan reject plan-002 -r "Needs more detail"
✗ Plan plan-002 rejected

$ ./agentmux plan list
ID        TITLE                    STATUS      CREATED
plan-001  Add authentication       ✓ approved  2026-03-01 03:15
plan-002  Refactor database layer  ✗ rejected  2026-03-01 03:15

$ ./agentmux plan delete plan-002
✓ Plan plan-002 deleted
```

### Tests: 9 new (60/60 total pass)
- TestCreatePlan, TestCreatePlanEmptyTitle
- TestListPlans, TestGetPlan
- TestApprovePlan, TestRejectPlan, TestApproveNonDraft
- TestDeletePlan, TestPlanIDSequence

---

## Step 8: TUI Dashboard ✅ (2026-03-01)

### What was built
- **Dependencies:** `bubbletea` v1.3.10, `lipgloss` v1.1.0, `bubbles` v1.0.0
- `internal/tui/styles.go` — Dark theme: violet primary, cyan secondary, green/red status, rounded borders
- `internal/tui/components/agent_list.go` — Sidebar component:
  - Status icons (● running / ○ stopped), selection highlighting
  - Agent type + uptime display, keyboard navigation (up/down)
- `internal/tui/components/log_viewer.go` — Main log panel:
  - Scrollable output viewer, auto-scroll to bottom on new content
  - Placeholder states for no agent / no output
  - ScrollUp/ScrollDown with pgup/pgdn
- `internal/tui/components/status_bar.go` — Bottom bar:
  - Key bindings help on left, agent count on right
- `internal/tui/app.go` — Main bubbletea model:
  - 1s tick refresh for agent list from session manager
  - Live output streaming via monitor.Watcher events channel
  - Keyboard: ↑/k ↓/j navigate, pgup/pgdn scroll, `a` attach, `d` stop, `q` quit
  - Responsive layout adapting to terminal size
  - Log buffer capped at 1000 lines per agent
- `cmd/dashboard.go` — `agentmux dashboard` with alt-screen + mouse support

### Tests: 60/60 total pass (no regressions)

---

## Step 9: DevContainer & Cross-Platform ✅ (2026-03-01)

### What was built
- `internal/devcontainer/generator.go` — Devcontainer config generation:
  - `DefaultConfig()` — Ubuntu image, Go + tmux features, VS Code extensions
  - `ConfigWithAgents()` — adds agent CLI install commands (claude, aider, codex)
  - `Generate()` — writes `.devcontainer/devcontainer.json`
- `internal/platform/platform.go` — Platform detection:
  - `Detect()` — returns OS, arch, WSL status (via env + `/proc/version`), Docker status
- `cmd/init.go` — `agentmux init`:
  - Creates `~/.agentmux/config.yaml` + sample `example.md` agent definition
  - `--devcontainer` generates `.devcontainer/devcontainer.json`
  - `--force` overwrites existing files, shows platform info
- `cmd/completion.go` — `agentmux completion <bash|zsh|fish|powershell>`

### CLI smoke test
```
$ ./agentmux init
✓ Created /home/cqi/.agentmux/config.yaml
✓ Created /home/cqi/.agentmux/agents/example.md
  Platform: linux/amd64
✓ Initialized 2 file(s)

$ ./agentmux init --devcontainer --force
✓ Created .devcontainer/devcontainer.json
✓ Initialized 3 file(s)

$ ./agentmux completion bash > /dev/null && echo "bash OK"
bash OK
```

### Tests: 4 new (64/64 total pass)
- 3 devcontainer tests: default config, generate, agent installs
- 1 platform test: detect smoke test

---

## Step 10: Polish, Testing & Documentation ✅ (2026-03-01)

### What was built
- `README.md` — Full documentation: features, quick start, commands reference, agent types, configuration, dashboard preview, architecture
- `Makefile` — Targets: build (with ldflags), test, lint, clean, install, fmt, deps
- `scripts/install.sh` — Curl-able installer with OS/arch detection, release download, source fallback
- `.goreleaser.yaml` — Cross-platform releases: linux/darwin × amd64/arm64, checksums, changelog
- `tests/integration_test.go` — 6 end-to-end tests:
  - Agent lifecycle (start → send keys → capture → stop)
  - Monitor flow (write → read logs)
  - Workflow flow (create → approve → reject → delete)
  - Devcontainer generation
  - Platform detection
  - Config + agent definitions

### Tests: 6 new E2E (70/70 total pass)

### Verification
```
$ make test    ✓ 70/70 pass
$ make lint    ✓ go vet clean
$ make build   ✓ binary with ldflags
```

---

## 🎉 Project Complete

**AgentMux** — 10/10 steps built across 3 sessions.

| Metric | Value |
|--------|-------|
| Total tests | 70 |
| Packages | 10 |
| CLI commands | 15 |
| Sessions | 3 |

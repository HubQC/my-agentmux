# Codex Integration & MCP Server Visualization

## Goal Description
Enhance AgentMux with deep integration for Codex users (specifically targeting `gpt-5.3-codex` and multi-agent setups). The goal is to accurately parse the `~/.codex/config.toml` file so the AgentMux TUI can directly visualize active MCP servers, Codex profiles, reasoning efforts, and sub-agent roles natively.

## Proposed Changes

### Configuration Parsing Layer
We need to parse Codex's TOML format natively since AgentMux currently only uses YAML for its own configs.

#### [NEW] `internal/codex/config.go`
- Create a new package to parse `~/.codex/config.toml` targeting `go-toml/v2`.
- Define structs to map:
  - `mcp_servers` (name, type, command)
  - `profiles` (name, model, personality, model_reasoning_effort)
  - `agents` (task_planner, supervisor)
  - Global `model`, `profile`, `multi_agent` boolean.
- Provide a `LoadConfig()` utility.

---

### Session Management Layer
The session layer needs to augment Codex sessions with the parsed metadata.

#### [MODIFY] `internal/session/agent.go`
- Inject the parsed Codex config metadata into the Agent session state when an agent of type `codex` is started or monitored.
- Attach specific labels/tags to the session struct (e.g., `Profile: gpt-5.3-codex`, `MCPs: filesystem, chrome-devtools`).

---

### TUI Dashboard Layer
The dashboard is the main visual interface where the user wants to see the context and MCP tools.

#### [MODIFY] `internal/tui/model.go` & `internal/tui/session_view.go`
- Update the drawing logic (`View()`) to render Codex-specific metadata.
- For a Codex agent, replace or augment the basic status line with:
  - **Profile**: `gpt-5.3-codex (Reasoning: high)`
  - **Mode**: `[Multi-Agent]` (if `task_planner` or `supervisor` configs are active)
  - **Tools**: List active MCP servers (e.g., `🔌 MCP: filesystem, chrome-devtools, sqlcl`)

## Verification Plan

### Automated Tests
- Run unit tests on the new TOML parser.
- `go test ./internal/codex/... -v`
- `make build` to ensure no compile errors.

### Manual Verification
1. Open the AgentMux dashboard: `./agentmux dashboard`
2. Start a codex session: `./agentmux start my-coder --agent-type codex`
3. Verify that the TUI explicitly shows the Codex profile `gpt-5.3-codex`, MCP servers `[filesystem, chrome-devtools, sqlcl]`, and advanced context indicators.

# Dashboard Modifications Design

## Problem
The current AgentMux TUI dashboard gives a read-only list of tracked sessions along with basic controls via shortcuts (e.g. `d` to destroy, `m` for action menu, `/` for find), but lacks creation flow entirely and forces users out of the immediate dashboard context to view detailed logs.

## Feature 1: "Create Agent" Flow directly from TUI
**Goal:** Allow users to press `n` (New) in the dashboard to pop up a quick-add modal to define and launch a new agent session.

**Implementation Details:**
- We can integrate `huh` form elements as bubbletea components seamlessly into our existing `app.go`. 
- Since `app.go` currently intercepts keypresses, `n` will set a `creatingAgent = true` state.
- In `Update()`, key events are re-routed to a embedded `huh.Form` until submitted or discarded (ESC).
- Upon submission, it triggers the `Launch()` command programmatically (similar to how `cmd/start.go` works) and inserts the new session into the dashboard list dynamically.

## Feature 2: Side-by-Side inline Log Previews
**Goal:** Prevent context-switching for users who just want to glance at why an agent is running or stuck.

**Implementation Details:**
- Currently `agentmux dashboard` provides a simple vertical stack layout or single column. We can partition the terminal into two panes (using `lipgloss` functions like `JoinHorizontal`).
- Left pane (30%): The existing `components.SessionTree`.
- Right pane (70%): A new `components.LogViewer` (which actually seems to be stubbed in the code based on previous exploration in `app.go`, but perhaps isn't fully utilized or doesn't support interactive tailing natively).
- When a user moves up/down in the session tree (`k`/`j`), the right pane strictly subscribes or displays the tail of the log file for that active session (`~/.agentmux/logs/<agent>.log`).
- The log pane can be toggled via a shortcut (e.g., `l` for logs).

## Trade-offs
- **Pros**: Reduces reliance on dropping to the shell, builds a complete ecosystem feel, keeps the user in flow.
- **Cons**: Subscribing to log file updates per session requires careful file watching (e.g., `fsnotify` or polling tail) so we don't block the main TUI render loop.

## Next Steps
We will modify `internal/tui/app.go` to intercept `n` for creation, and ensure the log pane updates reactively based on the active selection.

# Interactive CLI Wizard for `agentmux start`

## Problem
Currently, users must know the specific flags for the `start` command to use AgentMux effectively (e.g. `agentmux start <name> -t <type>`). Missing arguments typically result in a hard failure or unhelpful error without guiding the user. 
This can be intimidating for new users or cumbersome for frequent users who want a quick guided flow when creating an agent.

## Proposed Solution
Introduce an Interactive CLI Wizard for the `start` command using charmbracelet's `huh` library (or native `bubbletea` if necessary, since it's already in the dependency tree).

When a user runs `agentmux start` *without* the required `<agent-name>` argument, instead of failing, we launch the wizard.

### Flow
The wizard will guide the user through the following steps:
1. **Agent Name**: "What should we name this agent?" (Text Input)
2. **Agent Type**: "Which preset or template do you want to use?" (Select List: `claude`, `aider`, `codex`, `gemini`, etc.) 
3. **Workspace**: "Where should this agent run?" (Directory Picker or Text Input defaulting to `./`)
4. **Group (Optional)**: "Assign to a project group?" (Text Input or Select from existing)

Once the wizard completes, it seamlessly executes the command using the defined variables, exactly as if the user passed them via flags.

## Approach & Implementation

### 1. Dependency
Add `github.com/charmbracelet/huh` to `go.mod`. It is built on top of `bubbletea` and `lipgloss` (which the project already uses), so it is lightweight and visually consistent with the existing dashboard.

### 2. Modify `cmd/start.go`
- Change `Args: cobra.ExactArgs(1)` to `Args: cobra.MaximumNArgs(1)`.
- In `RunE`:
  ```go
  if len(args) == 0 {
      // Launch Interactive Wizard
      opts, err := RunStartWizard(activeCfg)
      if err != nil { return err }
      // Execute launch with opts
  } else {
      // Existing flag parsing logic
  }
  ```

### 3. Build the Wizard (`internal/wizard/start.go`)
- Use `huh.NewForm()` to build the multi-step prompt.
- Fetch available presets using `agent.AvailablePresets()` to populate the "Agent Type" select list.

## Trade-offs
- **Pros**: Drastically improves onboarding; feels modern and "magical" out of the box; discoverability of features (like templates) increases.
- **Cons**: Adds a small dependency footprint (`huh`), though it shares the `bubbletea` base.

## Security & Reliability
- Exiting the wizard (Ctrl+C) gracefully aborts the command without side effects.
- Input validation (e.g., ensuring name is not empty or doesn't containt spaces if not supported) is handled inline by `huh`.

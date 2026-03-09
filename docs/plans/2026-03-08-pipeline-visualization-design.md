# Advanced Pipeline Visualization Design

## Problem
AgentMux supports pipeline workflows (sequences of agents running sequentially or in parallel, defined in yaml configs), but visibility into their execution is limited to text logs. Users cannot easily see what stage a pipeline is currently blocked on, or a summarized graph of its execution.

## Proposed Solution
Create a new Pipeline Visualizer view in the dashboard (or a dedicated `agentmux pipeline view <name>` command). It will display the directed acyclic graph (DAG) of the pipeline logic.

### Visual Flow Representation
For a pipeline like: "claude runs -> then aider AND codex run in parallel -> then github-copilot-cli runs":

```
    [claude] ✓
        |
   +----+----+
   |         |
[aider] ⏳ [codex] ⚙️
   |         |
   +----+----+
        |
   [copilot] ⏸️
```

### Implementation Details:
1. **Pipeline State Tracking:**
   Currently, the orchestrator (`internal/orchestrator`) manages pipeline execution. We need to expose its live state (Pending, Running, Succeeded, Failed) via a channel or shared memory that the TUI can subscribe to.
2. **Data Structure:**
   The `Pipeline` struct in `internal/config` will need to retain graph-level metadata or ordering index so the UI knows how to render trees or sequential stages.
3. **Rendering:**
   Create a new component `internal/tui/components/pipeline_graph.go`. It can use standard Unicode box-drawing characters (or `lipgloss` tools) to render the graph structure dynamically when the user selects a running pipeline from the main session tree.

## Trade-offs
- **Pros**: Immediate situational awareness for complex, multi-agent builds. 
- **Cons**: Representing complex or highly branching DAGs in text-based terminals gets visually messy very quickly. We should enforce maximum horizontal parallel columns (e.g., max 3) or fall back to a simple vertical sequence list with indentation for parallel jobs.

## Extension Idea (Not Day 1)
Using the ANSI-rendering we can even make the edges (lines) light up Green, Gray, or Red to denote data flow success/failure states.

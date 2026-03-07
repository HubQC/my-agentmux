# Advanced Gemini Integration

AgentMux provides native, deep integration with the Google Gemini CLI. It automatically reads your `~/.gemini/settings.json` configuration file to bring your active MCP (Model Context Protocol) servers into the AgentMux workflow and dashboard.

## 1. Launching Gemini Sessions

When you start a session with the Gemini CLI agent, AgentMux automatically parses your configured MCP servers and binds their metadata to the running AgentMux session.

```bash
# General syntax
agentmux start <session-name> -t gemini

# Example: Starting a new Gemini session
agentmux start coding-assistant -t gemini
```

### Pro-Tip: Create Custom Reusable Agents

You can define a custom Gemini agent with specific prompts or environment configurations. Create a Markdown file at `~/.agentmux/agents/gemini-reviewer.md`:

```yaml
---
description: Code reviewer using Gemini
agent_type: gemini
env:
  REVIEW_MODE: "strict"
---

You are an expert code reviewer. Please review my code thoroughly.
```

Then start the agent with:
```bash
agentmux start reviewer -t gemini-reviewer
```

## 2. Inspecting Your Configurations

Unsure which MCP servers are actively hooked into your Gemini environment? AgentMux provides a dedicated command to interrogate your setup.

### `agentmux gemini`
Run this command for a detailed introspection of your `~/.gemini/settings.json` configurations. It lists out all configured MCP servers along with their launch commands.

```text
$ agentmux gemini

🚀 Gemini Environment Overview
========================================

🔌 MCP Servers (6 configured):
  --------------------------------------
  SERVER NAME           COMMAND   ARGS
  chrome-devtools       node      ...
  filesystem            node      ...
  github                node      ...

💡 To launch an agent session with Gemini:
   agentmux launch my-gemini --agent gemini

   Your specific MCP settings are passed implicitly.
```

## 3. Real-Time Dashboard

When you run `agentmux dashboard`, AgentMux intercepts running Gemini agent processes and automatically decorates the UI with their precise capabilities metrics. If you launched a Gemini session, the dashboard will actively render the configured MCP contexts:

```text
▼ gemini-project
  ● coding-assistant
    🔌 MCP: filesystem, chrome-devtools, github
```

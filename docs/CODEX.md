# Advanced Codex Integration

AgentMux provides native, deep integration with OpenAI Codex CLI. It can read your `~/.codex/config.toml` (and project-level `.codex/config.toml` overrides) to bring Codex's profile system and multi-agent roles directly into the AgentMux workflow and dashboard.

## 1. Launching Codex with Customized Profiles

Codex allows you to map combinations of models, reasoning efforts, and personalities to **Profiles** (e.g., `gpt-5.3-codex`). 

To launch a Codex session utilizing a specific profile, pass the `--profile` argument to the internal Codex binary using the `-a` (args) flag:

```bash
# General syntax
agentmux start <session-name> -t codex -a "--profile <profile-name>"

# Example: Starting a deep-reasoning GPT-5.3 session
agentmux start advanced-coder -t codex -a "--profile gpt-5.3-codex"
```

### Pro-Tip: Create Custom Reusable Agents

Constantly typing arguments is tedious. You can map your favorite Codex profiles into first-class custom agents in AgentMux. Create a Markdown file at `~/.agentmux/agents/reviewer.md` (or in your project's `.agentmux/config.yaml`):

```yaml
---
description: High reasoning code reviewer using GPT-5.3
agent_type: codex
args:
  - "--profile"
  - "gpt-5.3-codex"
---

You are an expert, meticulous code reviewer focusing on correctness.
```

Now you can launch it instantly:
```bash
agentmux start pr-reviewer -t reviewer
```

## 2. Choosing the Right Configuration

Unsure which profile or sub-agent to use? AgentMux provides two ways to inspect your current Codex configurations:

### `agentmux agents`
This command prints a table of all available base agents, followed by a tabulated view of all your Codex profiles and sub-agents. 

### `agentmux codex`
Run this dedicated command for a deep-dive, advisory view of your Codex configuration. It evaluates your available profiles and recommends usage based on reasoning effort and description.

```text
$ agentmux codex

💡 To start a session with a profile, run:
   agentmux start my-session -t codex -a "--profile <profile_name>"
```

## 3. Real-Time Dashboard

When you run `agentmux dashboard`, AgentMux intercepts running Codex agent processes and automatically decorates the UI with their profile metrics. If you launched codex with `--profile gpt-5.3-codex`, the dashboard will highlight this, along with active MCP modules and multi-agent roles:

```text
▼ codex-project
  ● deep-coder
    ↳ [gpt-5.3-codex] (Reasoning: high) 🤖 Multi-Agent
      🔌 MCP: filesystem, chrome-devtools, github
```

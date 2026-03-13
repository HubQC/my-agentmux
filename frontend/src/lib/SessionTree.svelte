<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let sessions = [];
  export let selectedSession = null;

  $: groupedSessions = sessions.reduce((groups, session) => {
    const group = session.group || 'Default';
    if (!groups[group]) groups[group] = [];
    groups[group].push(session);
    return groups;
  }, {});

  function selectSession(session) {
    dispatch('select', session);
  }
</script>

<div class="session-tree">
  {#each Object.entries(groupedSessions) as [group, groupSessions]}
    <div class="group">
      <div class="group-header">{group}</div>
      <div class="group-content">
        {#each groupSessions as session}
          <div 
            class="session-node" 
            class:selected={selectedSession && selectedSession.name === session.name}
            on:click={() => selectSession(session)}
          >
            <span class="status-badge {session.status}"></span>
            <span class="session-name">{session.name}</span>
            <span class="agent-type">{session.agent_type}</span>
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .session-tree {
    background: #1b2636;
    height: 100%;
    overflow-y: auto;
    border-right: 1px solid #2d3e50;
    color: #cbd5e1;
    font-family: ui-sans-serif, system-ui, -apple-system, sans-serif;
  }

  .group-header {
    padding: 8px 12px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    color: #64748b;
    background: #16202c;
  }

  .session-node {
    display: flex;
    align-items: center;
    padding: 8px 16px;
    cursor: pointer;
    gap: 10px;
    transition: background 0.2s;
    font-size: 0.875rem;
  }

  .session-node:hover {
    background: #2d3e50;
  }

  .session-node.selected {
    background: #3b82f6;
    color: white;
  }

  .status-badge {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-badge.running { background: #22c55e; }
  .status-badge.stopped { background: #94a3b8; }
  .status-badge.error { background: #ef4444; }

  .session-name {
    flex-grow: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .agent-type {
    font-size: 0.75rem;
    background: #0f172a;
    padding: 2px 6px;
    border-radius: 4px;
    color: #94a3b8;
  }

  .selected .agent-type {
    background: rgba(0, 0, 0, 0.2);
    color: white;
  }
</style>

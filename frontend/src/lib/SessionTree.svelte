<script>
  import { createEventDispatcher } from 'svelte';
  import { pinnedSessions, togglePin } from '../stores/pinned';
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

  function handleDragStart(event, session) {
    event.dataTransfer.setData('text/plain', session.name);
    event.dataTransfer.effectAllowed = 'move';
  }

  async function handleDrop(event, targetGroup) {
    event.preventDefault();
    const sessionName = event.dataTransfer.getData('text/plain');
    if (window.go && window.go.desktop) {
      await window.go.desktop.SessionService.MoveToGroup(sessionName, targetGroup);
    }
  }

  function handleDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }

  async function stopGroup(groupName) {
    if (confirm(`Stop all agents in "${groupName}"?`)) {
      if (window.go && window.go.desktop) {
        await window.go.desktop.SessionService.StopGroup(groupName);
      }
    }
  }
</script>

<div class="session-tree">
  {#each Object.entries(groupedSessions) as [group, groupSessions]}
    <div 
      class="group" 
      on:dragover={handleDragOver} 
      on:drop={(e) => handleDrop(e, group)}
      role="group"
      aria-label="Session Group"
    >
      <div class="group-header">
        <span class="group-name">{group}</span>
        <button class="group-action" on:click|stopPropagation={() => stopGroup(group)} title="Stop All">
          ⏹
        </button>
      </div>
      <div class="group-content">
        {#each groupSessions as session}
          <div 
            class="session-node" 
            class:selected={selectedSession && selectedSession.name === session.name}
            draggable="true"
            on:dragstart={(e) => handleDragStart(e, session)}
            on:click={() => selectSession(session)}
            on:keydown={(e) => e.key === 'Enter' && selectSession(session)}
            role="button"
            tabindex="0"
          >
            <span class="status-badge {session.status}"></span>
            <span class="session-name">{session.name}</span>
            
            <div class="node-actions">
              <button 
                class="pin-btn" 
                class:pinned={$pinnedSessions.includes(session.name)}
                on:click|stopPropagation={() => togglePin(session.name)}
                title="Pin to Grid"
              >
                📌
              </button>
              <span class="agent-type">{session.agent_type}</span>
            </div>
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
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .group-action {
    background: none;
    border: none;
    color: #64748b;
    cursor: pointer;
    font-size: 0.75rem;
    opacity: 0;
    transition: opacity 0.2s;
  }

  .group:hover .group-action {
    opacity: 1;
  }

  .group-action:hover {
    color: #ef4444;
  }

  .session-node {
    display: flex;
    align-items: center;
    padding: 8px 16px;
    cursor: pointer;
    gap: 10px;
    transition: background 0.2s;
    font-size: 0.875rem;
    outline: none;
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

  .node-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pin-btn {
    background: none;
    border: none;
    font-size: 0.75rem;
    cursor: pointer;
    opacity: 0.3;
    transition: opacity 0.2s, transform 0.2s;
    filter: grayscale(1);
  }

  .session-node:hover .pin-btn, .pin-btn.pinned {
    opacity: 1;
  }

  .pin-btn.pinned {
    filter: none;
    transform: rotate(-45deg);
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

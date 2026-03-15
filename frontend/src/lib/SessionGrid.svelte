<script>
  import { pinnedSessions } from '../stores/pinned';
  import Terminal from './Terminal.svelte';

  export let activeSessionName = '';

  $: gridColumns = $pinnedSessions.length > 2 ? 2 : $pinnedSessions.length;
  $: gridRows = Math.ceil($pinnedSessions.length / 2);
</script>

<div 
  class="session-grid" 
  style="--cols: {gridColumns}; --rows: {gridRows}"
>
  {#each $pinnedSessions as name}
    <div class="grid-item" class:active={activeSessionName === name}>
      <div class="item-header">
        <span class="session-name">{name}</span>
      </div>
      <div class="terminal-container">
        <Terminal sessionName={name} />
      </div>
    </div>
  {/each}
  
  {#if $pinnedSessions.length === 0}
    <div class="empty-grid">
      <p>No sessions pinned. Click the 📌 icon in the sidebar to add sessions to this grid.</p>
    </div>
  {/if}
</div>

<style>
  .session-grid {
    display: grid;
    grid-template-columns: repeat(var(--cols, 1), 1fr);
    grid-template-rows: repeat(var(--rows, 1), 1fr);
    gap: 2px;
    background: #2d3e50;
    height: 100%;
    width: 100%;
    overflow: hidden;
  }

  .grid-item {
    background: #0f172a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    border: 2px solid transparent;
  }

  .grid-item.active {
    border-color: #3b82f6;
  }

  .item-header {
    background: #16202c;
    padding: 4px 12px;
    font-size: 0.75rem;
    font-weight: 600;
    color: #64748b;
    border-bottom: 1px solid #2d3e50;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .terminal-container {
    flex-grow: 1;
    overflow: hidden;
  }

  .empty-grid {
    grid-column: 1 / -1;
    grid-row: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #64748b;
    background: #0f172a;
  }
</style>

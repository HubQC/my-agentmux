<script>
  import { onMount, onDestroy } from 'svelte';
  import SessionTree from './lib/SessionTree.svelte';
  import Terminal from './lib/Terminal.svelte';
  import SessionGrid from './lib/SessionGrid.svelte';
  import { sessions, startPolling, stopPolling, loading } from './stores/sessions';
  import { pinnedSessions } from './stores/pinned';

  let selectedSession = null;
  let viewMode = 'focused'; // 'focused' or 'grid'

  onMount(() => {
    // Wait for Wails to be ready before starting polling
    const checkWails = setInterval(() => {
      if (window.go && window.go.desktop) {
        clearInterval(checkWails);
        startPolling();
      }
    }, 100);
  });

  onDestroy(() => {
    stopPolling();
  });

  function handleSelect(event) {
    selectedSession = event.detail;
    if (viewMode === 'grid') {
      // In grid mode, we don't automatically switch back to focused
      // but we update selectedSession for potential highlighting
    }
  }

  function setViewMode(mode) {
    viewMode = mode;
  }
</script>

<main>
  <div class="sidebar">
    <div class="logo">
      AgentMux
      <span class="version">v0.1.0</span>
    </div>
    
    <div class="tree-container">
      {#if $loading}
        <div class="loading">Loading sessions...</div>
      {:else}
        <SessionTree 
          sessions={$sessions} 
          {selectedSession} 
          on:select={handleSelect} 
        />
      {/if}
    </div>

    <div class="sidebar-footer">
      <button class="new-btn" on:click={() => alert('New Agent creation coming soon')}>
        + New Agent
      </button>
    </div>
  </div>

  <div class="content">
    <header>
      <div class="header-left">
        {#if viewMode === 'focused' && selectedSession}
          <div class="session-title">
            <span class="status-dot {selectedSession.status}"></span>
            <h2>{selectedSession.name}</h2>
            <span class="type-pill">{selectedSession.agent_type}</span>
          </div>
        {:else if viewMode === 'grid'}
          <div class="session-title">
            <h2>Command Center</h2>
            <span class="type-pill">{$pinnedSessions.length} Pinned</span>
          </div>
        {:else}
          <div class="logo-text">AgentMux</div>
        {/if}
      </div>

      <div class="header-center">
        <div class="view-toggle">
          <button 
            class:active={viewMode === 'focused'} 
            on:click={() => setViewMode('focused')}
          >
            Focus
          </button>
          <button 
            class:active={viewMode === 'grid'} 
            on:click={() => setViewMode('grid')}
          >
            Grid
          </button>
        </div>
      </div>

      <div class="header-right">
        {#if viewMode === 'focused' && selectedSession}
          <div class="actions">
            <button class="action-btn stop" on:click={() => window.go.desktop.SessionService.StopSession(selectedSession.name)}>
              Stop
            </button>
          </div>
        {/if}
      </div>
    </header>

    <div class="main-view">
      {#if viewMode === 'grid'}
        <SessionGrid activeSessionName={selectedSession?.name} />
      {:else if selectedSession}
        <div class="terminal-wrapper">
          <Terminal sessionName={selectedSession.name} />
        </div>
      {:else}
        <div class="empty-state">
          <div class="empty-icon">⌘</div>
          <h3>No Session Selected</h3>
          <p>Select an agent from the sidebar or start a new one to begin.</p>
        </div>
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    background: #0f172a;
    color: #cbd5e1;
    font-family: ui-sans-serif, system-ui, -apple-system, sans-serif;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }

  main {
    display: flex;
    height: 100vh;
    width: 100vw;
    text-align: left;
  }

  .sidebar {
    width: 260px;
    display: flex;
    flex-direction: column;
    border-right: 1px solid #2d3e50;
    background: #1b2636;
    flex-shrink: 0;
    height: 100%;
  }

  .logo {
    padding: 20px;
    font-size: 1.25rem;
    font-weight: 800;
    color: #3b82f6;
    display: flex;
    align-items: center;
    gap: 8px;
    border-bottom: 1px solid #2d3e50;
    flex-shrink: 0;
  }

  .tree-container {
    flex-grow: 1;
    overflow-y: auto;
  }

  .version {
    font-size: 0.625rem;
    background: #0f172a;
    padding: 2px 6px;
    border-radius: 999px;
    color: #64748b;
  }

  .content {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    background: #0f172a;
    overflow: hidden;
    height: 100%;
  }

  header {
    padding: 0 20px;
    border-bottom: 1px solid #2d3e50;
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 50px;
    background: #1b2636;
    flex-shrink: 0;
  }

  .header-left, .header-right {
    flex: 1;
    display: flex;
    align-items: center;
  }

  .header-right {
    justify-content: flex-end;
  }

  .session-title {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .session-title h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
  }

  .logo-text {
    font-weight: 700;
    color: #64748b;
  }

  .status-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
  }

  .status-dot.running { background: #22c55e; }
  .status-dot.stopped { background: #94a3b8; }

  .type-pill {
    font-size: 0.75rem;
    background: #0f172a;
    padding: 2px 8px;
    border-radius: 4px;
    color: #64748b;
  }

  .view-toggle {
    display: flex;
    background: #0f172a;
    padding: 2px;
    border-radius: 6px;
    border: 1px solid #2d3e50;
  }

  .view-toggle button {
    background: none;
    border: none;
    color: #64748b;
    padding: 4px 16px;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .view-toggle button.active {
    background: #3b82f6;
    color: white;
  }

  .main-view {
    flex-grow: 1;
    position: relative;
    overflow: hidden;
  }

  .terminal-wrapper {
    height: 100%;
    width: 100%;
  }

  .empty-state {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 40px;
    color: #64748b;
  }

  .empty-icon {
    font-size: 4rem;
    margin-bottom: 20px;
    opacity: 0.2;
  }

  .sidebar-footer {
    padding: 16px;
    border-top: 1px solid #2d3e50;
    flex-shrink: 0;
  }

  .new-btn {
    width: 100%;
    background: #3b82f6;
    color: white;
    border: none;
    padding: 10px;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .new-btn:hover {
    background: #2563eb;
  }

  .action-btn {
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid transparent;
  }

  .action-btn.stop {
    background: #450a0a;
    color: #f87171;
    border-color: #7f1d1d;
  }

  .action-btn.stop:hover {
    background: #7f1d1d;
    color: white;
  }

  .loading {
    padding: 20px;
    text-align: center;
    color: #64748b;
    font-style: italic;
  }
</style>

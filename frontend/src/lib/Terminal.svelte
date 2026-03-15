<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from 'xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import 'xterm/css/xterm.css';

  export let sessionName = '';
  
  let terminalElement;
  let term;
  let fitAddon;
  let unsubscribe;
  let resizeObserver;

  $: if (sessionName && term) {
    handleSessionChange(sessionName);
  }

  function handleSessionChange(name) {
    term.clear();
    term.reset();
    
    if (window.go && window.go.desktop) {
      window.go.desktop.TerminalService.AttachTerminal(name);
    }
    
    const eventName = `terminal:output:${name}`;
    if (unsubscribe) unsubscribe();
    
    if (window.runtime) {
      unsubscribe = window.runtime.EventsOn(eventName, (data) => {
        term.write('\x1b[H\x1b[J' + data);
      });
    }
  }

  onMount(() => {
    term = new Terminal({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: '"MesloLGS NF", "JetBrainsMono Nerd Font", "Menlo", "Monaco", "Courier New", monospace',
      theme: {
        background: '#0f172a',
        foreground: '#cbd5e1',
        cursor: '#3b82f6',
        selectionBackground: 'rgba(59, 130, 246, 0.3)',
        black: '#000000',
        red: '#ff5555',
        green: '#50fa7b',
        yellow: '#f1fa8c',
        blue: '#6272a4',
        magenta: '#ff79c6',
        cyan: '#8be9fd',
        white: '#bbbbbb',
        brightBlack: '#44475a',
        brightRed: '#ff6e6e',
        brightGreen: '#69ff94',
        brightYellow: '#ffffa5',
        brightBlue: '#d6acff',
        brightMagenta: '#ff92df',
        brightCyan: '#a4ffff',
        brightWhite: '#ffffff'
      },
      convertEol: true,
      allowProposedApi: true
    });

    fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(new WebLinksAddon());

    term.open(terminalElement);
    
    // Initial fit
    setTimeout(() => fitAddon.fit(), 50);

    // Use ResizeObserver for more robust resizing than window.resize
    resizeObserver = new ResizeObserver(() => {
      if (fitAddon) {
        fitAddon.fit();
      }
    });
    resizeObserver.observe(terminalElement);

    // Listen for data from frontend (typing)
    term.onData((data) => {
      if (sessionName && window.go && window.go.desktop) {
        window.go.desktop.TerminalService.SendInput(sessionName, data);
      }
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
    if (resizeObserver) resizeObserver.disconnect();
    if (sessionName && window.go && window.go.desktop) {
      window.go.desktop.TerminalService.DetachTerminal(sessionName);
    }
    if (term) term.dispose();
  });
</script>

<div class="terminal-container">
  <div bind:this={terminalElement} class="terminal-instance"></div>
</div>

<style>
  .terminal-container {
    flex-grow: 1;
    background: #0f172a;
    padding: 0; /* Remove padding to ensure xterm fills everything */
    height: 100%;
    width: 100%;
    overflow: hidden;
    display: flex; /* Flexbox to ensure instance fills container */
  }

  .terminal-instance {
    height: 100%;
    width: 100%;
    text-align: left; /* Ensure text is left-aligned */
  }

  :global(.xterm) {
    padding: 8px; /* Padding inside the terminal itself */
  }

  :global(.xterm-viewport) {
    background-color: #0f172a !important;
  }
</style>

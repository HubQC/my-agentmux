import { writable } from 'svelte/store';

export const sessions = writable([]);
export const loading = writable(true);
export const error = writable(null);

let pollInterval;

export function startPolling() {
  if (pollInterval) return;

  const fetchSessions = async () => {
    try {
      if (window.go && window.go.desktop && window.go.desktop.SessionService) {
        const list = await window.go.desktop.SessionService.ListSessions();
        sessions.set(list || []);
        loading.set(false);
      }
    } catch (e) {
      console.error('Failed to fetch sessions:', e);
      error.set(e.message);
    }
  };

  fetchSessions();
  pollInterval = setInterval(fetchSessions, 2000);
}

export function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
}

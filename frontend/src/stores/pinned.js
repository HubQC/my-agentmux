import { writable } from 'svelte/store';

export const pinnedSessions = writable([]);

export function togglePin(sessionName) {
  pinnedSessions.update(list => {
    if (list.includes(sessionName)) {
      return list.filter(n => n !== sessionName);
    }
    return [...list, sessionName];
  });
}

# Design: Desktop Grid View and Advanced Tree Organization

**Date:** 2026-03-12
**Status:** Approved
**Topic:** Workflow management improvements for the AgentMux Desktop App.

## 1. Overview
Enhance the desktop application to support managing multiple agents simultaneously through an auto-tiling grid and improved sidebar organization.

## 2. Goals
- Allow users to monitor and interact with multiple terminal sessions at once.
- Provide intuitive drag-and-drop organization for agent groups.
- Enable bulk actions on groups of agents.

## 3. Architecture & Components

### 3.1 Frontend (Svelte)
- **`stores/pinned.js`**: A new store to track the list of "pinned" session names that should appear in the Grid View.
- **`lib/SessionGrid.svelte`**: A wrapper component that renders multiple `Terminal.svelte` instances in a responsive CSS Grid layout.
- **`lib/SessionTree.svelte` (Enhanced)**:
    - Add HTML5 Drag and Drop attributes (`draggable`, `ondragover`, `ondrop`).
    - Add "Pin" toggle icons to session nodes.
    - Add hover action buttons to group headers.
- **`App.svelte` (Update)**: Add a "View Mode" toggle (Single vs. Grid) in the header.

### 3.2 Backend (Go)
- **`SessionService` (Update)**:
    - Add `StopGroup(groupName string)` to kill all tmux sessions in a specific group.
    - Add `MoveToGroup(sessionName, newGroup string)` to update session metadata.

## 4. Key Workflows

### 4.1 Grid Tiling
1. User clicks the "Pin" icon next to "Agent A" and "Agent B".
2. User toggles "Grid View" in the header.
3. `SessionGrid` calculates a 1x2 or 2x1 layout based on container dimensions.
4. Two `Terminal` instances are rendered, each attached to its respective session.

### 4.2 Group Management
1. User drags "Agent C" node in the tree.
2. User drops it onto "Project Alpha" group header.
3. Frontend calls `go.desktop.SessionService.MoveToGroup`.
4. Sidebar re-renders with the updated hierarchy.

## 5. Visual Design
- **Grid Layout**: Auto-adjusts (1 tile = 100%, 2 tiles = 50/50, 3-4 tiles = 2x2).
- **Active Focus**: The active terminal tile will have a subtle blue border to indicate where keyboard input is directed.
- **DND Feedback**: Group headers will highlight when a session is dragged over them.

## 6. Verification Plan
- **Manual Test**: Verify that pinning 4 sessions renders a 2x2 grid.
- **Manual Test**: Verify that dragging a session to a new group persists after a manual refresh.
- **Build**: Ensure `make desktop` continues to produce a functional binary.

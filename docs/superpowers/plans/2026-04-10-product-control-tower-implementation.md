# MobileCoder Product Control Tower Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reposition MobileCoder around task supervision, then implement the product surfaces and backend state model needed to support a mobile control tower for AI coding tasks.

**Architecture:** Keep the existing `agent -> cloud -> web/mobile` structure, but insert a product layer centered on tasks, task states, task summaries, and notifications. The first implementation should map tasks onto existing sessions, then gradually add richer state, events, and action flows without rewriting the transport layer.

**Tech Stack:** Go, Gorilla WebSocket, React, Next.js, React Router, Capacitor, Supabase REST API

---

### Task 1: Rewrite product-facing messaging

**Files:**
- Modify: `README.md`
- Modify: `docs/user-manual.md`
- Create: `docs/product/product-positioning.md`

- [ ] **Step 1: Replace "remote terminal" messaging with "mobile control tower" messaging**
- [ ] **Step 2: Define target user, core job to be done, and product boundaries**
- [ ] **Step 3: Add a short product page style summary in `docs/product/product-positioning.md`**
- [ ] **Step 4: Review all three documents for consistent language: task, state, supervision, intervention**

### Task 2: Introduce a task domain model

**Files:**
- Create: `cloud/internal/service/task_service.go`
- Modify: `cloud/internal/service/device_service.go`
- Modify: `cloud/internal/db/supabase.go`
- Create: `chat/src/lib/tasks.ts`
- Create: `mobile-app/src/services/tasks.ts`

- [ ] **Step 1: Define a task DTO that wraps session data with product-facing fields**
- [ ] **Step 2: Map existing sessions into task objects on the backend**
- [ ] **Step 3: Expose a backend API that returns task-oriented payloads**
- [ ] **Step 4: Add frontend service helpers that consume tasks instead of raw sessions**
- [ ] **Step 5: Verify tasks can be listed without breaking current device/session flows**

### Task 3: Add a task state system

**Files:**
- Create: `cloud/internal/service/task_state.go`
- Modify: `cloud/internal/handler/ws_handler.go`
- Modify: `cloud/internal/ws/hub.go`
- Create: `cloud/internal/service/task_state_test.go`

- [ ] **Step 1: Define the initial state enum: Running, WaitingForInput, Completed, Failed, Disconnected, IdleTooLong, NeedsAttention**
- [ ] **Step 2: Write failing tests for state derivation from existing session and terminal events**
- [ ] **Step 3: Implement minimal state mapping using available signals**
- [ ] **Step 4: Update WebSocket handling so new outputs refresh task state**
- [ ] **Step 5: Re-run focused Go tests and verify the state model is stable**

### Task 4: Build a task-first homepage

**Files:**
- Modify: `chat/src/app/page.tsx`
- Modify: `chat/src/app/devices/page.tsx`
- Create: `chat/src/app/tasks/page.tsx`
- Create: `chat/src/components/tasks/task-card.tsx`
- Create: `chat/src/components/tasks/task-list.tsx`
- Create: `chat/src/components/tasks/task-filters.tsx`
- Modify: `mobile-app/src/App.tsx`
- Create: `mobile-app/src/pages/TasksPage.tsx`

- [ ] **Step 1: Move the primary navigation entry from devices to tasks**
- [ ] **Step 2: Build task cards showing title, repo, tool, state, summary, last activity**
- [ ] **Step 3: Add basic filters for running, waiting, failed, completed**
- [ ] **Step 4: Keep devices accessible as a secondary management page**
- [ ] **Step 5: Verify desktop web and mobile app both load into task-first navigation**

### Task 5: Build task detail as the main work surface

**Files:**
- Create: `chat/src/app/tasks/[taskId]/page.tsx`
- Create: `chat/src/components/tasks/task-detail.tsx`
- Create: `chat/src/components/tasks/task-timeline.tsx`
- Modify: `chat/src/app/terminal/page.tsx`
- Create: `mobile-app/src/pages/TaskDetailPage.tsx`
- Modify: `mobile-app/src/pages/TerminalPage.tsx`

- [ ] **Step 1: Create a task detail page above the terminal layer**
- [ ] **Step 2: Add status header, recent events, summary, and quick actions**
- [ ] **Step 3: Embed terminal output as a lower-priority section**
- [ ] **Step 4: Preserve a clear "open full terminal" path for expert use**
- [ ] **Step 5: Verify users can understand task state without reading raw logs**

### Task 6: Add lightweight intervention actions

**Files:**
- Modify: `chat/src/app/components/Terminal.tsx`
- Create: `chat/src/components/tasks/task-actions.tsx`
- Modify: `mobile-app/src/components/Terminal.tsx`
- Create: `mobile-app/src/components/TaskActions.tsx`
- Modify: `cloud/internal/handler/ws_handler.go`

- [ ] **Step 1: Define the first action set: Continue, Stop, Retry, Reply, Open terminal**
- [ ] **Step 2: Reuse existing terminal input transport for quick actions where possible**
- [ ] **Step 3: Add action buttons to task detail pages**
- [ ] **Step 4: Make actions clearly separated from full manual terminal control**
- [ ] **Step 5: Verify quick actions work without forcing full terminal entry**

### Task 7: Add notification infrastructure

**Files:**
- Create: `cloud/internal/service/notification_service.go`
- Modify: `cloud/internal/service/device_service.go`
- Modify: `mobile-app/capacitor.config.ts`
- Create: `mobile-app/src/services/notifications.ts`
- Modify: `docs/superpowers/specs/2026-04-04-android-app-design.md`

- [ ] **Step 1: Define a notification event model tied to task state changes**
- [ ] **Step 2: Add storage for push-capable device tokens**
- [ ] **Step 3: Wire mobile app notification registration**
- [ ] **Step 4: Trigger notifications on completed, waiting, stalled, disconnected**
- [ ] **Step 5: Verify notifications deep-link back into the relevant task**

### Task 8: Add task history and multi-task management

**Files:**
- Modify: `cloud/internal/db/supabase.go`
- Modify: `cloud/internal/service/task_service.go`
- Create: `chat/src/app/tasks/history/page.tsx`
- Create: `mobile-app/src/pages/TaskHistoryPage.tsx`

- [ ] **Step 1: Persist enough task metadata to query recent task history**
- [ ] **Step 2: Add views for active, attention-needed, and completed tasks**
- [ ] **Step 3: Support archive or hide completed tasks**
- [ ] **Step 4: Add recent-history browsing for replay and continuation**
- [ ] **Step 5: Verify multi-task users can manage more than one concurrent task cleanly**

### Task 9: Prepare cross-tool abstraction

**Files:**
- Modify: `agent/cmd/client/main.go`
- Create: `cloud/internal/service/tool_identity.go`
- Modify: `chat/src/components/tasks/task-card.tsx`
- Modify: `mobile-app/src/components/TaskActions.tsx`

- [ ] **Step 1: Standardize how the agent reports Claude, Codex, and Cursor identity**
- [ ] **Step 2: Surface tool identity in task views**
- [ ] **Step 3: Normalize tool-specific session metadata into one task format**
- [ ] **Step 4: Verify users can filter tasks by AI tool**

### Task 10: Final product validation

**Files:**
- Modify: `README.md`
- Modify: `docs/user-manual.md`
- Create: `docs/product/task-supervision-playbook.md`

- [ ] **Step 1: Validate that onboarding, homepage, task detail, and notifications all tell one coherent story**
- [ ] **Step 2: Document the intended control-tower workflow for internal use and demos**
- [ ] **Step 3: Review milestone sequencing and confirm there are no dependency gaps**
- [ ] **Step 4: Hand off for execution starting from Task 1, then Task 2, then Task 3**

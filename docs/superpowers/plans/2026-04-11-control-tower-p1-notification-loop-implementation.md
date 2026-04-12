# MobileCoder Control Tower P1 Notification Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first usable notification loop to MobileCoder so a user can receive in-app notifications and Android local notifications for completed, waiting, stalled, and disconnected tasks.

**Architecture:** Persist notification events in `cloud`, expose a small authenticated notification API, and let the mobile app run an app-mediated delivery loop that fetches unread items and converts newly delivered events into Android local notifications. Keep web and mobile aligned through the same backend notification records and deep-link all notifications into task detail.

**Tech Stack:** Go, Supabase REST API, React, React Router, Capacitor, Android local notifications

---

### Task 1: Persist notification records in `cloud`

**Files:**
- Create: `cloud/internal/service/notification_service.go`
- Create: `cloud/internal/service/notification_service_test.go`
- Modify: `cloud/internal/db/supabase.go`

- [ ] **Step 1: Add notification DB model support in `cloud/internal/db/supabase.go`**
- [ ] **Step 2: Add failing tests for create/list/mark-read behavior in `cloud/internal/service/notification_service_test.go`**
- [ ] **Step 3: Implement a notification service that creates persisted records and enforces the four allowed event types**
- [ ] **Step 4: Add simple retention and dedupe helpers in the service**
- [ ] **Step 5: Run `env GOCACHE=/tmp/mobilecoder-go-cache go test ./internal/service ./internal/db` in `cloud`**
- [ ] **Step 6: Commit with message `feat: persist control tower notifications`**

### Task 2: Emit notification events from task and device state changes

**Files:**
- Modify: `cloud/internal/service/task_service.go`
- Modify: `cloud/internal/ws/hub.go`
- Modify: `cloud/internal/service/device_service.go`
- Modify: `cloud/internal/service/task_service_test.go`
- Modify: `cloud/internal/ws/hub_test.go`

- [ ] **Step 1: Add failing tests for the four notification triggers: completed, waiting, idle-too-long, disconnected**
- [ ] **Step 2: Extend task state enrichment so event classification can surface notification-worthy state transitions**
- [ ] **Step 3: Emit persisted notification records only on transition edges, not on repeated unchanged states**
- [ ] **Step 4: Add idle-too-long detection using a fixed 15 minute threshold**
- [ ] **Step 5: Re-run focused Go tests for task service and hub**
- [ ] **Step 6: Commit with message `feat: emit task notification events`**

### Task 3: Add authenticated notification API routes

**Files:**
- Create: `cloud/internal/handler/notification_handler.go`
- Create: `cloud/internal/handler/notification_handler_test.go`
- Modify: `cloud/cmd/server/main.go`
- Modify: `cloud/internal/handler/auth_helpers.go` if shared auth helpers are needed

- [ ] **Step 1: Add failing handler tests for list, mark-read, and mark-all-read endpoints**
- [ ] **Step 2: Implement `GET /api/notifications`, `POST /api/notifications/read`, and `POST /api/notifications/read-all`**
- [ ] **Step 3: Add support for incremental fetch using an optional `since` query parameter**
- [ ] **Step 4: Verify only the authenticated user's records are returned or mutated**
- [ ] **Step 5: Run `env GOCACHE=/tmp/mobilecoder-go-cache go test ./...` in `cloud`**
- [ ] **Step 6: Commit with message `feat: add notification api`**

### Task 4: Build mobile notification runtime and Android local alerts

**Files:**
- Create: `mobile-app/src/services/notifications.ts`
- Create: `mobile-app/src/components/NotificationBell.tsx`
- Create: `mobile-app/src/pages/NotificationsPage.tsx`
- Modify: `mobile-app/src/App.tsx`
- Modify: `mobile-app/src/pages/TasksPage.tsx`
- Modify: `mobile-app/src/pages/TaskDetailPage.tsx`
- Modify: `mobile-app/capacitor.config.ts`
- Modify: `mobile-app/package.json` only if a Capacitor plugin dependency is required

- [ ] **Step 1: Add a notification client that fetches incrementally from `/api/notifications` and tracks unread state**
- [ ] **Step 2: Add Android local notification permission and trigger plumbing using Capacitor-compatible APIs**
- [ ] **Step 3: Add a notification center page plus unread badge entry in the mobile task-first chrome**
- [ ] **Step 4: Deep-link notification taps into `/tasks/:taskId`**
- [ ] **Step 5: Ensure the runtime degrades to in-app notifications if local notification permission is denied**
- [ ] **Step 6: Run `npm run build` in `mobile-app`**
- [ ] **Step 7: Commit with message `feat: add mobile notification loop`**

### Task 5: Add web notification center and unread badge

**Files:**
- Create: `chat/src/lib/notifications.ts`
- Create: `chat/src/app/notifications/page.tsx`
- Create: `chat/src/components/notifications/notification-bell.tsx`
- Create: `chat/src/components/notifications/notification-list.tsx`
- Modify: `chat/src/app/tasks/page.tsx`
- Modify: `chat/src/app/tasks/[taskId]/page.tsx`
- Modify: `chat/src/app/page.tsx` if root routing needs notification entry exposure

- [ ] **Step 1: Add notification API client helpers for web**
- [ ] **Step 2: Add a minimal notification list page with unread and read states**
- [ ] **Step 3: Add an unread badge entry from the task-first surfaces**
- [ ] **Step 4: Route clicks into task detail and mark items read on open**
- [ ] **Step 5: Run `npm run build` in `chat`**
- [ ] **Step 6: Commit with message `feat: add web notification center`**

### Task 6: Final integration, QA, and cleanup

**Files:**
- Modify: `README.md` if notification usage needs a short update
- Modify: `docs/user-manual.md` if the notification flow needs documentation

- [ ] **Step 1: Start `cloud`, `chat`, and `mobile-app` on isolated local ports**
- [ ] **Step 2: Create or reuse a test account, device, and sessions to generate notification events**
- [ ] **Step 3: Verify the four supported event types appear in-app and deep-link correctly**
- [ ] **Step 4: Verify Android local notification behavior while the app is alive in foreground/background-resident mode**
- [ ] **Step 5: Clean generated build artifacts and verify `git status --short` only shows intended source/doc changes**
- [ ] **Step 6: Commit with message `feat: ship p1 notification loop`**

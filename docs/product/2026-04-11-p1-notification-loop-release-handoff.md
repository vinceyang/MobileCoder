# MobileCoder P1 Notification Loop Release Handoff

## Scope

This release finishes the first usable notification loop for the control tower direction:

- persisted notifications in `cloud`
- transition-based event emission for:
  - task completed
  - waiting for input
  - idle too long
  - agent disconnected
- authenticated notification APIs
- web notification center and unread badge
- mobile notification center and local-notification bridge

Branch:

- `docs/product-control-tower`

Key commits:

- `eb79b48` `feat: ship p1 notification loop`
- `720ee9e` `feat: emit task notification events`
- `e5673b9` `feat: add notification api`
- `728b979` `feat: persist control tower notifications`

## Database Change

This release requires the notifications table schema:

- [cloud/sql/2026-04-11_notifications.sql](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/cloud/sql/2026-04-11_notifications.sql)

Status at handoff:

- applied to the current Supabase project

## Environment Expectations

`cloud` still expects:

- `SUPABASE_PROJECT_URL`
- `SUPABASE_API_KEY`
- `JWT_SECRET`
- `PORT`

`mobile-app` notification bridge depends on:

- `@capacitor/local-notifications`

For web preview and local dev, the notification runtime gracefully falls back to in-app only.

## Verification Completed

Backend:

- `cd cloud && env GOCACHE=/tmp/mobilecoder-go-cache go test ./...`

Web:

- `cd chat && npm run build`

Mobile:

- `cd mobile-app && npm install @capacitor/local-notifications@^8.0.2 && npm run build`

End-to-end smoke verification completed against local `cloud` on `8091` with real Supabase data:

- task list loads for authenticated user
- `GET /api/notifications` returns persisted records
- `POST /api/notifications/read` works
- `POST /api/notifications/read-all` works
- all four notification event types were generated and persisted

Smoke account used during verification:

- `qa-notify-20260411@example.com`

Smoke task used during verification:

- `qa25dev000000001:notify-smoke`

## Release Notes

User-visible changes:

- task-first surfaces now expose a notification entry
- web has a notification center page
- mobile has a notification center page
- mobile can trigger local notifications while the app is alive
- notification clicks route back into task detail

## Known Non-Blocking Issues

- password hashing is still bare `sha256`; acceptable for current private/small-circle release, not for broader public rollout
- `chat` build still shows existing React hooks lint warnings
- Next.js still prints the existing lockfile patch warning during build, but does not block output
- mobile local notifications are designed for foreground/background-resident operation, not hard-killed delivery

## Deployment Order

1. deploy `cloud`
2. deploy `chat`
3. deploy/update `mobile-app`
4. verify login
5. verify `/tasks`
6. verify `/notifications`
7. verify one real task can produce at least one notification

## Post-Deploy Smoke Checks

Minimum checks:

- register or log in successfully
- open `/tasks` and confirm authenticated task list loads
- open `/notifications` and confirm page loads
- mark a notification read
- mark all notifications read
- create one waiting/completed/disconnected scenario and verify a new notification appears

## Files Most Relevant To This Release

Backend:

- [cloud/internal/service/notification_service.go](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/cloud/internal/service/notification_service.go)
- [cloud/internal/service/task_service.go](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/cloud/internal/service/task_service.go)
- [cloud/internal/handler/notification_handler.go](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/cloud/internal/handler/notification_handler.go)
- [cloud/cmd/server/main.go](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/cloud/cmd/server/main.go)

Web:

- [chat/src/lib/notifications.ts](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/chat/src/lib/notifications.ts)
- [chat/src/app/notifications/page.tsx](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/chat/src/app/notifications/page.tsx)

Mobile:

- [mobile-app/src/services/notifications.ts](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app/src/services/notifications.ts)
- [mobile-app/src/pages/NotificationsPage.tsx](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app/src/pages/NotificationsPage.tsx)
- [mobile-app/src/components/NotificationBell.tsx](/Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app/src/components/NotificationBell.tsx)

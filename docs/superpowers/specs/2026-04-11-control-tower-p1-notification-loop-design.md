# MobileCoder Control Tower P1 Notification Loop Design

**Date:** 2026-04-11

## Summary

This document defines the first notification phase for the MobileCoder control tower direction.

P1 exists to close the biggest product gap left after P0:

> A user can see task state when they open the app, but the app still does not reliably pull them back when a task finishes, stalls, disconnects, or needs input.

This phase is explicitly scoped for personal use and a small number of familiar users.

It is not a team collaboration phase.
It is not a large-scale push infrastructure phase.
It is not a perfect background delivery phase.

## Product Goal

When a task reaches a meaningful supervision boundary, MobileCoder should notify the user quickly enough that they do not need to keep checking the app manually.

At the end of P1, a user should be able to:

- leave the app and still get pulled back when a task matters
- see an in-app notification feed with unread state
- tap a notification and land in the relevant task detail
- distinguish between finished work, waiting work, stalled work, and disconnected work

## Scope Boundary

### Included

- in-app notification center and unread indicator
- Android local notifications triggered by app-side delivery
- backend notification event model
- notification delivery for active mobile clients using the existing cloud service
- deep-link from notification to task detail
- four notification event types only

### Excluded

- team notifications
- iOS push work
- FCM or APNS integration
- guaranteed delivery after the app is fully killed by the OS
- complex notification rules or per-event tuning
- notification batching, digesting, or prioritization systems
- historical analytics dashboards

## Target User

This phase serves the same narrow user as P0:

> A heavy AI coding user who already runs long tasks and wants lightweight supervision while away from the computer.

This user is willing to keep the Android app installed and often leaves it in the foreground or background while work is running.

## Product Thesis

P0 made the app readable.
P1 makes it interrupt-driven.

The product should stop depending on user polling and start closing the loop around the moments that actually matter.

For this user segment, "good enough background delivery while the app is alive" is more valuable right now than waiting for a full FCM integration.

## Event Model

P1 only supports four notification events:

- `task_completed`
- `task_waiting_for_input`
- `task_idle_too_long`
- `agent_disconnected`

Every notification event must include:

- notification id
- task id
- device id
- session name
- event type
- title
- body
- created at
- read at, nullable

## Event Semantics

### `task_completed`

Trigger when:

- a task transitions into the current `completed` state
- or the latest classified event is `completed`

User value:

- they know they can return and inspect or archive the result

Example copy:

- "任务已完成"
- "release-train 已结束，可查看结果"

### `task_waiting_for_input`

Trigger when:

- task state becomes `waiting`
- or recent event classification indicates explicit user confirmation is needed

User value:

- they can intervene before the task sits idle indefinitely

Example copy:

- "任务正在等你回复"
- "fix-auth 需要确认下一步操作"

### `task_idle_too_long`

Trigger when:

- a task remains in an active session state
- but there is no recent output beyond a fixed threshold

P1 threshold:

- 15 minutes with no new activity

User value:

- they find stalled work before losing a long block of time

Example copy:

- "任务长时间无新输出"
- "nightly-tests 15 分钟没有新进展"

### `agent_disconnected`

Trigger when:

- the device or agent connection becomes unavailable while the task is still active or recently active

User value:

- they can decide whether the problem is network, agent crash, or machine sleep

Example copy:

- "Agent 已断开"
- "Acceptance Mac 已离线，任务可能中断"

## Delivery Model

P1 uses a lightweight self-hosted delivery design:

1. backend detects state-changing task events
2. backend persists notification records
3. mobile app keeps a lightweight notification delivery loop while foregrounded or background-resident
4. on receipt, the app writes the item into its local in-app feed and also triggers an Android local notification

This is not true remote push.
It is app-mediated delivery.

That is acceptable for P1 because the explicit product target is personal use and a few familiar users, not large-scale public reliability.

## Why This Delivery Model

### Recommended approach: persisted event queue plus app-side delivery

The backend owns truth.
The app owns presentation and local device notification.

This gives P1:

- one consistent notification record per event
- an in-app notification center without extra modeling
- local Android notification capability without requiring FCM
- a migration path to future real push infrastructure

### Rejected approach: direct WebSocket-only notification stream

This would feel simple at first, but it is too dependent on continuous socket health and gives no durable queue by default.

That is a poor base for unread state and late pickup.

### Rejected approach: periodic blind polling only

Polling alone is acceptable as a transport fallback, but not as the main mental model.

The backend still needs persisted events and delivery semantics.

## Architecture

### Backend

Add a notification service inside `cloud`.

Responsibilities:

- receive task state transitions and classified events
- decide whether a new notification should be emitted
- persist a notification record
- expose unread and recent notification queries
- expose read and mark-all-read mutations

Suggested backend surfaces:

- `GET /api/notifications`
- `POST /api/notifications/read`
- `POST /api/notifications/read-all`

Suggested future-compatible shape:

- optional `since` query parameter for incremental fetch
- optional `unread_only` query parameter

### Mobile app

Add a notification runtime service.

Responsibilities:

- register Android local notification permissions
- keep a delivery loop alive while the app is active or background-resident
- fetch new notification records incrementally
- trigger local notifications for newly delivered unread items
- store unread count in app state
- deep-link taps into task detail

### Web app

Web does not need system push in P1.

It does need:

- in-app notification list
- unread count surface
- open notification -> task detail path

That keeps web and mobile consistent even though Android gets the stronger notification experience first.

## Persistence Model

P1 should persist notifications in backend storage instead of deriving them only at render time.

Each record should include:

- immutable event identity
- task reference
- user reference
- event type
- rendered title/body
- read status
- created timestamp

P1 does not need:

- deletion
- archival policy beyond simple retention
- notification grouping

Suggested retention:

- keep the latest 200 notifications per user

## Deduplication Rules

P1 needs lightweight dedupe so users do not get spammed.

Rules:

- emit only when entering a meaningful new event state
- do not repeatedly emit `task_idle_too_long` for the same task until activity resumes and stalls again
- do not emit duplicate `task_waiting_for_input` without a new waiting edge
- do not re-emit `task_completed` unless a new run meaningfully restarts the task

## Deep-Link Behavior

Every notification click should land at the relevant task detail route.

Required behavior:

- Android local notification tap opens the app to `/tasks/:taskId`
- in-app notification list item tap also opens `/tasks/:taskId`
- if the task no longer exists, route to tasks home and show a soft failure message

## In-App Notification UX

P1 notification UI should stay minimal.

Required elements:

- bell or notification entry in primary app chrome where feasible
- unread count badge
- recent notification list ordered newest first
- unread/read visual distinction
- event type chip
- title, body, timestamp
- tap to open task
- mark single item as read
- mark all as read

P1 does not need:

- categories
- filtering
- search
- notification settings matrix

## Android Notification UX

P1 Android behavior:

- request notification permission when the user reaches an active supervision flow
- show a local notification for each newly delivered unread event
- use one stable channel name for control tower alerts
- tapping the notification opens the relevant task detail

P1 channel copy should be direct:

- channel name: `任务提醒`
- channel description: `任务完成、等待输入、卡住和断线提醒`

## State Inputs

P1 notification triggering depends on existing task state and event classification.

This phase does not require a brand new state engine.

It should reuse:

- current task state mapping
- recent timeline events
- session activity timestamps
- device connectivity changes where already observable

The quality bar is useful and quiet, not perfect semantic understanding.

## Success Criteria

P1 is successful if a user can:

1. leave the app and receive a visible Android notification for one of the four supported event types while the app remains alive
2. open the app and see a durable in-app notification feed with unread state
3. tap any notification and land on the correct task detail
4. avoid repeated duplicate alerts for the same unchanged condition

## Failure Modes To Handle

- app is alive but network is temporarily unavailable
- backend emits duplicate candidate events
- notification references a missing task
- local notification permission is denied
- app resumes after inactivity and must catch up missed notifications

Expected behavior:

- fetch loop resumes cleanly
- missed events still appear in-app because they are persisted
- Android local notifications only fire for newly delivered unread items
- permission denial degrades gracefully to in-app feed only

## Testing Requirements

### Backend

- notification creation on each of the four event types
- dedupe behavior for repeated unchanged states
- unread and mark-read flows
- task link data integrity

### Mobile

- permission request flow
- incremental fetch behavior
- local notification trigger path
- tap deep-link into task detail
- unread badge updates

### Web

- notification list rendering
- unread badge rendering
- open-notification deep-link flow

## Rollout Plan

P1 rollout should happen in this order:

1. backend notification persistence and API
2. mobile in-app feed
3. Android local notification trigger path
4. web in-app feed
5. field testing with personal use and a small number of familiar users

## Non-Goals

P1 is not the final notification architecture.

It intentionally avoids:

- FCM
- APNS
- public-scale delivery guarantees
- team collaboration
- cross-user routing
- advanced preferences

Those belong to later phases once the minimal supervision loop proves daily value.

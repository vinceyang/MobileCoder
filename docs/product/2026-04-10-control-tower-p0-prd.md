# MobileCoder Control Tower P0 PRD

**Date:** 2026-04-10

**Status:** Draft for execution

## Purpose

This document defines the first shippable phase of the MobileCoder control tower direction.

P0 is not the full product vision.

P0 exists to answer one question:

> Will heavy AI coding users come back to MobileCoder because it helps them supervise long-running tasks better than watching a terminal on their laptop?

If P0 cannot answer that question, later roadmap work is premature.

## P0 Product Goal

Turn MobileCoder from a device/session viewer into a task-first supervision tool for active AI coding work.

At the end of P0, a user should be able to:

- open MobileCoder and immediately see active work
- know which task is healthy, blocked, or completed
- identify when a task needs attention
- jump into the correct terminal only when needed

P0 does **not** need to solve:

- push notifications
- full task history
- team sharing
- polished multi-tool abstraction
- complete action system

## Target User

P0 is for one user type only:

> A developer already running long AI coding sessions who wants to supervise them from a phone without reopening the full desktop workflow.

This user likely:

- already trusts Claude Code, Codex, or Cursor enough to delegate real work
- runs tasks longer than 5 minutes
- uses more than one repo or session in a day
- leaves the desk frequently

## P0 User Stories

1. As a heavy AI coding user, I want to see all active tasks in one list, so I know what is happening without clicking through devices.

2. As a user supervising a long-running task, I want a clear state label, so I can tell whether the task is running, waiting, stalled, or done.

3. As a user with multiple sessions, I want the most urgent work highlighted first, so I do not waste time checking the wrong task.

4. As a user who only needs to intervene occasionally, I want to read a short task summary before opening the terminal, so I can decide whether intervention is necessary.

5. As a user who needs exact control, I want a clean path into the full terminal, so power-user flows still work.

## P0 Scope

### Included

- task-first homepage
- task view derived from existing session data
- explicit task states
- simple task summaries
- task detail page
- terminal as secondary entry point
- mobile app and web app both aligned to the same task-first model

### Excluded

- push notifications
- task history and archive
- retry/continue/stop action framework
- FCM/APNS work
- team collaboration
- billing or monetization
- iOS-specific work

## P0 Success Criteria

P0 is successful if a user can do these four things:

1. Open the app and identify the most important active task in under 5 seconds.
2. Distinguish `running`, `waiting`, `completed`, and `attention needed` without reading the raw terminal log.
3. Navigate from the task list into the right task detail and terminal in one step.
4. Supervise multiple sessions without thinking in terms of devices first.

## P0 Screens

### 1. Tasks Home

This becomes the primary entry point.

Required elements:

- page title: "Tasks"
- segmented filters: All, Running, Waiting, Attention, Completed
- task list
- secondary link to Devices

Each task card must include:

- task title
- repo or project path label
- AI tool label if available
- task state badge
- recent summary line
- last activity time
- open task detail affordance

### 2. Task Detail

This becomes the main reading surface.

Required elements:

- task title
- state badge
- repo / device / session metadata
- recent summary
- latest key output snippet
- open terminal button
- back to task list

### 3. Terminal

Terminal remains available, but visually and navigationally becomes subordinate to task detail.

Required elements:

- clear back path to task detail or task list
- connection status
- existing terminal output
- existing input and shortcut controls

### 4. Devices

Devices remain in the product, but are no longer the default home surface.

Required elements:

- list devices
- open device details
- maintain existing binding and session browsing flows

## P0 State Model

P0 only needs a small state system.

### Required states

- `running`
- `waiting`
- `completed`
- `attention`

### Mapping rules

#### `running`

Use when:

- session is active
- recent terminal output continues to arrive
- no obvious user response is required

#### `waiting`

Use when:

- terminal output indicates the agent is waiting for confirmation or input
- known keywords or patterns suggest user intervention is needed

#### `completed`

Use when:

- task output reaches a recognizable completion state
- session becomes inactive after apparent successful completion

#### `attention`

Use when:

- session disconnects unexpectedly
- output stalls for too long
- failure-like terminal conditions are detected

P0 does not need perfect classification. It needs useful classification.

## P0 Summary Model

Each task should expose one summary line.

Sources may include:

- recent terminal output
- session metadata
- simple heuristics from latest lines

P0 summary examples:

- "Running tests in `/Users/me/repo`"
- "Waiting for confirmation before deleting files"
- "Completed and returned to prompt"
- "No output for 12 minutes, check task"

P0 should not depend on advanced LLM summarization.

## Backend Changes Required For P0

### New product DTO

Introduce a task-oriented response object derived from current session data.

Suggested fields:

- `id`
- `title`
- `device_id`
- `device_name`
- `session_name`
- `project_path`
- `tool`
- `state`
- `summary`
- `last_activity_at`

### New API surface

Minimum new endpoints:

- `GET /api/tasks`
- `GET /api/tasks/:id` or an equivalent query-based detail endpoint

These can be wrappers around current session/device services.

### State derivation

State should be computed server-side so web and mobile clients stay consistent.

## Frontend Changes Required For P0

### Web

- add task list route
- route `/` to tasks
- task detail page above terminal
- preserve current device/session screens under secondary navigation

### Mobile

- add `TasksPage`
- route `/` to tasks
- add `TaskDetailPage`
- keep current terminal page intact but change entry flow

### Shared requirements

- same state labels
- same badge colors
- same task card mental model
- same summary structure

## Explicit Non-Goals

To keep P0 sharp, do **not** add:

- push notifications
- FCM token storage
- archive or history UI
- action buttons beyond opening terminal
- natural language task explanation layers
- team or sharing workflows

If a feature does not help the user answer "which task matters now?", it is out of P0.

## Risks

### Risk 1: State detection is noisy

Mitigation:
keep state model small and manually review common output patterns before expanding.

### Risk 2: Task abstraction feels fake

Mitigation:
map tasks directly to existing sessions in P0 and avoid inventing complex new lifecycle rules.

### Risk 3: Product still feels like terminal software

Mitigation:
make tasks the default route and terminal the secondary route.

### Risk 4: Scope creep

Mitigation:
defer notifications and actions until task/state/read flow is strong.

## P0 Execution Checklist

- [ ] Replace homepage routing so tasks become the primary landing page in web and mobile
- [ ] Define backend task DTO mapped from sessions
- [ ] Add backend task list endpoint
- [ ] Add backend task detail endpoint
- [ ] Add state derivation module for `running`, `waiting`, `completed`, `attention`
- [ ] Add summary derivation from recent session/output context
- [ ] Build web task list page
- [ ] Build web task detail page
- [ ] Move devices to secondary navigation in web
- [ ] Build mobile tasks page
- [ ] Build mobile task detail page
- [ ] Keep terminal available as power-user mode
- [ ] Verify the same task model appears in both web and mobile
- [ ] Validate the user can identify the most important active task without opening terminal output

## Release Decision Rule

P0 is ready to ship to early users when:

- task list works on web and mobile
- task state labels are visibly useful
- task detail reduces the need to jump immediately into terminal view
- users can still reach the original terminal workflow without friction

If any of those are missing, do not start P1.

# MobileCoder Product Control Tower Design

**Date:** 2026-04-10

## Summary

MobileCoder should stop presenting itself as "mobile remote control for Claude Code" and start positioning itself as the mobile control tower for AI coding tasks.

The product is not fundamentally about terminals, devices, or tmux. It is about helping developers supervise, interrupt, resume, and finish long-running AI coding work when they are away from the computer.

This is the product shift:

- Old framing: mobile terminal for Claude Code
- New framing: mobile supervision layer for AI coding agents

That shift changes the homepage, object model, notifications, success metrics, and roadmap.

## Problem

AI coding tools create a new kind of workflow problem.

Developers can now delegate meaningful work to an agent, but they still need to intermittently supervise that work:

- Did the agent finish or stall?
- Is it waiting for a reply?
- Is it blocked on permissions, tests, or a bad loop?
- Which of several running tasks needs attention now?

Today, most users solve this badly:

- Keep the laptop open and stare at the terminal
- Repeatedly walk back to the computer to check status
- Lose time when an agent silently stalls
- Avoid running longer jobs because they feel risky and opaque

MobileCoder exists to remove that supervision bottleneck.

## Product Thesis

The winning product is not "phone access to a terminal".

The winning product is:

> Let my AI coding tasks keep moving when I leave my computer, and pull me back only when my judgment is needed.

If MobileCoder does this well, it becomes a daily operating layer for heavy AI coding users.

## Target Users

### Primary user

A heavy AI coding user who already trusts Claude Code, Codex, or Cursor to do meaningful work and regularly runs long tasks.

Common traits:

- runs multi-minute or multi-hour agent tasks
- works across multiple repos or sessions
- leaves the desk often, but wants work to continue
- wants to supervise without reopening the whole workstation

### Best early adopters

1. Independent developers shipping solo
2. Small startup founders who code every day
3. Developers running multiple AI sessions in parallel
4. Power users who already live in tmux, worktrees, and terminal workflows

### Users to ignore for now

- casual ChatGPT users
- developers who do not trust agentic coding yet
- teams that want full browser IDE collaboration first

Those markets are broader, but weaker as an initial wedge.

## Core Job To Be Done

When I am away from my computer, help me understand whether my AI coding task is healthy, blocked, or done, and let me intervene in seconds when needed.

## Product Positioning

### One-line positioning

MobileCoder is the mobile control tower for AI coding agents.

### Expanded positioning

MobileCoder helps developers supervise and steer long-running AI coding tasks from a phone. It shows which tasks are running, stuck, waiting, or finished, sends notifications when attention is needed, and gives a fast path to intervene without reopening the full desktop environment.

### Positioning boundaries

MobileCoder is not:

- a general-purpose SSH client
- a remote desktop tool
- a mobile IDE
- a Claude Code clone

MobileCoder is:

- a supervision layer
- a task management surface
- a notification and intervention loop

## Product Principles

1. **Task-first**
   Users care about work units, not device topology.

2. **State over logs**
   Raw terminal output matters, but the product must interpret state before showing raw noise.

3. **Mobile is for supervision**
   Fast understanding, quick confirmation, light intervention. Not full editing.

4. **Push beats polling**
   If the user must keep refreshing, the product failed.

5. **Terminal is the fallback, not the homepage**
   The terminal remains a power-user escape hatch.

## Product Model

The current implementation thinks in terms of `device -> session -> terminal`.

The product should think in terms of `task -> state -> action`.

### Task

A task is the user-visible representation of one unit of AI coding work.

Suggested fields:

- task title
- repo or project name
- AI tool
- device
- session name
- current state
- recent summary
- last activity time
- requires user input flag

### State

The product should classify tasks into a limited state model:

- Running
- WaitingForInput
- Completed
- Failed
- Disconnected
- IdleTooLong
- NeedsAttention

State must be visible before the user opens the terminal.

### Action

Mobile actions should stay lightweight:

- Continue
- Stop
- Retry
- Reply
- Open terminal
- Switch session
- Archive

## Experience Design

## 1. Home

The homepage should become the task stream.

The user should be able to answer these questions in under five seconds:

- What is running?
- What is blocked?
- What just finished?
- Which task deserves attention right now?

### Home modules

- Active tasks
- Needs attention
- Recently completed
- Device health summary
- Quick filters: running, waiting, failed, completed

### Task card contents

- task title
- repo name
- AI tool badge
- state badge
- recent summary
- last activity timestamp
- one primary action button

## 2. Task Detail

Task detail should become the main work surface.

### Sections

- state header
- recent timeline
- generated summary
- quick actions
- latest terminal output
- open full terminal

The goal is to let the user act without reading the whole log.

## 3. Terminal

The terminal remains necessary, but its role changes.

It becomes:

- expert mode
- exact command surface
- full-context debugging view

It should not be the primary surface for product understanding.

## 4. Devices

Devices should move to secondary navigation.

Users care about devices when:

- binding a new machine
- debugging an offline agent
- checking which machine owns a task

That is important, but not first-screen important.

## 5. Notifications

Notifications are a core loop, not a side feature.

### P0 notifications

- task completed
- task waiting for confirmation
- task appears stalled
- agent disconnected

### Notification design rule

Every notification should answer:

- what happened
- why it matters
- what action the user can take

## Differentiation

The long-term moat is not terminal rendering.

The moat is being the best supervision layer for agentic coding.

That means:

- state awareness
- good interruption handling
- multi-task visibility
- fast handoff between async and live control
- cross-tool abstraction across Claude, Codex, Cursor

## Success Metrics

Do not optimize for vanity metrics like installs.

Track product behavior:

- weekly active supervised tasks
- tasks opened from push notifications
- rate of tasks successfully recovered after `WaitingForInput` or `IdleTooLong`
- average concurrent tasks per active user
- percentage of sessions that receive mobile intervention before failure

## Milestones

### Milestone 1: Task-first product

Shift from device list to task stream.

Success condition:
Users can understand ongoing work without opening a terminal.

### Milestone 2: State-aware system

Introduce explicit state classification and task summaries.

Success condition:
Users can distinguish healthy, blocked, and finished work at a glance.

### Milestone 3: Notification loop

Add push notifications for completion, waiting, stall, and disconnect events.

Success condition:
Users return because the product tells them when attention is needed.

### Milestone 4: Lightweight intervention

Add quick actions and fast reply flows.

Success condition:
Most agent stalls can be handled on mobile in under ten seconds.

### Milestone 5: Multi-task management

Support true parallel AI supervision.

Success condition:
Users actively monitor multiple concurrent coding tasks.

### Milestone 6: History and replay

Make tasks persistent and reviewable.

Success condition:
Past AI work becomes reusable operational history, not disposable output.

### Milestone 7: Cross-tool control plane

Unify Claude, Codex, and Cursor task surfaces.

Success condition:
Users think in tasks, not vendor-specific sessions.

### Milestone 8: Team workflows

Introduce shared visibility and cooperative supervision.

Success condition:
A small team can supervise the same pool of AI work.

## What To Delay

These should not lead the roadmap:

- deep theming
- broad customization
- full editing on mobile
- multi-language UI
- iOS-specific polish before the core loop is proven
- generic collaboration before single-user retention is strong

## Strategic Conclusion

MobileCoder has a strong wedge if it commits to being the mobile control tower for AI coding tasks.

If it stays framed as remote terminal access, it will feel like a clever utility.

If it becomes the place where developers supervise, prioritize, and rescue AI coding work, it can become a real product category.

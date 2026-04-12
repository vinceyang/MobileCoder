# Mobile Control Tower UI Polish Design

**Date:** 2026-04-10

## Purpose

This document defines the visual direction for polishing the mobile app after the task-first control tower flow has been implemented.

The goal is not to change product scope.

The goal is to make the existing `Tasks -> Task Detail -> Terminal` flow feel like a true mobile supervision console rather than a direct port of the web UI.

## Approved Direction

The approved visual direction is:

- high-tension runtime console
- neon command deck
- balanced operational emphasis
- dense information layout

This means the mobile app should feel like a live system monitor for active AI coding work.

It should communicate:

- motion
- urgency
- progress
- intervention readiness

It should not feel like:

- a generic settings-style list
- a flat CRUD mobile app
- a mini web admin dashboard

## Product Intent

The mobile app is the user’s pocket control surface.

The user opens it to answer:

- what is actively moving
- what needs intervention
- what just completed
- which task deserves immediate attention

The UI should optimize for rapid scanning under mobile constraints.

## Screen Priorities

### Tasks Home

This screen should feel like an active queue rather than a passive list.

It must make three things clear at a glance:

1. which task is the current hot lane
2. which tasks need intervention
3. which tasks are progressing normally

### Task Detail

This screen should feel like a live mission panel.

It must help the user understand:

- current state
- why the state is what it is
- recent event progression
- whether terminal takeover is necessary

### Terminal

Terminal remains secondary.

Its visual role is utility, not identity.

The app identity should be established by Tasks Home and Task Detail.

## Visual System

### Overall Mood

The app should resemble a live operational control surface.

The visual language should combine:

- deep blue-black backgrounds
- cyan and sky neon accents
- localized glow
- hard-edged cards with high contrast
- compact but readable spacing

The feeling should be “agent work in flight,” not “soft modern SaaS.”

### Color Strategy

Base:

- near-black navy backgrounds
- layered blue-gray surfaces

Primary accent:

- cyan / electric blue

Semantic accents:

- green for stable progress and test success
- amber for waiting or intervention-needed
- rose/red for failure or blocked state
- cyan for completed or high-confidence finished state

Color should be used for state emphasis, not decoration alone.

### Lighting and Atmosphere

Use restrained glow in a focused way:

- hero task cards may have subtle radial illumination
- status signals can emit localized glow
- active dots and event indicators should feel alive

Avoid turning the UI into diffuse blur soup.

The glow should sharpen hierarchy rather than soften it.

### Density

The app should remain dense by mobile standards.

That means:

- multiple tasks visible in one viewport
- compact metadata rows
- condensed labels
- event chips embedded into cards

But density should not collapse readability.

The pattern is:

- strong title
- one-line path
- one-line recent event
- one metadata row

## Tasks Home Design

### Structure

Top area:

- eyebrow label: `Control Tower`
- large `Tasks` heading
- one short operational summary
- compact live signal card on the right

Filter row:

- pill filters with clear active state
- horizontally scrollable if needed

Main section:

- priority queue framing
- stacked dense task cards

### Task Card Design

Each task card should contain:

- task title
- project path or repo label
- task state badge
- recent event chip
- recent event copy
- device label
- last activity time

The hottest or most relevant task may get a more illuminated card treatment.

### Hierarchy Rules

Within each card:

1. title
2. state
3. recent event kind + recent event line
4. metadata

Recent event should visually outrank generic summary text.

This keeps the home screen aligned with the supervision use case.

### Task State and Event Kind Pairing

The screen should show both:

- high-level task state
- recent event kind

For example:

- `运行中` + `步骤`
- `等待输入` + `待确认`
- `需关注` + `异常`
- `已完成` + `完成`

This pairing makes the list feel operational rather than static.

## Task Detail Design

### Structure

Header:

- eyebrow label
- task title
- short state reason
- state badge

Metadata grid:

- state reason
- last active
- device
- session

Timeline section:

- kind-aware recent timeline
- newest event first
- colored dots and chips
- short timestamp on each item

Action row:

- strong primary button for terminal takeover
- secondary refresh button

### Timeline Design

Timeline is the core visual identity of the detail page.

It should feel like an event stream from a running system.

Each event should include:

- colored dot
- event kind chip
- one-line event copy
- timestamp

Timeline should be dense but highly scannable.

### Action Design

Primary action:

- full-width or dominant-width takeover button
- bright cyan gradient or equivalent

Secondary action:

- subdued refresh button

The button hierarchy should make “open terminal only when needed” feel deliberate.

## Devices Screen Role

Devices remains in the app, but the UI should clearly signal it is secondary.

Required changes:

- less visual emphasis than tasks
- clear route back to tasks
- supporting view treatment rather than hero treatment

## Terminal Screen Role

Terminal should keep its existing utility but inherit the control tower framing:

- back path should prioritize task detail
- header should remain compact and utilitarian
- do not spend polish budget making terminal the visual center of the app

## Motion

Motion should be minimal but meaningful.

Allowed:

- subtle pulse on live indicator
- slight emphasis glow on active task card
- quick filter state transitions

Not allowed:

- decorative floating effects
- noisy looping animations
- slow cinematic transitions that hurt responsiveness

## Implementation Scope

This polish pass applies to:

- `mobile-app/src/pages/TasksPage.tsx`
- `mobile-app/src/pages/TaskDetailPage.tsx`
- `mobile-app/src/pages/DevicesPage.tsx`
- `mobile-app/src/pages/TerminalPage.tsx`
- shared task visual mappings in `mobile-app/src/services/tasks.ts`
- global mobile styling in `mobile-app/src/index.css`

This pass does not add:

- new backend APIs
- push notifications
- new task actions
- history
- new data model fields

## Success Criteria

The polish pass is successful when:

1. Tasks Home feels like a live runtime queue, not a generic list
2. Task Detail feels like a mission panel, not a plain metadata page
3. Waiting and attention states are more visually urgent than normal progress
4. Dense information remains readable on a phone
5. Terminal clearly reads as secondary to task supervision

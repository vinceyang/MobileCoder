# Mobile Control Tower UI Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle the mobile app task-first flow into the approved high-tension neon control-tower direction without changing product scope or backend APIs.

**Architecture:** Keep the current `mobile-app` routing and data model intact, and concentrate changes in page composition, shared task visual mappings, and global mobile styling. Polish `TasksPage` and `TaskDetailPage` first, then make secondary adjustments in `DevicesPage` and `TerminalPage` so the app reads as one coherent control surface.

**Tech Stack:** React 18, React Router, TypeScript, Vite, Tailwind v4 utility classes via `index.css`

---

## File Structure

### Existing files to modify

- `mobile-app/src/index.css`
  - Define the global visual substrate: background, text defaults, neon surface helpers if needed, and any reusable utility classes that reduce duplication across pages.
- `mobile-app/src/services/tasks.ts`
  - Keep data fetching unchanged but tighten shared label/style mappings so task state and event kind presentation stay consistent across pages.
- `mobile-app/src/pages/TasksPage.tsx`
  - Rebuild the home screen into a dense runtime queue with stronger hierarchy, live summary framing, and a visually dominant hot-lane card.
- `mobile-app/src/pages/TaskDetailPage.tsx`
  - Recast detail as a mission panel: stronger header, denser metadata panel, more deliberate timeline, clearer primary action.
- `mobile-app/src/pages/DevicesPage.tsx`
  - De-emphasize devices so it reads as a secondary management view rather than a competing home screen.
- `mobile-app/src/pages/TerminalPage.tsx`
  - Keep utility-first structure, but align the header and return path visually with the polished task flow.

### No new backend work

- Do not touch `cloud/`
- Do not add new routes or task fields
- Do not add notification plumbing or action framework

### Verification target

- `npm run build` in `mobile-app/`

---

### Task 1: Establish the Neon Control-Surface Foundation

**Files:**
- Modify: `mobile-app/src/index.css`
- Test: `mobile-app/package.json`

- [ ] **Step 1: Write the failing change by referencing the missing global design substrate**

Document the intended missing primitives in `mobile-app/src/index.css` before editing:

```css
/* Missing today:
   - layered app background beyond flat #111827
   - stronger text hierarchy defaults
   - reusable glow/surface feel for the control-tower UI
*/
```

Expected gap:

- Current pages rely on ad-hoc `bg-gray-*` utilities and do not share a coherent mobile control-surface foundation.

- [ ] **Step 2: Verify current app build before global CSS changes**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS
- Existing production build succeeds before polish work starts

- [ ] **Step 3: Add the minimal global styling substrate**

Update `mobile-app/src/index.css` to set the app-wide tone:

```css
@import "tailwindcss";

:root {
  color-scheme: dark;
}

* {
  -webkit-tap-highlight-color: transparent;
  box-sizing: border-box;
}

html,
body,
#root {
  min-height: 100%;
}

body {
  margin: 0;
  background:
    radial-gradient(circle at top right, rgba(34, 211, 238, 0.14), transparent 24%),
    radial-gradient(circle at top left, rgba(56, 189, 248, 0.08), transparent 28%),
    linear-gradient(180deg, #040b16 0%, #09111d 48%, #050b15 100%);
  color: #f8fafc;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

button,
input,
textarea,
select {
  font: inherit;
}
```

- [ ] **Step 4: Run the build to verify the foundation is clean**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS
- No TypeScript or Vite regressions from global CSS changes

- [ ] **Step 5: Commit**

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git add mobile-app/src/index.css
git commit -m "style: add mobile control tower visual foundation"
```

---

### Task 2: Normalize Shared Task Presentation Tokens

**Files:**
- Modify: `mobile-app/src/services/tasks.ts`
- Test: `mobile-app/package.json`

- [ ] **Step 1: Write the failing change by identifying shared token gaps**

Capture the current gap in `mobile-app/src/services/tasks.ts` before editing:

```ts
// Missing today:
// - more expressive mobile-oriented labels for event kinds
// - one shared helper for event dot / badge styling if pages need both
```

Expected gap:

- Pages currently own too much visual interpretation logic.

- [ ] **Step 2: Run the build to confirm the unpolished shared mapping is the current baseline**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 3: Extend shared mappings without changing the API contract**

Update `mobile-app/src/services/tasks.ts` so it exports any style/label helpers needed by both task pages. Keep the fetch functions unchanged. Add only small shared UI helpers such as:

```ts
export const taskEventKindDotStyles: Record<TaskEventKind, string> = {
  info: 'bg-slate-400',
  needs_input: 'bg-amber-400 shadow-[0_0_16px_rgba(245,158,11,0.35)]',
  error: 'bg-rose-400 shadow-[0_0_16px_rgba(251,113,133,0.35)]',
  test_result: 'bg-emerald-400 shadow-[0_0_16px_rgba(52,211,153,0.35)]',
  completed: 'bg-cyan-400 shadow-[0_0_16px_rgba(34,211,238,0.35)]',
  tool_step: 'bg-sky-400 shadow-[0_0_16px_rgba(56,189,248,0.35)]',
}
```

Do not move fetch logic or rename the exported types.

- [ ] **Step 4: Run the build to verify the shared token layer**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git add mobile-app/src/services/tasks.ts
git commit -m "style: add shared mobile task presentation tokens"
```

---

### Task 3: Rebuild Tasks Home as a Dense Runtime Queue

**Files:**
- Modify: `mobile-app/src/pages/TasksPage.tsx`
- Test: `mobile-app/package.json`

- [ ] **Step 1: Write the failing change by identifying the current home-screen mismatch**

Capture the mismatch in `mobile-app/src/pages/TasksPage.tsx`:

```tsx
// Current gap:
// - reads like a basic list page
// - does not establish a live runtime queue feeling
// - hot-lane task does not stand out enough
```

Expected gap:

- The approved design calls for a high-tension neon command deck, but the page is still mostly a straightforward feed.

- [ ] **Step 2: Run the current build as the baseline**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 3: Replace the page shell with the approved layout**

Refactor `mobile-app/src/pages/TasksPage.tsx` to include:

- a `Control Tower` eyebrow
- a larger `Tasks` heading with one-line operational summary
- a compact live signal panel
- dense horizontally scrollable filter pills
- a `priority queue` section label

Use utility classes in the style of:

```tsx
<header className="px-4 pt-5 pb-4 border-b border-cyan-400/10">
  <div className="flex items-start justify-between gap-4">
    <div>
      <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">Control Tower</p>
      <h1 className="text-3xl font-black tracking-tight mt-3">Tasks</h1>
      <p className="text-sm text-slate-400 mt-2">4 tasks active, 1 waiting, 1 attention needed</p>
    </div>
    <div className="rounded-2xl border border-cyan-400/10 bg-cyan-400/5 px-3 py-2">
      <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">hot lane</p>
      <p className="text-base font-black text-cyan-300 mt-2">release</p>
    </div>
  </div>
</header>
```

- [ ] **Step 4: Restyle each task card around event-first hierarchy**

Within the task list:

- make the first/highest-priority visible card visually hotter
- keep title and state on the first row
- place event-kind chip and recent-event copy above metadata
- show last activity time in metadata

Use a denser card pattern like:

```tsx
<button className="w-full rounded-[22px] border border-cyan-400/10 bg-slate-950/80 p-4 text-left">
  <div className="flex items-start justify-between gap-3">
    <div className="min-w-0">
      <h3 className="text-base font-extrabold text-slate-50 truncate">{task.title}</h3>
      <p className="text-xs text-slate-500 mt-1 truncate">{task.project_path || task.device_name}</p>
    </div>
    <span className={`px-2 py-1 rounded-full text-[11px] ${taskStateStyles[task.state]}`}>
      {taskStateLabels[task.state]}
    </span>
  </div>
  <div className="mt-3 flex items-center gap-2 flex-wrap">
    <span className={`text-[11px] px-2 py-1 rounded-full ${taskEventKindStyles[latestEventKind]}`}>
      {taskEventKindLabels[latestEventKind]}
    </span>
    <p className="text-sm text-slate-200">{task.recent_event || task.summary}</p>
  </div>
</button>
```

- [ ] **Step 5: Run the build to verify the polished home screen**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git add mobile-app/src/pages/TasksPage.tsx
git commit -m "style: polish mobile task home into runtime queue"
```

---

### Task 4: Recast Task Detail as a Mission Panel

**Files:**
- Modify: `mobile-app/src/pages/TaskDetailPage.tsx`
- Test: `mobile-app/package.json`

- [ ] **Step 1: Write the failing change by identifying the current detail-page weakness**

Document the gap in `mobile-app/src/pages/TaskDetailPage.tsx`:

```tsx
// Current gap:
// - metadata reads as stacked cards, but not yet as a mission panel
// - timeline is functional, but not visually central enough
// - primary takeover action needs stronger dominance
```

- [ ] **Step 2: Run the current build as a baseline**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 3: Strengthen the detail header and metadata block**

Update the top of `TaskDetailPage.tsx` so the page opens with:

- eyebrow
- stronger title
- concise state reason copy near the title
- dense but elevated metadata block

Use a structure like:

```tsx
<header className="px-4 pt-5 pb-4 border-b border-cyan-400/10">
  <button onClick={() => navigate('/tasks')} className="text-slate-400 mb-4">← 返回任务列表</button>
  <div className="flex items-start justify-between gap-3">
    <div>
      <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">Task Detail</p>
      <h1 className="text-3xl font-black tracking-tight mt-3">{task.title}</h1>
      <p className="text-sm text-slate-400 mt-2">{task.state_reason}</p>
    </div>
    <span className={`px-2 py-1 rounded-full text-[11px] ${taskStateStyles[task.state]}`}>
      {taskStateLabels[task.state]}
    </span>
  </div>
</header>
```

- [ ] **Step 4: Make the timeline the center of gravity**

Update the timeline block to:

- use stronger section framing
- use glowing colored dots
- tighten vertical rhythm
- keep event-kind chip and copy on the same visual row when space allows

The timeline item target pattern:

```tsx
<div className="grid grid-cols-[12px_1fr] gap-3">
  <div className={`mt-1 h-2.5 w-2.5 rounded-full ${taskEventKindDotStyles[event.kind]}`} />
  <div>
    <div className="flex items-center gap-2 flex-wrap">
      <span className={`text-[11px] px-2 py-1 rounded-full ${taskEventKindStyles[event.kind]}`}>
        {taskEventKindLabels[event.kind]}
      </span>
      <p className="text-sm text-slate-100">{event.summary}</p>
    </div>
    <p className="text-[11px] text-slate-500 mt-1">{formatActivityLabel(event.timestamp)}</p>
  </div>
</div>
```

- [ ] **Step 5: Rebalance the action row**

Make the terminal takeover button visually dominant and the refresh button clearly secondary:

```tsx
<div className="flex gap-3 pt-2">
  <button className="flex-1 rounded-2xl bg-gradient-to-br from-cyan-300 to-sky-400 px-4 py-3 font-extrabold text-sky-950">
    打开终端接管
  </button>
  <button className="rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-4 py-3 font-semibold text-slate-200">
    刷新
  </button>
</div>
```

- [ ] **Step 6: Run the build to verify the detail-page polish**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git add mobile-app/src/pages/TaskDetailPage.tsx
git commit -m "style: polish mobile task detail into mission panel"
```

---

### Task 5: De-Emphasize Devices and Align Terminal Framing

**Files:**
- Modify: `mobile-app/src/pages/DevicesPage.tsx`
- Modify: `mobile-app/src/pages/TerminalPage.tsx`
- Test: `mobile-app/package.json`

- [ ] **Step 1: Write the failing change by identifying the remaining visual mismatch**

Document the gap in each file:

```tsx
// DevicesPage gap:
// - still too visually similar to a primary surface
//
// TerminalPage gap:
// - routing is correct, but header still feels generic rather than connected to the control tower flow
```

- [ ] **Step 2: Run the current build as the baseline**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 3: Restyle DevicesPage as a secondary management surface**

Update `DevicesPage.tsx` to:

- keep a smaller header hierarchy than Tasks
- preserve the `Tasks` return affordance
- make cards simpler and less visually dominant than task cards

Target shell:

```tsx
<header className="px-4 pt-5 pb-4 border-b border-cyan-400/10">
  <p className="text-[11px] uppercase tracking-[0.22em] text-slate-500">Secondary View</p>
  <div className="flex items-start justify-between mt-3 gap-4">
    <div>
      <h1 className="text-2xl font-black">设备</h1>
      <p className="text-sm text-slate-400 mt-2">管理连接中的机器和会话入口。</p>
    </div>
    <button onClick={() => navigate('/tasks')} className="text-cyan-300 text-sm">Tasks</button>
  </div>
</header>
```

- [ ] **Step 4: Tighten TerminalPage’s header**

Update `TerminalPage.tsx` so the header feels utilitarian but still related to the control tower:

- smaller uppercase label
- clearer task-oriented back affordance
- preserve current routing behavior exactly

Target shell:

```tsx
<header className="flex items-center px-4 py-3 border-b border-cyan-400/10 bg-slate-950/90">
  <button ... className="text-slate-400 text-xl mr-4">←</button>
  <div>
    <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500">Terminal</p>
    <span className="text-sm text-slate-300">专家接管模式</span>
  </div>
</header>
```

- [ ] **Step 5: Run the build to verify supporting-page alignment**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git add mobile-app/src/pages/DevicesPage.tsx mobile-app/src/pages/TerminalPage.tsx
git commit -m "style: align mobile secondary surfaces with control tower"
```

---

### Task 6: Final Integration Verification

**Files:**
- Modify: none
- Test: `mobile-app/package.json`

- [ ] **Step 1: Run the full mobile production build**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower/mobile-app
npm run build
```

Expected:

- PASS
- `vite` prints a successful production bundle

- [ ] **Step 2: Review changed files and verify no build artifacts are left dirty**

Run:

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git status --short
```

Expected:

- only intended source-file changes are present
- no accidental `mobile-app/dist` changes remain

- [ ] **Step 3: Commit the final integration**

```bash
cd /Users/yangxq/Code/MobileCoder/.worktrees/product-control-tower
git add mobile-app/src/index.css mobile-app/src/services/tasks.ts mobile-app/src/pages/TasksPage.tsx mobile-app/src/pages/TaskDetailPage.tsx mobile-app/src/pages/DevicesPage.tsx mobile-app/src/pages/TerminalPage.tsx
git commit -m "style: polish mobile control tower surfaces"
```

---

## Self-Review

### Spec coverage

- Global neon runtime feel: covered in Task 1
- Shared task/event presentation consistency: covered in Task 2
- Dense runtime queue home: covered in Task 3
- Mission-panel detail page: covered in Task 4
- Secondary device view and utilitarian terminal framing: covered in Task 5
- Verification and clean source-only finish: covered in Task 6

No spec sections are left without an implementing task.

### Placeholder scan

- No `TODO`/`TBD`
- Every task has exact file paths
- Every code-changing step shows concrete target code patterns
- Every verification step has an exact command and expected result

### Type consistency

- Shared mappings continue to come from `mobile-app/src/services/tasks.ts`
- Pages continue to consume the existing `Task`, `TaskState`, and `TaskEventKind` types
- No new backend types or route changes are introduced

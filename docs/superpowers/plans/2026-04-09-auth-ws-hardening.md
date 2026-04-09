# Auth And WS Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current fake token flow with signed tokens, enforce device/session ownership checks on REST and WebSocket paths, and unify frontend server configuration so the system is safe enough for continued development.

**Architecture:** The cloud service will own a small shared auth module that signs and verifies tokens using an HMAC secret. Device and session handlers will require a verified user identity before returning data or accepting viewer connections, while agent-only flows stay explicitly unauthenticated. Frontend clients will stop hardcoding server addresses and derive HTTP/WS endpoints from shared configuration.

**Tech Stack:** Go, net/http, Gorilla WebSocket, React, Next.js, Vite, Capacitor

---

### Task 1: Add auth verification tests

**Files:**
- Create: `cloud/internal/auth/token_test.go`
- Modify: `cloud/internal/handler/device_handler_test.go`
- Modify: `cloud/internal/handler/ws_handler_test.go`

- [ ] **Step 1: Write failing tests for signed token verification**
- [ ] **Step 2: Run the focused Go tests and confirm they fail for the current fake token implementation**
- [ ] **Step 3: Add minimal auth package implementation**
- [ ] **Step 4: Re-run the focused Go tests and confirm token verification passes**

### Task 2: Enforce REST authorization

**Files:**
- Create: `cloud/internal/handler/auth_helpers.go`
- Modify: `cloud/internal/handler/auth_handler.go`
- Modify: `cloud/internal/handler/device_handler.go`
- Modify: `cloud/internal/service/device_service.go`
- Modify: `cloud/internal/db/supabase.go`

- [ ] **Step 1: Write failing handler tests for unauthorized device/session access**
- [ ] **Step 2: Run the handler tests and confirm they fail because ownership is not enforced**
- [ ] **Step 3: Implement shared auth extraction plus device/session ownership checks**
- [ ] **Step 4: Re-run the focused handler tests and confirm they pass**

### Task 3: Enforce WebSocket authorization

**Files:**
- Modify: `cloud/internal/handler/ws_handler.go`
- Modify: `cloud/internal/ws/hub.go`
- Modify: `cloud/internal/handler/ws_handler_test.go`

- [ ] **Step 1: Write failing WebSocket tests for unauthorized viewer connections**
- [ ] **Step 2: Run the WebSocket tests and confirm they fail with the current permissive behavior**
- [ ] **Step 3: Implement verified viewer auth while keeping agent connections explicit**
- [ ] **Step 4: Re-run the focused WebSocket tests and confirm they pass**

### Task 4: Unify frontend endpoint configuration

**Files:**
- Create: `mobile-app/src/config/api.ts`
- Modify: `mobile-app/src/services/auth.ts`
- Modify: `mobile-app/src/services/device.ts`
- Modify: `mobile-app/src/components/Terminal.tsx`
- Modify: `chat/src/app/components/Terminal.tsx`
- Modify: `chat/src/app/login/page.tsx`
- Modify: `chat/src/app/devices/page.tsx`
- Modify: `chat/src/app/devices/[deviceId]/page.tsx`

- [ ] **Step 1: Write or adapt minimal frontend tests where available, otherwise validate through TypeScript builds**
- [ ] **Step 2: Replace hardcoded hosts with shared config derived from environment and current origin**
- [ ] **Step 3: Update REST and WS callers to use the shared config**
- [ ] **Step 4: Run the relevant build/test commands and confirm both frontends still compile**

### Task 5: Verify and document boundaries

**Files:**
- Modify: `README.md`
- Modify: `docs/user-manual.md`

- [ ] **Step 1: Run focused verification for Go handlers plus frontend builds**
- [ ] **Step 2: Update docs only if endpoint or login behavior changed materially**
- [ ] **Step 3: Summarize any remaining security gaps that are intentionally deferred**

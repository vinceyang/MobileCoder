# MobileCoder Android App Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create Android app using Capacitor to package H5 code, enabling users to manage devices and control terminal from mobile.

**Architecture:** Create new `mobile-app/` project with Capacitor wrapping React H5 code. App loads remote H5 from `http://121.41.69.142:3001`. Native capabilities (push notifications, storage) added via Capacitor plugins.

**Tech Stack:** React 18 + TypeScript + Vite + Capacitor 7 + Tailwind CSS

---

## Phase 1: Project Setup

### Task 1: Create Capacitor Project

**Files:**
- Create: `mobile-app/package.json`
- Create: `mobile-app/tsconfig.json`
- Create: `mobile-app/vite.config.ts`
- Create: `mobile-app/index.html`
- Create: `mobile-app/capacitor.config.ts`

- [ ] **Step 1: Create project directory and initialize npm**

```bash
cd /Users/yangxq/Code/MobileCoder
mkdir -p mobile-app
cd mobile-app
npm init -y
```

- [ ] **Step 2: Install core dependencies**

```bash
npm install react@18 react-dom@18 react-router-dom@6
npm install -D typescript@5 vite@5 @vitejs/plugin-react@4 @types/react@18 @types/react-dom@18
npm install @capacitor/core @capacitor/cli @capacitor/android
```

- [ ] **Step 3: Create vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: './',
  build: {
    outDir: 'dist',
  },
})
```

- [ ] **Step 4: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

- [ ] **Step 5: Create tsconfig.node.json**

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
```

- [ ] **Step 6: Create index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover, user-scalable=no" />
    <meta name="theme-color" content="#1f2937" />
    <title>MobileCoder</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 7: Create capacitor.config.ts**

```typescript
import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.mobilecoder.app',
  appName: 'MobileCoder',
  webDir: 'dist',
  server: {
    url: 'http://121.41.69.142:3001',
    cleartext: true,
  },
  android: {
    backgroundColor: '#1f2937',
  },
};

export default config;
```

- [ ] **Step 8: Initialize Capacitor**

```bash
npx cap init MobileCoder com.mobilecoder.app --web-dir=dist
```

- [ ] **Step 9: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): create Capacitor project structure"
```

---

### Task 2: Setup Tailwind CSS

**Files:**
- Create: `mobile-app/tailwind.config.js`
- Create: `mobile-app/postcss.config.js`
- Create: `mobile-app/src/index.css`

- [ ] **Step 1: Install Tailwind CSS**

```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

- [ ] **Step 2: Configure tailwind.config.js**

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

- [ ] **Step 3: Create src/index.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

* {
  -webkit-tap-highlight-color: transparent;
}

body {
  @apply bg-gray-900 text-white antialiased;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
```

- [ ] **Step 4: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): add Tailwind CSS configuration"
```

---

### Task 3: Create App Entry Point and Routing

**Files:**
- Create: `mobile-app/src/main.tsx`
- Create: `mobile-app/src/App.tsx`
- Create: `mobile-app/src/pages/LoginPage.tsx`
- Create: `mobile-app/src/pages/DevicesPage.tsx`
- Create: `mobile-app/src/pages/DeviceDetailPage.tsx`
- Create: `mobile-app/src/pages/TerminalPage.tsx`

- [ ] **Step 1: Create src/main.tsx**

```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

- [ ] **Step 2: Create src/App.tsx**

```tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import DevicesPage from './pages/DevicesPage'
import DeviceDetailPage from './pages/DeviceDetailPage'
import TerminalPage from './pages/TerminalPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/devices" element={
          <ProtectedRoute>
            <DevicesPage />
          </ProtectedRoute>
        } />
        <Route path="/devices/:deviceId" element={
          <ProtectedRoute>
            <DeviceDetailPage />
          </ProtectedRoute>
        } />
        <Route path="/terminal" element={
          <ProtectedRoute>
            <TerminalPage />
          </ProtectedRoute>
        } />
        <Route path="/" element={<Navigate to="/devices" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
```

- [ ] **Step 3: Create placeholder pages**

Create each page with basic structure:

```tsx
// src/pages/LoginPage.tsx - Login form with email/password
// src/pages/DevicesPage.tsx - Device list
// src/pages/DeviceDetailPage.tsx - Device detail + sessions
// src/pages/TerminalPage.tsx - Terminal interface
```

- [ ] **Step 4: Build and verify**

```bash
npm run build
npx cap add android
npx cap sync android
```

- [ ] **Step 5: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): add routing and page structure"
```

---

## Phase 2: Core Pages

### Task 4: Implement Login Page

**Files:**
- Create: `mobile-app/src/pages/LoginPage.tsx`
- Create: `mobile-app/src/services/auth.ts`
- Create: `mobile-app/src/components/LoadingButton.tsx`

- [ ] **Step 1: Create auth service**

```typescript
// src/services/auth.ts
const API_URL = 'http://121.41.69.142:8080'

export interface LoginResponse {
  token: string
  user_id: number
  email: string
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${API_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || 'Login failed')
  }
  return res.json()
}

export async function register(email: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${API_URL}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || 'Registration failed')
  }
  return res.json()
}
```

- [ ] **Step 2: Create LoginPage with full implementation**

```tsx
// src/pages/LoginPage.tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login, register } from '../services/auth'

export default function LoginPage() {
  const [isLogin, setIsLogin] = useState(true)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const data = isLogin
        ? await login(email, password)
        : await register(email, password)

      localStorage.setItem('token', data.token)
      localStorage.setItem('user_id', String(data.user_id))
      localStorage.setItem('email', data.email)

      navigate('/devices')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <h1 className="text-2xl font-bold text-center mb-8">MobileCoder</h1>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">邮箱</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white"
              placeholder="email@example.com"
              required
            />
          </div>

          <div>
            <label className="block text-sm text-gray-400 mb-1">密码</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white"
              placeholder="••••••••"
              required
            />
          </div>

          {error && (
            <p className="text-red-500 text-sm">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 bg-blue-600 text-white rounded-lg font-medium disabled:opacity-50"
          >
            {loading ? '处理中...' : (isLogin ? '登录' : '注册')}
          </button>
        </form>

        <p className="text-center mt-4 text-gray-400 text-sm">
          {isLogin ? '没有账号？' : '已有账号？'}
          <button
            onClick={() => setIsLogin(!isLogin)}
            className="text-blue-400 ml-1"
          >
            {isLogin ? '注册' : '登录'}
          </button>
        </p>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): implement login page with auth service"
```

---

### Task 5: Implement Devices List Page

**Files:**
- Create: `mobile-app/src/services/device.ts`
- Modify: `mobile-app/src/pages/DevicesPage.tsx`

- [ ] **Step 1: Create device service**

```typescript
// src/services/device.ts
const API_URL = 'http://121.41.69.142:8080'

function getToken() {
  return localStorage.getItem('token') || ''
}

export interface Device {
  id: number
  device_id: string
  device_name: string
  status: string
  user_id: number
}

export interface Session {
  id: number
  device_id: string
  session_name: string
  project_path: string
  status: string
}

export async function getDevices(): Promise<Device[]> {
  const token = getToken()
  const res = await fetch(`${API_URL}/api/devices`, {
    headers: { 'Authorization': token },
  })
  if (!res.ok) throw new Error('Failed to fetch devices')
  const data = await res.json()
  return data.devices || []
}

export async function getDeviceSessions(deviceId: string): Promise<Session[]> {
  const token = getToken()
  const res = await fetch(`${API_URL}/api/devices/sessions?device_id=${deviceId}`, {
    headers: { 'Authorization': token },
  })
  if (!res.ok) throw new Error('Failed to fetch sessions')
  const data = await res.json()
  return data.sessions || []
}
```

- [ ] **Step 2: Implement DevicesPage**

```tsx
// src/pages/DevicesPage.tsx
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getDevices, Device } from '../services/device'

export default function DevicesPage() {
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    loadDevices()
  }, [])

  const loadDevices = async () => {
    try {
      const data = await getDevices()
      setDevices(data)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">我的设备</h1>
        <button onClick={handleLogout} className="text-gray-400 text-sm">
          退出
        </button>
      </header>

      <div className="p-4">
        {loading ? (
          <p className="text-gray-400 text-center">加载中...</p>
        ) : devices.length === 0 ? (
          <div className="text-center text-gray-400 mt-8">
            <p>暂无设备</p>
            <p className="text-sm mt-2">请在 Desktop Agent 中绑定设备</p>
          </div>
        ) : (
          <div className="space-y-3">
            {devices.map((device) => (
              <div
                key={device.device_id}
                onClick={() => navigate(`/devices/${device.device_id}`)}
                className="bg-gray-800 p-4 rounded-lg cursor-pointer active:bg-gray-700"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium">{device.device_name}</h3>
                    <p className="text-sm text-gray-400">{device.device_id}</p>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs ${
                    device.status === 'online' ? 'bg-green-600' : 'bg-gray-600'
                  }`}>
                    {device.status === 'online' ? '在线' : '离线'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): implement devices list page"
```

---

### Task 6: Implement Device Detail Page

**Files:**
- Modify: `mobile-app/src/pages/DeviceDetailPage.tsx`

- [ ] **Step 1: Implement DeviceDetailPage with sessions**

```tsx
// src/pages/DeviceDetailPage.tsx
import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getDeviceSessions, Session } from '../services/device'

export default function DeviceDetailPage() {
  const { deviceId } = useParams<{ deviceId: string }>()
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    if (deviceId) {
      loadSessions()
    }
  }, [deviceId])

  const loadSessions = async () => {
    if (!deviceId) return
    try {
      const data = await getDeviceSessions(deviceId)
      setSessions(data)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const connectSession = (sessionName: string) => {
    localStorage.setItem('device_id', deviceId!)
    localStorage.setItem('session_name', sessionName)
    navigate(`/terminal?device_id=${deviceId}&session_name=${encodeURIComponent(sessionName)}`)
  }

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="flex items-center p-4 border-b border-gray-800">
        <button onClick={() => navigate('/devices')} className="text-gray-400 mr-4">
          ←
        </button>
        <div>
          <h1 className="text-lg font-bold">设备详情</h1>
          <p className="text-xs text-gray-400">{deviceId}</p>
        </div>
      </header>

      <div className="p-4">
        <h2 className="text-lg font-medium mb-3">Sessions</h2>

        {loading ? (
          <p className="text-gray-400">加载中...</p>
        ) : sessions.length === 0 ? (
          <p className="text-gray-400 text-center mt-4">暂无活跃 Session</p>
        ) : (
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                onClick={() => connectSession(session.session_name)}
                className="bg-gray-800 p-4 rounded-lg cursor-pointer active:bg-gray-700"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium">{session.session_name}</h3>
                    <p className="text-sm text-gray-400">{session.project_path}</p>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs ${
                    session.status === 'active' ? 'bg-green-600' : 'bg-gray-600'
                  }`}>
                    {session.status === 'active' ? '活跃' : '非活跃'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): implement device detail page"
```

---

### Task 7: Implement Terminal Page

**Files:**
- Create: `mobile-app/src/components/Terminal.tsx`
- Modify: `mobile-app/src/pages/TerminalPage.tsx`

- [ ] **Step 1: Create Terminal component**

Copy and adapt from existing H5 Terminal component. Key changes:
- Adjust WebSocket URL for remote server
- Optimize for mobile touch input
- Add mobile-specific keyboard handling

```tsx
// src/components/Terminal.tsx
// Based on chat/src/app/components/Terminal.tsx
// Key differences:
// - Use remote WebSocket: ws://121.41.69.142:8080/ws
// - Larger touch targets
// - Simplified toolbar
```

- [ ] **Step 2: Create TerminalPage**

```tsx
// src/pages/TerminalPage.tsx
import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import Terminal from '../components/Terminal'

export default function TerminalPage() {
  const [searchParams] = useSearchParams()
  const [deviceId, setDeviceId] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    const deviceIdParam = searchParams.get('device_id')
    const sessionNameParam = searchParams.get('session_name')

    const id = deviceIdParam || localStorage.getItem('device_id')
    if (!id) {
      navigate('/devices')
      return
    }

    setDeviceId(id)
    if (sessionNameParam) {
      localStorage.setItem('session_name', sessionNameParam)
    }
  }, [searchParams])

  if (!deviceId) {
    return null
  }

  return <Terminal deviceId={deviceId} />
}
```

- [ ] **Step 3: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): implement terminal page with WebSocket"
```

---

## Phase 3: Capacitor Native Features

### Task 8: Add Capacitor Plugins

**Files:**
- Modify: `mobile-app/package.json`
- Modify: `mobile-app/capacitor.config.ts`

- [ ] **Step 1: Install Capacitor plugins**

```bash
npm install @capacitor/preferences @capacitor/haptics
```

- [ ] **Step 2: Update capacitor.config.ts**

```typescript
import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.mobilecoder.app',
  appName: 'MobileCoder',
  webDir: 'dist',
  server: {
    url: 'http://121.41.69.142:3001',
    cleartext: true,
  },
  android: {
    backgroundColor: '#1f2937',
  },
  plugins: {
    Preferences: {},
    Haptics: {
      style: 'light',
    },
  },
};

export default config;
```

- [ ] **Step 3: Create storage service**

```typescript
// src/services/storage.ts
import { Preferences } from '@capacitor/preferences'

export async function setItem(key: string, value: string) {
  await Preferences.set({ key, value })
}

export async function getItem(key: string): Promise<string | null> {
  const { value } = await Preferences.get({ key })
  return value
}

export async function removeItem(key: string) {
  await Preferences.remove({ key })
}
```

- [ ] **Step 4: Create haptic feedback utility**

```typescript
// src/utils/haptics.ts
import { Haptics, ImpactStyle } from '@capacitor/haptics'

export async function lightImpact() {
  try {
    await Haptics.impact({ style: ImpactStyle.Light })
  } catch (e) {
    // Haptics not available
  }
}

export async function mediumImpact() {
  try {
    await Haptics.impact({ style: ImpactStyle.Medium })
  } catch (e) {
    // Haptics not available
  }
}
```

- [ ] **Step 5: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): add Capacitor plugins for native features"
```

---

### Task 9: Build and Test APK

**Files:**
- None (verification task)

- [ ] **Step 1: Build web app**

```bash
cd mobile-app
npm run build
```

- [ ] **Step 2: Sync to Android**

```bash
npx cap sync android
```

- [ ] **Step 3: Build debug APK**

```bash
cd android
./gradlew assembleDebug
```

- [ ] **Step 4: Verify APK exists**

```bash
ls -la android/app/build/outputs/apk/debug/
```

Expected output: `app-debug.apk`

- [ ] **Step 5: Commit**

```bash
git add mobile-app/
git commit -m "feat(mobile): build debug APK"
```

---

## Summary

| Task | Description | Status |
|------|-------------|--------|
| 1 | Create Capacitor Project | ⬜ |
| 2 | Setup Tailwind CSS | ⬜ |
| 3 | Create App Entry and Routing | ⬜ |
| 4 | Implement Login Page | ⬜ |
| 5 | Implement Devices List Page | ⬜ |
| 6 | Implement Device Detail Page | ⬜ |
| 7 | Implement Terminal Page | ⬜ |
| 8 | Add Capacitor Plugins | ⬜ |
| 9 | Build and Test APK | ⬜ |

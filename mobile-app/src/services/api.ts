import { getApiBaseUrl } from '../config/api'

const AUTH_EXPIRED_EVENT = 'mobilecoder-auth-expired'

function getToken() {
  return localStorage.getItem('token') || ''
}

function setToken(token: string) {
  localStorage.setItem('token', token)
}

export function clearAuthState() {
  localStorage.removeItem('token')
  localStorage.removeItem('user_id')
  localStorage.removeItem('email')
}

export function notifyAuthExpired() {
  clearAuthState()
  window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
}

export function onAuthExpired(callback: () => void) {
  window.addEventListener(AUTH_EXPIRED_EVENT, callback)
  return () => window.removeEventListener(AUTH_EXPIRED_EVENT, callback)
}

async function refreshToken(): Promise<boolean> {
  const token = getToken()
  if (!token) return false

  try {
    const res = await fetch(`${getApiBaseUrl()}/api/auth/refresh`, {
      method: 'POST',
      headers: { Authorization: token },
    })
    if (!res.ok) return false

    const data = await res.json()
    if (!data.token) return false

    setToken(data.token)
    if (data.user_id) localStorage.setItem('user_id', String(data.user_id))
    if (data.email) localStorage.setItem('email', data.email)
    return true
  } catch {
    return false
  }
}

function withAuthHeader(init: RequestInit = {}): RequestInit {
  const headers = new Headers(init.headers)
  const token = getToken()
  if (token) headers.set('Authorization', token)
  return { ...init, headers }
}

export async function authFetch(input: string, init: RequestInit = {}): Promise<Response> {
  const first = await fetch(input, withAuthHeader(init))
  if (first.status !== 401) return first

  const refreshed = await refreshToken()
  if (!refreshed) {
    notifyAuthExpired()
    return first
  }

  const second = await fetch(input, withAuthHeader(init))
  if (second.status === 401) notifyAuthExpired()
  return second
}

import { getApiBaseUrl } from '../config/api'

export interface LoginResponse {
  token: string
  user_id: number
  email: string
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${getApiBaseUrl()}/api/auth/login`, {
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
  const res = await fetch(`${getApiBaseUrl()}/api/auth/register`, {
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

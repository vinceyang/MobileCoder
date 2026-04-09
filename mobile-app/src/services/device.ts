import { getApiBaseUrl } from '../config/api'

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
  const res = await fetch(`${getApiBaseUrl()}/api/devices`, {
    headers: { 'Authorization': token },
  })
  if (!res.ok) throw new Error('Failed to fetch devices')
  const data = await res.json()
  return data.devices || []
}

export async function getDeviceSessions(deviceId: string): Promise<Session[]> {
  const token = getToken()
  const res = await fetch(`${getApiBaseUrl()}/api/devices/sessions?device_id=${deviceId}`, {
    headers: { 'Authorization': token },
  })
  if (!res.ok) throw new Error('Failed to fetch sessions')
  const data = await res.json()
  return data.sessions || []
}

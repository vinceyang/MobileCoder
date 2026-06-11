import { getApiBaseUrl } from '../config/api'
import { authFetch } from './api'

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
  const res = await authFetch(`${getApiBaseUrl()}/api/devices`)
  if (!res.ok) throw new Error('Failed to fetch devices')
  const data = await res.json()
  return data.devices || []
}

export async function getDeviceSessions(deviceId: string): Promise<Session[]> {
  const res = await authFetch(`${getApiBaseUrl()}/api/devices/sessions?device_id=${deviceId}`)
  if (!res.ok) throw new Error('Failed to fetch sessions')
  const data = await res.json()
  return data.sessions || []
}

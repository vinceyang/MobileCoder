const DEFAULT_API_PORT = '8080'

function normalizeHttpBaseUrl(value: string): string {
  return value.replace(/\/+$/, '')
}

export function getApiBaseUrl(): string {
  const configured = import.meta.env.VITE_API_URL
  if (configured) {
    return normalizeHttpBaseUrl(configured)
  }

  if (typeof window !== 'undefined') {
    const { protocol, hostname } = window.location
    if (protocol === 'http:' || protocol === 'https:') {
      return `${protocol}//${hostname}:${DEFAULT_API_PORT}`
    }
  }

  return `http://localhost:${DEFAULT_API_PORT}`
}

export function getWsBaseUrl(): string {
  const apiUrl = getApiBaseUrl()
  if (apiUrl.startsWith('https://')) {
    return `wss://${apiUrl.slice('https://'.length)}`
  }
  if (apiUrl.startsWith('http://')) {
    return `ws://${apiUrl.slice('http://'.length)}`
  }
  return apiUrl
}

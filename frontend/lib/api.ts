const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:3000'
const PROXY_HOSTS = ['cdninstagram.com', 'fbcdn.net', 'tiktokcdn.com', 'googlevideo.com']

export type DisplayItem = {
  id: string
  sourceType: 'instagram' | 'tiktok' | 'youtube' | 'image' | 'text'
  igUrl: string
  igImageUrl: string
  igUsername: string
  videoUrl: string
  audioUrl: string
  message: string
  status: 'queued' | 'displaying' | 'skipped' | 'displayed'
  displayMinutes: number
  createdAt: string
  displayedAt: string | null
  finishedAt: string | null
}

export type PreviewItem = {
  igUrl: string
  igImageUrl: string
  igUsername: string
  videoUrl: string
  audioUrl: string
}

export type CreateDisplayItemInput =
  | {
      sourceType: 'instagram' | 'tiktok' | 'youtube'
      url: string
      message: string
    }
  | {
      sourceType: 'image'
      image: File
      message: string
    }
  | {
      sourceType: 'text'
      message: string
    }

export type AdjustTimeInput = {
  minutes: number
}

export type LoginInput = {
  username: string
  password: string
}

export type RegisterInput = LoginInput

export type AuthUser = {
  id: string
  username: string
}

export type Major = {
  id: string
  name: string
  code: string
}

export type CreateMajorInput = {
  name: string
  code: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const isFormData = typeof FormData !== 'undefined' && init?.body instanceof FormData
  const headers = new Headers(init?.headers)
  if (!isFormData && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    let message = text
    try {
      const data = JSON.parse(text) as { error?: string }
      message = data.error || text
    } catch {
      // Keep the raw response body when the backend does not return JSON.
    }
    throw new Error(`${res.status} ${res.statusText}: ${message}`)
  }
  if (res.status === 204) return undefined as T
  const data = await res.json()
  return withProxiedMedia(data) as T
}

function withProxiedMedia(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(withProxiedMedia)
  if (!value || typeof value !== 'object') return value

  const out: Record<string, unknown> = {}
  for (const [key, item] of Object.entries(value)) {
    if (typeof item === 'string' && (key === 'igImageUrl' || key === 'videoUrl' || key === 'audioUrl')) {
      out[key] = toMediaProxyURL(item)
    } else {
      out[key] = withProxiedMedia(item)
    }
  }
  return out
}

function toMediaProxyURL(mediaUrl: string) {
  if (!mediaUrl) return mediaUrl

  try {
    const parsed = new URL(mediaUrl)
    const apiURL = new URL(API_URL)
    if (parsed.origin === apiURL.origin && parsed.pathname === '/items/media') return mediaUrl

    const allowed =
      PROXY_HOSTS.some((host) => parsed.hostname === host || parsed.hostname.endsWith(`.${host}`)) ||
      isInstagramMediaImageURL(parsed)
    if (!allowed) return mediaUrl

    return `${apiURL.origin}/items/media?src=${encodeURIComponent(mediaUrl)}`
  } catch {
    return mediaUrl
  }
}

function isInstagramMediaImageURL(url: URL) {
  if (url.hostname !== 'instagram.com' && url.hostname !== 'www.instagram.com') return false

  const parts = url.pathname.split('/').filter(Boolean)
  return parts.length === 3 && ['p', 'reel', 'tv'].includes(parts[0]) && parts[2] === 'media'
}

export function getWebSocketURL() {
  const url = new URL(API_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/ws'
  url.search = ''
  return url.toString()
}

export type Config = {
  guestUrl: string
}

export const getConfig = () => request<Config>('/config')

export const previewItem = (url: string) =>
  request<PreviewItem>('/items/preview', {
    method: 'POST',
    body: JSON.stringify({ url }),
  })

export const createDisplayItem = (input: CreateDisplayItemInput) => {
  if (input.sourceType === 'image') {
    const body = new FormData()
    body.append('sourceType', input.sourceType)
    body.append('message', input.message)
    body.append('image', input.image)

    return request<DisplayItem>('/items', {
      method: 'POST',
      body,
    })
  }

  return request<DisplayItem>('/items', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export const getAdminItems = () => request<DisplayItem[]>('/admin/items')

export const getCurrentDisplay = () => request<DisplayItem | null>('/display/current')

export const getDisplayQueue = () => request<DisplayItem[]>('/display/queue')

export const showDisplayItem = (id: string) =>
  request<DisplayItem | null>(`/admin/items/${id}/show`, {
    method: 'POST',
  })

export const skipCurrentDisplayItem = () =>
  request<DisplayItem | null>('/admin/items/current/skip', {
    method: 'POST',
  })

export const finishCurrentDisplayItem = () =>
  request<DisplayItem | null>('/admin/items/current/finish', {
    method: 'POST',
  })

export const addTimeToCurrentDisplayItem = (input: AdjustTimeInput) =>
  request<DisplayItem>('/admin/items/current/add-time', {
    method: 'POST',
    body: JSON.stringify(input),
  })

export const reduceTimeFromCurrentDisplayItem = (input: AdjustTimeInput) =>
  request<DisplayItem>('/admin/items/current/reduce-time', {
    method: 'POST',
    body: JSON.stringify(input),
  })

export const deleteDisplayItem = (id: string) =>
  request<void>(`/admin/items/${id}`, { method: 'DELETE' })

export const loginAdmin = (input: LoginInput) =>
  request<AuthUser>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  })

export const registerAdmin = (input: RegisterInput) =>
  request<AuthUser>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(input),
  })

export const getMajors = () => request<Major[]>('/majors')

export const getMajor = (id: string) => request<Major>(`/majors/${id}`)

export const createMajor = (input: CreateMajorInput) =>
  request<Major>('/majors', {
    method: 'POST',
    body: JSON.stringify(input),
  })

export const deleteMajor = (id: string) =>
  request<void>(`/majors/${id}`, { method: 'DELETE' })

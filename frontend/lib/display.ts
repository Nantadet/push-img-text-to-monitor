import type { DisplayItem } from './api'

export function getRemainingSeconds(item: DisplayItem | null, now = Date.now()) {
  if (!item || item.status !== 'displaying' || !item.displayedAt) return 0
  const start = new Date(item.displayedAt).getTime()
  const end = start + item.displayMinutes * 60_000
  return Math.max(0, Math.ceil((end - now) / 1000))
}

export function formatClock(totalSeconds: number) {
  const safe = Math.max(0, totalSeconds)
  const minutes = Math.floor(safe / 60)
  const seconds = safe % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

export function formatDateTime(value: string | null) {
  if (!value) return '-'
  return new Date(value).toLocaleString('th-TH', {
    dateStyle: 'short',
    timeStyle: 'medium',
  })
}

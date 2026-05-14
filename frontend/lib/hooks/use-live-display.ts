'use client'

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  addTimeToCurrentDisplayItem,
  createDisplayItem,
  deleteDisplayItem,
  finishCurrentDisplayItem,
  getAdminItems,
  getCurrentDisplay,
  getDisplayQueue,
  getWebSocketURL,
  loginAdmin,
  previewItem,
  registerAdmin,
  reduceTimeFromCurrentDisplayItem,
  showDisplayItem,
  skipCurrentDisplayItem,
  type DisplayItem,
} from '@/lib/api'
import { getRemainingSeconds } from '@/lib/display'

function useInvalidateDisplayState() {
  const qc = useQueryClient()

  return () => {
    qc.invalidateQueries({ queryKey: ['items', 'all'] })
    qc.invalidateQueries({ queryKey: ['items', 'current'] })
    qc.invalidateQueries({ queryKey: ['items', 'queue'] })
  }
}

export function useRealtimeDisplaySync(enabled = true) {
  const invalidate = useInvalidateDisplayState()

  useEffect(() => {
    if (!enabled) return

    let socket: WebSocket | null = null
    let disposed = false
    let retryTimer: ReturnType<typeof setTimeout> | null = null

    const connect = () => {
      if (disposed) return
      socket = new WebSocket(getWebSocketURL())
      socket.onmessage = () => invalidate()
      socket.onclose = () => {
        if (disposed) return
        retryTimer = setTimeout(connect, 1500)
      }
      socket.onerror = () => socket?.close()
    }

    connect()

    return () => {
      disposed = true
      if (retryTimer) clearTimeout(retryTimer)
      socket?.close()
    }
  }, [enabled, invalidate])
}

export function useAdminItems() {
  return useQuery({
    queryKey: ['items', 'all'],
    queryFn: getAdminItems,
    refetchInterval: 15000,
  })
}

export function useCurrentDisplay() {
  return useQuery({
    queryKey: ['items', 'current'],
    queryFn: getCurrentDisplay,
    refetchInterval: 5000,
  })
}

export function useDisplayQueue() {
  return useQuery({
    queryKey: ['items', 'queue'],
    queryFn: getDisplayQueue,
    refetchInterval: 5000,
  })
}

export function usePreviewDisplayItem() {
  return useMutation({
    mutationFn: previewItem,
  })
}

export function useCreateDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: createDisplayItem,
    onSuccess: invalidate,
  })
}

export function useAdminLogin() {
  return useMutation({
    mutationFn: loginAdmin,
  })
}

export function useAdminRegister() {
  return useMutation({
    mutationFn: registerAdmin,
  })
}

export function useShowDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: showDisplayItem,
    onSuccess: invalidate,
  })
}

export function useSkipCurrentDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: skipCurrentDisplayItem,
    onSuccess: invalidate,
  })
}

export function useFinishCurrentDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: finishCurrentDisplayItem,
    onSuccess: invalidate,
  })
}

export function useAddTimeToCurrentDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: addTimeToCurrentDisplayItem,
    onSuccess: invalidate,
  })
}

export function useReduceTimeFromCurrentDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: reduceTimeFromCurrentDisplayItem,
    onSuccess: invalidate,
  })
}

export function useDeleteDisplayItem() {
  const invalidate = useInvalidateDisplayState()
  return useMutation({
    mutationFn: deleteDisplayItem,
    onSuccess: invalidate,
  })
}

export function useDisplayCountdown(item: DisplayItem | null) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    setNow(Date.now())
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [item?.id, item?.displayedAt, item?.displayMinutes, item?.status])

  return useMemo(() => getRemainingSeconds(item, now), [item, now])
}

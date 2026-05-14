'use client'

import { useEffect, useRef } from 'react'
import {
  useCurrentDisplay,
  useDisplayCountdown,
  useFinishCurrentDisplayItem,
  useRealtimeDisplaySync,
} from '@/lib/hooks/use-live-display'

export default function DisplayPage() {
  useRealtimeDisplaySync(true)

  const currentQuery = useCurrentDisplay()
  const finishMutation = useFinishCurrentDisplayItem()
  const current = currentQuery.data ?? null
  const remaining = useDisplayCountdown(current)
  const finishingIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (!current || current.status !== 'displaying') {
      finishingIdRef.current = null
      return
    }

    if (finishingIdRef.current && finishingIdRef.current !== current.id) {
      finishingIdRef.current = null
    }

    if (remaining > 0) return
    if (finishingIdRef.current === current.id) return

    finishingIdRef.current = current.id
    finishMutation.mutate()
  }, [current, finishMutation, remaining])

  if (!current) {
    return (
      <main className="grid h-screen w-screen place-items-center bg-black">
        <div className="text-center text-white">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-white/60">display idle</p>
          <h1 className="mt-4 text-4xl md:text-6xl">Waiting for content...</h1>
        </div>
      </main>
    )
  }

  return (
    <main className="flex h-screen w-screen flex-col overflow-hidden bg-black">
      {/* Image area — centered, contained, never breaks aspect ratio */}
      <div className="flex flex-1 items-center justify-center overflow-hidden px-4 pt-4 md:px-8 md:pt-8">
        {current.igImageUrl ? (
          <img
            src={current.igImageUrl}
            alt={current.message}
            className="max-h-full max-w-full rounded-2xl object-contain shadow-2xl"
          />
        ) : (
          <div className="grid h-full w-full place-items-center">
            <h1 className="max-w-4xl px-8 text-center text-4xl leading-tight text-white md:text-6xl">
              {current.message}
            </h1>
          </div>
        )}
      </div>

      {/* Bottom info bar */}
      <div className="shrink-0 px-6 pb-6 pt-4 text-center text-white md:px-10 md:pb-8 md:pt-6">
        <div className="mx-auto flex max-w-4xl flex-col items-center gap-3">
          {/* Instagram handle */}
          {current.igUsername && (
            <p className="text-lg font-semibold tracking-wide text-white/90 md:text-xl">
              @{current.igUsername}
            </p>
          )}

          {/* Message */}
          {current.message && (
            <h2 className="max-w-2xl text-xl leading-snug text-white/80 md:text-2xl">
              {current.message}
            </h2>
          )}

          {/* QR placeholder — user will customize */}
          <div className="mt-2 flex flex-col items-center gap-2">
            <div className="flex h-24 w-24 items-center justify-center rounded-xl bg-white/10 backdrop-blur-sm md:h-28 md:w-28">
              <span className="text-xs text-white/40">QR</span>
            </div>
            <span className="text-xs text-white/50">Scan to submit</span>
          </div>
        </div>
      </div>
    </main>
  )
}

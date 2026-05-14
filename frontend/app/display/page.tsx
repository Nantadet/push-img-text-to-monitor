'use client'

import { useEffect, useRef } from 'react'
import {
  useCurrentDisplay,
  useDisplayCountdown,
  useFinishCurrentDisplayItem,
  useRealtimeDisplaySync,
} from '@/lib/hooks/use-live-display'

/*
  To change the display background color, edit the className below.
  Examples:
    bg-black
    bg-stone-900
    bg-[#1a1a1a]
    bg-gradient-to-br from-purple-900 to-black
*/
const DISPLAY_BG = 'bg-black'

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
      <main className={`grid h-screen w-screen place-items-center ${DISPLAY_BG}`}>
        <div className="text-center text-white">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-white/60">display idle</p>
          <h1 className="mt-4 text-4xl md:text-6xl">Waiting for content...</h1>
        </div>
      </main>
    )
  }

  const isInstagram = current.sourceType === 'instagram'

  return (
    <main className={`flex h-screen w-screen items-center justify-center gap-10 md:gap-16 ${DISPLAY_BG} px-8 md:px-16`}>
      {/* LEFT — Fixed-size image */}
      <div className="flex-shrink-0">
        <div className="relative h-[70vh] w-[52vh] overflow-hidden rounded-2xl shadow-2xl">
          {current.igImageUrl ? (
            <img
              src={current.igImageUrl}
              alt={current.message}
              className="h-full w-full object-cover"
            />
          ) : (
            <div className="grid h-full w-full place-items-center bg-gradient-to-br from-stone-800 to-stone-900">
              <p className="text-white/60">No image</p>
            </div>
          )}
        </div>
      </div>

      {/* RIGHT — Big text */}
      <div className="flex flex-1 flex-col items-start justify-center gap-6 text-white">
        {/* Instagram handle with logo */}
        {isInstagram && current.igUsername && (
          <div className="flex items-center gap-3">
            <InstagramIcon className="h-10 w-10 text-pink-400 md:h-14 md:w-14" />
            <p className="text-3xl font-semibold tracking-wide md:text-5xl lg:text-6xl">
              @{current.igUsername}
            </p>
          </div>
        )}

        {/* Message — very big for far-away viewers */}
        {current.message && (
          <h2 className="max-w-2xl text-5xl font-bold leading-tight md:text-7xl lg:text-8xl">
            {current.message}
          </h2>
        )}

        {/* QR placeholder — user will customize */}
        <div className="mt-4 flex flex-col items-start gap-2">
          <div className="flex h-28 w-28 items-center justify-center rounded-xl bg-white/10 backdrop-blur-sm md:h-32 md:w-32">
            <span className="text-sm text-white/40">QR</span>
          </div>
          <span className="text-base text-white/50 md:text-lg">Scan to submit</span>
        </div>
      </div>
    </main>
  )
}

function InstagramIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <rect width="20" height="20" x="2" y="2" rx="5" ry="5" />
      <path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z" />
      <line x1="17.5" x2="17.51" y1="6.5" y2="6.5" />
    </svg>
  )
}

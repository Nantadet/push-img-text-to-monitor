'use client'

import Link from 'next/link'
import { useEffect, useRef } from 'react'
import { formatClock } from '@/lib/display'
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

  return (
    <main className="min-h-screen px-4 py-4 md:px-6">
      <div className="panel grid min-h-[calc(100vh-2rem)] overflow-hidden rounded-[32px] lg:grid-cols-[1.35fr_0.65fr]">
        {current ? (
          <>
            <div className="relative min-h-[420px] overflow-hidden">
              {current.igImageUrl ? (
                <>
                  <img
                    src={current.igImageUrl}
                    alt={current.message}
                    className="h-full min-h-[420px] w-full object-cover"
                  />
                  <div className="absolute inset-0 bg-gradient-to-t from-black/55 via-black/5 to-transparent" />
                  <div className="absolute bottom-0 left-0 right-0 p-6 text-white md:p-8">
                    <p className="font-mono text-xs uppercase tracking-[0.2em] text-white/75">now showing</p>
                    <h1 className="mt-3 max-w-3xl text-3xl leading-tight md:text-5xl">{current.message}</h1>
                  </div>
                </>
              ) : (
                <div className="grid min-h-[420px] place-items-center bg-[var(--panel-strong)] p-8 text-center">
                  <div className="max-w-3xl">
                    <p className="font-mono text-xs uppercase tracking-[0.2em] text-[var(--accent-deep)]">
                      now showing
                    </p>
                    <h1 className="mt-5 text-4xl leading-tight md:text-6xl">{current.message}</h1>
                  </div>
                </div>
              )}
            </div>

            <div className="flex flex-col justify-between gap-8 p-6 md:p-8">
              <div>
                <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">display control</p>
                <div className="mt-5 rounded-[28px] bg-[var(--accent-soft)] px-6 py-5 text-[var(--accent-deep)]">
                  <p className="font-mono text-sm uppercase tracking-[0.18em]">time left</p>
                  <p className="mt-2 text-6xl leading-none">{formatClock(remaining)}</p>
                  <p className="mt-3 text-sm">Duration: {current.displayMinutes} minute(s)</p>
                </div>
              </div>

              <div className="space-y-4">
                <div className="rounded-[24px] border border-[var(--line)] bg-white/75 p-5">
                  <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">source</p>
                  {current.igUrl ? (
                    <a
                      href={current.igUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="mt-3 block text-lg text-stone-900 underline decoration-[var(--accent)] underline-offset-4"
                    >
                      {current.igUsername || current.igUrl}
                    </a>
                  ) : (
                    <p className="mt-3 text-lg capitalize text-stone-900">{current.sourceType}</p>
                  )}
                </div>

                <div className="rounded-[24px] border border-[var(--line)] bg-white/75 p-5">
                  <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">sync</p>
                  <p className="mt-3 text-sm leading-6 text-stone-700">
                    This page counts down locally from displayedAt plus displayMinutes. When time reaches zero, it asks
                    backend to mark the item as displayed and promote the oldest queued item.
                  </p>
                </div>
              </div>

              <div className="flex flex-wrap gap-3">
                <Link href="/admin" className="rounded-full border border-[var(--line)] px-4 py-2 text-sm">
                  Open Admin
                </Link>
                <Link href="/guest" className="rounded-full border border-[var(--line)] px-4 py-2 text-sm">
                  Open Guest
                </Link>
              </div>
            </div>
          </>
        ) : (
          <div className="grid min-h-[calc(100vh-2rem)] place-items-center p-6">
            <div className="max-w-xl text-center">
              <p className="font-mono text-xs uppercase tracking-[0.2em] text-[var(--muted)]">display idle</p>
              <h1 className="mt-4 text-5xl">Waiting for the next queued item</h1>
              <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
                When backend has no active item, the next queued item with the oldest createdAt will be promoted as
                soon as one is available.
              </p>
            </div>
          </div>
        )}
      </div>
    </main>
  )
}

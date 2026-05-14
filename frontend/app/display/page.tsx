'use client'

import { useEffect, useRef, useState } from 'react'
import QRCode from 'react-qr-code'
import { useQuery } from '@tanstack/react-query'
import { getConfig } from '@/lib/api'
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

  const configQuery = useQuery({
    queryKey: ['config'],
    queryFn: getConfig,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  })

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
  const isTiktok = current.sourceType === 'tiktok'
  const isYoutube = current.sourceType === 'youtube'
  const hasVideo = Boolean(current.videoUrl)
  const hasAudio = Boolean(current.audioUrl)
  const guestUrl = configQuery.data?.guestUrl ?? ''

  return (
    <main className={`flex h-screen w-screen items-center justify-center gap-10 md:gap-16 ${DISPLAY_BG} px-8 md:px-16`}>
      {/* LEFT — Media (image / video / audio visualizer) */}
      <div className="flex-shrink-0">
        <div className="relative h-[70vh] w-[52vh] overflow-hidden rounded-2xl shadow-2xl">
          {hasVideo ? (
            <video
              src={current.videoUrl}
              autoPlay
              loop
              muted
              playsInline
              className="h-full w-full object-contain bg-black"
            />
          ) : hasAudio ? (
            <AudioVisualizer audioUrl={current.audioUrl} thumbnailUrl={current.igImageUrl} />
          ) : current.igImageUrl ? (
            <img
              src={current.igImageUrl}
              alt={current.message}
              className="h-full w-full object-contain bg-black"
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
        {/* Platform handle with logo */}
        {isInstagram && current.igUsername && (
          <div className="flex items-center gap-3">
            <InstagramIcon className="h-10 w-10 text-pink-400 md:h-14 md:w-14" />
            <p className="text-3xl font-semibold tracking-wide md:text-5xl lg:text-6xl">
              @{current.igUsername}
            </p>
          </div>
        )}
        {isTiktok && current.igUsername && (
          <div className="flex items-center gap-3">
            <TikTokIcon className="h-10 w-10 text-white md:h-14 md:w-14" />
            <p className="text-3xl font-semibold tracking-wide md:text-5xl lg:text-6xl">
              @{current.igUsername}
            </p>
          </div>
        )}
        {isYoutube && (
          <div className="flex items-center gap-3">
            <YouTubeIcon className="h-10 w-10 text-red-500 md:h-14 md:w-14" />
            <p className="text-3xl font-semibold tracking-wide md:text-5xl lg:text-6xl">
              YouTube
            </p>
          </div>
        )}

        {/* Message — very big for far-away viewers */}
        {current.message && (
          <h2 className="max-w-2xl text-5xl font-bold leading-tight md:text-7xl lg:text-8xl">
            {current.message}
          </h2>
        )}

        {/* QR Code — scan to open guest page */}
        {guestUrl && (
          <div className="mt-4 flex flex-col items-start gap-3">
            <div className="rounded-xl bg-white p-3">
              <QRCode value={guestUrl} size={140} />
            </div>
            <span className="text-base text-white/70 md:text-lg">Scan to submit</span>
          </div>
        )}
      </div>
    </main>
  )
}

function AudioVisualizer({ audioUrl, thumbnailUrl }: { audioUrl: string; thumbnailUrl: string }) {
  const audioRef = useRef<HTMLAudioElement>(null)

  useEffect(() => {
    if (audioRef.current) {
      audioRef.current.play().catch(() => {
        // Autoplay may be blocked; user interaction required on some browsers
      })
    }
  }, [audioUrl])

  return (
    <div className="relative flex h-full w-full flex-col items-center justify-center bg-gradient-to-br from-stone-900 to-black">
      {/* Thumbnail background */}
      {thumbnailUrl && (
        <div
          className="absolute inset-0 opacity-20 blur-sm"
          style={{
            backgroundImage: `url(${thumbnailUrl})`,
            backgroundSize: 'cover',
            backgroundPosition: 'center',
          }}
        />
      )}

      {/* Animated bars */}
      <div className="relative z-10 flex items-end gap-1.5">
        {Array.from({ length: 12 }).map((_, i) => (
          <div
            key={i}
            className="w-3 rounded-full bg-red-500"
            style={{
              height: `${20 + Math.random() * 60}%`,
              animation: `audioBar 0.6s ease-in-out ${i * 0.08}s infinite alternate`,
            }}
          />
        ))}
      </div>

      <p className="relative z-10 mt-6 text-xl font-medium text-white/80">Now Playing</p>

      <audio ref={audioRef} src={audioUrl} autoPlay loop className="hidden" />

      <style jsx>{`
        @keyframes audioBar {
          0% {
            height: 20%;
          }
          100% {
            height: 85%;
          }
        }
      `}</style>
    </div>
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

function TikTokIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
    >
      <path d="M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-5.2 1.74 2.89 2.89 0 0 1 2.31-4.64 2.93 2.93 0 0 1 .88.13V9.4a6.84 6.84 0 0 0-1-.05A6.33 6.33 0 0 0 5 20.1a6.34 6.34 0 0 0 10.86-4.43v-7a8.16 8.16 0 0 0 4.77 1.52v-3.4a4.85 4.85 0 0 1-1-.1z" />
    </svg>
  )
}

function YouTubeIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
    >
      <path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z" />
    </svg>
  )
}

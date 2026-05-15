'use client'

import { useEffect, useState } from 'react'
import { useCreateDisplayItem, usePreviewDisplayItem } from '@/lib/hooks/use-live-display'

type SubmitMode = 'link' | 'image' | 'text'
type Platform = 'instagram' | 'tiktok' | 'youtube' | 'unknown'

const submitModes: Array<{ value: SubmitMode; label: string }> = [
  { value: 'link', label: 'Link + Message' },
  { value: 'image', label: 'Image + Message' },
  { value: 'text', label: 'Message Only' },
]

function detectPlatform(url: string): Platform {
  try {
    const u = new URL(url)
    const host = u.hostname.toLowerCase()
    if (host === 'instagram.com' || host === 'www.instagram.com') return 'instagram'
    if (['tiktok.com', 'www.tiktok.com', 'm.tiktok.com', 'vm.tiktok.com', 'vt.tiktok.com'].includes(host)) return 'tiktok'
    if (host === 'youtube.com' || host === 'www.youtube.com' || host === 'youtu.be' || host === 'music.youtube.com') return 'youtube'
    return 'unknown'
  } catch {
    return 'unknown'
  }
}

function platformLabel(platform: Platform): string {
  switch (platform) {
    case 'instagram': return 'Instagram'
    case 'tiktok': return 'TikTok'
    case 'youtube': return 'YouTube'
    default: return 'Link'
  }
}

export default function GuestPage() {
  const previewMutation = usePreviewDisplayItem()
  const createMutation = useCreateDisplayItem()

  const [mode, setMode] = useState<SubmitMode>('link')
  const [url, setUrl] = useState('')
  const [message, setMessage] = useState('')
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [imageInputKey, setImageInputKey] = useState(0)
  const [imagePreviewURL, setImagePreviewURL] = useState('')
  const [previewKey, setPreviewKey] = useState('')

  const trimmedUrl = url.trim()
  const trimmedMessage = message.trim()
  const platform = detectPlatform(trimmedUrl)
  const linkPreview =
    mode === 'link' && previewMutation.data && previewKey === trimmedUrl ? previewMutation.data : null
  const previewImageURL = mode === 'link' ? linkPreview?.igImageUrl : imagePreviewURL
  const canSubmit = Boolean(
    trimmedMessage.length > 0 &&
      ((mode === 'link' && trimmedUrl.length > 0) || (mode === 'image' && imageFile) || mode === 'text'),
  )

  useEffect(() => {
    if (!imageFile) {
      setImagePreviewURL('')
      return
    }

    const nextURL = URL.createObjectURL(imageFile)
    setImagePreviewURL(nextURL)
    return () => URL.revokeObjectURL(nextURL)
  }, [imageFile])

  function handleModeChange(nextMode: SubmitMode) {
    setMode(nextMode)
    createMutation.reset()
    if (nextMode !== 'link') {
      setPreviewKey('')
      previewMutation.reset()
    }
  }

  async function handlePreview() {
    if (!trimmedUrl) return false

    setPreviewKey('')
    previewMutation.reset()
    try {
      await previewMutation.mutateAsync(trimmedUrl)
      setPreviewKey(trimmedUrl)
      return true
    } catch {
      return false
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!trimmedMessage) return

    try {
      if (mode === 'link') {
        if (!trimmedUrl) return
        if (!linkPreview) {
          const previewLoaded = await handlePreview()
          if (!previewLoaded) return
        }
        await createMutation.mutateAsync({
          sourceType: platform === 'unknown' ? 'instagram' : platform,
          url: trimmedUrl,
          message: trimmedMessage,
        })
        setUrl('')
        setPreviewKey('')
        previewMutation.reset()
      } else if (mode === 'image') {
        if (!imageFile) return
        await createMutation.mutateAsync({
          sourceType: 'image',
          image: imageFile,
          message: trimmedMessage,
        })
        setImageFile(null)
        setImageInputKey((value) => value + 1)
      } else {
        await createMutation.mutateAsync({
          sourceType: 'text',
          message: trimmedMessage,
        })
      }

      setMessage('')
    } catch {
      return
    }
  }

  const hasVideoPreview = Boolean(linkPreview?.videoUrl)
  const hasAudioPreview = Boolean(linkPreview?.audioUrl)

  return (
    <main className="min-h-screen px-6 py-10 md:px-10">
      <div className="mx-auto grid max-w-6xl gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="panel rounded-[28px] p-6 md:p-8">
          <div className="mb-8">
            <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--accent-deep)]">guest submit</p>
            <h1 className="mt-3 text-4xl md:text-5xl">Send to live display</h1>
            <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
              Choose the content type, write the message, then send it into the live queue.
            </p>
          </div>

          <form id="guest-form" onSubmit={handleSubmit} className="space-y-5">
            <div>
              <span className="mb-2 block text-sm font-medium text-stone-800">Submit Type</span>
              <div className="flex flex-wrap gap-2 rounded-2xl border border-[var(--line)] bg-white/65 p-1">
                {submitModes.map((item) => (
                  <button
                    key={item.value}
                    type="button"
                    onClick={() => handleModeChange(item.value)}
                    className={`rounded-xl px-4 py-2 text-sm transition ${
                      mode === item.value ? 'bg-stone-900 text-white' : 'text-stone-700 hover:bg-white'
                    }`}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
            </div>

            {mode === 'link' ? (
              <label className="block">
                <span className="mb-2 block text-sm font-medium text-stone-800">
                  {platformLabel(platform)} URL
                </span>
                <input
                  value={url}
                  onChange={(e) => {
                    setUrl(e.target.value)
                    setPreviewKey('')
                  }}
                  placeholder="https://www.instagram.com/...  or  tiktok.com/...  or  youtube.com/..."
                  className="w-full rounded-2xl border border-[var(--line)] bg-[var(--panel-strong)] px-4 py-3 text-sm outline-none transition focus:border-[var(--accent)]"
                />
              </label>
            ) : null}

            {mode === 'image' ? (
              <label className="block">
                <span className="mb-2 block text-sm font-medium text-stone-800">Image</span>
                <input
                  key={imageInputKey}
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
                  className="w-full rounded-2xl border border-[var(--line)] bg-[var(--panel-strong)] px-4 py-3 text-sm outline-none file:mr-4 file:rounded-full file:border-0 file:bg-stone-900 file:px-4 file:py-2 file:text-sm file:text-white"
                />
              </label>
            ) : null}

            <label className="block">
              <span className="mb-2 block text-sm font-medium text-stone-800">Message</span>
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                maxLength={200}
                placeholder="Write the message that should appear on the display screen."
                className="min-h-32 w-full resize-y rounded-2xl border border-[var(--line)] bg-[var(--panel-strong)] px-4 py-3 text-sm outline-none transition focus:border-[var(--accent)]"
              />
              <span className="mt-1 block text-right text-xs text-[var(--muted)]">{message.length}/200</span>
            </label>

            <div className="flex flex-wrap gap-3">
              {mode === 'link' ? (
                <button
                  type="button"
                  onClick={handlePreview}
                  disabled={previewMutation.isPending || !trimmedUrl}
                  className="rounded-full bg-stone-900 px-5 py-3 text-sm font-medium text-white disabled:opacity-50"
                >
                  {previewMutation.isPending ? 'Loading preview...' : 'Load Preview'}
                </button>
              ) : null}
            </div>

            {previewMutation.error ? <p className="text-sm text-rose-700">{previewMutation.error.message}</p> : null}
            {createMutation.error ? <p className="text-sm text-rose-700">{createMutation.error.message}</p> : null}
            {createMutation.isSuccess ? (
              <p className="text-sm text-emerald-700">Submission added to the live queue.</p>
            ) : null}
          </form>
        </section>

        <aside className="panel flex flex-col rounded-[28px] p-6 md:p-8">
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">preview</p>
          <div className="mt-5 flex-1 overflow-hidden rounded-[24px] border border-[var(--line)] bg-white">
            {linkPreview?.embedUrl ? (
              <>
                <iframe
                  src={linkPreview.embedUrl}
                  className="h-[360px] w-full"
                  allow="autoplay; encrypted-media"
                  sandbox="allow-scripts allow-same-origin allow-presentation"
                />
                <PreviewMessage
                  title={linkPreview?.igUsername || platformLabel(platform)}
                  message={trimmedMessage}
                  igUrl={linkPreview?.igUrl}
                />
              </>
            ) : hasVideoPreview ? (
              <>
                <video
                  src={linkPreview?.videoUrl}
                  autoPlay
                  loop
                  muted
                  playsInline
                  className="h-[360px] w-full object-cover"
                />
                <PreviewMessage
                  title={linkPreview?.igUsername || platformLabel(platform)}
                  message={trimmedMessage}
                  igUrl={linkPreview?.igUrl}
                />
              </>
            ) : hasAudioPreview ? (
              <div className="flex h-[360px] flex-col items-center justify-center gap-4 bg-gradient-to-br from-stone-900 to-black">
                {previewImageURL && (
                  <img src={previewImageURL} alt="" className="h-48 w-48 rounded-xl object-cover" />
                )}
                <div className="flex items-center gap-2 text-white/80">
                  <svg className="h-6 w-6" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z" />
                  </svg>
                  <span className="text-lg font-medium">Audio Preview</span>
                </div>
                <PreviewMessage
                  title={platformLabel(platform)}
                  message={trimmedMessage}
                  igUrl={linkPreview?.igUrl}
                />
              </div>
            ) : previewImageURL ? (
              <>
                <img
                  src={previewImageURL}
                  alt={trimmedMessage || linkPreview?.igUsername || 'submission preview'}
                  className="h-[360px] w-full object-cover"
                />
                <PreviewMessage
                  title={
                    mode === 'link'
                      ? linkPreview?.igUsername || platformLabel(platform)
                      : imageFile?.name || 'Uploaded Image'
                  }
                  message={trimmedMessage}
                  igUrl={mode === 'link' ? linkPreview?.igUrl : ''}
                />
              </>
            ) : mode === 'text' ? (
              <div className="flex min-h-[480px] items-center justify-center p-6 text-center">
                <div>
                  <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--accent-deep)]">
                    message only
                  </p>
                  <p className="mt-4 text-3xl leading-tight">{trimmedMessage || 'Your message will appear here.'}</p>
                </div>
              </div>
            ) : (
              <div className="flex h-[480px] flex-col justify-between p-5">
                <div>
                  <p className="text-2xl">
                    {mode === 'link' ? 'Waiting for link preview' : 'Waiting for image'}
                  </p>
                  <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
                    {mode === 'link'
                      ? 'Load preview to check the content before sending.'
                      : 'Choose an image to preview it before sending.'}
                  </p>
                </div>
                <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">
                  {mode === 'link' ? 'instagram / tiktok / youtube' : 'jpg / png / webp / gif'}
                </p>
              </div>
            )}
          </div>

          {/* Submit button moved below preview */}
          <button
            form="guest-form"
            type="submit"
            disabled={createMutation.isPending || !canSubmit}
            className="mt-5 ml-auto block rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-medium text-white disabled:opacity-50"
          >
            {createMutation.isPending ? 'Sending to queue...' : 'Send To Queue'}
          </button>
        </aside>
      </div>
    </main>
  )
}

function PreviewMessage({ title, message, igUrl }: { title: string; message: string; igUrl?: string }) {
  return (
    <div className="space-y-3 p-5">
      <p className="font-mono text-xs uppercase tracking-[0.16em] text-[var(--accent-deep)]">{title}</p>
      <p className="text-xl leading-snug">{message || 'Your message will appear here.'}</p>
      {igUrl ? (
        <a
          href={igUrl}
          target="_blank"
          rel="noreferrer"
          className="inline-flex rounded-full border border-[var(--line)] px-4 py-2 text-sm text-stone-700 hover:border-[var(--accent)]"
        >
          Open Link
        </a>
      ) : null}
    </div>
  )
}

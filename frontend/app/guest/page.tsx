'use client'

import { useEffect, useState } from 'react'
import { useCreateDisplayItem, usePreviewDisplayItem } from '@/lib/hooks/use-live-display'

type SubmitMode = 'instagram' | 'image' | 'text'

const submitModes: Array<{ value: SubmitMode; label: string }> = [
  { value: 'instagram', label: 'Instagram + Message' },
  { value: 'image', label: 'Image + Message' },
  { value: 'text', label: 'Message Only' },
]

export default function GuestPage() {
  const previewMutation = usePreviewDisplayItem()
  const createMutation = useCreateDisplayItem()

  const [mode, setMode] = useState<SubmitMode>('instagram')
  const [igUrl, setIGUrl] = useState('')
  const [message, setMessage] = useState('')
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [imageInputKey, setImageInputKey] = useState(0)
  const [imagePreviewURL, setImagePreviewURL] = useState('')
  const [previewKey, setPreviewKey] = useState('')

  const trimmedIGUrl = igUrl.trim()
  const trimmedMessage = message.trim()
  const instagramPreview =
    mode === 'instagram' && previewMutation.data && previewKey === trimmedIGUrl ? previewMutation.data : null
  const previewImageURL = mode === 'instagram' ? instagramPreview?.igImageUrl : imagePreviewURL
  const canSubmit = Boolean(
    trimmedMessage.length > 0 &&
      ((mode === 'instagram' && trimmedIGUrl.length > 0) || (mode === 'image' && imageFile) || mode === 'text'),
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
    if (nextMode !== 'instagram') {
      setPreviewKey('')
      previewMutation.reset()
    }
  }

  async function handlePreview() {
    if (!trimmedIGUrl) return false

    setPreviewKey('')
    previewMutation.reset()
    try {
      await previewMutation.mutateAsync(trimmedIGUrl)
      setPreviewKey(trimmedIGUrl)
      return true
    } catch {
      return false
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!trimmedMessage) return

    try {
      if (mode === 'instagram') {
        if (!trimmedIGUrl) return
        if (!instagramPreview) {
          const previewLoaded = await handlePreview()
          if (!previewLoaded) return
        }
        await createMutation.mutateAsync({
          sourceType: 'instagram',
          igUrl: trimmedIGUrl,
          message: trimmedMessage,
        })
        setIGUrl('')
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

          <form onSubmit={handleSubmit} className="space-y-5">
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

            {mode === 'instagram' ? (
              <label className="block">
                <span className="mb-2 block text-sm font-medium text-stone-800">Instagram URL</span>
                <input
                  value={igUrl}
                  onChange={(e) => {
                    setIGUrl(e.target.value)
                    setPreviewKey('')
                  }}
                  placeholder="https://www.instagram.com/..."
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
              {mode === 'instagram' ? (
                <button
                  type="button"
                  onClick={handlePreview}
                  disabled={previewMutation.isPending || !trimmedIGUrl}
                  className="rounded-full bg-stone-900 px-5 py-3 text-sm font-medium text-white disabled:opacity-50"
                >
                  {previewMutation.isPending ? 'Loading preview...' : 'Load Preview'}
                </button>
              ) : null}
              <button
                type="submit"
                disabled={createMutation.isPending || !canSubmit}
                className="rounded-full bg-[var(--accent)] px-5 py-3 text-sm font-medium text-white disabled:opacity-50"
              >
                {createMutation.isPending ? 'Sending to queue...' : 'Send To Queue'}
              </button>
            </div>

            {previewMutation.error ? <p className="text-sm text-rose-700">{previewMutation.error.message}</p> : null}
            {createMutation.error ? <p className="text-sm text-rose-700">{createMutation.error.message}</p> : null}
            {createMutation.isSuccess ? (
              <p className="text-sm text-emerald-700">Submission added to the live queue.</p>
            ) : null}
          </form>
        </section>

        <aside className="panel rounded-[28px] p-6 md:p-8">
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">preview</p>
          <div className="mt-5 overflow-hidden rounded-[24px] border border-[var(--line)] bg-white">
            {previewImageURL ? (
              <>
                <img
                  src={previewImageURL}
                  alt={trimmedMessage || instagramPreview?.igUsername || 'submission preview'}
                  className="h-[360px] w-full object-cover"
                />
                <PreviewMessage
                  title={
                    mode === 'instagram'
                      ? instagramPreview?.igUsername || 'Instagram'
                      : imageFile?.name || 'Uploaded Image'
                  }
                  message={trimmedMessage}
                  igUrl={mode === 'instagram' ? instagramPreview?.igUrl : ''}
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
                    {mode === 'instagram' ? 'Waiting for Instagram preview' : 'Waiting for image'}
                  </p>
                  <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
                    {mode === 'instagram'
                      ? 'Load preview to check the Instagram image before sending.'
                      : 'Choose an image to preview it before sending.'}
                  </p>
                </div>
                <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">
                  {mode === 'instagram' ? 'public profile / post / reel' : 'jpg / png / webp / gif'}
                </p>
              </div>
            )}
          </div>
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
          Open Instagram
        </a>
      ) : null}
    </div>
  )
}

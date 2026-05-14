'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import type { DisplayItem } from '@/lib/api'
import { formatClock, formatDateTime } from '@/lib/display'
import {
  useAddTimeToCurrentDisplayItem,
  useAdminItems,
  useAdminLogin,
  useAdminRegister,
  useCurrentDisplay,
  useDeleteDisplayItem,
  useDisplayCountdown,
  useDisplayQueue,
  useRealtimeDisplaySync,
  useReduceTimeFromCurrentDisplayItem,
  useShowDisplayItem,
  useSkipCurrentDisplayItem,
} from '@/lib/hooks/use-live-display'

const SESSION_KEY = 'live-display-admin'

type AuthMode = 'login' | 'register'

export default function AdminPage() {
  const [mounted, setMounted] = useState(false)
  const [session, setSession] = useState<string | null>(null)
  const [mode, setMode] = useState<AuthMode>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  useEffect(() => {
    setMounted(true)
    setSession(window.localStorage.getItem(SESSION_KEY))
  }, [])

  if (!mounted) return null

  if (!session) {
    return (
      <AdminAuth
        mode={mode}
        username={username}
        password={password}
        onModeChange={setMode}
        onUsernameChange={setUsername}
        onPasswordChange={setPassword}
        onSuccess={(nextUsername) => {
          window.localStorage.setItem(SESSION_KEY, nextUsername)
          setSession(nextUsername)
        }}
      />
    )
  }

  return (
    <AdminDashboard
      session={session}
      onLogout={() => {
        window.localStorage.removeItem(SESSION_KEY)
        setSession(null)
      }}
    />
  )
}

function AdminAuth(props: {
  mode: AuthMode
  username: string
  password: string
  onModeChange: (value: AuthMode) => void
  onUsernameChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onSuccess: (username: string) => void
}) {
  const loginMutation = useAdminLogin()
  const registerMutation = useAdminRegister()
  const activeMutation = props.mode === 'login' ? loginMutation : registerMutation

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    const payload = {
      username: props.username.trim(),
      password: props.password,
    }

    const user =
      props.mode === 'login'
        ? await loginMutation.mutateAsync(payload)
        : await registerMutation.mutateAsync(payload)

    props.onSuccess(user.username)
  }

  return (
    <main className="min-h-screen px-6 py-10 md:px-10">
      <div className="mx-auto max-w-3xl">
        <div className="panel rounded-[28px] p-6 md:p-8">
          <div className="mb-8 flex items-start justify-between gap-4">
            <div>
              <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--accent-deep)]">
                admin access
              </p>
              <h1 className="mt-3 text-4xl md:text-5xl">
                {props.mode === 'login' ? 'Login to dashboard' : 'Create admin account'}
              </h1>
              <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
                Register once if the database has no admin yet. After that, log in and control the live queue.
              </p>
            </div>
            <Link href="/" className="font-mono text-sm text-[var(--muted)] underline-offset-4 hover:underline">
              home
            </Link>
          </div>

          <div className="mb-6 inline-flex rounded-full border border-[var(--line)] bg-white/70 p-1 text-sm">
            <button
              type="button"
              onClick={() => props.onModeChange('login')}
              className={`rounded-full px-4 py-2 ${props.mode === 'login' ? 'bg-stone-900 text-white' : 'text-stone-700'}`}
            >
              Login
            </button>
            <button
              type="button"
              onClick={() => props.onModeChange('register')}
              className={`rounded-full px-4 py-2 ${props.mode === 'register' ? 'bg-stone-900 text-white' : 'text-stone-700'}`}
            >
              Register
            </button>
          </div>

          <form onSubmit={handleSubmit} className="grid gap-4">
            <input
              value={props.username}
              onChange={(e) => props.onUsernameChange(e.target.value)}
              placeholder="username"
              className="rounded-2xl border border-[var(--line)] bg-[var(--panel-strong)] px-4 py-3 text-sm outline-none focus:border-[var(--accent)]"
            />
            <input
              type="password"
              value={props.password}
              onChange={(e) => props.onPasswordChange(e.target.value)}
              placeholder="password"
              className="rounded-2xl border border-[var(--line)] bg-[var(--panel-strong)] px-4 py-3 text-sm outline-none focus:border-[var(--accent)]"
            />
            <button
              type="submit"
              disabled={activeMutation.isPending}
              className="rounded-full bg-stone-900 px-5 py-3 text-sm font-medium text-white disabled:opacity-50"
            >
              {activeMutation.isPending
                ? props.mode === 'login'
                  ? 'Signing in...'
                  : 'Creating account...'
                : props.mode === 'login'
                  ? 'Login'
                  : 'Register'}
            </button>
          </form>

          {activeMutation.error ? (
            <p className="mt-4 text-sm text-rose-700">{activeMutation.error.message}</p>
          ) : null}
        </div>
      </div>
    </main>
  )
}

function AdminDashboard({ session, onLogout }: { session: string; onLogout: () => void }) {
  useRealtimeDisplaySync(true)

  const itemsQuery = useAdminItems()
  const queueQuery = useDisplayQueue()
  const currentQuery = useCurrentDisplay()
  const showMutation = useShowDisplayItem()
  const skipMutation = useSkipCurrentDisplayItem()
  const addTimeMutation = useAddTimeToCurrentDisplayItem()
  const reduceTimeMutation = useReduceTimeFromCurrentDisplayItem()
  const deleteMutation = useDeleteDisplayItem()

  const current = currentQuery.data ?? null
  const remaining = useDisplayCountdown(current)
  const items = itemsQuery.data ?? []
  const queuedItems = queueQuery.data ?? []

  return (
    <main className="min-h-screen px-6 py-10 md:px-10">
      <div className="mx-auto max-w-7xl space-y-6">
        <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--accent-deep)]">dashboard</p>
            <h1 className="mt-3 text-4xl md:text-5xl">Control what the audience sees.</h1>
            <p className="mt-3 max-w-3xl text-sm leading-6 text-[var(--muted)]">
              Review all submissions, manage the queue, and change the current display item in real time.
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link href="/display" className="rounded-full border border-[var(--line)] px-4 py-2 text-sm">
              Open Display
            </Link>
            <button onClick={onLogout} className="rounded-full bg-stone-900 px-4 py-2 text-sm text-white">
              Logout {session}
            </button>
          </div>
        </div>

        <section className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
          <div className="panel rounded-[28px] p-6 md:p-8">
            <div className="mb-5 flex items-start justify-between gap-4">
              <div>
                <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">current display</p>
                <h2 className="mt-3 text-3xl">Now showing</h2>
              </div>
              <div className="rounded-full bg-[var(--accent-soft)] px-4 py-2 font-mono text-lg text-[var(--accent-deep)]">
                {formatClock(remaining)}
              </div>
            </div>

            {current ? (
              <div className="grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
                <ItemPreviewTile item={current} className="h-[320px] w-full rounded-[24px]" />
                <div className="flex flex-col justify-between gap-4">
                  <div>
                    <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--accent-deep)]">
                      {current.status}
                    </p>
                    <p className="mt-3 text-2xl leading-tight">{current.message}</p>
                    <div className="mt-4">
                      <ItemSource item={current} />
                    </div>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2">
                    <button
                      onClick={() => skipMutation.mutate()}
                      disabled={skipMutation.isPending}
                      className="rounded-2xl bg-stone-900 px-4 py-3 text-sm text-white disabled:opacity-50"
                    >
                      Skip Current
                    </button>
                    <button
                      onClick={() => addTimeMutation.mutate({ minutes: 1 })}
                      disabled={addTimeMutation.isPending}
                      className="rounded-2xl border border-[var(--line)] px-4 py-3 text-sm disabled:opacity-50"
                    >
                      Add 1m
                    </button>
                    <button
                      onClick={() => addTimeMutation.mutate({ minutes: 3 })}
                      disabled={addTimeMutation.isPending}
                      className="rounded-2xl border border-[var(--line)] px-4 py-3 text-sm disabled:opacity-50"
                    >
                      Add 3m
                    </button>
                    <button
                      onClick={() => reduceTimeMutation.mutate({ minutes: 1 })}
                      disabled={reduceTimeMutation.isPending}
                      className="rounded-2xl border border-[var(--line)] px-4 py-3 text-sm disabled:opacity-50"
                    >
                      Reduce 1m
                    </button>
                  </div>

                  <div className="grid gap-1 text-sm text-[var(--muted)]">
                    <p>display length: {current.displayMinutes} minute(s)</p>
                    <p>created: {formatDateTime(current.createdAt)}</p>
                    <p>displayed: {formatDateTime(current.displayedAt)}</p>
                    <p>finished: {formatDateTime(current.finishedAt)}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="rounded-[24px] border border-dashed border-[var(--line)] p-8 text-sm text-[var(--muted)]">
                No item is being displayed right now.
              </div>
            )}
          </div>

          <div className="panel rounded-[28px] p-6 md:p-8">
            <div className="mb-5">
              <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">queue</p>
              <h2 className="mt-3 text-3xl">Oldest queued first</h2>
            </div>
            <div className="space-y-3">
              {queuedItems.length === 0 ? (
                <p className="rounded-[22px] border border-dashed border-[var(--line)] p-5 text-sm text-[var(--muted)]">
                  Queue is empty.
                </p>
              ) : (
                queuedItems.map((item) => (
                  <article key={item.id} className="rounded-[22px] border border-[var(--line)] bg-white/75 p-4">
                    <div className="flex gap-4">
                      <ItemPreviewTile item={item} className="h-24 w-24 shrink-0 rounded-2xl" />
                      <div className="min-w-0 flex-1">
                        <p className="text-base leading-6">{item.message}</p>
                        <p className="mt-2 font-mono text-xs text-[var(--muted)]">
                          {sourceLabel(item)}
                        </p>
                        <p className="mt-1 text-xs text-[var(--muted)]">created {formatDateTime(item.createdAt)}</p>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-wrap gap-2">
                      <button
                        onClick={() => showMutation.mutate(item.id)}
                        disabled={showMutation.isPending}
                        className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm text-white disabled:opacity-50"
                      >
                        Show Now
                      </button>
                      <button
                        onClick={() => deleteMutation.mutate(item.id)}
                        disabled={deleteMutation.isPending}
                        className="rounded-full border border-[var(--line)] px-4 py-2 text-sm disabled:opacity-50"
                      >
                        Delete
                      </button>
                    </div>
                  </article>
                ))
              )}
            </div>
            {queueQuery.error ? <p className="mt-4 text-sm text-rose-700">{queueQuery.error.message}</p> : null}
          </div>
        </section>

        <section className="panel rounded-[28px] p-6 md:p-8">
          <div className="mb-5">
            <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--muted)]">submissions</p>
            <h2 className="mt-3 text-3xl">All user data</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full border-separate border-spacing-y-3 text-left text-sm">
              <thead>
                <tr className="text-[var(--muted)]">
                  <th className="px-3">status</th>
                  <th className="px-3">message</th>
                  <th className="px-3">source</th>
                  <th className="px-3">minutes</th>
                  <th className="px-3">created</th>
                  <th className="px-3">displayed</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.id} className="rounded-2xl bg-white/70">
                    <td className="rounded-l-2xl px-3 py-3 font-mono text-xs uppercase text-[var(--accent-deep)]">
                      {item.status}
                    </td>
                    <td className="px-3 py-3">{item.message}</td>
                    <td className="px-3 py-3">
                      <ItemSource item={item} />
                    </td>
                    <td className="px-3 py-3">{item.displayMinutes}</td>
                    <td className="px-3 py-3">{formatDateTime(item.createdAt)}</td>
                    <td className="rounded-r-2xl px-3 py-3">{formatDateTime(item.displayedAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {itemsQuery.error ? <p className="mt-4 text-sm text-rose-700">{itemsQuery.error.message}</p> : null}
        </section>
      </div>
    </main>
  )
}

function ItemPreviewTile({ item, className }: { item: DisplayItem; className: string }) {
  if (item.igImageUrl) {
    return <img src={item.igImageUrl} alt={item.message} className={`${className} object-cover`} />
  }

  return (
    <div className={`${className} grid place-items-center bg-[var(--panel-strong)] p-4 text-center`}>
      <p className="max-h-full overflow-hidden text-sm leading-5 text-stone-800">{item.message}</p>
    </div>
  )
}

function ItemSource({ item }: { item: DisplayItem }) {
  if (item.igUrl) {
    return (
      <a
        href={item.igUrl}
        target="_blank"
        rel="noreferrer"
        className="text-sm text-[var(--accent-deep)] underline"
      >
        {item.igUsername || 'Open Instagram'}
      </a>
    )
  }

  return <span className="text-sm capitalize text-[var(--muted)]">{sourceLabel(item)}</span>
}

function sourceLabel(item: DisplayItem) {
  if (item.sourceType === 'instagram') return item.igUsername || 'Instagram'
  if (item.sourceType === 'image') return 'Uploaded image'
  return 'Message only'
}

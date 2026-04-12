import { useEffect, useId, useRef, useState } from 'react'
import { useAuth } from '../contexts/useAuth'
import { useTheme } from '../contexts/useTheme'
import { Button } from './ui/button'

export function UserAccountMenu() {
  const { user, logout, refreshUser } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const [open, setOpen] = useState(false)
  const wrapRef = useRef<HTMLDivElement>(null)
  const panelTitleId = useId()
  const panelId = useId()

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (wrapRef.current?.contains(e.target as Node)) return
      setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open])

  if (!user) return null

  const displayName = user.name?.trim() || user.email

  return (
    <div className="relative" ref={wrapRef}>
      <button
        type="button"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-controls={open ? panelId : undefined}
        onClick={() => {
          setOpen((v) => !v)
          void refreshUser()
        }}
        className="flex max-w-[min(100vw-8rem,14rem)] items-center gap-1.5 rounded-full border border-blue-600 bg-blue-600 px-3 py-1.5 text-left text-sm font-medium text-white shadow-sm transition-colors hover:border-blue-500 hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      >
        <span className="min-w-0 flex-1 truncate">{displayName}</span>
        <span className="shrink-0 text-[10px] text-white/85" aria-hidden>
          ▾
        </span>
      </button>

      {open ? (
        <div
          id={panelId}
          role="dialog"
          aria-labelledby={panelTitleId}
          className="absolute right-0 z-[100] mt-2 w-[min(calc(100vw-2rem),18rem)] rounded-xl border border-border bg-panel p-4 shadow-[0_16px_48px_rgba(0,0,0,0.45)]"
        >
          <p
            id={panelTitleId}
            className="text-[11px] font-medium uppercase tracking-wide text-fg-soft"
          >
            Your account
          </p>
          <div className="mt-3 space-y-1">
            <p className="text-xs text-fg-muted">Name</p>
            <p className="text-sm font-medium text-fg">{user.name?.trim() || user.email}</p>
          </div>
          <div className="mt-3 space-y-1">
            <p className="text-xs text-fg-muted">Email</p>
            <p className="break-all text-sm text-fg">{user.email}</p>
          </div>
          <div className="mt-3 space-y-1">
            <p className="text-xs text-fg-muted">User ID</p>
            <p className="break-all font-mono text-[11px] leading-snug text-fg-muted" title={user.id}>
              {user.id}
            </p>
          </div>
          <div className="mt-4 space-y-3 border-t border-border pt-4">
            <div>
              <p className="text-xs text-fg-muted">Theme</p>
              <Button
                type="button"
                variant="secondary"
                className="mt-1.5 w-full"
                onClick={() => {
                  toggleTheme()
                }}
              >
                {theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              </Button>
            </div>
            <Button
              type="button"
              variant="secondary"
              className="w-full"
              onClick={() => {
                setOpen(false)
                logout()
              }}
            >
              Log out
            </Button>
          </div>
        </div>
      ) : null}
    </div>
  )
}

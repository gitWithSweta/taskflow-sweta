import { useState } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/useAuth'
import { useFormFields } from '../hooks/useFormFields'
import { requiredTrimmed } from '../lib/validators'
import { ThemeToggle } from '../components/ThemeToggle'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'

export function LoginPage() {
  const { login, token } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const { fields, setFields, error, clearErrors, applyApiError } = useFormFields()

  if (token) {
    return <Navigate to="/project" replace />
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    clearErrors()
    const f: Record<string, string> = {}
    const e1 = requiredTrimmed(email, 'is required')
    if (e1) f.email = e1
    const e2 = requiredTrimmed(password, 'is required')
    if (e2) f.password = e2
    if (Object.keys(f).length) {
      setFields(f)
      return
    }
    setLoading(true)
    try {
      await login(email.trim(), password)
    } catch (err) {
      applyApiError(err, 'Could not sign in', (msg) =>
        msg === 'unauthorized' ? 'Invalid email or password' : msg,
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative flex min-h-screen flex-col items-center justify-center px-4">
      <div className="absolute right-4 top-4 sm:right-6 sm:top-6">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-md rounded-2xl border border-border bg-surface-2/60 p-8 shadow-xl backdrop-blur">
        <h1 className="text-2xl font-semibold text-fg">Sign in</h1>
        <p className="mt-1 text-sm text-fg-muted">
          New here?{' '}
          <Link
            to="/register"
            className="font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)] hover:underline"
          >
            Create an account
          </Link>
        </p>
        <form onSubmit={onSubmit} className="mt-8 space-y-4">
          {error ? (
            <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-300">{error}</p>
          ) : null}
          <div>
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1"
            />
            {fields.email ? <p className="mt-1 text-sm text-red-400">{fields.email}</p> : null}
          </div>
          <div>
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="mt-1"
            />
            {fields.password ? (
              <p className="mt-1 text-sm text-red-400">{fields.password}</p>
            ) : null}
          </div>
          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>
        <p className="mt-4 text-center text-xs leading-relaxed text-fg-soft">
          Using CSV demo users? Password for all{' '}
          <span className="font-mono text-fg-muted">seed*@taskflow.demo</span> accounts is{' '}
          <span className="font-mono font-medium text-fg-dim">Testpass1</span> (capital T, ends with digit 1). They
          are created the first time the API starts against an empty database.
        </p>
      </div>
    </div>
  )
}

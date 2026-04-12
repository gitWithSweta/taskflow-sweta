import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Button } from './ui/button'

type Props = { children: ReactNode }

type State = { error: Error | null }

function ErrorFallback({ error, onRetry }: { error: Error; onRetry: () => void }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 py-12">
      <div className="w-full max-w-lg rounded-2xl border border-border bg-[var(--color-surface-2)]/80 p-8 shadow-xl">
        <h1 className="text-xl font-semibold text-fg">Something went wrong</h1>
        <p className="mt-2 text-sm text-fg-muted">
          The app hit an unexpected error. You can try again or go back to your projects.
        </p>
        <pre className="mt-4 max-h-40 overflow-auto rounded-lg bg-black/30 p-3 text-left text-xs text-red-200/90">
          {error.message}
        </pre>
        <div className="mt-6 flex flex-wrap gap-2">
          <Button type="button" onClick={onRetry}>
            Try again
          </Button>
          <Button type="button" variant="secondary" asChild>
            <Link to="/project">All projects</Link>
          </Button>
        </div>
      </div>
    </div>
  )
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <ErrorFallback error={this.state.error} onRetry={() => this.setState({ error: null })} />
      )
    }
    return this.props.children
  }
}

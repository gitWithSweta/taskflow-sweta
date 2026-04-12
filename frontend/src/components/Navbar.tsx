import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/useAuth'
import { ThemeToggle } from './ThemeToggle'
import { UserAccountMenu } from './UserAccountMenu'

export function Navbar() {
  const { user } = useAuth()
  return (
    <header className="border-b border-border bg-surface/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
        <Link to="/project" className="text-lg font-semibold tracking-tight text-fg">
          TaskFlow
        </Link>
        <div className="flex shrink-0 items-center gap-1">
          <ThemeToggle size="sm" />
          {user ? <UserAccountMenu /> : null}
        </div>
      </div>
    </header>
  )
}

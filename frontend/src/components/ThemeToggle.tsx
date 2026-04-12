import { clsx } from 'clsx'
import { useTheme } from '../contexts/useTheme'
import { Button } from './ui/button'

type Props = {
  size?: 'sm' | 'md'
  className?: string
}

function SunIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
    </svg>
  )
}

function MoonIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
  )
}

export function ThemeToggle({ size = 'md', className }: Props) {
  const { theme, toggleTheme } = useTheme()
  const isDark = theme === 'dark'
  const label = isDark ? 'Switch to light theme' : 'Switch to dark theme'
  const iconClass = size === 'sm' ? 'size-[1.125rem] shrink-0' : 'size-5 shrink-0'

  return (
    <Button
      type="button"
      variant="ghost"
      size={size === 'sm' ? 'sm' : 'md'}
      className={clsx(
        'border border-border/70 bg-surface-2/50 text-fg hover:border-border hover:bg-[var(--tf-ghost-hover)] hover:text-fg',
        className,
      )}
      onClick={toggleTheme}
      aria-label={label}
      title={label}
    >
      {isDark ? <SunIcon className={iconClass} /> : <MoonIcon className={iconClass} />}
      <span className="text-sm font-medium">{isDark ? 'Light' : 'Dark'}</span>
    </Button>
  )
}

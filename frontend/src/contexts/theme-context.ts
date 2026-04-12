import { createContext } from 'react'

export type Theme = 'light' | 'dark'

export const DEFAULT_THEME: Theme = 'dark'

export type ThemeContextValue = {
  theme: Theme
  setTheme: (t: Theme) => void
  toggleTheme: () => void
}

export const ThemeContext = createContext<ThemeContextValue | null>(null)

export const THEME_STORAGE_KEY = 'taskflow-theme'

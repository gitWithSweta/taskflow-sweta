import { useCallback, useMemo, useState, type ReactNode } from 'react'
import { api, clearSession, getStoredUser, getToken, setSession } from '../api/client'
import { AuthContext } from './auth-context'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState(() => getStoredUser())
  const [token, setToken] = useState(() => getToken())

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.login({ email, password })
    setSession(res.token, res.user)
    setToken(res.token)
    setUser(res.user)
  }, [])

  const register = useCallback(async (name: string, email: string, password: string) => {
    const res = await api.register({ name, email, password })
    setSession(res.token, res.user)
    setToken(res.token)
    setUser(res.user)
  }, [])

  const logout = useCallback(() => {
    clearSession()
    setToken(null)
    setUser(null)
  }, [])

  const refreshUser = useCallback(async () => {
    if (!getToken()) return
    try {
      const { user: u } = await api.me()
      setUser(u)
      localStorage.setItem('taskflow_user', JSON.stringify(u))
    } catch {
      logout()
    }
  }, [logout])

  const value = useMemo(
    () => ({
      user,
      token,
      login,
      register,
      logout,
      refreshUser,
    }),
    [user, token, login, register, logout, refreshUser],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

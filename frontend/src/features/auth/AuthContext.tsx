import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

import { ApiError, apiClient } from '../../api/client'
import { authService } from '../../services/auth'
import type { LoginInput, User } from '../../types/api'

export type AuthPhase = 'checking' | 'authenticated' | 'unauthenticated'

interface AuthState {
  phase: AuthPhase
  user: User | null
  verificationError: ApiError | null
  sessionExpired: boolean
  logoutCompleted: boolean
  refresh: () => Promise<void>
  login: (input: LoginInput) => Promise<User>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [phase, setPhase] = useState<AuthPhase>('checking')
  const [user, setUser] = useState<User | null>(null)
  const [verificationError, setVerificationError] = useState<ApiError | null>(null)
  const [sessionExpired, setSessionExpired] = useState(false)
  const [logoutCompleted, setLogoutCompleted] = useState(false)

  const loadCurrentUser = useCallback(async (signal?: AbortSignal) => {
    setPhase('checking')
    setVerificationError(null)
    try {
      const currentUser = await authService.currentUser(signal)
      if (signal?.aborted) return
      setUser(currentUser)
      setSessionExpired(false)
      setLogoutCompleted(false)
      setPhase('authenticated')
    } catch (cause) {
      if (cause instanceof DOMException && cause.name === 'AbortError') return
      if (cause instanceof ApiError && cause.status === 401) {
        setUser(null)
        setSessionExpired(false)
        setLogoutCompleted(false)
        setPhase('unauthenticated')
      } else {
        setVerificationError(
          cause instanceof ApiError
            ? cause
            : new ApiError({ status: 0, code: 'unknown_error', message: 'Não foi possível verificar sua sessão.' }),
        )
        setPhase('checking')
      }
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    void loadCurrentUser(controller.signal)
    return () => controller.abort()
  }, [loadCurrentUser])

  useEffect(() => {
    apiClient.setUnauthorizedHandler(() => {
      setUser(null)
      setVerificationError(null)
      setSessionExpired(true)
      setLogoutCompleted(false)
      setPhase('unauthenticated')
    })
    return () => apiClient.setUnauthorizedHandler()
  }, [])

  const refresh = useCallback(() => loadCurrentUser(), [loadCurrentUser])
  const login = useCallback(async (input: LoginInput) => {
    const authenticatedUser = await authService.login(input)
    setUser(authenticatedUser)
    setVerificationError(null)
    setSessionExpired(false)
    setLogoutCompleted(false)
    setPhase('authenticated')
    return authenticatedUser
  }, [])
  const logout = useCallback(async () => {
    await authService.logout()
    setUser(null)
    setVerificationError(null)
    setSessionExpired(false)
    setLogoutCompleted(true)
    setPhase('unauthenticated')
  }, [])

  const value = useMemo(
    () => ({ phase, user, verificationError, sessionExpired, logoutCompleted, refresh, login, logout }),
    [phase, user, verificationError, sessionExpired, logoutCompleted, refresh, login, logout],
  )
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth deve ser usado dentro de AuthProvider')
  return context
}

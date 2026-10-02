import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { ErrorState } from '../../components/feedback/ErrorState'
import { LoadingState } from '../../components/feedback/LoadingState'
import { useAuth } from '../../features/auth/AuthContext'

export function ProtectedRoute() {
  const auth = useAuth()
  const location = useLocation()

  if (auth.phase === 'checking' && !auth.verificationError) return <LoadingState label="Verificando sua sessão…" fullPage />
  if (auth.verificationError) {
    return <ErrorState title="Não foi possível verificar sua sessão" message={auth.verificationError.message} onRetry={auth.refresh} fullPage />
  }
  if (auth.phase === 'unauthenticated') {
    return <Navigate to="/login" replace state={{ from: `${location.pathname}${location.search}`, reason: auth.sessionExpired ? 'session_expired' : undefined }} />
  }
  return <Outlet />
}

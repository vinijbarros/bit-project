import { useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'

import { useAuth } from '../../features/auth/AuthContext'
import { AlertMessage } from '../feedback/AlertMessage'

const navigation = [
  { to: '/dashboard', label: 'Dashboard', end: true },
  { to: '/solicitacoes', label: 'Solicitações', end: true },
  { to: '/solicitacoes/nova', label: 'Nova solicitação', end: true },
]

export function AppLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [loggingOut, setLoggingOut] = useState(false)
  const [logoutError, setLogoutError] = useState<string | null>(null)

  async function handleLogout() {
    setLoggingOut(true)
    setLogoutError(null)
    try {
      await logout()
      navigate('/login', { replace: true })
    } catch {
      setLogoutError('Não foi possível encerrar a sessão. Tente novamente.')
    } finally {
      setLoggingOut(false)
    }
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="brand-block">
          <span className="brand-mark" aria-hidden="true">bit</span>
          <div>
            <strong>Portal Interno</strong>
            <span>Solicitações</span>
          </div>
        </div>
        <div className="user-actions">
          <span className="user-name">{user?.display_name}</span>
          <button className="button button--quiet" type="button" onClick={handleLogout} disabled={loggingOut}>
            {loggingOut ? 'Saindo…' : 'Sair'}
          </button>
        </div>
      </header>

      <div className="app-body">
        <nav className="main-navigation" aria-label="Navegação principal">
          {navigation.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.end} className={({ isActive }) => `nav-link${isActive ? ' nav-link--active' : ''}`}>
              {item.label}
            </NavLink>
          ))}
        </nav>
        <main className="app-content" id="conteudo-principal">
          {logoutError && <AlertMessage tone="error">{logoutError}</AlertMessage>}
          <Outlet />
        </main>
      </div>
    </div>
  )
}

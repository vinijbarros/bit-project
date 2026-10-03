import { useState } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'

import { useAuth } from '../../features/auth/AuthContext'
import { AlertMessage } from '../feedback/AlertMessage'

const navigation = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/solicitacoes', label: 'Solicitações' },
  { to: '/solicitacoes/nova', label: 'Nova solicitação' },
]

function isCurrentPage(pathname: string, destination: string): boolean {
  if (destination === '/solicitacoes') {
    return pathname === destination || (pathname.startsWith('/solicitacoes/') && pathname !== '/solicitacoes/nova')
  }
  return pathname === destination
}

export function AppLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
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
      <a className="skip-link" href="#conteudo-principal">Pular para o conteúdo principal</a>
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
          {navigation.map((item) => {
            const active = isCurrentPage(location.pathname, item.to)
            return (
              <Link key={item.to} to={item.to} aria-current={active ? 'page' : undefined} className={`nav-link${active ? ' nav-link--active' : ''}`}>
                {item.label}
              </Link>
            )
          })}
        </nav>
        <main className="app-content" id="conteudo-principal">
          {logoutError && <AlertMessage tone="error">{logoutError}</AlertMessage>}
          <Outlet />
        </main>
      </div>
    </div>
  )
}

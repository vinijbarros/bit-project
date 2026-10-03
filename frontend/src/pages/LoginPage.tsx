import { useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import { ApiError } from '../api/client'
import { AlertMessage } from '../components/feedback/AlertMessage'
import { FormField } from '../components/forms/FormField'
import { ThemeToggle } from '../components/theme/ThemeToggle'
import { useAuth } from '../features/auth/AuthContext'
import { safeInternalDestination } from '../app/routes/destination'

interface LoginLocationState {
  from?: unknown
  reason?: unknown
  message?: unknown
}

export function LoginPage() {
  const auth = useAuth()
  const location = useLocation()
  const navigate = useNavigate()
  const state = (location.state ?? {}) as LoginLocationState
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<{ username?: string; password?: string }>({})

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting) return

    const errors: { username?: string; password?: string } = {}
    if (!username.trim()) errors.username = 'Informe o usuário.'
    if (!password) errors.password = 'Informe a senha.'
    setFieldErrors(errors)
    setFormError(null)
    if (Object.keys(errors).length > 0) return

    setSubmitting(true)
    try {
      await auth.login({ username, password })
      setPassword('')
      navigate(safeInternalDestination(state.from), { replace: true })
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) {
        setFormError('Usuário ou senha inválidos.')
      } else if (cause instanceof ApiError) {
        setFormError(cause.message)
        if (cause.fields?.username?.[0] || cause.fields?.password?.[0]) {
          setFieldErrors({ username: cause.fields.username?.[0], password: cause.fields.password?.[0] })
        }
      } else {
        setFormError('Não foi possível entrar. Tente novamente.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="public-shell">
      <div className="public-theme-action">
        <ThemeToggle />
      </div>
      <section className="auth-card" aria-labelledby="login-title">
        <span className="brand-mark brand-mark--large" aria-hidden="true">b1t</span>
        <span className="eyebrow">Portal de Solicitações Internas</span>
        <h1 id="login-title">Acesso ao portal</h1>
        <p>Entre com as credenciais fornecidas para o ambiente.</p>

        {(state.reason === 'session_expired' || auth.sessionExpired) && (
          <AlertMessage tone="error">Sua sessão expirou. Entre novamente para continuar.</AlertMessage>
        )}
        {(auth.logoutCompleted || typeof state.message === 'string') && (
          <AlertMessage tone="success">{typeof state.message === 'string' ? state.message : 'Sessão encerrada com segurança.'}</AlertMessage>
        )}
        {formError && <AlertMessage tone="error">{formError}</AlertMessage>}

        <form onSubmit={handleSubmit} noValidate>
          <FormField id="username" label="Usuário" error={fieldErrors.username} required>
            <input
              name="username"
              type="text"
              autoComplete="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              disabled={submitting}
            />
          </FormField>
          <FormField id="password" label="Senha" error={fieldErrors.password} required>
            <input
              name="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              disabled={submitting}
            />
          </FormField>
          <button className="button button--full" type="submit" disabled={submitting}>
            {submitting ? 'Entrando…' : 'Entrar'}
          </button>
        </form>
        <p className="scope-note">Não há cadastro público nem recuperação de senha neste portal.</p>
      </section>
    </main>
  )
}

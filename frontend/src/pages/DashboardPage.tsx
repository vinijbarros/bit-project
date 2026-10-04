import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { ApiError } from '../api/client'
import { AlertMessage } from '../components/feedback/AlertMessage'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { PageHeader } from '../components/layout/PageHeader'
import { getDashboard } from '../services/dashboard'
import type { Dashboard, Status } from '../types/api'

interface DashboardIndicator {
  key: keyof Dashboard
  label: string
  status?: Status
}

const indicators: DashboardIndicator[] = [
  { key: 'total', label: 'Quantidade total de solicitações' },
  { key: 'abertas', label: 'Abertas', status: 'aberto' },
  { key: 'em_atendimento', label: 'Em Atendimento', status: 'em_atendimento' },
  { key: 'concluidas', label: 'Concluídas', status: 'concluido' },
]

function dashboardErrorMessage(cause: unknown): string {
  if (cause instanceof ApiError) return cause.message
  return 'Não foi possível carregar os indicadores. Verifique sua conexão e tente novamente.'
}

export function DashboardPage() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>()
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    setLoading(true)
    setError(undefined)
    void getDashboard(controller.signal)
      .then((value) => {
        if (active) setDashboard(value)
      })
      .catch((cause: unknown) => {
        if (!active || (cause instanceof DOMException && cause.name === 'AbortError')) return
        setDashboard(null)
        setError(dashboardErrorMessage(cause))
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [reloadKey])

  return (
    <section aria-labelledby="dashboard-title">
      <PageHeader
        eyebrow="Visão geral"
        title="Dashboard"
        titleId="dashboard-title"
        description="Indicadores globais de todas as solicitações. Os filtros da listagem não alteram estes números."
        actions={<Link className="button" to="/solicitacoes/nova">Nova solicitação</Link>}
      />

      {loading ? (
        <LoadingState label={dashboard ? 'Atualizando indicadores…' : 'Carregando indicadores…'} />
      ) : error || !dashboard ? (
        <ErrorState
          title="Não foi possível carregar o dashboard"
          message={error ?? 'A API retornou uma resposta inválida.'}
          onRetry={() => setReloadKey((value) => value + 1)}
        />
      ) : (
        <>
          <ul className="dashboard-grid" aria-label="Indicadores globais de solicitações">
            {indicators.map((indicator) => {
              const destination = indicator.status ? `/solicitacoes?status=${indicator.status}` : '/solicitacoes'
              return (
                <li key={indicator.key}>
                  <Link className={`dashboard-card dashboard-card--${indicator.status ?? 'total'}`} to={destination}>
                    <span className="dashboard-card__label">{indicator.label}</span>
                    <strong className="dashboard-card__value">{dashboard[indicator.key]}</strong>
                    <span className="dashboard-card__action">
                      {indicator.status ? `Ver solicitações: ${indicator.label}` : 'Ver todas as solicitações'}
                    </span>
                  </Link>
                </li>
              )
            })}
          </ul>
          {dashboard.total === 0 && (
            <AlertMessage>Não há solicitações cadastradas. Os quatro indicadores globais estão zerados.</AlertMessage>
          )}
        </>
      )}
    </section>
  )
}

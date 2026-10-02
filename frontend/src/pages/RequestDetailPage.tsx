import { useEffect, useState } from 'react'
import { Link, useLocation, useParams } from 'react-router-dom'

import { ApiError } from '../api/client'
import { CategoryBadge, StatusBadge } from '../components/badges/ValueBadge'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { PageHeader } from '../components/layout/PageHeader'
import { formatDateTimeSaoPaulo } from '../features/requests/dateTime'
import { requestsService } from '../services/requests'
import type { Request } from '../types/api'

function listReturnPath(value: unknown): string {
  if (typeof value === 'string' && (value === '/solicitacoes' || value.startsWith('/solicitacoes?'))) return value
  return '/solicitacoes'
}

export function RequestDetailPage() {
  const { id } = useParams()
  const location = useLocation()
  const numericID = Number(id)
  const validID = Number.isSafeInteger(numericID) && numericID > 0
  const [request, setRequest] = useState<Request | null>(null)
  const [loading, setLoading] = useState(validID)
  const [error, setError] = useState<string>()
  const [reloadKey, setReloadKey] = useState(0)
  const state = (location.state ?? {}) as { from?: unknown }
  const returnPath = listReturnPath(state.from)

  useEffect(() => {
    if (!validID) return
    const controller = new AbortController()
    let active = true
    setLoading(true)
    setError(undefined)
    void requestsService.get(numericID, controller.signal)
      .then((value) => {
        if (active) setRequest(value)
      })
      .catch((cause: unknown) => {
        if (!active || (cause instanceof DOMException && cause.name === 'AbortError')) return
        setRequest(null)
        if (cause instanceof ApiError && cause.status === 404) setError('A solicitação informada não existe ou foi excluída.')
        else if (cause instanceof ApiError) setError(cause.message)
        else setError('Não foi possível carregar os detalhes. Tente novamente.')
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [numericID, reloadKey, validID])

  if (!validID) {
    return <ErrorState title="Solicitação inválida" message="O código informado na URL não é válido." />
  }
  if (loading) return <LoadingState label="Carregando detalhes da solicitação…" />
  if (error || !request) {
    return <ErrorState title="Não foi possível consultar a solicitação" message={error ?? 'Resposta inválida da API.'} onRetry={() => setReloadKey((value) => value + 1)} />
  }

  return (
    <section aria-labelledby="request-detail-title">
      <PageHeader
        eyebrow={request.code}
        title={request.title}
        titleId="request-detail-title"
        description="Detalhes completos da solicitação."
        actions={<Link className="button button--quiet" to={returnPath}>Voltar para a listagem</Link>}
      />
      <div className="request-detail-card">
        <dl className="request-detail-grid">
          <div><dt>Categoria</dt><dd><CategoryBadge category={request.category} /></dd></div>
          <div><dt>Status</dt><dd><StatusBadge status={request.status} /></dd></div>
          <div><dt>Solicitante</dt><dd>{request.requester.display_name} <span className="requester-username">@{request.requester.username}</span></dd></div>
          <div><dt>Data de abertura</dt><dd>{formatDateTimeSaoPaulo(request.created_at)}</dd></div>
          <div><dt>Última atualização</dt><dd>{formatDateTimeSaoPaulo(request.updated_at)}</dd></div>
        </dl>
        <section className="request-description" aria-labelledby="request-description-title">
          <h2 id="request-description-title">Descrição</h2>
          <p>{request.description}</p>
        </section>
      </div>
    </section>
  )
}

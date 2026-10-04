import { useEffect, useState } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'

import { ApiError } from '../api/client'
import { CategoryBadge, StatusBadge } from '../components/badges/ValueBadge'
import { ConfirmDialog } from '../components/confirmation/ConfirmDialog'
import { AlertMessage } from '../components/feedback/AlertMessage'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { PageHeader } from '../components/layout/PageHeader'
import { formatDateTimeSaoPaulo } from '../features/requests/dateTime'
import { useRequestMetadata } from '../features/requests/useRequestMetadata'
import { requestsService } from '../services/requests'
import type { Request, Status } from '../types/api'

function listReturnPath(value: unknown): string {
  if (typeof value === 'string' && (value === '/solicitacoes' || value.startsWith('/solicitacoes?'))) return value
  return '/solicitacoes'
}

interface LoadError {
  kind: 'not_found' | 'infrastructure'
  message: string
}

function detailLoadError(cause: unknown): LoadError {
  if (cause instanceof ApiError && cause.status === 404) {
    return { kind: 'not_found', message: 'A solicitação informada não existe ou foi excluída.' }
  }
  if (cause instanceof ApiError) return { kind: 'infrastructure', message: cause.message }
  return { kind: 'infrastructure', message: 'Não foi possível carregar os detalhes. Verifique sua conexão e tente novamente.' }
}

export function RequestDetailPage() {
  const { id } = useParams()
  const location = useLocation()
  const navigate = useNavigate()
  const numericID = Number(id)
  const validID = Number.isSafeInteger(numericID) && numericID > 0
  const state = (location.state ?? {}) as { from?: unknown; message?: unknown }
  const returnPath = listReturnPath(state.from)
  const initialMessage = typeof state.message === 'string' ? state.message : undefined
  const { metadata, loading: loadingMetadata, error: metadataError, retry: retryMetadata } = useRequestMetadata()
  const [request, setRequest] = useState<Request | null>(null)
  const [selectedStatus, setSelectedStatus] = useState<Status>('aberto')
  const [loading, setLoading] = useState(validID)
  const [refreshing, setRefreshing] = useState(false)
  const [loadError, setLoadError] = useState<LoadError>()
  const [reloadKey, setReloadKey] = useState(0)
  const [feedback, setFeedback] = useState(initialMessage)
  const [mutationError, setMutationError] = useState<string>()
  const [statusSubmitting, setStatusSubmitting] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteSubmitting, setDeleteSubmitting] = useState(false)
  const [deleteError, setDeleteError] = useState<string>()

  useEffect(() => {
    if (!validID) return
    const controller = new AbortController()
    let active = true
    setLoading(true)
    setLoadError(undefined)
    void requestsService.get(numericID, controller.signal)
      .then((value) => {
        if (!active) return
        setRequest(value)
        setSelectedStatus(value.status.value)
      })
      .catch((cause: unknown) => {
        if (!active || (cause instanceof DOMException && cause.name === 'AbortError')) return
        setRequest(null)
        setLoadError(detailLoadError(cause))
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [numericID, reloadKey, validID])

  async function refreshAfterConflict(message: string) {
    setRefreshing(true)
    try {
      const current = await requestsService.get(numericID)
      setRequest(current)
      setSelectedStatus(current.status.value)
      setMutationError(message)
    } catch (cause: unknown) {
      setRequest(null)
      setLoadError(detailLoadError(cause))
    } finally {
      setRefreshing(false)
    }
  }

  async function changeStatus() {
    if (!request || selectedStatus === request.status.value || statusSubmitting) return
    setStatusSubmitting(true)
    setFeedback(undefined)
    setMutationError(undefined)
    try {
      const updated = await requestsService.updateStatus(request.id, selectedStatus)
      setRequest(updated)
      setSelectedStatus(updated.status.value)
      setFeedback(`Status alterado para ${updated.status.label}.`)
    } catch (cause: unknown) {
      if (cause instanceof ApiError && cause.status === 404) {
        setRequest(null)
        setLoadError({ kind: 'not_found', message: 'A solicitação foi removida por outra sessão.' })
      } else if (cause instanceof ApiError && (cause.status === 403 || cause.status === 409)) {
        await refreshAfterConflict('A solicitação mudou enquanto você alterava o status. Os dados atuais foram recarregados.')
      } else {
        setMutationError(cause instanceof ApiError ? cause.message : 'Não foi possível alterar o status. Tente novamente.')
      }
    } finally {
      setStatusSubmitting(false)
    }
  }

  function openDeleteDialog() {
    setDeleteError(undefined)
    setMutationError(undefined)
    setDeleteOpen(true)
  }

  async function confirmDelete() {
    if (!request || deleteSubmitting) return
    setDeleteSubmitting(true)
    setDeleteError(undefined)
    try {
      await requestsService.remove(request.id)
      navigate(returnPath, {
        replace: true,
        state: { message: `${request.code} foi excluída com sucesso.` },
      })
    } catch (cause: unknown) {
      if (cause instanceof ApiError && cause.status === 404) {
        setDeleteOpen(false)
        setRequest(null)
        setLoadError({ kind: 'not_found', message: 'A solicitação já foi removida por outra sessão.' })
      } else if (cause instanceof ApiError && (cause.status === 403 || cause.status === 409)) {
        setDeleteOpen(false)
        await refreshAfterConflict(cause.status === 409
          ? 'A solicitação não está mais aberta e não pode ser excluída. Os dados atuais foram recarregados.'
          : 'Sua permissão para excluir esta solicitação mudou. Os dados atuais foram recarregados.')
      } else {
        setDeleteError(cause instanceof ApiError ? cause.message : 'Não foi possível excluir. Verifique sua conexão e tente novamente.')
      }
    } finally {
      setDeleteSubmitting(false)
    }
  }

  if (!validID) {
    return <ErrorState title="Solicitação inválida" message="O código informado na URL não é válido." />
  }
  if (loading || loadingMetadata) return <LoadingState label="Carregando detalhes da solicitação…" />
  if (metadataError || !metadata) {
    return <ErrorState title="Não foi possível carregar as opções de status" message={metadataError ?? 'Resposta inválida da API.'} onRetry={retryMetadata} />
  }
  if (loadError || !request) {
    const notFound = loadError?.kind === 'not_found'
    return (
      <ErrorState
        title={notFound ? 'Solicitação não encontrada' : 'Não foi possível consultar a solicitação'}
        message={loadError?.message ?? 'Resposta inválida da API.'}
        onRetry={notFound ? () => navigate(returnPath) : () => setReloadKey((value) => value + 1)}
        retryLabel={notFound ? 'Voltar para a listagem' : 'Tentar novamente'}
      />
    )
  }

  return (
    <section aria-labelledby="request-detail-title">
      <PageHeader
        eyebrow={request.code}
        title={request.title}
        titleId="request-detail-title"
        description="Detalhes completos da solicitação."
        actions={
          <>
            {request.permissions.can_edit && <Link className="button" to={`/solicitacoes/${request.id}/editar`} state={{ from: returnPath }}>Editar solicitação</Link>}
            {request.permissions.can_delete && <button className="button button--danger" type="button" onClick={openDeleteDialog}>Excluir solicitação</button>}
            <Link className="button button--quiet" to={returnPath}>Voltar para a listagem</Link>
          </>
        }
      />
      {feedback && <AlertMessage tone="success">{feedback}</AlertMessage>}
      {mutationError && <AlertMessage tone="error">{mutationError}</AlertMessage>}
      <div className="request-detail-card">
        <dl className="request-detail-grid">
          <div><dt>Categoria</dt><dd><CategoryBadge category={request.category} /></dd></div>
          <div><dt>Status atual</dt><dd><StatusBadge status={request.status} /></dd></div>
          <div><dt>Solicitante</dt><dd>{request.requester.display_name} <span className="requester-username">@{request.requester.username}</span></dd></div>
          <div><dt>Data de criação</dt><dd>{formatDateTimeSaoPaulo(request.created_at)}</dd></div>
          <div><dt>Última atualização</dt><dd>{formatDateTimeSaoPaulo(request.updated_at)}</dd></div>
        </dl>
        <section className="request-description" aria-labelledby="request-description-title">
          <h2 id="request-description-title">Descrição</h2>
          <p>{request.description}</p>
        </section>
      </div>

      <section className="status-management" aria-labelledby="status-management-title">
        <div>
          <h2 id="status-management-title">Gerenciar status</h2>
          <p>Qualquer usuário autenticado pode alterar o status, inclusive reabrir uma solicitação.</p>
        </div>
        <div className="status-management__controls">
          <label htmlFor="request-status">Novo status</label>
          <select id="request-status" value={selectedStatus} onChange={(event) => setSelectedStatus(event.target.value as Status)} disabled={statusSubmitting || refreshing}>
            {metadata.statuses.map((status) => <option key={status.value} value={status.value}>{status.label}</option>)}
          </select>
          <button className="button" type="button" onClick={() => void changeStatus()} disabled={selectedStatus === request.status.value || statusSubmitting || refreshing}>
            {statusSubmitting ? 'Alterando status…' : refreshing ? 'Atualizando dados…' : 'Alterar status'}
          </button>
        </div>
      </section>

      <ConfirmDialog
        open={deleteOpen}
        title="Excluir solicitação?"
        description={`A solicitação ${request.code} — ${request.title} será removida definitivamente. Esta ação não pode ser desfeita.`}
        confirmLabel="Excluir definitivamente"
        busy={deleteSubmitting}
        destructive
        error={deleteError}
        onCancel={() => setDeleteOpen(false)}
        onConfirm={() => void confirmDelete()}
      />
    </section>
  )
}

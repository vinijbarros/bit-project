import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { ApiError } from '../api/client'
import { CategoryBadge, StatusBadge } from '../components/badges/ValueBadge'
import { EmptyState } from '../components/feedback/EmptyState'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { FormField } from '../components/forms/FormField'
import { PageHeader } from '../components/layout/PageHeader'
import {
  draftFromSearchParams,
  emptyRequestFilterDraft,
  filtersFromSearchParams,
  hasRequestFilters,
  searchParamsFromDraft,
  validateRequestPeriod,
  type RequestFilterDraft,
} from '../features/requests/listQuery'
import { formatDateTimeSaoPaulo } from '../features/requests/dateTime'
import { getMetadata } from '../services/metadata'
import { requestsService } from '../services/requests'
import type { Metadata, PaginatedRequests, RequestListItem } from '../types/api'

function requestErrorMessage(cause: unknown): string {
  if (cause instanceof ApiError) return cause.message
  return 'Não foi possível carregar as solicitações. Tente novamente.'
}

function RequestsTable({ items, returnPath }: { items: RequestListItem[]; returnPath: string }) {
  return (
    <div className="table-region" role="region" aria-label="Resultados das solicitações" tabIndex={0}>
      <table className="requests-table">
        <thead>
          <tr>
            <th scope="col">Código</th>
            <th scope="col">Título</th>
            <th scope="col">Categoria</th>
            <th scope="col">Solicitante</th>
            <th scope="col">Data de abertura</th>
            <th scope="col">Status</th>
            <th scope="col"><span className="visually-hidden">Ações</span></th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td data-label="Código"><strong>{item.code}</strong></td>
              <td data-label="Título">{item.title}</td>
              <td data-label="Categoria"><CategoryBadge category={item.category} /></td>
              <td data-label="Solicitante">
                <span className="requester-name">{item.requester.display_name}</span>
                <span className="requester-username">@{item.requester.username}</span>
              </td>
              <td data-label="Data de abertura">{formatDateTimeSaoPaulo(item.created_at)}</td>
              <td data-label="Status"><StatusBadge status={item.status} /></td>
              <td className="table-action"><Link to={`/solicitacoes/${item.id}`} state={{ from: returnPath }} aria-label={`Ver detalhes de ${item.code}`}>Ver detalhes</Link></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function RequestsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const queryKey = searchParams.toString()
  const filters = useMemo(() => filtersFromSearchParams(searchParams), [searchParams])
  const [draft, setDraft] = useState<RequestFilterDraft>(() => draftFromSearchParams(searchParams))
  const [periodError, setPeriodError] = useState<string>()
  const [metadata, setMetadata] = useState<Metadata | null>(null)
  const [result, setResult] = useState<PaginatedRequests | null>(null)
  const [loadingMetadata, setLoadingMetadata] = useState(true)
  const [loadingRequests, setLoadingRequests] = useState(true)
  const [loadError, setLoadError] = useState<string>()
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    setDraft(draftFromSearchParams(searchParams))
    setPeriodError(undefined)
  }, [searchParams])

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    setLoadingMetadata(true)
    setLoadError(undefined)
    void getMetadata(controller.signal)
      .then((value) => {
        if (active) setMetadata(value)
      })
      .catch((cause: unknown) => {
        if (active && !(cause instanceof DOMException && cause.name === 'AbortError')) {
          setMetadata(null)
          setLoadError(requestErrorMessage(cause))
        }
      })
      .finally(() => {
        if (active) setLoadingMetadata(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [reloadKey])

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    setLoadingRequests(true)
    setLoadError(undefined)
    void requestsService.list(filters, controller.signal)
      .then((value) => {
        if (active) setResult(value)
      })
      .catch((cause: unknown) => {
        if (active && !(cause instanceof DOMException && cause.name === 'AbortError')) {
          setResult(null)
          setLoadError(requestErrorMessage(cause))
        }
      })
      .finally(() => {
        if (active) setLoadingRequests(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [filters, queryKey, reloadKey])

  function updateDraft(field: keyof RequestFilterDraft, value: string) {
    setDraft((current) => ({ ...current, [field]: value }))
  }

  function applyFilters(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const validationError = validateRequestPeriod(draft)
    setPeriodError(validationError)
    if (validationError) return
    setSearchParams(searchParamsFromDraft(draft), { replace: false })
  }

  function clearFilters() {
    setPeriodError(undefined)
    setDraft(emptyRequestFilterDraft)
    setSearchParams(new URLSearchParams(), { replace: false })
  }

  function changePage(page: number) {
    setSearchParams(searchParamsFromDraft(draftFromSearchParams(searchParams), page), { replace: false })
  }

  const isLoading = loadingMetadata || loadingRequests
  const activeFilters = hasRequestFilters(searchParams)

  return (
    <section aria-labelledby="requests-title">
      <PageHeader
        eyebrow="Atendimento interno"
        title="Solicitações"
        titleId="requests-title"
        description="Consulte solicitações de todos os usuários e filtre pelo período de abertura."
        actions={<Link className="button" to="/solicitacoes/nova">Nova solicitação</Link>}
      />

      <form className="filters-panel" onSubmit={applyFilters} noValidate aria-label="Filtros de solicitações">
        <div className="filters-grid">
          <FormField id="filter-date-from" label="Data inicial" hint="Abertura a partir deste dia">
            <input type="date" value={draft.dateFrom} onChange={(event) => updateDraft('dateFrom', event.target.value)} />
          </FormField>
          <FormField id="filter-date-to" label="Data final" hint="Abertura até o fim deste dia" error={periodError}>
            <input type="date" value={draft.dateTo} onChange={(event) => updateDraft('dateTo', event.target.value)} />
          </FormField>
          <FormField id="filter-category" label="Categoria">
            <select value={draft.category} onChange={(event) => updateDraft('category', event.target.value)} disabled={!metadata}>
              <option value="">Todas as categorias</option>
              {metadata?.categories.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
          </FormField>
          <FormField id="filter-status" label="Status">
            <select value={draft.status} onChange={(event) => updateDraft('status', event.target.value)} disabled={!metadata}>
              <option value="">Todos os status</option>
              {metadata?.statuses.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
          </FormField>
          <FormField id="filter-query" label="Texto no título">
            <input type="search" maxLength={150} value={draft.query} onChange={(event) => updateDraft('query', event.target.value)} placeholder="Ex.: acesso ao sistema" />
          </FormField>
        </div>
        <div className="filter-actions">
          <button className="button" type="submit">Aplicar filtros</button>
          <button className="button button--quiet" type="button" onClick={clearFilters}>Limpar filtros</button>
        </div>
      </form>

      {loadError && !isLoading ? (
        <ErrorState title="Não foi possível carregar a listagem" message={loadError} onRetry={() => setReloadKey((value) => value + 1)} />
      ) : isLoading ? (
        <LoadingState label={result ? 'Atualizando solicitações…' : 'Carregando solicitações…'} />
      ) : result && result.items.length > 0 ? (
        <>
          <p className="results-summary" aria-live="polite">{result.pagination.total_items} solicitação(ões) encontrada(s).</p>
          <RequestsTable items={result.items} returnPath={`/solicitacoes${queryKey ? `?${queryKey}` : ''}`} />
          <nav className="pagination" aria-label="Paginação das solicitações">
            <button className="button button--quiet" type="button" onClick={() => changePage(result.pagination.page - 1)} disabled={result.pagination.page <= 1}>Página anterior</button>
            <span>Página <strong>{result.pagination.page}</strong> de <strong>{result.pagination.total_pages}</strong></span>
            <button className="button button--quiet" type="button" onClick={() => changePage(result.pagination.page + 1)} disabled={result.pagination.page >= result.pagination.total_pages}>Próxima página</button>
          </nav>
        </>
      ) : result && result.pagination.total_items > 0 ? (
        <EmptyState title="Esta página não possui resultados" action={<button className="button" type="button" onClick={() => changePage(1)}>Voltar à primeira página</button>}>
          <p>Existem solicitações para esta consulta, mas não na página informada.</p>
        </EmptyState>
      ) : result && activeFilters ? (
        <EmptyState title="Nenhuma solicitação encontrada" action={<button className="button button--quiet" type="button" onClick={clearFilters}>Limpar filtros</button>}>
          <p>Revise os filtros aplicados ao período de abertura, categoria, status ou título.</p>
        </EmptyState>
      ) : (
        <EmptyState title="Ainda não há solicitações" action={<Link className="button" to="/solicitacoes/nova">Criar primeira solicitação</Link>}>
          <p>Quando uma solicitação for criada, ela aparecerá aqui.</p>
        </EmptyState>
      )}
    </section>
  )
}

import type { Category, RequestFilters, Status } from '../../types/api'

export const REQUESTS_PAGE_SIZE = 20

export interface RequestFilterDraft {
  dateFrom: string
  dateTo: string
  category: string
  status: string
  query: string
}

export const emptyRequestFilterDraft: RequestFilterDraft = {
  dateFrom: '',
  dateTo: '',
  category: '',
  status: '',
  query: '',
}

function positiveInteger(value: string | null, fallback: number): number {
  if (!value) return fallback
  const parsed = Number(value)
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : fallback
}

export function draftFromSearchParams(params: URLSearchParams): RequestFilterDraft {
  return {
    dateFrom: params.get('date_from') ?? '',
    dateTo: params.get('date_to') ?? '',
    category: params.get('category') ?? '',
    status: params.get('status') ?? '',
    query: params.get('q') ?? '',
  }
}

export function filtersFromSearchParams(params: URLSearchParams): RequestFilters {
  const draft = draftFromSearchParams(params)
  return {
    date_from: draft.dateFrom || undefined,
    date_to: draft.dateTo || undefined,
    category: (draft.category || undefined) as Category | undefined,
    status: (draft.status || undefined) as Status | undefined,
    q: draft.query.trim() || undefined,
    page: positiveInteger(params.get('page'), 1),
    page_size: REQUESTS_PAGE_SIZE,
  }
}

export function searchParamsFromDraft(draft: RequestFilterDraft, page = 1): URLSearchParams {
  const params = new URLSearchParams()
  if (draft.dateFrom) params.set('date_from', draft.dateFrom)
  if (draft.dateTo) params.set('date_to', draft.dateTo)
  if (draft.category) params.set('category', draft.category)
  if (draft.status) params.set('status', draft.status)
  const query = draft.query.trim()
  if (query) params.set('q', query)
  if (page > 1) params.set('page', String(page))
  return params
}

export function hasRequestFilters(params: URLSearchParams): boolean {
  return ['date_from', 'date_to', 'category', 'status', 'q'].some((name) => Boolean(params.get(name)?.trim()))
}

export function validateRequestPeriod(draft: RequestFilterDraft): string | undefined {
  if (draft.dateFrom && draft.dateTo && draft.dateFrom > draft.dateTo) {
    return 'A data final deve ser igual ou posterior à data inicial.'
  }
  return undefined
}

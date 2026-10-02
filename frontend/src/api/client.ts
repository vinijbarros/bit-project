import type { ApiErrorPayload, FieldErrors } from '../types/api'

type UnauthorizedHandler = () => void

export interface ApiRequestOptions extends Omit<RequestInit, 'body'> {
  json?: unknown
  handleUnauthorized?: boolean
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields?: FieldErrors
  readonly requestId?: string
  readonly retryAfter?: string

  constructor(options: {
    status: number
    code: string
    message: string
    fields?: FieldErrors
    requestId?: string
    retryAfter?: string
    cause?: unknown
  }) {
    super(options.message, { cause: options.cause })
    this.name = 'ApiError'
    this.status = options.status
    this.code = options.code
    this.fields = options.fields
    this.requestId = options.requestId
    this.retryAfter = options.retryAfter
  }
}

export interface ApiClient {
  request<T>(path: string, options?: ApiRequestOptions): Promise<T>
  setUnauthorizedHandler(handler?: UnauthorizedHandler): void
}

function normalizeBasePath(value: string | undefined): string {
  const basePath = value?.trim() || '/api/v1'
  if (!basePath.startsWith('/') || basePath.startsWith('//') || basePath.includes('://')) {
    throw new Error('VITE_API_BASE_PATH deve ser um caminho relativo iniciado por /')
  }
  return basePath.replace(/\/+$/, '')
}

function isApiErrorPayload(value: unknown): value is ApiErrorPayload {
  if (!value || typeof value !== 'object' || !('error' in value)) return false
  const error = (value as { error?: unknown }).error
  if (!error || typeof error !== 'object') return false
  const detail = error as { code?: unknown; message?: unknown; fields?: unknown }
  return typeof detail.code === 'string' && typeof detail.message === 'string'
}

function fallbackMessage(status: number): string {
  const messages: Record<number, string> = {
    400: 'Não foi possível validar os dados enviados.',
    401: 'Sua sessão não é válida. Entre novamente.',
    403: 'Você não tem permissão para realizar esta ação.',
    404: 'O recurso solicitado não foi encontrado.',
    409: 'A operação conflita com o estado atual do recurso.',
    415: 'O formato enviado não é compatível com a API.',
    429: 'Muitas tentativas. Aguarde antes de tentar novamente.',
    500: 'Ocorreu um erro interno. Tente novamente mais tarde.',
    503: 'O serviço está temporariamente indisponível.',
  }
  return messages[status] ?? 'Não foi possível concluir a solicitação.'
}

export function createApiClient(
  basePath = normalizeBasePath(import.meta.env.VITE_API_BASE_PATH),
  fetchImplementation: typeof fetch = fetch,
): ApiClient {
  let unauthorizedHandler: UnauthorizedHandler | undefined

  return {
    setUnauthorizedHandler(handler) {
      unauthorizedHandler = handler
    },

    async request<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
      const normalizedPath = path.startsWith('/') ? path : `/${path}`
      const { json, handleUnauthorized = true, ...requestOptions } = options
      const headers = new Headers(requestOptions.headers)
      headers.set('Accept', 'application/json')

      let body: BodyInit | undefined
      if (json !== undefined) {
        headers.set('Content-Type', 'application/json')
        body = JSON.stringify(json)
      }

      let response: Response
      try {
        response = await fetchImplementation(`${basePath}${normalizedPath}`, {
          ...requestOptions,
          body,
          credentials: 'same-origin',
          headers,
        })
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') throw error
        throw new ApiError({
          status: 0,
          code: 'network_error',
          message: 'Não foi possível se conectar ao servidor.',
          cause: error,
        })
      }

      if (response.status === 204) return undefined as T

      const text = await response.text()
      let payload: unknown
      if (text) {
        try {
          payload = JSON.parse(text) as unknown
        } catch (error) {
          if (response.ok) {
            throw new ApiError({
              status: response.status,
              code: 'invalid_response',
              message: 'O servidor retornou uma resposta inválida.',
              cause: error,
            })
          }
        }
      }

      if (!response.ok) {
        const errorPayload = isApiErrorPayload(payload) ? payload : undefined
        const apiError = new ApiError({
          status: response.status,
          code: errorPayload?.error.code ?? 'http_error',
          message: errorPayload?.error.message ?? fallbackMessage(response.status),
          fields: errorPayload?.error.fields,
          requestId: errorPayload?.request_id,
          retryAfter: response.headers.get('Retry-After') ?? undefined,
        })
        if (response.status === 401 && handleUnauthorized) unauthorizedHandler?.()
        throw apiError
      }

      if (payload === undefined) {
        throw new ApiError({
          status: response.status,
          code: 'invalid_response',
          message: 'O servidor retornou uma resposta vazia inesperada.',
        })
      }
      return payload as T
    },
  }
}

export const apiClient = createApiClient()

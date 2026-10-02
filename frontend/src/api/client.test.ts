import { describe, expect, it, vi } from 'vitest'

import { ApiError, createApiClient } from './client'

describe('apiClient', () => {
  it('envia JSON e cookie pela URL relativa', async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ data: { id: 1 } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    const client = createApiClient('/api/v1', fetchMock)

    await expect(client.request('/items', { method: 'POST', json: { title: 'Teste' } })).resolves.toEqual({
      data: { id: 1 },
    })
    expect(fetchMock).toHaveBeenCalledOnce()
    const [url, options] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/items')
    expect(options?.credentials).toBe('same-origin')
    expect(options?.body).toBe('{"title":"Teste"}')
    expect(new Headers(options?.headers).get('Content-Type')).toBe('application/json')
  })

  it('aceita 204 sem tentar interpretar JSON', async () => {
    const client = createApiClient('/api/v1', vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 })))
    await expect(client.request<void>('/auth/logout', { method: 'POST' })).resolves.toBeUndefined()
  })

  it('preserva erro de campo e notifica sessão inválida', async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          error: { code: 'validation_failed', message: 'Corrija os campos.', fields: { title: ['Inválido.'] } },
          request_id: 'req-1',
        }),
        { status: 401 },
      ),
    )
    const unauthorized = vi.fn()
    const client = createApiClient('/api/v1', fetchMock)
    client.setUnauthorizedHandler(unauthorized)

    const error = await client.request('/requests').catch((cause: unknown) => cause)
    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 401, code: 'validation_failed', requestId: 'req-1' })
    expect((error as ApiError).fields).toEqual({ title: ['Inválido.'] })
    expect(unauthorized).toHaveBeenCalledOnce()
  })

  it('não trata 401 de login como expiração de sessão', async () => {
    const client = createApiClient(
      '/api/v1',
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(JSON.stringify({ error: { code: 'invalid_credentials', message: 'Credenciais inválidas.' } }), { status: 401 }),
      ),
    )
    const unauthorized = vi.fn()
    client.setUnauthorizedHandler(unauthorized)
    await expect(client.request('/auth/login', { method: 'POST', handleUnauthorized: false })).rejects.toMatchObject({
      status: 401,
      code: 'invalid_credentials',
    })
    expect(unauthorized).not.toHaveBeenCalled()
  })

  it.each([403, 404, 409])('mantém status e código do backend para HTTP %s', async (status) => {
    const client = createApiClient(
      '/api/v1',
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(JSON.stringify({ error: { code: `error_${status}`, message: 'Mensagem segura.' } }), { status }),
      ),
    )
    await expect(client.request('/requests/99')).rejects.toMatchObject({ status, code: `error_${status}` })
  })

  it('converte erro não JSON e falha de rede sem repetir a chamada', async () => {
    const nonJsonFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response('gateway offline', { status: 503 }))
    await expect(createApiClient('/api/v1', nonJsonFetch).request('/dashboard')).rejects.toMatchObject({
      status: 503,
      code: 'http_error',
    })
    expect(nonJsonFetch).toHaveBeenCalledOnce()

    const networkFetch = vi.fn<typeof fetch>().mockRejectedValue(new TypeError('failed'))
    await expect(createApiClient('/api/v1', networkFetch).request('/dashboard')).rejects.toMatchObject({
      status: 0,
      code: 'network_error',
    })
    expect(networkFetch).toHaveBeenCalledOnce()
  })

  it('propaga cancelamento para impedir atualização com resposta antiga', async () => {
    const abort = new DOMException('cancelado', 'AbortError')
    const client = createApiClient('/api/v1', vi.fn<typeof fetch>().mockRejectedValue(abort))
    await expect(client.request('/requests', { signal: AbortSignal.abort() })).rejects.toBe(abort)
  })
})

// @vitest-environment jsdom

import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'

import { RequestDetailPage } from '../../pages/RequestDetailPage'
import type { Request } from '../../types/api'

const serviceMocks = vi.hoisted(() => ({ get: vi.fn() }))

vi.mock('../../services/requests', () => ({ requestsService: { get: serviceMocks.get } }))

const request: Request = {
  id: 9,
  code: 'SOL-000009',
  title: 'Acesso ao sistema',
  description: 'Liberação necessária para executar as atividades internas.',
  category: { value: 'ti', label: 'TI' },
  status: { value: 'em_atendimento', label: 'Em Atendimento' },
  requester: { id: 2, username: 'colaborador2', display_name: 'Colaborador 2' },
  created_at: '2026-10-02T02:30:00Z',
  updated_at: '2026-10-02T15:00:00Z',
  permissions: { can_edit: false, can_delete: false },
}

function renderDetail(path: string, state?: unknown) {
  const router = createMemoryRouter([
    { path: '/solicitacoes/:id', element: <RequestDetailPage /> },
    { path: '/solicitacoes', element: <h1>Listagem</h1> },
  ], { initialEntries: [{ pathname: path, state }] })
  render(<RouterProvider router={router} />)
}

beforeEach(() => serviceMocks.get.mockReset().mockResolvedValue(request))
afterEach(cleanup)

describe('detalhes da solicitação', () => {
  it('consulta a API, mostra dados completos e preserva o retorno filtrado', async () => {
    renderDetail('/solicitacoes/9', { from: '/solicitacoes?category=ti&page=2' })

    expect(screen.getByText('Carregando detalhes da solicitação…')).toBeTruthy()
    expect(await screen.findByRole('heading', { name: 'Acesso ao sistema' })).toBeTruthy()
    expect(serviceMocks.get).toHaveBeenCalledWith(9, expect.any(AbortSignal))
    expect(screen.getByText('SOL-000009')).toBeTruthy()
    expect(screen.getByText('Liberação necessária para executar as atividades internas.')).toBeTruthy()
    expect(screen.getByText('Colaborador 2')).toBeTruthy()
    expect(screen.getByText('@colaborador2')).toBeTruthy()
    expect(screen.getByText('Em Atendimento')).toBeTruthy()
    expect(screen.getByText(/01\/10\/2026.*23:30/)).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Voltar para a listagem' }).getAttribute('href')).toBe('/solicitacoes?category=ti&page=2')
    expect(screen.queryByRole('button', { name: /editar|excluir|status/i })).toBeNull()
  })

  it('rejeita ID inválido sem chamar a API', () => {
    renderDetail('/solicitacoes/invalido')

    expect(screen.getByText('Solicitação inválida')).toBeTruthy()
    expect(serviceMocks.get).not.toHaveBeenCalled()
  })
})

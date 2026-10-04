// @vitest-environment jsdom

import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'

import { ApiError } from '../../api/client'
import { RequestsPage } from '../../pages/RequestsPage'
import type { Metadata, PaginatedRequests, RequestFilters, RequestListItem } from '../../types/api'

const serviceMocks = vi.hoisted(() => ({
  metadata: vi.fn(),
  list: vi.fn(),
}))

vi.mock('../../services/metadata', () => ({ getMetadata: serviceMocks.metadata }))
vi.mock('../../services/requests', () => ({ requestsService: { list: serviceMocks.list } }))

const metadata: Metadata = {
  categories: [
    { value: 'ti', label: 'TI' },
    { value: 'rh', label: 'RH' },
    { value: 'compras', label: 'Compras' },
    { value: 'financeiro', label: 'Financeiro' },
    { value: 'infraestrutura', label: 'Infraestrutura' },
  ],
  statuses: [
    { value: 'aberto', label: 'Aberto' },
    { value: 'em_atendimento', label: 'Em Atendimento' },
    { value: 'concluido', label: 'Concluído' },
  ],
}

const firstItem: RequestListItem = {
  id: 9,
  code: 'SOL-000009',
  title: 'Acesso ao sistema',
  category: { value: 'ti', label: 'TI' },
  requester: { id: 1, username: 'colaborador1', display_name: 'Colaborador 1' },
  created_at: '2026-10-02T02:30:00Z',
  status: { value: 'aberto', label: 'Aberto' },
}

function page(items: RequestListItem[], current = 1, totalItems = items.length, totalPages = totalItems ? 1 : 0): PaginatedRequests {
  return { items, pagination: { page: current, page_size: 20, total_items: totalItems, total_pages: totalPages } }
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((promiseResolve) => { resolve = promiseResolve })
  return { promise, resolve }
}

function renderPage(initialEntry = '/solicitacoes') {
  const router = createMemoryRouter([
    { path: '/solicitacoes', element: <RequestsPage /> },
    { path: '/solicitacoes/nova', element: <h1>Nova solicitação</h1> },
    { path: '/solicitacoes/:id', element: <h1>Detalhe da solicitação</h1> },
  ], { initialEntries: [initialEntry] })
  render(<RouterProvider router={router} />)
  return router
}

beforeEach(() => {
  serviceMocks.metadata.mockReset().mockResolvedValue(metadata)
  serviceMocks.list.mockReset().mockResolvedValue(page([firstItem]))
})

afterEach(cleanup)

describe('listagem de solicitações', () => {
  it('consome URL e metadata, exibe todos os campos e formata abertura em São Paulo', async () => {
    serviceMocks.list.mockResolvedValueOnce(page([firstItem], 2, 21, 2))
    const router = renderPage('/solicitacoes?date_from=2026-10-01&category=ti&status=aberto&q=acesso&page=2')

    expect(await screen.findByText('SOL-000009')).toBeTruthy()
    expect(serviceMocks.list).toHaveBeenCalledWith({
      date_from: '2026-10-01', date_to: undefined, category: 'ti', status: 'aberto', q: 'acesso', page: 2, page_size: 20,
    }, expect.any(AbortSignal))
    expect((screen.getByLabelText('Data inicial') as HTMLInputElement).value).toBe('2026-10-01')
    expect((screen.getByLabelText('Categoria') as HTMLSelectElement).value).toBe('ti')
    expect(screen.getByText('Acesso ao sistema')).toBeTruthy()
    expect(screen.getByText('Colaborador 1')).toBeTruthy()
    expect(screen.getByText('@colaborador1')).toBeTruthy()
    expect(screen.getByText(/01\/10\/2026.*23:30/)).toBeTruthy()
    expect(screen.getAllByText('Aberto').length).toBeGreaterThan(0)
    expect(screen.getByRole('link', { name: 'Ver detalhes de SOL-000009' }).getAttribute('href')).toBe('/solicitacoes/9')
    expect(screen.getByRole('link', { name: 'Nova solicitação' })).toBeTruthy()
    expect((screen.getByRole('button', { name: 'Página anterior' }) as HTMLButtonElement).disabled).toBe(false)
    expect((screen.getByRole('button', { name: 'Próxima página' }) as HTMLButtonElement).disabled).toBe(true)

    fireEvent.click(screen.getByRole('button', { name: 'Página anterior' }))
    await waitFor(() => expect(serviceMocks.list).toHaveBeenCalledTimes(2))
    expect(router.state.location.search).not.toContain('page=')
    expect(router.state.location.search).toContain('category=ti')
  })

  it('aplica filtros somente no submit, volta à primeira página e limpa URL/formulário', async () => {
    const interaction = userEvent.setup()
    const router = renderPage('/solicitacoes?page=4')
    await screen.findByText('SOL-000009')
    expect(serviceMocks.list).toHaveBeenCalledTimes(1)

    fireEvent.change(screen.getByLabelText('Data inicial'), { target: { value: '2026-10-01' } })
    fireEvent.change(screen.getByLabelText('Data final'), { target: { value: '2026-10-31' } })
    await interaction.selectOptions(screen.getByLabelText('Categoria'), 'compras')
    await interaction.selectOptions(screen.getByLabelText('Status'), 'concluido')
    await interaction.type(screen.getByLabelText('Texto no título'), 'nota fiscal')
    expect(serviceMocks.list).toHaveBeenCalledTimes(1)

    await interaction.click(screen.getByRole('button', { name: 'Aplicar filtros' }))
    await waitFor(() => expect(serviceMocks.list).toHaveBeenCalledTimes(2))
    expect(router.state.location.search).toBe('?date_from=2026-10-01&date_to=2026-10-31&category=compras&status=concluido&q=nota+fiscal')
    expect(serviceMocks.list).toHaveBeenLastCalledWith({
      date_from: '2026-10-01', date_to: '2026-10-31', category: 'compras', status: 'concluido', q: 'nota fiscal', page: 1, page_size: 20,
    }, expect.any(AbortSignal))

    await interaction.click(screen.getByRole('button', { name: 'Limpar filtros' }))
    await waitFor(() => expect(serviceMocks.list).toHaveBeenCalledTimes(3))
    expect(router.state.location.search).toBe('')
    expect((screen.getByLabelText('Texto no título') as HTMLInputElement).value).toBe('')

    await act(async () => { await router.navigate(-1) })
    await waitFor(() => expect(serviceMocks.list).toHaveBeenCalledTimes(4))
    expect(router.state.location.search).toContain('category=compras')
    expect((screen.getByLabelText('Texto no título') as HTMLInputElement).value).toBe('nota fiscal')
  }, 10_000)

  it('impede intervalo invertido sem substituir a validação da API', async () => {
    const interaction = userEvent.setup()
    renderPage()
    await screen.findByText('SOL-000009')
    fireEvent.change(screen.getByLabelText('Data inicial'), { target: { value: '2026-10-03' } })
    fireEvent.change(screen.getByLabelText('Data final'), { target: { value: '2026-10-02' } })

    await interaction.click(screen.getByRole('button', { name: 'Aplicar filtros' }))

    expect(screen.getByRole('alert').textContent).toMatch(/data final deve ser igual ou posterior/i)
    expect(serviceMocks.list).toHaveBeenCalledTimes(1)
  })

  it('aborta consulta anterior e não deixa resposta antiga sobrescrever a URL atual', async () => {
    const oldRequest = deferred<PaginatedRequests>()
    const newRequest = deferred<PaginatedRequests>()
    let oldSignal: AbortSignal | undefined
    serviceMocks.list.mockImplementation((filters: RequestFilters, signal: AbortSignal) => {
      if (filters.q === 'nova') return newRequest.promise
      oldSignal = signal
      return oldRequest.promise
    })
    const router = renderPage()
    await waitFor(() => expect(serviceMocks.list).toHaveBeenCalledTimes(1))

    await act(async () => { await router.navigate('/solicitacoes?q=nova') })
    await waitFor(() => expect(serviceMocks.list).toHaveBeenCalledTimes(2))
    expect(oldSignal?.aborted).toBe(true)

    const newItem = { ...firstItem, id: 10, code: 'SOL-000010', title: 'Consulta nova' }
    await act(async () => newRequest.resolve(page([newItem])))
    expect(await screen.findByText('Consulta nova')).toBeTruthy()
    await act(async () => oldRequest.resolve(page([{ ...firstItem, title: 'Resposta antiga' }])))
    expect(screen.queryByText('Resposta antiga')).toBeNull()
    expect(screen.getByText('Consulta nova')).toBeTruthy()
  })

  it('distingue base vazia, filtro sem resultado, página além do fim e erro', async () => {
    serviceMocks.list.mockResolvedValueOnce(page([]))
    const router = renderPage()
    expect(await screen.findByText('Ainda não há solicitações')).toBeTruthy()

    serviceMocks.list.mockResolvedValueOnce(page([]))
    await act(async () => { await router.navigate('/solicitacoes?status=concluido') })
    expect(await screen.findByText('Nenhuma solicitação encontrada')).toBeTruthy()

    serviceMocks.list.mockResolvedValueOnce(page([], 9, 3, 1))
    await act(async () => { await router.navigate('/solicitacoes?page=9') })
    expect(await screen.findByText('Esta página não possui resultados')).toBeTruthy()

    serviceMocks.list.mockRejectedValueOnce(new ApiError({ status: 500, code: 'internal_error', message: 'Erro interno inesperado.' }))
    await act(async () => { await router.navigate('/solicitacoes?page=2') })
    expect(await screen.findByText('Não foi possível carregar a listagem')).toBeTruthy()
    expect(screen.queryByText(/0 solicitação/)).toBeNull()
    expect(screen.getByRole('button', { name: 'Tentar novamente' })).toBeTruthy()
  }, 10_000)
})

// @vitest-environment jsdom

import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, RouterProvider, useLocation } from 'react-router-dom'

import { ApiError } from '../../api/client'
import { RequestCreatePage } from '../../pages/RequestCreatePage'
import { RequestEditPage } from '../../pages/RequestEditPage'
import type { Metadata, Request } from '../../types/api'

const serviceMocks = vi.hoisted(() => ({
  metadata: vi.fn(),
  get: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
}))

vi.mock('../../services/metadata', () => ({ getMetadata: serviceMocks.metadata }))
vi.mock('../../services/requests', () => ({
  requestsService: {
    get: serviceMocks.get,
    create: serviceMocks.create,
    update: serviceMocks.update,
  },
}))

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

const editableRequest: Request = {
  id: 12,
  code: 'SOL-000012',
  title: 'Notebook não inicializa',
  description: 'Equipamento apresenta tela preta ao iniciar.',
  category: { value: 'ti', label: 'TI' },
  status: { value: 'aberto', label: 'Aberto' },
  requester: { id: 1, username: 'colaborador1', display_name: 'Colaborador 1' },
  created_at: '2026-10-02T12:00:00Z',
  updated_at: '2026-10-02T12:00:00Z',
  permissions: { can_edit: true, can_delete: true },
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((promiseResolve) => { resolve = promiseResolve })
  return { promise, resolve }
}

function Destination() {
  const location = useLocation()
  const state = (location.state ?? {}) as { message?: string }
  return <><h1>Destino de detalhes</h1>{state.message && <p>{state.message}</p>}</>
}

function renderForm(path: string) {
  const router = createMemoryRouter([
    { path: '/solicitacoes/nova', element: <RequestCreatePage /> },
    { path: '/solicitacoes/:id/editar', element: <RequestEditPage /> },
    { path: '/solicitacoes/:id', element: <Destination /> },
    { path: '/solicitacoes', element: <h1>Listagem</h1> },
  ], { initialEntries: [path] })
  render(<RouterProvider router={router} />)
  return router
}

beforeEach(() => {
  serviceMocks.metadata.mockReset().mockResolvedValue(metadata)
  serviceMocks.get.mockReset().mockResolvedValue(editableRequest)
  serviceMocks.create.mockReset().mockResolvedValue(editableRequest)
  serviceMocks.update.mockReset().mockResolvedValue(editableRequest)
  vi.spyOn(window, 'confirm').mockReturnValue(true)
})

afterEach(() => {
  vi.restoreAllMocks()
  cleanup()
})

describe('nova solicitação', () => {
  it('valida vazios no campo sem chamar a API', async () => {
    const interaction = userEvent.setup()
    renderForm('/solicitacoes/nova')
    await screen.findByRole('heading', { name: 'Nova solicitação' })

    await interaction.click(screen.getByRole('button', { name: 'Criar solicitação' }))

    expect(screen.getByText('Informe um título entre 3 e 150 caracteres.')).toBeTruthy()
    expect(screen.getByText('Informe uma descrição entre 10 e 5000 caracteres.')).toBeTruthy()
    expect(screen.getByText('Escolha uma categoria válida.')).toBeTruthy()
    expect(serviceMocks.create).not.toHaveBeenCalled()
  })

  it('envia somente campos normalizados, bloqueia duplicação e navega com feedback', async () => {
    const pending = deferred<Request>()
    serviceMocks.create.mockReturnValue(pending.promise)
    const interaction = userEvent.setup()
    renderForm('/solicitacoes/nova')
    await screen.findByRole('heading', { name: 'Nova solicitação' })
    fireEvent.change(screen.getByLabelText(/Título/), { target: { value: '  Acesso ao sistema  ' } })
    fireEvent.change(screen.getByLabelText(/Descrição/), { target: { value: '  Primeira linha\nSegunda linha  ' } })
    await interaction.selectOptions(screen.getByLabelText(/Categoria/), 'ti')

    await interaction.click(screen.getByRole('button', { name: 'Criar solicitação' }))
    expect(serviceMocks.create).toHaveBeenCalledTimes(1)
    expect(serviceMocks.create).toHaveBeenCalledWith({
      title: 'Acesso ao sistema', description: 'Primeira linha\nSegunda linha', category: 'ti',
    })
    expect((screen.getByRole('button', { name: 'Salvando…' }) as HTMLButtonElement).disabled).toBe(true)
    await interaction.click(screen.getByRole('button', { name: 'Salvando…' }))
    expect(serviceMocks.create).toHaveBeenCalledTimes(1)

    await act(async () => pending.resolve(editableRequest))
    expect(await screen.findByText('Solicitação criada com sucesso.')).toBeTruthy()
  }, 10_000)

  it('preserva conteúdo após falha de rede e permite cancelar sem mutação', async () => {
    serviceMocks.create.mockRejectedValue(new ApiError({ status: 0, code: 'network_error', message: 'Sem conexão.' }))
    const interaction = userEvent.setup()
    const router = renderForm('/solicitacoes/nova')
    await screen.findByRole('heading', { name: 'Nova solicitação' })
    fireEvent.change(screen.getByLabelText(/Título/), { target: { value: 'Solicitação preservada' } })
    fireEvent.change(screen.getByLabelText(/Descrição/), { target: { value: 'Descrição suficientemente longa.' } })
    await interaction.selectOptions(screen.getByLabelText(/Categoria/), 'rh')
    await interaction.click(screen.getByRole('button', { name: 'Criar solicitação' }))

    expect(await screen.findByText('Sem conexão.')).toBeTruthy()
    expect((screen.getByLabelText(/Título/) as HTMLInputElement).value).toBe('Solicitação preservada')
    expect((screen.getByLabelText(/Descrição/) as HTMLTextAreaElement).value).toBe('Descrição suficientemente longa.')

    vi.mocked(window.confirm).mockReturnValueOnce(false)
    await interaction.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(router.state.location.pathname).toBe('/solicitacoes/nova')
    vi.mocked(window.confirm).mockReturnValueOnce(true)
    await interaction.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(router.state.location.pathname).toBe('/solicitacoes')
    expect(serviceMocks.create).toHaveBeenCalledTimes(1)
  }, 10_000)
})

describe('editar solicitação', () => {
  it('carrega o recurso do autor e envia somente o campo alterado', async () => {
    const interaction = userEvent.setup()
    renderForm('/solicitacoes/12/editar')
    expect(await screen.findByDisplayValue('Notebook não inicializa')).toBeTruthy()
    fireEvent.change(screen.getByLabelText(/Título/), { target: { value: 'Notebook reinicia sozinho' } })

    await interaction.click(screen.getByRole('button', { name: 'Salvar alterações' }))

    expect(serviceMocks.get).toHaveBeenCalledWith(12, expect.any(AbortSignal))
    expect(serviceMocks.update).toHaveBeenCalledWith(12, { title: 'Notebook reinicia sozinho' })
    expect(await screen.findByText('Solicitação atualizada com sucesso.')).toBeTruthy()
  })

  it.each([
    [{ ...editableRequest, permissions: { can_edit: false, can_delete: false } }, 'Somente o autor pode editar esta solicitação.'],
    [{ ...editableRequest, status: { value: 'concluido' as const, label: 'Concluído' }, permissions: { can_edit: false, can_delete: false } }, 'Somente solicitações com status Aberto podem ser editadas.'],
  ])('bloqueia URL direta conforme permissão atual do backend', async (resource, message) => {
    serviceMocks.get.mockResolvedValue(resource)
    renderForm('/solicitacoes/12/editar')

    expect(await screen.findByText(message)).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'Salvar alterações' })).toBeNull()
    expect(serviceMocks.update).not.toHaveBeenCalled()
  })

  it.each([
    [403, 'forbidden', 'Você não tem permissão para editar esta solicitação.'],
    [409, 'request_not_open', 'A solicitação mudou de status e não pode mais ser editada.'],
  ])('trata resposta %s do PATCH sem perder alterações', async (status, code, message) => {
    serviceMocks.update.mockRejectedValue(new ApiError({ status, code, message: 'Falha do backend.' }))
    const interaction = userEvent.setup()
    renderForm('/solicitacoes/12/editar')
    await screen.findByDisplayValue('Notebook não inicializa')
    fireEvent.change(screen.getByLabelText(/Título/), { target: { value: 'Alteração ainda não salva' } })

    await interaction.click(screen.getByRole('button', { name: 'Salvar alterações' }))

    expect(await screen.findByText(new RegExp(message))).toBeTruthy()
    expect((screen.getByLabelText(/Título/) as HTMLInputElement).value).toBe('Alteração ainda não salva')
    if (status === 409) expect(screen.getByRole('link', { name: 'Recarregar detalhes' })).toBeTruthy()
  })
})

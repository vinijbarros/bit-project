// @vitest-environment jsdom

import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, RouterProvider, useLocation } from 'react-router-dom'

import { ApiError } from '../../api/client'
import { RequestDetailPage } from '../../pages/RequestDetailPage'
import type { Metadata, Request } from '../../types/api'

const serviceMocks = vi.hoisted(() => ({
  get: vi.fn(),
  updateStatus: vi.fn(),
  remove: vi.fn(),
  metadata: vi.fn(),
}))

vi.mock('../../services/metadata', () => ({ getMetadata: serviceMocks.metadata }))
vi.mock('../../services/requests', () => ({
  requestsService: {
    get: serviceMocks.get,
    updateStatus: serviceMocks.updateStatus,
    remove: serviceMocks.remove,
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

const request: Request = {
  id: 9,
  code: 'SOL-000009',
  title: 'Acesso ao sistema',
  description: 'Primeira linha\n<script>alert("não executar")</script>\nÚltima linha',
  category: { value: 'ti', label: 'TI' },
  status: { value: 'aberto', label: 'Aberto' },
  requester: { id: 1, username: 'colaborador1', display_name: 'Colaborador 1' },
  created_at: '2026-10-02T02:30:00Z',
  updated_at: '2026-10-02T15:00:00Z',
  permissions: { can_edit: true, can_delete: true },
}

function changedRequest(status: Request['status'], permissions = { can_edit: false, can_delete: false }): Request {
  return { ...request, status, permissions, updated_at: '2026-10-03T12:00:00Z' }
}

function Destination() {
  const location = useLocation()
  const state = (location.state ?? {}) as { message?: string }
  return <><h1>Listagem</h1>{state.message && <p>{state.message}</p>}</>
}

function renderDetail(path = '/solicitacoes/9', state?: unknown) {
  const router = createMemoryRouter([
    { path: '/solicitacoes/:id', element: <RequestDetailPage /> },
    { path: '/solicitacoes/:id/editar', element: <h1>Editar</h1> },
    { path: '/solicitacoes', element: <Destination /> },
  ], { initialEntries: [{ pathname: path, state }] })
  const result = render(<RouterProvider router={router} />)
  return { router, ...result }
}

function apiError(status: number, message: string) {
  return new ApiError({ status, code: `error_${status}`, message })
}

beforeEach(() => {
  serviceMocks.get.mockReset().mockResolvedValue(request)
  serviceMocks.updateStatus.mockReset()
  serviceMocks.remove.mockReset()
  serviceMocks.metadata.mockReset().mockResolvedValue(metadata)
  Object.defineProperty(HTMLDialogElement.prototype, 'showModal', {
    configurable: true,
    value(this: HTMLDialogElement) { this.open = true },
  })
  Object.defineProperty(HTMLDialogElement.prototype, 'close', {
    configurable: true,
    value(this: HTMLDialogElement) { this.open = false },
  })
})

afterEach(cleanup)

describe('detalhes e gerenciamento da solicitação', () => {
  it('consulta a URL direta, mostra todos os dados e renderiza a descrição somente como texto', async () => {
    const { container } = renderDetail('/solicitacoes/9', { from: '/solicitacoes?category=ti&page=2' })

    expect(screen.getByText('Carregando detalhes da solicitação…')).toBeTruthy()
    expect(await screen.findByRole('heading', { name: 'Acesso ao sistema' })).toBeTruthy()
    expect(serviceMocks.get).toHaveBeenCalledWith(9, expect.any(AbortSignal))
    expect(screen.getByText('SOL-000009')).toBeTruthy()
    expect(screen.getByText(/Primeira linha/).textContent).toContain('<script>alert("não executar")</script>')
    expect(container.querySelector('.request-description script')).toBeNull()
    expect(screen.getByText('Colaborador 1')).toBeTruthy()
    expect(screen.getByText('@colaborador1')).toBeTruthy()
    expect(screen.getAllByText('Aberto').length).toBeGreaterThan(0)
    expect(screen.getByText(/01\/10\/2026.*23:30/)).toBeTruthy()
    expect(screen.getByText(/02\/10\/2026.*12:00/)).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Voltar para a listagem' }).getAttribute('href')).toBe('/solicitacoes?category=ti&page=2')
  })

  it('fecha e reabre pelo servidor, recalculando edição e exclusão para o autor', async () => {
    const interaction = userEvent.setup()
    const concluded = changedRequest({ value: 'concluido', label: 'Concluído' })
    const reopened = changedRequest({ value: 'aberto', label: 'Aberto' }, { can_edit: true, can_delete: true })
    let resolveStatus!: (value: Request) => void
    serviceMocks.updateStatus
      .mockReturnValueOnce(new Promise<Request>((resolve) => { resolveStatus = resolve }))
      .mockResolvedValueOnce(reopened)
    renderDetail()
    await screen.findByRole('heading', { name: 'Acesso ao sistema' })

    expect(screen.getByRole('link', { name: 'Editar solicitação' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Excluir solicitação' })).toBeTruthy()
    await interaction.selectOptions(screen.getByLabelText('Novo status'), 'concluido')
    await interaction.click(screen.getByRole('button', { name: 'Alterar status' }))
    expect((screen.getByRole('button', { name: 'Alterando status…' }) as HTMLButtonElement).disabled).toBe(true)
    expect(serviceMocks.updateStatus).toHaveBeenCalledTimes(1)

    await act(async () => { resolveStatus(concluded) })
    expect(await screen.findByText('Status alterado para Concluído.')).toBeTruthy()
    expect(screen.queryByRole('link', { name: 'Editar solicitação' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'Excluir solicitação' })).toBeNull()

    await interaction.selectOptions(screen.getByLabelText('Novo status'), 'aberto')
    await interaction.click(screen.getByRole('button', { name: 'Alterar status' }))
    expect(await screen.findByText('Status alterado para Aberto.')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Editar solicitação' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Excluir solicitação' })).toBeTruthy()
    expect(serviceMocks.updateStatus).toHaveBeenNthCalledWith(1, 9, 'concluido')
    expect(serviceMocks.updateStatus).toHaveBeenNthCalledWith(2, 9, 'aberto')
  })

  it('permite que outro usuário altere status, mas não oferece edição nem exclusão', async () => {
    const interaction = userEvent.setup()
    const otherUserRequest = { ...request, permissions: { can_edit: false, can_delete: false } }
    serviceMocks.get.mockResolvedValue(otherUserRequest)
    serviceMocks.updateStatus.mockResolvedValue(changedRequest({ value: 'em_atendimento', label: 'Em Atendimento' }))
    renderDetail()
    await screen.findByRole('heading', { name: 'Acesso ao sistema' })

    expect(screen.queryByRole('link', { name: 'Editar solicitação' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'Excluir solicitação' })).toBeNull()
    await interaction.selectOptions(screen.getByLabelText('Novo status'), 'em_atendimento')
    await interaction.click(screen.getByRole('button', { name: 'Alterar status' }))
    expect(await screen.findByText('Status alterado para Em Atendimento.')).toBeTruthy()
    expect(serviceMocks.updateStatus).toHaveBeenCalledWith(9, 'em_atendimento')
  })

  it('cancela por botão ou Escape sem excluir e devolve o foco ao acionador', async () => {
    const interaction = userEvent.setup()
    renderDetail()
    await screen.findByRole('heading', { name: 'Acesso ao sistema' })
    const trigger = screen.getByRole('button', { name: 'Excluir solicitação' })

    await interaction.click(trigger)
    let dialog = screen.getByRole('dialog', { name: 'Excluir solicitação?' })
    expect(dialog.textContent).toContain('SOL-000009 — Acesso ao sistema')
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Cancelar' }))
    await interaction.click(screen.getByRole('button', { name: 'Cancelar' }))
    await waitFor(() => expect(document.activeElement).toBe(trigger))

    await interaction.click(trigger)
    dialog = screen.getByRole('dialog', { name: 'Excluir solicitação?' })
    fireEvent(dialog, new Event('cancel', { bubbles: false, cancelable: true }))
    await waitFor(() => expect(document.activeElement).toBe(trigger))
    expect(serviceMocks.remove).not.toHaveBeenCalled()
  })

  it('confirma a exclusão apenas uma vez e navega após a resposta do servidor', async () => {
    const interaction = userEvent.setup()
    let resolveDelete!: () => void
    serviceMocks.remove.mockReturnValue(new Promise<void>((resolve) => { resolveDelete = resolve }))
    const { router } = renderDetail('/solicitacoes/9', { from: '/solicitacoes?status=aberto' })
    await screen.findByRole('heading', { name: 'Acesso ao sistema' })

    await interaction.click(screen.getByRole('button', { name: 'Excluir solicitação' }))
    await interaction.click(screen.getByRole('button', { name: 'Excluir definitivamente' }))
    expect((screen.getByRole('button', { name: 'Processando…' }) as HTMLButtonElement).disabled).toBe(true)
    expect(serviceMocks.remove).toHaveBeenCalledTimes(1)
    expect(router.state.location.pathname).toBe('/solicitacoes/9')

    await act(async () => { resolveDelete() })
    expect(await screen.findByRole('heading', { name: 'Listagem' })).toBeTruthy()
    expect(router.state.location.pathname + router.state.location.search).toBe('/solicitacoes?status=aberto')
    expect(screen.getByText('SOL-000009 foi excluída com sucesso.')).toBeTruthy()
  })

  it.each([
    [409, changedRequest({ value: 'concluido', label: 'Concluído' }), 'não está mais aberta'],
    [403, { ...request, permissions: { can_edit: false, can_delete: false } }, 'permissão para excluir'],
  ])('recarrega a representação após erro %i na exclusão', async (status, refreshed, expectedMessage) => {
    const interaction = userEvent.setup()
    serviceMocks.get.mockResolvedValueOnce(request).mockResolvedValueOnce(refreshed)
    serviceMocks.remove.mockRejectedValue(apiError(status, 'Conflito concorrente'))
    renderDetail()
    await screen.findByRole('heading', { name: 'Acesso ao sistema' })

    await interaction.click(screen.getByRole('button', { name: 'Excluir solicitação' }))
    await interaction.click(screen.getByRole('button', { name: 'Excluir definitivamente' }))
    expect(await screen.findByText(new RegExp(expectedMessage))).toBeTruthy()
    expect(serviceMocks.get).toHaveBeenCalledTimes(2)
    expect(screen.queryByRole('button', { name: 'Excluir solicitação' })).toBeNull()
  })

  it('mostra que o recurso foi removido quando o DELETE retorna 404', async () => {
    const interaction = userEvent.setup()
    serviceMocks.remove.mockRejectedValue(apiError(404, 'Não encontrado'))
    const { router } = renderDetail()
    await screen.findByRole('heading', { name: 'Acesso ao sistema' })

    await interaction.click(screen.getByRole('button', { name: 'Excluir solicitação' }))
    await interaction.click(screen.getByRole('button', { name: 'Excluir definitivamente' }))
    expect(await screen.findByRole('heading', { name: 'Solicitação não encontrada' })).toBeTruthy()
    expect(screen.getByText('A solicitação já foi removida por outra sessão.')).toBeTruthy()
    expect(router.state.location.pathname).toBe('/solicitacoes/9')
  })

  it('distingue recurso removido de falha de infraestrutura', async () => {
    serviceMocks.get.mockRejectedValueOnce(apiError(404, 'Não encontrado'))
    const { unmount } = renderDetail()
    expect(await screen.findByRole('heading', { name: 'Solicitação não encontrada' })).toBeTruthy()
    expect(screen.getByText('A solicitação informada não existe ou foi excluída.')).toBeTruthy()
    unmount()
    cleanup()

    serviceMocks.get.mockRejectedValueOnce(new Error('conexão encerrada'))
    renderDetail()
    expect(await screen.findByRole('heading', { name: 'Não foi possível consultar a solicitação' })).toBeTruthy()
    expect(screen.getByText(/Verifique sua conexão/)).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Tentar novamente' })).toBeTruthy()
  })

  it('rejeita ID inválido sem consultar detalhes', () => {
    renderDetail('/solicitacoes/invalido')
    expect(screen.getByText('Solicitação inválida')).toBeTruthy()
    expect(serviceMocks.get).not.toHaveBeenCalled()
  })
})

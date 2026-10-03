// @vitest-environment jsdom

import { act, cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'

import { ApiError } from '../../api/client'
import { DashboardPage } from '../../pages/DashboardPage'
import type { Dashboard } from '../../types/api'

const dashboardMock = vi.hoisted(() => vi.fn())

vi.mock('../../services/dashboard', () => ({ getDashboard: dashboardMock }))

function renderDashboard() {
  const router = createMemoryRouter([
    { path: '/dashboard', element: <DashboardPage /> },
    { path: '/solicitacoes', element: <h1>Solicitações</h1> },
    { path: '/solicitacoes/nova', element: <h1>Nova solicitação</h1> },
    { path: '/outra', element: <h1>Outra página</h1> },
  ], { initialEntries: ['/dashboard'] })
  render(<RouterProvider router={router} />)
  return router
}

beforeEach(() => dashboardMock.mockReset())
afterEach(cleanup)

describe('dashboard', () => {
  it('apresenta somente os quatro indicadores globais vindos da API', async () => {
    dashboardMock.mockResolvedValue({ total: 12, abertas: 5, em_atendimento: 4, concluidas: 3 } satisfies Dashboard)
    renderDashboard()

    expect(screen.getByText('Carregando indicadores…')).toBeTruthy()
    expect(await screen.findByText('Quantidade total de solicitações')).toBeTruthy()
    expect(dashboardMock).toHaveBeenCalledWith(expect.any(AbortSignal))
    expect(screen.getByText('12')).toBeTruthy()
    expect(screen.getByText('5')).toBeTruthy()
    expect(screen.getByText('4')).toBeTruthy()
    expect(screen.getByText('3')).toBeTruthy()
    expect(screen.getByText(/Indicadores globais de todas as solicitações/)).toBeTruthy()
    expect(screen.queryByText(/meta|gráfico/i)).toBeNull()
  })

  it('leva cada card à listagem com o filtro correspondente', async () => {
    dashboardMock.mockResolvedValue({ total: 8, abertas: 3, em_atendimento: 2, concluidas: 3 } satisfies Dashboard)
    renderDashboard()
    await screen.findByText('Quantidade total de solicitações')

    expect(screen.getByRole('link', { name: /Quantidade total de solicitações.*8.*Ver todas/ }).getAttribute('href')).toBe('/solicitacoes')
    expect(screen.getByRole('link', { name: /Abertas.*3.*Ver solicitações/ }).getAttribute('href')).toBe('/solicitacoes?status=aberto')
    expect(screen.getByRole('link', { name: /Em Atendimento.*2.*Ver solicitações/ }).getAttribute('href')).toBe('/solicitacoes?status=em_atendimento')
    expect(screen.getByRole('link', { name: /Concluídas.*3.*Ver solicitações/ }).getAttribute('href')).toBe('/solicitacoes?status=concluido')
  })

  it('exibe os quatro zeros verdadeiros quando a base está vazia', async () => {
    dashboardMock.mockResolvedValue({ total: 0, abertas: 0, em_atendimento: 0, concluidas: 0 } satisfies Dashboard)
    renderDashboard()

    expect(await screen.findByText('Não há solicitações cadastradas. Os quatro indicadores globais estão zerados.')).toBeTruthy()
    expect(screen.getAllByText('0')).toHaveLength(4)
  })

  it('diferencia erro de base vazia e permite repetir', async () => {
    dashboardMock
      .mockRejectedValueOnce(new ApiError({ status: 500, code: 'internal_error', message: 'Erro interno inesperado.' }))
      .mockResolvedValueOnce({ total: 2, abertas: 1, em_atendimento: 1, concluidas: 0 } satisfies Dashboard)
    const interaction = userEvent.setup()
    renderDashboard()

    expect(await screen.findByRole('heading', { name: 'Não foi possível carregar o dashboard' })).toBeTruthy()
    expect(screen.queryByText('Não há solicitações cadastradas. Os quatro indicadores globais estão zerados.')).toBeNull()
    await interaction.click(screen.getByRole('button', { name: 'Tentar novamente' }))
    expect(await screen.findByText('2')).toBeTruthy()
    expect(dashboardMock).toHaveBeenCalledTimes(2)
  })

  it('faz uma nova consulta ao voltar ao dashboard e não conserva contadores antigos', async () => {
    dashboardMock
      .mockResolvedValueOnce({ total: 1, abertas: 1, em_atendimento: 0, concluidas: 0 } satisfies Dashboard)
      .mockResolvedValueOnce({ total: 2, abertas: 0, em_atendimento: 1, concluidas: 1 } satisfies Dashboard)
    const router = renderDashboard()
    expect((await screen.findAllByText('1')).length).toBeGreaterThan(0)

    await act(async () => { await router.navigate('/outra') })
    await act(async () => { await router.navigate('/dashboard') })
    expect(await screen.findByText('2')).toBeTruthy()
    expect(dashboardMock).toHaveBeenCalledTimes(2)
  })
})

// @vitest-environment jsdom

import { act, cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router-dom'

import { ApiError } from '../../api/client'
import { ProtectedRoute } from '../../app/routes/ProtectedRoute'
import { PublicOnlyRoute } from '../../app/routes/PublicOnlyRoute'
import { AppLayout } from '../../components/layout/AppLayout'
import { safeInternalDestination } from '../../app/routes/destination'
import { LoginPage } from '../../pages/LoginPage'
import type { User } from '../../types/api'
import { AuthProvider } from './AuthContext'

const authMocks = vi.hoisted(() => ({
  currentUser: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  unauthorizedHandler: undefined as (() => void) | undefined,
}))

vi.mock('../../services/auth', () => ({
  authService: {
    currentUser: authMocks.currentUser,
    login: authMocks.login,
    logout: authMocks.logout,
  },
}))

vi.mock('../../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../api/client')>()
  return {
    ...actual,
    apiClient: {
      setUnauthorizedHandler(handler?: () => void) {
        authMocks.unauthorizedHandler = handler
      },
    },
  }
})

const user: User = { id: 1, username: 'colaborador1', display_name: 'Colaborador 1' }

function unauthenticatedError() {
  return new ApiError({ status: 401, code: 'authentication_required', message: 'Autenticação necessária.' })
}

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((promiseResolve, promiseReject) => {
    resolve = promiseResolve
    reject = promiseReject
  })
  return { promise, resolve, reject }
}

function renderAuthApp(initialEntry: string) {
  const router = createMemoryRouter([
    {
      element: <PublicOnlyRoute />,
      children: [{ path: '/login', element: <LoginPage /> }],
    },
    {
      element: <ProtectedRoute />,
      children: [
        {
          element: <AppLayout />,
          children: [
            { path: '/dashboard', element: <h1>Dashboard protegido</h1> },
            { path: '/solicitacoes/:id', element: <h1>Detalhe protegido</h1> },
          ],
        },
        { path: '/area-teste', element: <Outlet /> },
      ],
    },
  ], { initialEntries: [initialEntry] })
  render(<AuthProvider><RouterProvider router={router} /></AuthProvider>)
  return router
}

beforeEach(() => {
  authMocks.currentUser.mockReset()
  authMocks.login.mockReset()
  authMocks.logout.mockReset()
  authMocks.unauthorizedHandler = undefined
})

afterEach(cleanup)

describe('fluxo de autenticação', () => {
  it('valida preenchimento e mostra credenciais inválidas sem expirar uma sessão', async () => {
    authMocks.currentUser.mockRejectedValue(unauthenticatedError())
    authMocks.login.mockRejectedValue(new ApiError({ status: 401, code: 'invalid_credentials', message: 'Credenciais inválidas.' }))
    const interaction = userEvent.setup()
    renderAuthApp('/login')

    await screen.findByRole('heading', { name: 'Acesso ao portal' })
    await interaction.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(screen.getByText('Informe o usuário.')).toBeTruthy()
    expect(screen.getByText('Informe a senha.')).toBeTruthy()
    expect(authMocks.login).not.toHaveBeenCalled()

    await interaction.type(screen.getByLabelText(/Usuário/), 'colaborador1')
    await interaction.type(screen.getByLabelText(/Senha/), 'incorreta')
    await interaction.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(await screen.findByText('Usuário ou senha inválidos.')).toBeTruthy()
    expect(screen.queryByText(/sessão expirou/i)).toBeNull()
  }, 10_000)

  it('preserva destino interno, bloqueia submissão duplicada e entra com sucesso', async () => {
    authMocks.currentUser.mockRejectedValue(unauthenticatedError())
    const pendingLogin = deferred<User>()
    authMocks.login.mockReturnValue(pendingLogin.promise)
    const interaction = userEvent.setup()
    renderAuthApp('/solicitacoes/42?origem=lista')

    await screen.findByRole('heading', { name: 'Acesso ao portal' })
    await interaction.type(screen.getByLabelText(/Usuário/), 'colaborador1')
    await interaction.type(screen.getByLabelText(/Senha/), 'SenhaLocal-1')
    await interaction.click(screen.getByRole('button', { name: 'Entrar' }))
    expect((screen.getByRole('button', { name: 'Entrando…' }) as HTMLButtonElement).disabled).toBe(true)
    await interaction.click(screen.getByRole('button', { name: 'Entrando…' }))
    expect(authMocks.login).toHaveBeenCalledTimes(1)

    await act(async () => pendingLogin.resolve(user))
    expect(await screen.findByRole('heading', { name: 'Detalhe protegido' })).toBeTruthy()
    expect(screen.queryByLabelText(/Senha/)).toBeNull()
  })

  it('não mostra rota protegida antes de recuperar a sessão existente', async () => {
    const pendingSession = deferred<User>()
    authMocks.currentUser.mockReturnValue(pendingSession.promise)
    renderAuthApp('/solicitacoes/9')

    expect(screen.getByText('Verificando sua sessão…')).toBeTruthy()
    expect(screen.queryByText('Detalhe protegido')).toBeNull()
    await act(async () => pendingSession.resolve(user))
    expect(await screen.findByText('Detalhe protegido')).toBeTruthy()
    expect(screen.getByText('Colaborador 1')).toBeTruthy()
  })

  it('só limpa o estado após logout confirmado e protege o histórico', async () => {
    authMocks.currentUser.mockResolvedValue(user)
    const pendingLogout = deferred<void>()
    authMocks.logout.mockReturnValue(pendingLogout.promise)
    const interaction = userEvent.setup()
    const router = renderAuthApp('/dashboard')

    expect(await screen.findByText('Dashboard protegido')).toBeTruthy()
    await interaction.click(screen.getByRole('button', { name: 'Sair' }))
    expect((screen.getByRole('button', { name: 'Saindo…' }) as HTMLButtonElement).disabled).toBe(true)
    expect(screen.getByText('Dashboard protegido')).toBeTruthy()
    await act(async () => pendingLogout.resolve())
    expect(await screen.findByText('Sessão encerrada com segurança.')).toBeTruthy()

    await act(async () => router.navigate('/dashboard'))
    expect(await screen.findByRole('heading', { name: 'Acesso ao portal' })).toBeTruthy()
    expect(screen.queryByText('Dashboard protegido')).toBeNull()
  })

  it('mantém autenticação quando logout falha e permite repetir', async () => {
    authMocks.currentUser.mockResolvedValue(user)
    authMocks.logout
      .mockRejectedValueOnce(new ApiError({ status: 0, code: 'network_error', message: 'Sem conexão.' }))
      .mockResolvedValueOnce(undefined)
    const interaction = userEvent.setup()
    renderAuthApp('/dashboard')

    await screen.findByText('Dashboard protegido')
    await interaction.click(screen.getByRole('button', { name: 'Sair' }))
    expect(await screen.findByText('Não foi possível encerrar a sessão. Tente novamente.')).toBeTruthy()
    expect(screen.getByText('Dashboard protegido')).toBeTruthy()
    await interaction.click(screen.getByRole('button', { name: 'Sair' }))
    expect(await screen.findByText('Sessão encerrada com segurança.')).toBeTruthy()
    expect(authMocks.logout).toHaveBeenCalledTimes(2)
  })

  it('interrompe a área protegida e informa expiração após 401 operacional', async () => {
    authMocks.currentUser.mockResolvedValue(user)
    renderAuthApp('/solicitacoes/7')
    await screen.findByText('Detalhe protegido')

    act(() => authMocks.unauthorizedHandler?.())
    expect(await screen.findByText('Sua sessão expirou. Entre novamente para continuar.')).toBeTruthy()
    expect(screen.queryByText('Detalhe protegido')).toBeNull()
  })

  it('exibe indisponibilidade de /me e permite repetir sem fingir logout', async () => {
    authMocks.currentUser
      .mockRejectedValueOnce(new ApiError({ status: 500, code: 'internal_error', message: 'Erro interno inesperado.' }))
      .mockResolvedValueOnce(user)
    const interaction = userEvent.setup()
    renderAuthApp('/dashboard')

    expect(await screen.findByText('Não foi possível verificar sua sessão')).toBeTruthy()
    expect(screen.queryByText('Acesso ao portal')).toBeNull()
    await interaction.click(screen.getByRole('button', { name: 'Tentar novamente' }))
    expect(await screen.findByText('Dashboard protegido')).toBeTruthy()
  })
})

describe('destino pós-login', () => {
  it.each([
    ['https://evil.example', '/dashboard'],
    ['//evil.example/path', '/dashboard'],
    ['/rota-inexistente', '/dashboard'],
    ['/solicitacoes/42?origem=lista', '/solicitacoes/42?origem=lista'],
  ])('normaliza %s para %s', (candidate, expected) => {
    expect(safeInternalDestination(candidate)).toBe(expected)
  })
})

import { expect, request, test, type APIRequestContext, type Page } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL ?? 'http://127.0.0.1:8080'
const user1 = {
  username: process.env.E2E_USER1_USERNAME ?? 'colaborador1',
  password: process.env.E2E_USER1_PASSWORD ?? 'DemoLocal-Colaborador1',
}
const user2 = {
  username: process.env.E2E_USER2_USERNAME ?? 'colaborador2',
  password: process.env.E2E_USER2_PASSWORD ?? 'DemoLocal-Colaborador2',
}
const requestTitle = 'E2E Playwright - compra de periférico'
const editedTitle = 'E2E Playwright - compra de monitor'

async function login(page: Page, credentials = user1) {
  await page.goto('/login')
  await page.getByLabel('Usuário').fill(credentials.username)
  await page.getByLabel('Senha').fill(credentials.password)
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page.getByRole('button', { name: 'Sair' })).toBeVisible()
}

function saoPauloDate(instant: string): string {
  const parts = new Intl.DateTimeFormat('en', {
    timeZone: 'America/Sao_Paulo',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(new Date(instant))
  const value = (type: Intl.DateTimeFormatPartTypes) => parts.find((part) => part.type === type)?.value
  return `${value('year')}-${value('month')}-${value('day')}`
}

async function dashboardTotal(page: Page): Promise<number> {
  const card = page.getByRole('link', { name: /Quantidade total de solicitações/ })
  return Number(await card.locator('strong').innerText())
}

async function authenticatedAPI(): Promise<APIRequestContext> {
  const context = await request.newContext({ baseURL, extraHTTPHeaders: { Origin: baseURL } })
  const response = await context.post('/api/v1/auth/login', {
    data: { username: user1.username, password: user1.password },
  })
  expect(response.ok()).toBeTruthy()
  return context
}

async function removeStaleRequests() {
  const api = await authenticatedAPI()
  try {
    for (const title of [requestTitle, editedTitle]) {
      const response = await api.get(`/api/v1/requests?q=${encodeURIComponent(title)}&page_size=100`)
      expect(response.ok()).toBeTruthy()
      const payload = await response.json() as {
        items: Array<{ id: number; title: string; status: { value: string }; requester: { username: string } }>
      }
      for (const item of payload.items) {
        if (item.title !== title || item.requester.username !== user1.username) continue
        if (item.status.value !== 'aberto') {
          const reopened = await api.patch(`/api/v1/requests/${item.id}/status`, { data: { status: 'aberto' } })
          expect(reopened.ok()).toBeTruthy()
        }
        const removed = await api.delete(`/api/v1/requests/${item.id}`)
        expect(removed.status()).toBe(204)
      }
    }
  } finally {
    await api.post('/api/v1/auth/logout')
    await api.dispose()
  }
}

test.describe('fluxo integrado do portal', () => {
  test.beforeEach(async () => removeStaleRequests())
  test.afterEach(async () => removeStaleRequests())

  test('autentica, executa o ciclo completo da solicitação e encerra a sessão', async ({ page, browser }) => {
    await page.goto('/solicitacoes')
    await expect(page).toHaveURL(/\/login$/)

    await login(page)
    await page.goto('/dashboard')
    const initialTotal = await dashboardTotal(page)

    await page.getByRole('link', { name: 'Nova solicitação' }).first().click()
    await page.getByLabel('Título').fill(requestTitle)
    await page.getByLabel('Descrição').fill('Solicitação controlada pelo E2E para aquisição de um novo monitor.')
    await page.getByLabel('Categoria').selectOption({ label: 'Compras' })
    await page.getByRole('button', { name: 'Criar solicitação' }).click()
    await expect(page.getByRole('status')).toContainText('Solicitação criada com sucesso.')
    await expect(page.getByRole('heading', { name: requestTitle })).toBeVisible()
    const detailURL = page.url()
    const requestID = Number(new URL(detailURL).pathname.split('/').at(-1))
    expect(requestID).toBeGreaterThan(0)
    const createdAt = await page.evaluate(async (id) => {
      const response = await fetch(`/api/v1/requests/${id}`)
      const payload = await response.json() as { data: { created_at: string } }
      return payload.data.created_at
    }, requestID)

    await page.reload()
    await expect(page.getByRole('heading', { name: requestTitle })).toBeVisible()

    await page.getByRole('link', { name: 'Editar solicitação' }).click()
    await page.getByLabel('Título').fill(editedTitle)
    await page.getByRole('button', { name: 'Salvar alterações' }).click()
    await expect(page.getByRole('heading', { name: editedTitle })).toBeVisible()

    await page.getByRole('link', { name: 'Solicitações', exact: true }).click()
    const today = saoPauloDate(createdAt)
    await page.getByLabel('Data inicial').fill(today)
    await page.getByLabel('Data final').fill(today)
    await page.getByLabel('Categoria').selectOption({ label: 'Compras' })
    await page.getByLabel('Status').selectOption({ label: 'Aberto' })
    await page.getByLabel('Texto no título').fill('compra de monitor')
    await page.getByRole('button', { name: 'Aplicar filtros' }).click()
    await expect(page).toHaveURL(new RegExp(`date_from=${today}.*date_to=${today}.*category=compras.*status=aberto.*q=compra`))
    for (const heading of ['Código', 'Título', 'Categoria', 'Solicitante', 'Data de abertura', 'Status']) {
      await expect(page.getByRole('columnheader', { name: heading, exact: true })).toBeVisible()
    }
    const resultRow = page.getByRole('row').filter({ hasText: editedTitle })
    await expect(resultRow).toBeVisible()
    await resultRow.getByRole('link', { name: /Ver detalhes de SOL-/ }).click()

    const otherContext = await browser.newContext()
    const otherPage = await otherContext.newPage()
    try {
      await login(otherPage, user2)
      await otherPage.goto(`/solicitacoes/${requestID}`)
      await expect(otherPage.getByRole('heading', { name: editedTitle })).toBeVisible()
      await expect(otherPage.getByRole('link', { name: 'Editar solicitação' })).toHaveCount(0)
      await expect(otherPage.getByRole('button', { name: 'Excluir solicitação' })).toHaveCount(0)

      await otherPage.goto(`/solicitacoes/${requestID}/editar`)
      await expect(otherPage.getByRole('heading', { name: 'Edição não permitida' })).toBeVisible()
      const forbidden = await otherPage.evaluate(async (id) => {
        const edit = await fetch(`/api/v1/requests/${id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: 'Alteração indevida pelo segundo usuário' }),
        })
        const remove = await fetch(`/api/v1/requests/${id}`, { method: 'DELETE' })
        return { edit: edit.status, remove: remove.status }
      }, requestID)
      expect(forbidden).toEqual({ edit: 403, remove: 403 })

      await otherPage.goto(`/solicitacoes/${requestID}`)
      await otherPage.getByLabel('Novo status').selectOption('em_atendimento')
      await otherPage.getByRole('button', { name: 'Alterar status' }).click()
      await expect(otherPage.getByRole('status')).toContainText('Status alterado para Em Atendimento.')

      await page.reload()
      const currentStatus = page.getByText('Status atual').locator('..')
      await expect(currentStatus.getByText('Em Atendimento', { exact: true })).toBeVisible()
      await expect(page.getByRole('link', { name: 'Editar solicitação' })).toHaveCount(0)
      await expect(page.getByRole('button', { name: 'Excluir solicitação' })).toHaveCount(0)
      const closedConflict = await page.evaluate(async (id) => {
        const edit = await fetch(`/api/v1/requests/${id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: 'Alteração enquanto fechada' }),
        })
        const remove = await fetch(`/api/v1/requests/${id}`, { method: 'DELETE' })
        return { edit: edit.status, remove: remove.status }
      }, requestID)
      expect(closedConflict).toEqual({ edit: 409, remove: 409 })

      await otherPage.getByLabel('Novo status').selectOption('concluido')
      await otherPage.getByRole('button', { name: 'Alterar status' }).click()
      await expect(otherPage.getByRole('status')).toContainText('Status alterado para Concluído.')
      await otherPage.getByLabel('Novo status').selectOption('aberto')
      await otherPage.getByRole('button', { name: 'Alterar status' }).click()
      await expect(otherPage.getByRole('status')).toContainText('Status alterado para Aberto.')
      await otherPage.getByRole('button', { name: 'Sair' }).click()
      await expect(otherPage).toHaveURL(/\/login$/)
    } finally {
      await otherContext.close()
    }

    await page.reload()
    await expect(page.getByRole('link', { name: 'Editar solicitação' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Excluir solicitação' })).toBeVisible()

    await page.getByRole('link', { name: 'Dashboard' }).click()
    await expect.poll(() => dashboardTotal(page)).toBe(initialTotal + 1)
    await page.goto(`/solicitacoes/${requestID}`)
    await page.getByRole('button', { name: 'Excluir solicitação' }).click()
    const dialog = page.getByRole('dialog', { name: 'Excluir solicitação?' })
    await expect(dialog).toContainText(editedTitle)
    await dialog.getByRole('button', { name: 'Excluir definitivamente' }).click()
    await expect(page.getByRole('status').filter({ hasText: /SOL-\d+ foi excluída com sucesso/ })).toBeVisible()

    await page.getByRole('link', { name: 'Dashboard' }).click()
    await expect.poll(() => dashboardTotal(page)).toBe(initialTotal)
    await page.getByRole('button', { name: 'Sair' }).click()
    await expect(page).toHaveURL(/\/login$/)
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login$/)
    await expect(page.getByRole('heading', { name: 'Acesso ao portal' })).toBeVisible()
  })
})

for (const viewport of [
  { name: 'desktop', width: 1440, height: 900 },
  { name: 'celular', width: 375, height: 812 },
]) {
  test(`smoke responsivo em ${viewport.name}`, async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: viewport.width, height: viewport.height } })
    const page = await context.newPage()
    try {
      await login(page)
      await expect(page.getByRole('navigation', { name: 'Navegação principal' })).toBeVisible()
      await page.getByRole('link', { name: 'Solicitações', exact: true }).click()
      await expect(page.getByRole('link', { name: 'Nova solicitação' }).first()).toBeVisible()
      await expect(page.getByLabel('Data inicial')).toBeVisible()
      await expect(page.getByLabel('Texto no título')).toBeVisible()
      await expect(page.getByRole('button', { name: 'Aplicar filtros' })).toBeVisible()
      const dimensions = await page.evaluate(() => ({ body: document.body.scrollWidth, viewport: window.innerWidth }))
      expect(dimensions.body).toBeLessThanOrEqual(dimensions.viewport)
      await page.mouse.wheel(1000, 0)
      await expect.poll(() => page.evaluate(() => window.scrollX)).toBe(0)
      if (viewport.name === 'celular') {
        const table = page.getByRole('region', { name: 'Resultados das solicitações' })
        const tableScroll = await table.evaluate((element) => {
          element.scrollLeft = element.scrollWidth
          return { left: element.scrollLeft, client: element.clientWidth, content: element.scrollWidth }
        })
        expect(tableScroll.content).toBeGreaterThan(tableScroll.client)
        expect(tableScroll.left).toBeGreaterThan(0)
      }
      await page.getByRole('button', { name: 'Sair' }).click()
      await expect(page).toHaveURL(/\/login$/)
    } finally {
      await context.close()
    }
  })
}

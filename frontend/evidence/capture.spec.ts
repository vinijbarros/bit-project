import { expect, test } from '@playwright/test'
import path from 'node:path'

const evidenceDirectory = path.resolve('../docs/evidencias')
const username = process.env.E2E_USER1_USERNAME ?? 'colaborador1'
const password = process.env.E2E_USER1_PASSWORD ?? 'DemoLocal-Colaborador1'

async function capture(page: import('@playwright/test').Page, filename: string, fullPage = true) {
  await page.screenshot({
    path: path.join(evidenceDirectory, filename),
    fullPage,
    animations: 'disabled',
  })
}

test('gera evidências reais da entrega sem expor credenciais ou token', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByRole('heading', { name: 'Acesso ao portal' })).toBeVisible()
  await capture(page, '01-login.png')

  await page.getByLabel('Usuário').fill(username)
  await page.getByLabel('Senha').fill(password)
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByRole('link', { name: /Quantidade total de solicitações/ })).toBeVisible()
  await capture(page, '02-dashboard.png')

  await page.getByRole('link', { name: 'Solicitações', exact: true }).click()
  await expect(page.getByText('5 solicitação(ões) encontrada(s).')).toBeVisible()
  await capture(page, '03-listagem-completa.png')

  await page.getByLabel('Categoria').selectOption('infraestrutura')
  await page.getByLabel('Status').selectOption('em_atendimento')
  await page.getByLabel('Texto no título').fill('iluminação')
  await page.getByRole('button', { name: 'Aplicar filtros' }).click()
  await expect(page.getByText('1 solicitação(ões) encontrada(s).')).toBeVisible()
  await expect(page.getByText('Ajuste de iluminação da sala')).toBeVisible()
  await capture(page, '04-filtros-combinados.png')

  await page.getByRole('link', { name: 'Nova solicitação' }).first().click()
  await page.getByLabel('Título').fill('Evidência final - acesso à sala de projetos')
  await page.getByLabel('Descrição').fill('Solicitação sintética criada durante a validação final da entrega.')
  await page.getByLabel('Categoria').selectOption('infraestrutura')
  await page.getByRole('button', { name: 'Criar solicitação' }).click()
  await expect(page.getByRole('status')).toContainText('Solicitação criada com sucesso.')
  await capture(page, '05-criacao-sucesso.png')

  await page.reload()
  await expect(page.getByRole('heading', { name: 'Evidência final - acesso à sala de projetos' })).toBeVisible()
  await expect(page.getByText('Solicitação sintética criada durante a validação final da entrega.')).toBeVisible()
  await capture(page, '06-detalhes-aberto.png')

  await page.getByLabel('Novo status').selectOption('em_atendimento')
  await page.getByRole('button', { name: 'Alterar status' }).click()
  await expect(page.getByRole('status')).toContainText('Status alterado para Em Atendimento.')
  await capture(page, '07-status-em-atendimento.png')

  await page.getByLabel('Novo status').selectOption('concluido')
  await page.getByRole('button', { name: 'Alterar status' }).click()
  await expect(page.getByRole('status')).toContainText('Status alterado para Concluído.')
  await capture(page, '08-status-concluido.png')

  await page.setViewportSize({ width: 375, height: 812 })
  await page.goto('/solicitacoes')
  await expect(page.getByRole('button', { name: 'Aplicar filtros' })).toBeVisible()
  await capture(page, '09-layout-movel.png', false)

  await page.getByRole('button', { name: 'Sair' }).click()
  await expect(page).toHaveURL(/\/login$/)
})

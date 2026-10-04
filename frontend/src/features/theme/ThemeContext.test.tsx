// @vitest-environment jsdom

import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { ThemeToggle } from '../../components/theme/ThemeToggle'
import { THEME_STORAGE_KEY, ThemeProvider } from './ThemeContext'

beforeEach(() => {
  window.localStorage.clear()
  delete document.documentElement.dataset.theme
})

afterEach(() => {
  cleanup()
  window.localStorage.clear()
  delete document.documentElement.dataset.theme
})

describe('preferência de tema', () => {
  it('ativa o modo noturno e persiste a escolha', async () => {
    const interaction = userEvent.setup()
    const { unmount } = render(<ThemeProvider><ThemeToggle /></ThemeProvider>)

    const darkModeButton = screen.getByRole('button', { name: 'Ativar modo noturno' })
    expect(darkModeButton.getAttribute('aria-pressed')).toBe('false')
    expect(document.documentElement.dataset.theme).toBe('light')

    await interaction.click(darkModeButton)
    expect(screen.getByRole('button', { name: 'Ativar modo claro' }).getAttribute('aria-pressed')).toBe('true')
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe('dark')

    unmount()
    render(<ThemeProvider><ThemeToggle /></ThemeProvider>)
    expect(screen.getByRole('button', { name: 'Ativar modo claro' }).getAttribute('aria-pressed')).toBe('true')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })
})

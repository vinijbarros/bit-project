import { useTheme } from '../../features/theme/ThemeContext'

export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme()
  const darkModeEnabled = theme === 'dark'
  const label = darkModeEnabled ? 'Ativar modo claro' : 'Ativar modo noturno'

  return (
    <button
      className="theme-toggle"
      type="button"
      aria-label={label}
      aria-pressed={darkModeEnabled}
      onClick={toggleTheme}
    >
      <span className="theme-toggle__icon" aria-hidden="true">{darkModeEnabled ? '☀' : '☾'}</span>
      <span>{darkModeEnabled ? 'Modo claro' : 'Modo noturno'}</span>
    </button>
  )
}

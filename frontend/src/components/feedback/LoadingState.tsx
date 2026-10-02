export function LoadingState({ label = 'Carregando…', fullPage = false }: { label?: string; fullPage?: boolean }) {
  return <div className={`loading-state${fullPage ? ' loading-state--full' : ''}`} role="status" aria-live="polite"><span className="spinner" aria-hidden="true" />{label}</div>
}

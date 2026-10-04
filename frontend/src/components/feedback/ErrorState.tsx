interface ErrorStateProps {
  title: string
  message: string
  onRetry?: () => void | Promise<void>
  retryLabel?: string
  fullPage?: boolean
}

export function ErrorState({ title, message, onRetry, retryLabel = 'Tentar novamente', fullPage }: ErrorStateProps) {
  return (
    <section className={`state-card${fullPage ? ' state-card--full' : ''}`} role="alert">
      <span className="eyebrow">Não foi possível continuar</span>
      <h1>{title}</h1>
      <p>{message}</p>
      {onRetry && <button className="button" type="button" onClick={() => void onRetry()}>{retryLabel}</button>}
    </section>
  )
}

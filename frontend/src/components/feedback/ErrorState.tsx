interface ErrorStateProps {
  title: string
  message: string
  onRetry?: () => void | Promise<void>
  fullPage?: boolean
}

export function ErrorState({ title, message, onRetry, fullPage }: ErrorStateProps) {
  return (
    <section className={`state-card${fullPage ? ' state-card--full' : ''}`} role="alert">
      <span className="eyebrow">Não foi possível continuar</span>
      <h1>{title}</h1>
      <p>{message}</p>
      {onRetry && <button className="button" type="button" onClick={() => void onRetry()}>Tentar novamente</button>}
    </section>
  )
}

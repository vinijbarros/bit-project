export function safeInternalDestination(value: unknown): string {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//')) return '/dashboard'
  try {
    const destination = new URL(value, window.location.origin)
    if (destination.origin !== window.location.origin) return '/dashboard'
    const allowed = destination.pathname === '/dashboard'
      || destination.pathname === '/solicitacoes'
      || destination.pathname.startsWith('/solicitacoes/')
    return allowed ? `${destination.pathname}${destination.search}${destination.hash}` : '/dashboard'
  } catch {
    return '/dashboard'
  }
}

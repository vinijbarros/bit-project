import type { ReactNode } from 'react'

export function AlertMessage({ children, tone = 'info' }: { children: ReactNode; tone?: 'info' | 'success' | 'error' }) {
  return <div className={`alert alert--${tone}`} role={tone === 'error' ? 'alert' : 'status'}>{children}</div>
}

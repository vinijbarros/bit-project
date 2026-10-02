import type { ReactNode } from 'react'

export function EmptyState({ title, children, action }: { title: string; children: ReactNode; action?: ReactNode }) {
  return <section className="state-card"><h2>{title}</h2><div>{children}</div>{action}</section>
}

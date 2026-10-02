import type { ReactNode } from 'react'

export function PageHeader({ eyebrow, title, description, actions, titleId }: { eyebrow?: string; title: string; description?: string; actions?: ReactNode; titleId?: string }) {
  return (
    <header className="page-header">
      <div>
        {eyebrow && <span className="eyebrow">{eyebrow}</span>}
        <h1 id={titleId}>{title}</h1>
        {description && <p>{description}</p>}
      </div>
      {actions && <div className="page-actions">{actions}</div>}
    </header>
  )
}

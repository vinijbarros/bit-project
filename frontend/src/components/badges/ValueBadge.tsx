import type { Category, LabeledValue, Status } from '../../types/api'

export function StatusBadge({ status }: { status: LabeledValue<Status> }) {
  return <span className={`badge badge--status-${status.value}`}>{status.label}</span>
}

export function CategoryBadge({ category }: { category: LabeledValue<Category> }) {
  return <span className="badge badge--category">{category.label}</span>
}

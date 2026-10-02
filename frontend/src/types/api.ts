export type Category = 'ti' | 'rh' | 'compras' | 'financeiro' | 'infraestrutura'

export type Status = 'aberto' | 'em_atendimento' | 'concluido'

export interface User {
  id: number
  username: string
  display_name: string
}

export interface LabeledValue<T extends string = string> {
  value: T
  label: string
}

export interface MetadataOption<T extends string> {
  value: T
  label: string
}

export interface Metadata {
  categories: MetadataOption<Category>[]
  statuses: MetadataOption<Status>[]
}

export interface RequestPermissions {
  can_edit: boolean
  can_delete: boolean
}

export interface Request {
  id: number
  code: string
  title: string
  description: string
  category: LabeledValue<Category>
  status: LabeledValue<Status>
  requester: User
  created_at: string
  updated_at: string
  permissions: RequestPermissions
}

export interface RequestListItem {
  id: number
  code: string
  title: string
  category: LabeledValue<Category>
  status: LabeledValue<Status>
  requester: User
  created_at: string
}

export interface Pagination {
  page: number
  page_size: number
  total_items: number
  total_pages: number
}

export interface PaginatedRequests {
  items: RequestListItem[]
  pagination: Pagination
}

export interface Dashboard {
  total: number
  abertas: number
  em_atendimento: number
  concluidas: number
}

export type FieldErrors = Record<string, string[]>

export interface ApiErrorPayload {
  error: {
    code: string
    message: string
    fields?: FieldErrors
  }
  request_id?: string
}

export interface DataResponse<T> {
  data: T
}

export interface LoginInput {
  username: string
  password: string
}

export interface CreateRequestInput {
  title: string
  description: string
  category: Category
}

export interface UpdateRequestInput {
  title?: string
  description?: string
  category?: Category
}

export interface RequestFilters {
  date_from?: string
  date_to?: string
  category?: Category
  status?: Status
  q?: string
  page?: number
  page_size?: number
}

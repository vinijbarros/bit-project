import { apiClient } from '../api/client'
import type {
  CreateRequestInput,
  DataResponse,
  PaginatedRequests,
  Request,
  RequestFilters,
  Status,
  UpdateRequestInput,
} from '../types/api'

function requestQuery(filters: RequestFilters): string {
  const query = new URLSearchParams()
  for (const [name, value] of Object.entries(filters)) {
    if (value !== undefined && value !== '') query.set(name, String(value))
  }
  const encoded = query.toString()
  return encoded ? `?${encoded}` : ''
}

export const requestsService = {
  list(filters: RequestFilters = {}, signal?: AbortSignal): Promise<PaginatedRequests> {
    return apiClient.request<PaginatedRequests>(`/requests${requestQuery(filters)}`, { signal })
  },

  async get(id: number, signal?: AbortSignal): Promise<Request> {
    const response = await apiClient.request<DataResponse<Request>>(`/requests/${id}`, { signal })
    return response.data
  },

  async create(input: CreateRequestInput, signal?: AbortSignal): Promise<Request> {
    const response = await apiClient.request<DataResponse<Request>>('/requests', {
      method: 'POST',
      json: input,
      signal,
    })
    return response.data
  },

  async update(id: number, input: UpdateRequestInput, signal?: AbortSignal): Promise<Request> {
    const response = await apiClient.request<DataResponse<Request>>(`/requests/${id}`, {
      method: 'PATCH',
      json: input,
      signal,
    })
    return response.data
  },

  remove(id: number, signal?: AbortSignal): Promise<void> {
    return apiClient.request<void>(`/requests/${id}`, { method: 'DELETE', signal })
  },

  async updateStatus(id: number, status: Status, signal?: AbortSignal): Promise<Request> {
    const response = await apiClient.request<DataResponse<Request>>(`/requests/${id}/status`, {
      method: 'PATCH',
      json: { status },
      signal,
    })
    return response.data
  },
}

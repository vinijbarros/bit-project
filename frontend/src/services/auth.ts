import { apiClient } from '../api/client'
import type { DataResponse, LoginInput, User } from '../types/api'

export const authService = {
  async login(input: LoginInput, signal?: AbortSignal): Promise<User> {
    const response = await apiClient.request<DataResponse<User>>('/auth/login', {
      method: 'POST',
      json: input,
      signal,
      handleUnauthorized: false,
    })
    return response.data
  },

  async currentUser(signal?: AbortSignal): Promise<User> {
    const response = await apiClient.request<DataResponse<User>>('/auth/me', { signal, handleUnauthorized: false })
    return response.data
  },

  logout(signal?: AbortSignal): Promise<void> {
    return apiClient.request<void>('/auth/logout', { method: 'POST', signal })
  },
}

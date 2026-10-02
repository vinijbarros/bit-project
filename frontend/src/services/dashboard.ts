import { apiClient } from '../api/client'
import type { Dashboard, DataResponse } from '../types/api'

export async function getDashboard(signal?: AbortSignal): Promise<Dashboard> {
  const response = await apiClient.request<DataResponse<Dashboard>>('/dashboard', { signal })
  return response.data
}

import { apiClient } from '../api/client'
import type { DataResponse, Metadata } from '../types/api'

export async function getMetadata(signal?: AbortSignal): Promise<Metadata> {
  const response = await apiClient.request<DataResponse<Metadata>>('/metadata', { signal })
  return response.data
}

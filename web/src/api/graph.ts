import { apiClient } from './client'
import type { GraphData } from './types'

export async function getGraph(): Promise<GraphData> {
  return apiClient.get<GraphData>('/graph')
}

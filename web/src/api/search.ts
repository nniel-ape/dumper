import { apiClient } from './client'
import type { SearchResult, AskResponse } from './types'

export async function search(query: string): Promise<SearchResult[]> {
  return apiClient.get<SearchResult[]>(`/search?q=${encodeURIComponent(query)}`)
}

export async function ask(question: string): Promise<AskResponse> {
  return apiClient.post<AskResponse>('/ask', { question })
}

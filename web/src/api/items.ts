import { apiClient } from './client'
import type { Item, Stats } from './types'

export interface ListItemsParams {
  limit?: number
  offset?: number
  tag?: string
}

export async function listItems(params: ListItemsParams = {}): Promise<Item[]> {
  const searchParams = new URLSearchParams()
  if (params.limit) searchParams.set('limit', String(params.limit))
  if (params.offset) searchParams.set('offset', String(params.offset))
  if (params.tag) searchParams.set('tag', params.tag)

  const query = searchParams.toString()
  return apiClient.get<Item[]>(`/items${query ? `?${query}` : ''}`)
}

export async function getItem(id: string): Promise<Item> {
  return apiClient.get<Item>(`/items/${id}`)
}

export async function deleteItem(id: string): Promise<void> {
  return apiClient.delete(`/items/${id}`)
}

export async function getTags(): Promise<string[]> {
  return apiClient.get<string[]>('/tags')
}

export async function getStats(): Promise<Stats> {
  return apiClient.get<Stats>('/stats')
}

export function getExportUrl(): string {
  return apiClient.getExportUrl()
}

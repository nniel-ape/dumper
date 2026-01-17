export type ItemType = 'link' | 'note' | 'image' | 'search'

export interface Item {
  id: string
  type: ItemType
  url?: string
  title: string
  content?: string
  summary?: string
  image_path?: string
  tags: string[]
  created_at: string
  updated_at: string
}

export interface Relationship {
  id: number
  source_id: string
  target_id: string
  relation_type: string
  strength: number
}

export interface SearchResult {
  item: Item
  snippet?: string
  score: number
}

export interface GraphData {
  nodes: Item[]
  edges: Relationship[]
}

export interface Stats {
  items: number
  tags: number
}

export interface AskResponse {
  answer: string
  sources: SearchResult[]
}

export interface ApiError {
  error: string
}

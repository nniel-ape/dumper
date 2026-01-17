import { useQuery, useMutation } from '@tanstack/react-query'
import { search, ask } from '@/api'

const SEARCH_KEY = 'search'

export function useSearch(query: string) {
  return useQuery({
    queryKey: [SEARCH_KEY, query],
    queryFn: () => search(query),
    enabled: query.length > 0,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

export function useAsk() {
  return useMutation({
    mutationFn: ask,
  })
}

// Helper to detect question-like queries
export function isQuestion(query: string): boolean {
  const q = query.toLowerCase().trim()
  return (
    q.includes('?') ||
    q.startsWith('what ') ||
    q.startsWith('how ') ||
    q.startsWith('why ') ||
    q.startsWith('when ') ||
    q.startsWith('where ') ||
    q.startsWith('who ') ||
    q.startsWith('which ') ||
    q.startsWith('can ') ||
    q.startsWith('do ') ||
    q.startsWith('does ') ||
    q.startsWith('is ') ||
    q.startsWith('are ')
  )
}

import { useQuery } from '@tanstack/react-query'
import { getGraph } from '@/api'

const GRAPH_KEY = 'graph'

export function useGraph() {
  return useQuery({
    queryKey: [GRAPH_KEY],
    queryFn: getGraph,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

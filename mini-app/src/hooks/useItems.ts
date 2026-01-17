import {
  useQuery,
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from '@tanstack/react-query'
import {
  listItems,
  getItem,
  deleteItem,
  getTags,
  getStats,
} from '@/api'

const ITEMS_KEY = 'items'
const TAGS_KEY = 'tags'
const STATS_KEY = 'stats'
const PAGE_SIZE = 20

export function useItems(tag?: string) {
  return useInfiniteQuery({
    queryKey: [ITEMS_KEY, { tag }],
    queryFn: async ({ pageParam = 0 }) => {
      const result = await listItems({ limit: PAGE_SIZE, offset: pageParam, tag })
      return result ?? [] // Handle null response
    },
    getNextPageParam: (lastPage, allPages) => {
      if (!lastPage || lastPage.length < PAGE_SIZE) return undefined
      return allPages.length * PAGE_SIZE
    },
    initialPageParam: 0,
  })
}

export function useItem(id: string) {
  return useQuery({
    queryKey: [ITEMS_KEY, id],
    queryFn: () => getItem(id),
    enabled: !!id,
  })
}

export function useDeleteItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [ITEMS_KEY] })
      queryClient.invalidateQueries({ queryKey: [STATS_KEY] })
    },
  })
}

export function useTags() {
  return useQuery({
    queryKey: [TAGS_KEY],
    queryFn: getTags,
  })
}

export function useStats() {
  return useQuery({
    queryKey: [STATS_KEY],
    queryFn: getStats,
  })
}

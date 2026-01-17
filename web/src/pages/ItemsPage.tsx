import { useState, useRef, useCallback } from 'react'
import { useItems } from '@/hooks'
import { ItemCard } from '@/components/ItemCard'
import { ItemsFeedSkeleton } from '@/components/LoadingSkeleton'
import { EmptyState } from '@/components/EmptyState'
import { ErrorState } from '@/components/ErrorState'
import { hapticFeedback } from '@/lib/telegram'
import type { Item } from '@/api'

interface ItemsPageProps {
  filterTag?: string
  onTagClick?: (tag: string) => void
  onItemSelect?: (item: Item) => void
}

export function ItemsPage({ filterTag, onTagClick, onItemSelect }: ItemsPageProps) {
  const [isRefreshing, setIsRefreshing] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const touchStartY = useRef(0)

  const {
    data,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    refetch,
  } = useItems(filterTag)

  const items = data?.pages.flat() ?? []

  // Infinite scroll
  const handleScroll = useCallback(() => {
    if (!containerRef.current || isFetchingNextPage || !hasNextPage) return

    const { scrollTop, scrollHeight, clientHeight } = containerRef.current
    if (scrollHeight - scrollTop - clientHeight < 200) {
      fetchNextPage()
    }
  }, [fetchNextPage, hasNextPage, isFetchingNextPage])

  // Pull to refresh
  const handleTouchStart = (e: React.TouchEvent) => {
    touchStartY.current = e.touches[0].clientY
  }

  const handleTouchEnd = async (e: React.TouchEvent) => {
    const touchEndY = e.changedTouches[0].clientY
    const pullDistance = touchEndY - touchStartY.current

    if (
      containerRef.current?.scrollTop === 0 &&
      pullDistance > 100 &&
      !isRefreshing
    ) {
      setIsRefreshing(true)
      hapticFeedback('medium')
      await refetch()
      setIsRefreshing(false)
      hapticFeedback('success')
    }
  }

  const handleItemClick = (item: Item) => {
    hapticFeedback('light')
    onItemSelect?.(item)
  }

  if (isLoading) {
    return <ItemsFeedSkeleton />
  }

  if (isError) {
    return (
      <ErrorState
        message={error?.message || 'Failed to load items'}
        onRetry={() => refetch()}
      />
    )
  }

  if (items.length === 0) {
    return <EmptyState type="items" />
  }

  return (
    <div
      ref={containerRef}
      className="h-full overflow-y-auto"
      onScroll={handleScroll}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      {/* Pull to refresh indicator */}
      {isRefreshing && (
        <div className="flex justify-center py-4">
          <div className="animate-spin rounded-full h-6 w-6 border-2 border-tg-button border-t-transparent" />
        </div>
      )}

      {/* Filter indicator */}
      {filterTag && (
        <div className="px-4 py-2 bg-tg-secondary-bg border-b border-tg-hint/20 flex items-center justify-between">
          <span className="text-sm">
            Filtered by: <strong>{filterTag}</strong>
          </span>
          <button
            onClick={() => onTagClick?.('')}
            className="text-sm text-tg-link"
          >
            Clear
          </button>
        </div>
      )}

      {/* Items list */}
      {items.map((item) => (
        <ItemCard
          key={item.id}
          item={item}
          onClick={() => handleItemClick(item)}
          onTagClick={onTagClick}
        />
      ))}

      {/* Loading more indicator */}
      {isFetchingNextPage && (
        <div className="py-4">
          <ItemsFeedSkeleton count={2} />
        </div>
      )}

      {/* End of list */}
      {!hasNextPage && items.length > 0 && (
        <p className="text-center text-sm text-tg-hint py-8">
          No more items
        </p>
      )}
    </div>
  )
}

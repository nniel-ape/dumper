import { Skeleton } from '@/components/ui/skeleton'

export function ItemCardSkeleton() {
  return (
    <div className="mx-4 my-3 p-6 card-base">
      {/* Type badge + date row */}
      <div className="flex items-center justify-between mb-3">
        <Skeleton className="h-5 w-14 rounded-md" />
        <Skeleton className="h-4 w-16" />
      </div>
      {/* Title */}
      <Skeleton className="h-6 w-3/4 mb-2" />
      <Skeleton className="h-6 w-1/2 mb-3" />
      {/* Summary */}
      <Skeleton className="h-4 w-full mb-1" />
      <Skeleton className="h-4 w-2/3 mb-4" />
      {/* Tags */}
      <div className="flex gap-2">
        <Skeleton className="h-5 w-16 rounded-full" />
        <Skeleton className="h-5 w-20 rounded-full" />
      </div>
    </div>
  )
}

export function ItemsFeedSkeleton({ count = 5 }: { count?: number }) {
  return (
    <div className="pt-2">
      {Array.from({ length: count }).map((_, i) => (
        <ItemCardSkeleton key={i} />
      ))}
    </div>
  )
}

export function GraphSkeleton() {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <div className="rounded-full bg-accent-muted p-3 mx-auto mb-3 animate-pulse">
          <div className="h-6 w-6 rounded-full bg-accent/30" />
        </div>
        <Skeleton className="h-4 w-24 mx-auto" />
      </div>
    </div>
  )
}

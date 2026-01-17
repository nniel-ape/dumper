import { Skeleton } from '@/components/ui/skeleton'

export function ItemCardSkeleton() {
  return (
    <div className="mx-4 my-2 p-4 glass-card">
      <div className="flex items-start gap-3">
        <Skeleton className="h-8 w-8 rounded-lg" />
        <div className="flex-1">
          <Skeleton className="h-5 w-3/4 mb-2" />
          <Skeleton className="h-4 w-full mb-2" />
          <Skeleton className="h-4 w-2/3 mb-3" />
          <div className="flex gap-2">
            <Skeleton className="h-5 w-16 rounded-full" />
            <Skeleton className="h-5 w-20 rounded-full" />
          </div>
        </div>
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

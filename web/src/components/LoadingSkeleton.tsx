import { Skeleton } from '@/components/ui/skeleton'

export function ItemCardSkeleton() {
  return (
    <div className="p-4 border-b border-tg-hint/20">
      <Skeleton className="h-5 w-3/4 mb-2" />
      <Skeleton className="h-4 w-full mb-2" />
      <Skeleton className="h-4 w-2/3 mb-3" />
      <div className="flex gap-2">
        <Skeleton className="h-5 w-16 rounded-full" />
        <Skeleton className="h-5 w-20 rounded-full" />
      </div>
    </div>
  )
}

export function ItemsFeedSkeleton({ count = 5 }: { count?: number }) {
  return (
    <div>
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
        <Skeleton className="h-8 w-8 rounded-full mx-auto mb-2" />
        <Skeleton className="h-4 w-24" />
      </div>
    </div>
  )
}

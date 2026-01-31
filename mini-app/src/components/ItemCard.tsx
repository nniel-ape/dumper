import { TagPill } from './TagPill'
import { cn } from '@/lib/utils'
import type { Item } from '@/api'

interface ItemCardProps {
  item: Item
  onClick?: () => void
  onTagClick?: (tag: string) => void
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays < 7) return `${diffDays} days ago`

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
  })
}

export function ItemCard({ item, onClick, onTagClick }: ItemCardProps) {
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (onClick && (e.key === 'Enter' || e.key === ' ')) {
          onClick()
        }
      }}
      className={cn(
        'mx-4 my-3 p-6 card-base transition-all duration-150',
        'hover:shadow-md hover:border-accent active:bg-surface-elevated',
        onClick && 'cursor-pointer'
      )}
    >
      {/* Type badge + date row */}
      <div className="flex items-center justify-between mb-3">
        <span className="rounded-md bg-foreground text-background px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider font-mono">
          {item.type}
        </span>
        <span className="text-xs text-text-muted font-mono font-medium tracking-tight tabular-nums">
          {formatDate(item.created_at)}
        </span>
      </div>

      {/* Title */}
      <h3 className="font-bold text-lg leading-snug mb-1 line-clamp-2 text-foreground">
        {item.title || 'Untitled'}
      </h3>

      {/* Summary */}
      {item.summary && (
        <p className="text-sm text-muted-foreground leading-relaxed line-clamp-2 mb-3">
          {item.summary}
        </p>
      )}

      {/* Tags */}
      {item.tags.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          {item.tags.slice(0, 3).map((tag) => (
            <TagPill
              key={tag}
              tag={tag}
              onClick={
                onTagClick
                  ? (e) => {
                      e?.stopPropagation()
                      onTagClick(tag)
                    }
                  : undefined
              }
            />
          ))}
          {item.tags.length > 3 && (
            <span className="text-xs text-muted-foreground">
              +{item.tags.length - 3}
            </span>
          )}
        </div>
      )}
    </div>
  )
}

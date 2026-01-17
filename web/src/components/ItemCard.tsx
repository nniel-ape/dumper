import { Link2, FileText, Image, Search } from 'lucide-react'
import { TagPill } from './TagPill'
import { cn } from '@/lib/utils'
import type { Item } from '@/api'

interface ItemCardProps {
  item: Item
  onClick?: () => void
  onTagClick?: (tag: string) => void
}

const typeIcons: Record<string, React.ElementType> = {
  link: Link2,
  note: FileText,
  image: Image,
  search: Search,
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
  const Icon = typeIcons[item.type] || FileText

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
        'p-4 border-b border-tg-hint/20 active:bg-tg-secondary-bg transition-colors',
        onClick && 'cursor-pointer'
      )}
    >
      <div className="flex items-start gap-3">
        <div className="rounded-lg bg-tg-secondary-bg p-2 shrink-0">
          <Icon className="h-4 w-4 text-tg-hint" />
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="font-medium text-sm leading-tight mb-1 line-clamp-2">
            {item.title || 'Untitled'}
          </h3>
          {item.summary && (
            <p className="text-xs text-tg-hint line-clamp-2 mb-2">
              {item.summary}
            </p>
          )}
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
              <span className="text-xs text-tg-hint">
                +{item.tags.length - 3}
              </span>
            )}
            <span className="text-xs text-tg-hint ml-auto">
              {formatDate(item.created_at)}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

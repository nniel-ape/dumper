import { cn } from '@/lib/utils'

interface TagPillProps {
  tag: string
  onClick?: (e?: React.MouseEvent) => void
  className?: string
}

// Generate consistent color from tag name
function tagToColor(tag: string): string {
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = tag.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 70%, 85%)`
}

function tagToTextColor(tag: string): string {
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = tag.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 70%, 25%)`
}

export function TagPill({ tag, onClick, className }: TagPillProps) {
  return (
    <span
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onClick={onClick}
      onKeyDown={(e) => {
        if (onClick && (e.key === 'Enter' || e.key === ' ')) {
          onClick()
        }
      }}
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        onClick && 'cursor-pointer hover:opacity-80 active:opacity-60',
        className
      )}
      style={{
        backgroundColor: tagToColor(tag),
        color: tagToTextColor(tag),
      }}
    >
      {tag}
    </span>
  )
}

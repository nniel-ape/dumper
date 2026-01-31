import { cn } from '@/lib/utils'

interface TagPillProps {
  tag: string
  onClick?: (e?: React.MouseEvent) => void
  className?: string
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
        'bg-accent-muted text-accent border border-accent/25',
        'transition-colors duration-150',
        onClick && 'cursor-pointer hover:bg-accent/15',
        className
      )}
    >
      {tag}
    </span>
  )
}

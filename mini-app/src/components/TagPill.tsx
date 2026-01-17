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
        'bg-accent-muted text-accent backdrop-blur-sm',
        'border border-accent/20',
        'transition-all duration-200',
        onClick && 'cursor-pointer hover:bg-accent/20 hover:border-accent/30 active:scale-95',
        className
      )}
    >
      {tag}
    </span>
  )
}

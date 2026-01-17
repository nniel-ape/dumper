import { Inbox, Search, Share2 } from 'lucide-react'

type EmptyStateType = 'items' | 'search' | 'graph'

interface EmptyStateProps {
  type: EmptyStateType
  query?: string
}

const configs: Record<EmptyStateType, { icon: React.ElementType; title: string; description: string }> = {
  items: {
    icon: Inbox,
    title: 'No items yet',
    description: 'Forward links or send text to the bot to start building your knowledge vault.',
  },
  search: {
    icon: Search,
    title: 'No results found',
    description: 'Try a different search term or ask a question.',
  },
  graph: {
    icon: Share2,
    title: 'No connections yet',
    description: 'Save more items to discover relationships between them.',
  },
}

export function EmptyState({ type, query }: EmptyStateProps) {
  const config = configs[type]
  const Icon = config.icon

  return (
    <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
      <div className="rounded-full bg-tg-secondary-bg p-4 mb-4">
        <Icon className="h-8 w-8 text-tg-hint" />
      </div>
      <h3 className="text-lg font-medium mb-1">{config.title}</h3>
      <p className="text-sm text-tg-hint max-w-xs">
        {type === 'search' && query
          ? `No results for "${query}"`
          : config.description}
      </p>
    </div>
  )
}

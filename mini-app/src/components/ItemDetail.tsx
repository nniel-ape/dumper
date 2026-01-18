import { useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { ArrowLeft, ExternalLink, Trash2, Link2, FileText, Image, Search } from 'lucide-react'
import { TagPill } from './TagPill'
import { openLink, hapticFeedback, backButton } from '@/lib/telegram'
import { useDeleteItem } from '@/hooks'
import type { Item } from '@/api'

interface ItemDetailProps {
  item: Item
  onBack: () => void
  onTagClick?: (tag: string) => void
}

const typeIcons: Record<string, React.ElementType> = {
  link: Link2,
  note: FileText,
  image: Image,
  search: Search,
}

function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function ItemDetail({ item, onBack, onTagClick }: ItemDetailProps) {
  const deleteItem = useDeleteItem()
  const Icon = typeIcons[item.type] || FileText

  // Handle browser/Telegram back button
  useEffect(() => {
    // Push a state so we can intercept back
    window.history.pushState({ itemDetail: true }, '')

    const handlePopState = () => {
      onBack()
    }

    window.addEventListener('popstate', handlePopState)

    // Telegram back button (modern SDK)
    let cleanup: (() => void) | undefined
    try {
      if (backButton.isMounted() && backButton.show.isAvailable()) {
        backButton.show()
        cleanup = backButton.onClick(onBack)
      }
    } catch {
      // Not in Telegram
    }

    return () => {
      window.removeEventListener('popstate', handlePopState)
      try {
        cleanup?.()
        if (backButton.isMounted() && backButton.hide.isAvailable()) {
          backButton.hide()
        }
      } catch {
        // Not in Telegram
      }
    }
  }, [onBack])

  const handleBack = () => {
    // Go back in history if we pushed state
    if (window.history.state?.itemDetail) {
      window.history.back()
    } else {
      onBack()
    }
  }

  const handleDelete = async () => {
    if (!confirm('Delete this item?')) return

    hapticFeedback('medium')
    await deleteItem.mutateAsync(item.id)
    hapticFeedback('success')
    handleBack()
  }

  const handleOpenLink = () => {
    if (item.url) {
      hapticFeedback('light')
      openLink(item.url)
    }
  }

  const handleTagClick = (tag: string) => {
    if (onTagClick) {
      handleBack()
      setTimeout(() => onTagClick(tag), 50)
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-background overflow-hidden">
      {/* Header */}
      <header className="px-4 py-3 border-b border-border flex items-center gap-3">
        <button
          onClick={handleBack}
          className="p-1 -ml-1 text-foreground hover:text-accent transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div className="flex-1 min-w-0 flex items-center gap-3">
          <div className="rounded-lg bg-accent-muted p-2 shrink-0">
            <Icon className="h-4 w-4 text-accent" />
          </div>
          <div className="flex-1 min-w-0">
            <h1 className="text-base font-semibold truncate text-foreground">
              {item.title || 'Untitled'}
            </h1>
            <p className="text-xs text-muted-foreground font-mono">
              {formatDateTime(item.created_at)}
            </p>
          </div>
        </div>
        <button
          onClick={handleDelete}
          disabled={deleteItem.isPending}
          className="p-2 text-muted-foreground hover:text-destructive transition-colors disabled:opacity-50"
        >
          <Trash2 className="h-5 w-5" />
        </button>
      </header>

      {/* Content */}
      <main className="absolute top-14 bottom-0 left-0 right-0 overflow-y-auto p-4 space-y-4">
        {/* Tags */}
        {item.tags && item.tags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {item.tags.map((tag) => (
              <TagPill
                key={tag}
                tag={tag}
                onClick={onTagClick ? () => handleTagClick(tag) : undefined}
              />
            ))}
          </div>
        )}

        {/* Summary */}
        {item.summary && (
          <section>
            <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              Summary
            </h2>
            <p className="text-sm leading-relaxed text-foreground">{item.summary}</p>
          </section>
        )}

        {/* Content */}
        {item.content && (
          <section>
            <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              Content
            </h2>
            <p className="text-sm leading-relaxed whitespace-pre-wrap text-foreground">
              {item.content}
            </p>
          </section>
        )}

        {/* URL */}
        {item.url && (
          <section>
            <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              Source
            </h2>
            <button
              onClick={handleOpenLink}
              className="text-sm text-accent hover:text-accent-light break-all text-left transition-colors"
            >
              {item.url}
            </button>
          </section>
        )}

        {/* Open Original Button */}
        {item.url && (
          <Button
            variant="gradient"
            className="w-full"
            onClick={handleOpenLink}
          >
            <ExternalLink className="h-4 w-4 mr-2" />
            Open Original
          </Button>
        )}
      </main>
    </div>
  )
}

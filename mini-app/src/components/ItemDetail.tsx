import { useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { ArrowLeft, ExternalLink, Trash2 } from 'lucide-react'
import { TagPill } from './TagPill'
import { openLink, hapticFeedback, backButton } from '@/lib/telegram'
import { useDeleteItem } from '@/hooks'
import type { Item } from '@/api'

interface ItemDetailProps {
  item: Item
  onBack: () => void
  onTagClick?: (tag: string) => void
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
    <div className="fixed inset-0 z-50 bg-background flex flex-col">
      {/* Header */}
      <header className="shrink-0 px-4 py-3 safe-area-top border-b border-border flex items-center gap-3">
        <button
          onClick={handleBack}
          className="p-2 -ml-2 text-foreground hover:text-accent transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div className="flex-1 min-w-0 flex items-center gap-3">
          <span className="rounded-md bg-foreground text-background px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider font-mono shrink-0">
            {item.type}
          </span>
          <div className="flex-1 min-w-0">
            <h1 className="text-base font-bold truncate text-foreground">
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
      <main className="flex-1 min-h-0 overflow-y-auto p-5 safe-area-bottom">
        {/* Images */}
        {item.image_paths && item.image_paths.length > 0 ? (
          <section className="space-y-3">
            {item.image_paths.map((_, index) => (
              <img
                key={index}
                src={`/api/items/${item.id}/images/${index}`}
                alt={`${item.title || 'Item image'} (${index + 1}/${item.image_paths!.length})`}
                className="w-full rounded-xl border-2 border-border"
                onError={(e) => {
                  e.currentTarget.style.display = 'none'
                }}
                loading="lazy"
              />
            ))}
          </section>
        ) : item.image_path ? (
          <section>
            <img
              src={`/api/items/${item.id}/image`}
              alt={item.title || 'Item image'}
              className="w-full rounded-xl border-2 border-border"
              onError={(e) => {
                e.currentTarget.style.display = 'none'
              }}
              loading="lazy"
            />
          </section>
        ) : null}

        {/* Tags */}
        {item.tags && item.tags.length > 0 && (
          <section className={item.image_path ? 'catalog-divider' : 'mt-0'}>
            <div className="flex flex-wrap gap-2">
              {item.tags.map((tag) => (
                <TagPill
                  key={tag}
                  tag={tag}
                  onClick={onTagClick ? () => handleTagClick(tag) : undefined}
                />
              ))}
            </div>
          </section>
        )}

        {/* Summary */}
        {item.summary && (
          <section className="catalog-divider">
            <h2 className="text-sm font-bold text-foreground mb-3">
              Summary
            </h2>
            <p className="text-base leading-relaxed text-foreground">{item.summary}</p>
          </section>
        )}

        {/* Content */}
        {item.content && (
          <section className="catalog-divider">
            <h2 className="text-sm font-bold text-foreground mb-3">
              Content
            </h2>
            <p className="text-sm leading-relaxed whitespace-pre-wrap text-foreground">
              {item.content}
            </p>
          </section>
        )}

        {/* URL */}
        {item.url && (
          <section className="catalog-divider">
            <h2 className="text-sm font-bold text-foreground mb-3">
              Source
            </h2>
            <button
              onClick={handleOpenLink}
              className="text-sm text-accent hover:text-accent-hover break-all text-left transition-colors"
            >
              {item.url}
            </button>
          </section>
        )}

        {/* Open Original Button */}
        {item.url && (
          <div className="catalog-divider">
            <Button
              variant="default"
              className="w-full"
              onClick={handleOpenLink}
            >
              <ExternalLink className="h-4 w-4 mr-2" />
              Open Original
            </Button>
          </div>
        )}
      </main>
    </div>
  )
}

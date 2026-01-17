import { useState, useEffect, useRef } from 'react'
import { Search, X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { useSearch, useAsk, isQuestion } from '@/hooks'
import { ItemCard } from '@/components/ItemCard'
import { AIAnswerCard } from '@/components/AIAnswerCard'
import { ItemsFeedSkeleton } from '@/components/LoadingSkeleton'
import { EmptyState } from '@/components/EmptyState'
import type { Item, AskResponse } from '@/api'

interface SearchPageProps {
  onTagClick?: (tag: string) => void
  onItemSelect?: (item: Item) => void
}

export function SearchPage({ onTagClick, onItemSelect }: SearchPageProps) {
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [aiResponse, setAiResponse] = useState<AskResponse | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const { data: searchResults, isLoading: isSearching } = useSearch(debouncedQuery)
  const askMutation = useAsk()

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query.trim())
    }, 300)
    return () => clearTimeout(timer)
  }, [query])

  // Trigger AI when query looks like a question
  useEffect(() => {
    if (debouncedQuery && isQuestion(debouncedQuery) && !askMutation.isPending) {
      setAiResponse(null)
      askMutation.mutate(debouncedQuery, {
        onSuccess: setAiResponse,
      })
    } else if (!isQuestion(debouncedQuery)) {
      setAiResponse(null)
    }
  }, [debouncedQuery])

  const handleClear = () => {
    setQuery('')
    setAiResponse(null)
    inputRef.current?.focus()
  }

  const showAI = isQuestion(debouncedQuery)
  const hasResults = searchResults && searchResults.length > 0
  const showEmpty = debouncedQuery && !isSearching && !hasResults && !askMutation.isPending

  return (
    <div className="flex flex-col h-full">
      {/* Search input */}
      <div className="p-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            ref={inputRef}
            type="text"
            placeholder="Search or ask a question..."
            value={query}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setQuery(e.target.value)}
            className="pl-10 pr-10"
          />
          {query && (
            <button
              onClick={handleClear}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* Results area */}
      <div className="flex-1 overflow-y-auto">
        {/* AI Answer */}
        {showAI && (askMutation.isPending || aiResponse) && (
          <div className="px-4 pt-2">
            <AIAnswerCard
              response={aiResponse || { answer: '', sources: [] }}
              isLoading={askMutation.isPending}
              onSourceClick={onItemSelect}
            />
          </div>
        )}

        {/* Search results */}
        {isSearching && <ItemsFeedSkeleton count={3} />}

        {hasResults && (
          <div className="pt-2">
            {!showAI && (
              <p className="px-4 py-2 text-xs text-muted-foreground font-medium">
                {searchResults.length} result{searchResults.length !== 1 ? 's' : ''}
              </p>
            )}
            {searchResults.map((result) => (
              <ItemCard
                key={result.item.id}
                item={result.item}
                onClick={() => onItemSelect?.(result.item)}
                onTagClick={onTagClick}
              />
            ))}
          </div>
        )}

        {/* Empty state */}
        {showEmpty && <EmptyState type="search" query={debouncedQuery} />}

        {/* Initial state */}
        {!debouncedQuery && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <div className="rounded-full bg-accent-muted p-4 mb-4">
              <Search className="h-8 w-8 text-accent" />
            </div>
            <p className="text-sm text-muted-foreground">
              Search your vault or ask a question
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

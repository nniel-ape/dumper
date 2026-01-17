import { useState } from 'react'
import { Sparkles, ChevronDown, ChevronUp } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { ItemCard } from './ItemCard'
import type { AskResponse, Item } from '@/api'

interface AIAnswerCardProps {
  response: AskResponse
  isLoading?: boolean
  onSourceClick?: (item: Item) => void
}

export function AIAnswerCard({ response, isLoading, onSourceClick }: AIAnswerCardProps) {
  const [showSources, setShowSources] = useState(false)

  if (isLoading) {
    return (
      <Card className="border-accent/30 bg-gradient-to-br from-accent/10 to-accent/5">
        <CardContent className="p-4">
          <div className="flex items-center gap-2 mb-3">
            <Sparkles className="h-4 w-4 text-accent animate-pulse" />
            <span className="text-sm font-semibold text-accent">Thinking...</span>
          </div>
          <div className="space-y-2">
            <div className="h-4 bg-accent/20 rounded animate-pulse" />
            <div className="h-4 bg-accent/20 rounded animate-pulse w-3/4" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="border-accent/30 bg-gradient-to-br from-accent/10 to-accent/5">
      <CardContent className="p-4">
        <div className="flex items-center gap-2 mb-3">
          <Sparkles className="h-4 w-4 text-accent" />
          <span className="text-sm font-semibold text-accent">AI Answer</span>
        </div>

        <p className="text-sm leading-relaxed mb-3 text-foreground">{response.answer}</p>

        {response.sources.length > 0 && (
          <div>
            <button
              onClick={() => setShowSources(!showSources)}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              {showSources ? (
                <ChevronUp className="h-3 w-3" />
              ) : (
                <ChevronDown className="h-3 w-3" />
              )}
              {response.sources.length} source{response.sources.length !== 1 ? 's' : ''}
            </button>

            {showSources && (
              <div className="mt-2 -mx-4 border-t border-glass-border">
                {response.sources.map((result) => (
                  <ItemCard
                    key={result.item.id}
                    item={result.item}
                    onClick={() => onSourceClick?.(result.item)}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

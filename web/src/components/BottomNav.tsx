import { Home, Search, Share2, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import { hapticFeedback } from '@/lib/telegram'

export type TabId = 'items' | 'search' | 'graph' | 'settings'

interface BottomNavProps {
  activeTab: TabId
  onTabChange: (tab: TabId) => void
}

const tabs: Array<{ id: TabId; label: string; icon: React.ElementType }> = [
  { id: 'items', label: 'Items', icon: Home },
  { id: 'search', label: 'Search', icon: Search },
  { id: 'graph', label: 'Graph', icon: Share2 },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export function BottomNav({ activeTab, onTabChange }: BottomNavProps) {
  const handleTabClick = (tabId: TabId) => {
    if (tabId !== activeTab) {
      hapticFeedback('light')
      onTabChange(tabId)
    }
  }

  return (
    <nav className="fixed bottom-0 left-0 right-0 bg-tg-bg border-t border-tg-hint/20 safe-area-bottom">
      <div className="flex justify-around items-center h-14">
        {tabs.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            onClick={() => handleTabClick(id)}
            className={cn(
              'flex flex-col items-center justify-center flex-1 h-full transition-colors',
              activeTab === id
                ? 'text-tg-button'
                : 'text-tg-hint active:text-tg-text'
            )}
          >
            <Icon className="h-5 w-5 mb-0.5" />
            <span className="text-[10px] font-medium">{label}</span>
          </button>
        ))}
      </div>
    </nav>
  )
}

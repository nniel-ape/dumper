import { useState, useCallback } from 'react'
import { BottomNav, type TabId } from '@/components/BottomNav'
import { ItemDetail } from '@/components/ItemDetail'
import { AuroraBackground } from '@/components/AuroraBackground'
import { ItemsPage } from '@/pages/ItemsPage'
import { SearchPage } from '@/pages/SearchPage'
import { GraphPage } from '@/pages/GraphPage'
import { SettingsPage } from '@/pages/SettingsPage'
import type { Item } from '@/api'

export default function App() {
  const [activeTab, setActiveTab] = useState<TabId>('items')
  const [filterTag, setFilterTag] = useState<string | undefined>()
  const [selectedItem, setSelectedItem] = useState<Item | null>(null)

  const handleTagClick = useCallback((tag: string) => {
    if (tag) {
      setFilterTag(tag)
      setActiveTab('items')
    } else {
      setFilterTag(undefined)
    }
  }, [])

  const handleItemSelect = useCallback((item: Item) => {
    setSelectedItem(item)
  }, [])

  const handleBack = useCallback(() => {
    setSelectedItem(null)
  }, [])

  // Show full-screen item detail
  if (selectedItem) {
    return (
      <ItemDetail
        item={selectedItem}
        onBack={handleBack}
        onTagClick={handleTagClick}
      />
    )
  }

  return (
    <div className="h-full flex flex-col bg-background text-foreground">
      <AuroraBackground />

      {/* Header */}
      <header className="shrink-0 px-4 py-3 glass-elevated border-b border-glass-border">
        <h1 className="text-lg font-bold bg-gradient-to-r from-indigo-500 to-violet-500 bg-clip-text text-transparent">
          Dumper
        </h1>
      </header>

      {/* Content */}
      <main className="flex-1 min-h-0 pb-14">
        {activeTab === 'items' && (
          <ItemsPage
            filterTag={filterTag}
            onTagClick={handleTagClick}
            onItemSelect={handleItemSelect}
          />
        )}
        {activeTab === 'search' && (
          <SearchPage
            onTagClick={handleTagClick}
            onItemSelect={handleItemSelect}
          />
        )}
        {activeTab === 'graph' && (
          <GraphPage
            onTagClick={handleTagClick}
            onItemSelect={handleItemSelect}
          />
        )}
        {activeTab === 'settings' && <SettingsPage />}
      </main>

      {/* Navigation */}
      <BottomNav activeTab={activeTab} onTabChange={setActiveTab} />
    </div>
  )
}

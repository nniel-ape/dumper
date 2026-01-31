import { useState, useCallback } from 'react'
import { BottomNav, type TabId } from '@/components/BottomNav'
import { ItemDetail } from '@/components/ItemDetail'
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
      {/* Header */}
      <header className="shrink-0 px-5 py-5 safe-area-top bg-surface-elevated border-b-2 border-border-subtle shadow-sm">
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          Dumper
        </h1>
      </header>

      {/* Content */}
      <main
        className="flex-1 min-h-0 pb-14 relative"
        style={{ height: '100%' }}  // iOS Safari fix: explicit height for absolute children
      >
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

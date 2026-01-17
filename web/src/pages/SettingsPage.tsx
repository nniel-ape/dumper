import { Download, FileText, Tag, ExternalLink, Info } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useStats, useTags } from '@/hooks'
import { getExportUrl } from '@/api'
import { hapticFeedback, openLink } from '@/lib/telegram'

export function SettingsPage() {
  const { data: stats, isLoading: statsLoading } = useStats()
  const { data: tags, isLoading: tagsLoading } = useTags()

  const handleExport = () => {
    hapticFeedback('medium')
    const url = getExportUrl()
    openLink(url)
  }

  return (
    <div className="p-4 space-y-4 overflow-y-auto h-full">
      {/* Stats */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">Your Vault</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-tg-secondary-bg p-2">
                <FileText className="h-5 w-5 text-tg-hint" />
              </div>
              <div>
                {statsLoading ? (
                  <Skeleton className="h-6 w-12" />
                ) : (
                  <p className="text-2xl font-semibold">{stats?.items ?? 0}</p>
                )}
                <p className="text-xs text-tg-hint">Items</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-tg-secondary-bg p-2">
                <Tag className="h-5 w-5 text-tg-hint" />
              </div>
              <div>
                {tagsLoading ? (
                  <Skeleton className="h-6 w-12" />
                ) : (
                  <p className="text-2xl font-semibold">{tags?.length ?? 0}</p>
                )}
                <p className="text-xs text-tg-hint">Tags</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Export */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">Export</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-tg-hint mb-3">
            Download your vault as Obsidian-compatible markdown files.
          </p>
          <Button onClick={handleExport} className="w-full">
            <Download className="h-4 w-4 mr-2" />
            Export to Obsidian
          </Button>
        </CardContent>
      </Card>

      {/* About */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">About</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm text-tg-hint">Version</span>
            <span className="text-sm">1.0.0</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-tg-hint">Source</span>
            <button
              onClick={() => openLink('https://github.com/nerdneilsfield/dumper')}
              className="flex items-center gap-1 text-sm text-tg-link"
            >
              GitHub
              <ExternalLink className="h-3 w-3" />
            </button>
          </div>
        </CardContent>
      </Card>

      {/* Tips */}
      <Card>
        <CardContent className="py-4">
          <div className="flex gap-3">
            <Info className="h-5 w-5 text-tg-hint shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium mb-1">Pro tip</p>
              <p className="text-xs text-tg-hint">
                Forward links or send text messages to the bot to save them to your vault.
                The AI will automatically generate summaries and tags.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

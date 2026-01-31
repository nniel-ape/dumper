import { Download, ExternalLink, Sparkles, Sun, Moon, Monitor } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useStats, useTags, useTheme, type Theme } from '@/hooks'
import { getExportUrl } from '@/api'
import { hapticFeedback, openLink } from '@/lib/telegram'
import { cn } from '@/lib/utils'

const themeOptions: { value: Theme; label: string; icon: React.ElementType }[] = [
  { value: 'light', label: 'Light', icon: Sun },
  { value: 'dark', label: 'Dark', icon: Moon },
  { value: 'system', label: 'System', icon: Monitor },
]

export function SettingsPage() {
  const { data: stats, isLoading: statsLoading } = useStats()
  const { data: tags, isLoading: tagsLoading } = useTags()
  const { theme, setTheme } = useTheme()

  const handleExport = () => {
    hapticFeedback('medium')
    const url = getExportUrl()
    openLink(url)
  }

  const handleThemeChange = (newTheme: Theme) => {
    hapticFeedback('light')
    setTheme(newTheme)
  }

  return (
    <div className="p-4 space-y-5 overflow-y-auto h-full">
      {/* Stats */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-lg font-bold">Your Vault</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center py-4">
              {statsLoading ? (
                <Skeleton className="h-16 w-20 mx-auto" />
              ) : (
                <p className="text-6xl font-bold text-foreground font-mono tabular-nums leading-none">
                  {stats?.items ?? 0}
                </p>
              )}
              <p className="text-sm text-muted-foreground mt-2 font-medium">Items</p>
            </div>
            <div className="text-center py-4">
              {tagsLoading ? (
                <Skeleton className="h-16 w-20 mx-auto" />
              ) : (
                <p className="text-6xl font-bold text-foreground font-mono tabular-nums leading-none">
                  {tags?.length ?? 0}
                </p>
              )}
              <p className="text-sm text-muted-foreground mt-2 font-medium">Tags</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Theme */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-lg font-bold">Appearance</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            {themeOptions.map(({ value, label, icon: Icon }) => (
              <button
                key={value}
                onClick={() => handleThemeChange(value)}
                className={cn(
                  'flex-1 flex flex-col items-center gap-1.5 py-3 px-2 rounded-xl border transition-all duration-200',
                  theme === value
                    ? 'bg-accent-muted border-accent text-accent shadow-sm'
                    : 'bg-surface border-border-subtle text-muted-foreground hover:text-foreground hover:border-accent/30'
                )}
              >
                <Icon className="h-5 w-5" />
                <span className="text-xs font-medium">{label}</span>
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Export */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-lg font-bold">Export</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-3">
            Download your vault as Obsidian-compatible markdown files.
          </p>
          <Button variant="default" onClick={handleExport} className="w-full">
            <Download className="h-4 w-4 mr-2" />
            Export to Obsidian
          </Button>
        </CardContent>
      </Card>

      {/* About */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-lg font-bold">About</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Version</span>
            <span className="text-sm font-mono text-foreground">1.0.0</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Source</span>
            <button
              onClick={() => openLink('https://github.com/nerdneilsfield/dumper')}
              className="flex items-center gap-1 text-sm text-accent hover:text-accent-hover transition-colors"
            >
              GitHub
              <ExternalLink className="h-3 w-3" />
            </button>
          </div>
        </CardContent>
      </Card>

      {/* Tips */}
      <Card className="border-accent/20">
        <CardContent className="py-4">
          <div className="flex gap-3">
            <div className="rounded-lg bg-accent-muted p-2 h-fit">
              <Sparkles className="h-5 w-5 text-accent" />
            </div>
            <div>
              <p className="text-sm font-semibold mb-1 text-foreground">Pro tip</p>
              <p className="text-xs text-muted-foreground leading-relaxed">
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

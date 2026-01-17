import { memo, useMemo } from 'react'
import { Handle, Position } from '@xyflow/react'
import { Link2, FileText, Image, Search } from 'lucide-react'
import type { Item } from '@/api'

export interface ItemNodeData {
  item: Item
  connectionCount: number
}

interface ItemNodeProps {
  data: ItemNodeData
  selected?: boolean
}

// Type-based color schemes (defined outside component for performance)
const typeColors = {
  link: {
    bg: 'hsl(239 84% 74% / 0.15)',
    border: 'hsl(239 84% 74% / 0.4)',
    icon: 'hsl(239 84% 74%)',
  },
  note: {
    bg: 'hsl(142 76% 55% / 0.12)',
    border: 'hsl(142 76% 55% / 0.35)',
    icon: 'hsl(142 76% 55%)',
  },
  image: {
    bg: 'hsl(330 81% 60% / 0.12)',
    border: 'hsl(330 81% 60% / 0.35)',
    icon: 'hsl(330 81% 60%)',
  },
  search: {
    bg: 'hsl(258 90% 66% / 0.12)',
    border: 'hsl(258 90% 66% / 0.35)',
    icon: 'hsl(258 90% 66%)',
  },
} as const

// Map item types to icons
const typeIcons = {
  link: Link2,
  note: FileText,
  image: Image,
  search: Search,
} as const

type ItemType = keyof typeof typeColors

export const ItemNode = memo(function ItemNode({ data, selected }: ItemNodeProps) {
  const { item, connectionCount } = data

  // Determine item type with fallback
  const itemType: ItemType = useMemo(() => {
    const type = item.type?.toLowerCase() as ItemType
    return type in typeColors ? type : 'link'
  }, [item.type])

  // Size tier based on connections: leaf (0-2), connected (3-5), hub (6+)
  const sizeTier = useMemo(() => {
    if (connectionCount >= 6) return 'hub'
    if (connectionCount >= 3) return 'connected'
    return 'leaf'
  }, [connectionCount])

  const colors = typeColors[itemType]
  const IconComponent = typeIcons[itemType]

  // Width classes by tier
  const widthClass = {
    leaf: 'w-36',      // 144px
    connected: 'w-40', // 160px
    hub: 'w-48',       // 192px
  }[sizeTier]

  return (
    <>
      <Handle type="target" position={Position.Top} className="opacity-0" />
      <div
        className={`
          ${widthClass}
          rounded-2xl p-3 cursor-pointer
          transition-all duration-200
          hover:scale-105
          backdrop-blur-xl border shadow-lg
        `}
        style={{
          background: colors.bg,
          borderColor: colors.border,
          boxShadow: selected
            ? `0 0 0 2px ${colors.icon}, 0 8px 32px ${colors.bg}`
            : `0 4px 20px ${colors.bg}`,
        }}
      >
        <div className="flex items-start gap-2">
          {/* Type icon badge */}
          <div
            className="rounded-lg p-1.5 shrink-0"
            style={{ background: colors.border }}
          >
            <IconComponent
              className="h-3.5 w-3.5"
              style={{ color: colors.icon }}
            />
          </div>

          {/* Title */}
          <p className="text-xs font-medium leading-tight line-clamp-2 text-foreground min-w-0">
            {item.title || 'Untitled'}
          </p>
        </div>

        {/* Connection count for hub nodes */}
        {connectionCount >= 3 && (
          <div
            className="mt-2 text-[10px] font-mono"
            style={{ color: colors.icon }}
          >
            {connectionCount} links
          </div>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} className="opacity-0" />
    </>
  )
})

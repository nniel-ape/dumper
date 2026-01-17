import { memo } from 'react'
import { Handle, Position } from '@xyflow/react'
import type { Item } from '@/api'

export interface ItemNodeData {
  item: Item
  connectionCount: number
}

function tagToColor(tag: string): string {
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = tag.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 60%, 70%)`
}

interface ItemNodeProps {
  data: ItemNodeData
}

export const ItemNode = memo(function ItemNode({ data }: ItemNodeProps) {
  const { item, connectionCount } = data
  const primaryTag = item.tags[0]
  const bgColor = primaryTag ? tagToColor(primaryTag) : 'var(--tg-theme-secondary-bg-color)'

  // Size based on connection count
  const size = Math.min(80, 40 + connectionCount * 5)

  return (
    <>
      <Handle type="target" position={Position.Top} className="opacity-0" />
      <div
        className="rounded-lg p-2 shadow-md border border-white/20 cursor-pointer transition-transform hover:scale-105"
        style={{
          backgroundColor: bgColor,
          width: size,
          minHeight: size,
        }}
      >
        <p
          className="text-[10px] font-medium leading-tight text-center line-clamp-3"
          style={{ color: 'rgba(0,0,0,0.8)' }}
        >
          {item.title || 'Untitled'}
        </p>
      </div>
      <Handle type="source" position={Position.Bottom} className="opacity-0" />
    </>
  )
})

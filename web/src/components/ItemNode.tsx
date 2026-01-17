import { memo } from 'react'
import { Handle, Position } from '@xyflow/react'
import type { Item } from '@/api'

export interface ItemNodeData {
  item: Item
  connectionCount: number
}

interface ItemNodeProps {
  data: ItemNodeData
}

export const ItemNode = memo(function ItemNode({ data }: ItemNodeProps) {
  const { item, connectionCount } = data

  // Size based on connection count
  const size = Math.min(80, 40 + connectionCount * 5)

  return (
    <>
      <Handle type="target" position={Position.Top} className="opacity-0" />
      <div
        className="rounded-xl p-2 cursor-pointer transition-all duration-200 hover:scale-105 backdrop-blur-md border shadow-lg"
        style={{
          width: size,
          minHeight: size,
          background: 'hsl(var(--glass))',
          borderColor: 'hsl(var(--accent) / 0.3)',
          boxShadow: '0 4px 20px hsl(var(--accent) / 0.15)',
        }}
      >
        <p
          className="text-[10px] font-medium leading-tight text-center line-clamp-3 text-foreground"
        >
          {item.title || 'Untitled'}
        </p>
      </div>
      <Handle type="source" position={Position.Bottom} className="opacity-0" />
    </>
  )
})

import { useMemo, useCallback, useEffect } from 'react'
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  MarkerType,
  type Node,
  type Edge,
  type NodeTypes,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { useGraph } from '@/hooks'
import { ItemNode } from '@/components/ItemNode'
import { GraphSkeleton } from '@/components/LoadingSkeleton'
import { EmptyState } from '@/components/EmptyState'
import { ErrorState } from '@/components/ErrorState'
import type { Item } from '@/api'

const nodeTypes: NodeTypes = {
  item: ItemNode,
}

// Type-based colors for minimap (must match ItemNode colors)
const minimapTypeColors: Record<string, string> = {
  link: 'hsl(239 84% 74%)',
  note: 'hsl(142 76% 55%)',
  image: 'hsl(330 81% 60%)',
  search: 'hsl(258 90% 66%)',
}

// Default edge marker
const defaultMarker = {
  type: MarkerType.ArrowClosed,
  width: 16,
  height: 16,
  color: 'hsl(239 84% 74% / 0.6)',
}

interface GraphPageProps {
  onTagClick?: (tag: string) => void
  onItemSelect?: (item: Item) => void
}

export function GraphPage({ onItemSelect }: GraphPageProps) {
  const { data, isLoading, isError, error, refetch } = useGraph()

  // Convert graph data to React Flow format
  const { initialNodes, initialEdges } = useMemo(() => {
    if (!data) return { initialNodes: [], initialEdges: [] }

    const graphNodes = data.nodes ?? []
    const graphEdges = data.edges ?? []

    // Count connections per node
    const connectionCounts = new Map<string, number>()
    graphEdges.forEach((edge) => {
      connectionCounts.set(
        edge.source_id,
        (connectionCounts.get(edge.source_id) || 0) + 1
      )
      connectionCounts.set(
        edge.target_id,
        (connectionCounts.get(edge.target_id) || 0) + 1
      )
    })

    // Sort nodes by connection count (hubs first for better layout)
    const sortedNodes = [...graphNodes].sort((a, b) => {
      const countA = connectionCounts.get(a.id) || 0
      const countB = connectionCounts.get(b.id) || 0
      return countB - countA
    })

    // Position nodes in a grid with wider spacing
    const cols = Math.ceil(Math.sqrt(sortedNodes.length)) || 1
    const spacingX = 240
    const spacingY = 160

    const nodes: Node[] = sortedNodes.map((item, index) => ({
      id: item.id,
      type: 'item',
      position: {
        x: (index % cols) * spacingX,
        y: Math.floor(index / cols) * spacingY,
      },
      data: {
        item,
        connectionCount: connectionCounts.get(item.id) || 0,
      },
    }))

    const edges: Edge[] = graphEdges.map((rel) => ({
      id: `${rel.source_id}-${rel.target_id}`,
      source: rel.source_id,
      target: rel.target_id,
      type: 'smoothstep',
      label: rel.relation_type,
      animated: rel.strength > 0.7,
      markerEnd: defaultMarker,
      style: {
        strokeWidth: Math.max(1.5, rel.strength * 2.5),
        stroke: 'hsl(239 84% 74% / 0.6)',
      },
      pathOptions: {
        borderRadius: 20,
      },
    }))

    return { initialNodes: nodes, initialEdges: edges }
  }, [data])

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)

  // Update when data changes
  useEffect(() => {
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [initialNodes, initialEdges, setNodes, setEdges])

  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      const item = data?.nodes?.find((n) => n.id === node.id)
      if (item) onItemSelect?.(item)
    },
    [data, onItemSelect]
  )

  // MiniMap node color based on item type
  const getMinimapNodeColor = useCallback((node: Node) => {
    const item = node.data?.item as Item | undefined
    const type = item?.type?.toLowerCase() || 'link'
    return minimapTypeColors[type] || minimapTypeColors.link
  }, [])

  if (isLoading) {
    return <GraphSkeleton />
  }

  if (isError) {
    return (
      <ErrorState
        message={error?.message || 'Failed to load graph'}
        onRetry={() => refetch()}
      />
    )
  }

  if (!data || !data.nodes || data.nodes.length === 0) {
    return <EmptyState type="graph" />
  }

  return (
    <div className="absolute inset-0">
      {/* Aurora glow behind graph */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden z-0">
        <div className="absolute top-1/4 left-1/4 w-64 h-64 rounded-full aurora-orb-1 blur-3xl opacity-40" />
        <div className="absolute bottom-1/4 right-1/4 w-48 h-48 rounded-full aurora-orb-2 blur-3xl opacity-40" />
      </div>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        nodeTypes={nodeTypes}
        fitView
        minZoom={0.1}
        maxZoom={2}
        defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
      >
        <Background
          variant={BackgroundVariant.Dots}
          color="hsl(var(--muted-foreground) / 0.08)"
          gap={24}
          size={1}
        />
        <Controls
          showInteractive={false}
          className="react-flow-controls"
        />
        <MiniMap
          nodeColor={getMinimapNodeColor}
          className="react-flow-minimap"
          maskColor="hsl(var(--background) / 0.85)"
          pannable
          zoomable
        />
      </ReactFlow>
    </div>
  )
}

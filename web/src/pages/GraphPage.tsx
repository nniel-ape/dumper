import { useMemo, useCallback, useEffect } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
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

    // Position nodes in a grid (simple layout)
    const cols = Math.ceil(Math.sqrt(graphNodes.length)) || 1
    const spacing = 150

    const nodes: Node[] = graphNodes.map((item, index) => ({
      id: item.id,
      type: 'item',
      position: {
        x: (index % cols) * spacing + Math.random() * 30,
        y: Math.floor(index / cols) * spacing + Math.random() * 30,
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
      label: rel.relation_type,
      animated: rel.strength > 0.7,
      style: {
        strokeWidth: Math.max(1, rel.strength * 3),
        stroke: 'hsl(var(--accent) / 0.5)',
        opacity: 0.6,
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
    <div className="h-full w-full relative">
      {/* Aurora glow behind graph */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-1/4 left-1/4 w-64 h-64 rounded-full aurora-orb-1 blur-3xl opacity-50" />
        <div className="absolute bottom-1/4 right-1/4 w-48 h-48 rounded-full aurora-orb-2 blur-3xl opacity-50" />
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
        <Background color="hsl(var(--muted-foreground) / 0.2)" gap={20} />
        <Controls
          showInteractive={false}
          className="glass-card border-glass-border"
        />
        <MiniMap
          nodeColor={() => 'hsl(var(--accent))'}
          className="glass-card border-glass-border"
          maskColor="hsl(var(--background) / 0.8)"
        />
      </ReactFlow>
    </div>
  )
}

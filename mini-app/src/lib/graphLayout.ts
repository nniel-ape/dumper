import Dagre from '@dagrejs/dagre'
import type { Node, Edge } from '@xyflow/react'

interface LayoutOptions {
  direction?: 'TB' | 'BT' | 'LR' | 'RL'
  nodeWidth?: number
  nodeHeight?: number
  rankSep?: number
  nodeSep?: number
}

/**
 * Apply dagre hierarchical layout to React Flow nodes and edges.
 * Returns new arrays with updated node positions.
 */
export function getLayoutedElements(
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions = {}
): { nodes: Node[]; edges: Edge[] } {
  const {
    direction = 'TB',
    nodeWidth = 160,
    nodeHeight = 80,
    rankSep = 80,
    nodeSep = 40,
  } = options

  const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))

  g.setGraph({
    rankdir: direction,
    ranksep: rankSep,
    nodesep: nodeSep,
  })

  // Add nodes to dagre graph
  nodes.forEach((node) => {
    g.setNode(node.id, { width: nodeWidth, height: nodeHeight })
  })

  // Add edges to dagre graph
  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target)
  })

  // Run layout algorithm
  Dagre.layout(g)

  // Map dagre coordinates back to React Flow format
  // Dagre returns center coordinates, React Flow uses top-left
  const layoutedNodes = nodes.map((node) => {
    const dagreNode = g.node(node.id)
    return {
      ...node,
      position: {
        x: dagreNode.x - nodeWidth / 2,
        y: dagreNode.y - nodeHeight / 2,
      },
    }
  })

  return { nodes: layoutedNodes, edges }
}

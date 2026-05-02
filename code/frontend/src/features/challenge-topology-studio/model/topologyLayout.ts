import type { TopologyEditorDraft } from './topologyDraft'

export interface CanvasNodePosition {
  x: number
  y: number
}

export interface CanvasNode {
  key: string
  label: string
  tier: string
  isEntry: boolean
  injectFlag: boolean
  networks: string[]
  position: CanvasNodePosition
}

export interface CanvasEdge {
  id: string
  from: string
  to: string
  kind: 'link' | 'allow' | 'deny'
}

export interface TopologyCanvasGraph {
  nodes: CanvasNode[]
  edges: CanvasEdge[]
}

const CANVAS_WIDTH = 920
const CANVAS_HEIGHT = 600
const NODE_MIN_X = 52
const NODE_MAX_X = 868
const NODE_MIN_Y = 62
const NODE_MAX_Y = 540

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max)
}

function buildAutoPositions(nodeCount: number): CanvasNodePosition[] {
  if (nodeCount <= 1) {
    return [{ x: CANVAS_WIDTH / 2, y: CANVAS_HEIGHT / 2 }]
  }

  const columns = Math.min(3, Math.ceil(Math.sqrt(nodeCount)))
  const rows = Math.ceil(nodeCount / columns)
  const xStep = columns === 1 ? 0 : (NODE_MAX_X - NODE_MIN_X) / (columns - 1)
  const yStep = rows === 1 ? 0 : (NODE_MAX_Y - NODE_MIN_Y) / (rows - 1)

  return Array.from({ length: nodeCount }, (_, index) => {
    const column = index % columns
    const row = Math.floor(index / columns)
    return {
      x: clamp(NODE_MIN_X + column * xStep, NODE_MIN_X, NODE_MAX_X),
      y: clamp(NODE_MIN_Y + row * yStep, NODE_MIN_Y, NODE_MAX_Y),
    }
  })
}

export function buildTopologyCanvasGraph(
  draft: TopologyEditorDraft,
  positions: Record<string, CanvasNodePosition>
): TopologyCanvasGraph {
  const autoPositions = buildAutoPositions(draft.nodes.length)

  const nodes = draft.nodes.map((node, index) => ({
    key: node.key,
    label: node.name || node.key,
    tier: node.tier,
    isEntry: draft.entry_node_key === node.key,
    injectFlag: node.inject_flag,
    networks: node.network_keys,
    position: positions[node.key] || autoPositions[index],
  }))

  const edges: CanvasEdge[] = [
    ...draft.links.map((link, index) => ({
      id: `link-${index}-${link.from_node_key}-${link.to_node_key}`,
      from: link.from_node_key,
      to: link.to_node_key,
      kind: 'link' as const,
    })),
    ...draft.policies.map((policy, index) => ({
      id: `policy-${index}-${policy.source_node_key}-${policy.target_node_key}-${policy.action}`,
      from: policy.source_node_key,
      to: policy.target_node_key,
      kind: policy.action,
    })),
  ]

  return { nodes, edges }
}

export function normalizeCanvasPositions(
  draft: TopologyEditorDraft,
  current: Record<string, CanvasNodePosition>
): Record<string, CanvasNodePosition> {
  const autoPositions = buildAutoPositions(draft.nodes.length)
  const next: Record<string, CanvasNodePosition> = {}

  draft.nodes.forEach((node, index) => {
    next[node.key] = current[node.key] || autoPositions[index]
  })

  return next
}

export function clampCanvasPosition(position: CanvasNodePosition): CanvasNodePosition {
  return {
    x: clamp(position.x, NODE_MIN_X, NODE_MAX_X),
    y: clamp(position.y, NODE_MIN_Y, NODE_MAX_Y),
  }
}

<script setup lang="ts">
import { computed, onBeforeUnmount, useTemplateRef } from 'vue'

import type { CanvasNodePosition, TopologyCanvasGraph } from './topologyLayout'

interface CanvasEdgePath {
  id: string
  path: string
  kind: 'link' | 'allow' | 'deny'
}

export type CanvasInteractionMode = 'pan' | 'add-node' | 'link' | 'allow' | 'deny'

const props = defineProps<{
  graph: TopologyCanvasGraph
  selectedNodeKey?: string | null
  selectedEdgeId?: string | null
  interactionMode?: CanvasInteractionMode
  pendingSourceNodeKey?: string | null
}>()

const emit = defineEmits<{
  selectNode: [nodeKey: string]
  selectEdge: [edgeId: string]
  updatePosition: [payload: { nodeKey: string; position: CanvasNodePosition }]
  createNodeAt: [position: CanvasNodePosition]
  createEdge: [
    payload: { sourceNodeKey: string; targetNodeKey: string; kind: 'link' | 'allow' | 'deny' },
  ]
  clearPending: []
}>()

const svgRef = useTemplateRef<SVGSVGElement>('svgRef')

const nodeMap = computed(() =>
  props.graph.nodes.reduce<Record<string, (typeof props.graph.nodes)[number]>>((acc, node) => {
    acc[node.key] = node
    return acc
  }, {})
)

const edgePaths = computed<CanvasEdgePath[]>(() => {
  const items: CanvasEdgePath[] = []
  for (const edge of props.graph.edges) {
    const from = nodeMap.value[edge.from]
    const to = nodeMap.value[edge.to]
    if (!from || !to) {
      continue
    }

    const curveOffset = edge.kind === 'link' ? 0 : edge.kind === 'allow' ? -36 : 36
    const centerX = (from.position.x + to.position.x) / 2
    const centerY = (from.position.y + to.position.y) / 2 + curveOffset
    items.push({
      id: edge.id,
      kind: edge.kind,
      path: `M ${from.position.x} ${from.position.y} Q ${centerX} ${centerY} ${to.position.x} ${to.position.y}`,
    })
  }
  return items
})

let draggingNodeKey = ''
let pointerOffset = { x: 0, y: 0 }

function edgeStroke(kind: 'link' | 'allow' | 'deny') {
  if (kind === 'allow') {
    return '#3fb950'
  }
  if (kind === 'deny') {
    return '#f85149'
  }
  return 'rgba(148, 163, 184, 0.7)'
}

function startDrag(event: PointerEvent, nodeKey: string) {
  if (props.interactionMode && props.interactionMode !== 'pan') {
    return
  }
  const svg = svgRef.value
  const node = nodeMap.value[nodeKey]
  if (!svg || !node) {
    return
  }

  const rect = svg.getBoundingClientRect()
  draggingNodeKey = nodeKey
  pointerOffset = {
    x: event.clientX - rect.left - node.position.x,
    y: event.clientY - rect.top - node.position.y,
  }
  svg.setPointerCapture(event.pointerId)
  emit('selectNode', nodeKey)
}

function moveDrag(event: PointerEvent) {
  const svg = svgRef.value
  if (!svg || !draggingNodeKey) {
    return
  }

  const rect = svg.getBoundingClientRect()
  emit('updatePosition', {
    nodeKey: draggingNodeKey,
    position: {
      x: event.clientX - rect.left - pointerOffset.x,
      y: event.clientY - rect.top - pointerOffset.y,
    },
  })
}

function stopDrag(event?: PointerEvent) {
  if (svgRef.value && event && svgRef.value.hasPointerCapture(event.pointerId)) {
    svgRef.value.releasePointerCapture(event.pointerId)
  }
  draggingNodeKey = ''
}

function handleCanvasClick(event: MouseEvent) {
  if (props.interactionMode !== 'add-node') {
    return
  }
  const svg = svgRef.value
  if (!svg) {
    return
  }

  const pt = svg.createSVGPoint()
  pt.x = event.clientX
  pt.y = event.clientY

  // Transform screen coordinates to SVG coordinates
  const svgP = pt.matrixTransform(svg.getScreenCTM()?.inverse())

  emit('createNodeAt', {
    x: svgP.x,
    y: svgP.y,
  })
}

function handleNodeClick(nodeKey: string) {
  const mode = props.interactionMode || 'pan'
  if (mode === 'pan' || mode === 'add-node') {
    emit('selectNode', nodeKey)
    return
  }
  if (!props.pendingSourceNodeKey) {
    emit('selectNode', nodeKey)
    return
  }
  if (props.pendingSourceNodeKey === nodeKey) {
    emit('clearPending')
    return
  }
  emit('createEdge', {
    sourceNodeKey: props.pendingSourceNodeKey,
    targetNodeKey: nodeKey,
    kind: mode,
  })
}

onBeforeUnmount(() => {
  draggingNodeKey = ''
})
</script>

<template>
  <div
    class="topology-canvas-board__root overflow-hidden rounded-[28px] border border-border bg-[radial-gradient(circle_at_top,rgba(14,165,233,0.12),transparent_38%),linear-gradient(180deg,rgba(15,23,42,0.94),rgba(2,6,23,0.98))] p-4 shadow-[0_24px_60px_var(--color-shadow-soft)]"
  >
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-100/65">
          Graph Canvas
        </div>
        <div class="mt-2 text-lg font-semibold text-white">拖拽节点调整拓扑视图</div>
      </div>
      <div class="flex flex-wrap gap-2 text-xs">
        <span
          class="rounded-full border border-[var(--color-text-muted)]/30 bg-[var(--color-text-muted)]/10 px-3 py-1 text-[var(--color-text-primary)]"
        >
          灰线：逻辑连线
        </span>
        <span
          class="rounded-full border border-[var(--color-success)]/30 bg-[var(--color-success)]/10 px-3 py-1 text-[var(--color-success)]"
        >
          绿线：allow
        </span>
        <span class="rounded-full border border-[var(--color-danger)]/30 bg-[var(--color-danger)]/10 px-3 py-1 text-[var(--color-danger)]">
          红线：deny
        </span>
      </div>
    </div>

    <svg
      ref="svgRef"
      viewBox="0 0 920 600"
      class="topology-canvas-board__surface h-[600px] w-full rounded-[22px] border border-white/8 bg-[linear-gradient(180deg,rgba(15,23,42,0.88),rgba(2,6,23,0.96))]"
      @click="handleCanvasClick"
      @pointermove="moveDrag"
      @pointerup="stopDrag"
      @pointerleave="stopDrag"
    >
      <defs>
        <pattern id="topology-grid" width="36" height="36" patternUnits="userSpaceOnUse">
          <path
            d="M 36 0 L 0 0 0 36"
            fill="none"
            stroke="rgba(148,163,184,0.08)"
            stroke-width="1"
          />
        </pattern>
      </defs>

      <rect x="0" y="0" width="920" height="600" fill="url(#topology-grid)" />

      <path
        v-for="edge in edgePaths"
        :key="edge.id"
        :d="edge.path"
        fill="none"
        :stroke="edgeStroke(edge.kind)"
        :stroke-width="selectedEdgeId === edge.id ? 5 : edge.kind === 'link' ? 2 : 3"
        :stroke-dasharray="edge.kind === 'link' ? '6 6' : edge.kind === 'allow' ? '0' : '10 8'"
        stroke-linecap="round"
        :opacity="selectedEdgeId === edge.id ? '1' : '0.82'"
        class="cursor-pointer"
        @click.stop="emit('selectEdge', edge.id)"
      />

      <g
        v-for="node in graph.nodes"
        :key="node.key"
        class="cursor-grab active:cursor-grabbing"
        :transform="`translate(${node.position.x} ${node.position.y})`"
        @pointerdown="startDrag($event, node.key)"
        @click.stop="handleNodeClick(node.key)"
      >
        <circle
          r="52"
          :fill="selectedNodeKey === node.key ? 'rgba(8,145,178,0.28)' : 'rgba(15,23,42,0.95)'"
          :stroke="
            pendingSourceNodeKey === node.key
              ? '#fb923c'
              : selectedNodeKey === node.key
                ? '#38bdf8'
                : node.isEntry
                  ? '#f59e0b'
                  : 'rgba(148,163,184,0.45)'
          "
          stroke-width="3"
        />
        <circle
          r="40"
          :fill="
            node.tier === 'public'
              ? 'rgba(14,165,233,0.16)'
              : node.tier === 'internal'
                ? 'rgba(248,81,73,0.16)'
                : 'rgba(99,102,241,0.16)'
          "
          stroke="rgba(255,255,255,0.08)"
          stroke-width="1.5"
        />

        <text x="0" y="-8" text-anchor="middle" class="fill-white text-[14px] font-semibold">
          {{ node.label.slice(0, 14) }}
        </text>
        <text
          x="0"
          y="12"
          text-anchor="middle"
          class="fill-[var(--color-text-secondary)] text-[10px] uppercase tracking-[0.18em]"
        >
          {{ node.key }}
        </text>
        <text x="0" y="30" text-anchor="middle" class="fill-[var(--color-text-muted)] text-[9px]">
          {{ node.networks.join(', ') || 'no-network' }}
        </text>

        <g v-if="node.isEntry">
          <circle cx="38" cy="-36" r="10" fill="#f59e0b" />
          <text x="38" y="-32" text-anchor="middle" class="fill-[#0f172a] text-[9px] font-bold">
            IN
          </text>
        </g>
        <g v-if="node.injectFlag">
          <circle cx="-38" cy="-36" r="10" fill="#22c55e" />
          <text x="-38" y="-32" text-anchor="middle" class="fill-[#0f172a] text-[9px] font-bold">
            F
          </text>
        </g>
      </g>
    </svg>
  </div>
</template>

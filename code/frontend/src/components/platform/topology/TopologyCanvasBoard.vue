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
    return 'var(--topology-canvas-edge-allow)'
  }
  if (kind === 'deny') {
    return 'var(--topology-canvas-edge-deny)'
  }
  return 'var(--topology-canvas-edge-link)'
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
  <div class="topology-canvas-board__root">
    <div class="topology-canvas-board__head">
      <div>
        <div class="topology-canvas-board__eyebrow">
          Graph Canvas
        </div>
        <div class="topology-canvas-board__title">
          拖拽节点调整拓扑视图
        </div>
      </div>
      <div class="topology-canvas-board__legend">
        <span class="topology-canvas-board__legend-pill"> 灰线：逻辑连线 </span>
        <span class="topology-canvas-board__legend-pill topology-canvas-board__legend-pill--allow">
          绿线：allow
        </span>
        <span class="topology-canvas-board__legend-pill topology-canvas-board__legend-pill--deny">
          红线：deny
        </span>
      </div>
    </div>

    <svg
      ref="svgRef"
      viewBox="0 0 920 600"
      class="topology-canvas-board__surface"
      @click="handleCanvasClick"
      @pointermove="moveDrag"
      @pointerup="stopDrag"
      @pointerleave="stopDrag"
    >
      <defs>
        <pattern
          id="topology-grid"
          width="36"
          height="36"
          patternUnits="userSpaceOnUse"
        >
          <path
            d="M 36 0 L 0 0 0 36"
            fill="none"
            stroke="var(--topology-canvas-grid)"
            stroke-width="1"
          />
        </pattern>
      </defs>

      <rect
        x="0"
        y="0"
        width="920"
        height="600"
        fill="url(#topology-grid)"
      />

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
          :fill="
            selectedNodeKey === node.key
              ? 'var(--topology-canvas-node-selected)'
              : 'var(--topology-canvas-node-shell)'
          "
          :stroke="
            pendingSourceNodeKey === node.key
              ? 'var(--topology-canvas-node-pending)'
              : selectedNodeKey === node.key
                ? 'var(--topology-canvas-node-active)'
                : node.isEntry
                  ? 'var(--topology-canvas-node-entry)'
                  : 'var(--topology-canvas-node-ring)'
          "
          stroke-width="3"
        />
        <circle
          r="40"
          :fill="
            node.tier === 'public'
              ? 'var(--topology-canvas-tier-public)'
              : node.tier === 'internal'
                ? 'var(--topology-canvas-tier-internal)'
                : 'var(--topology-canvas-tier-service)'
          "
          stroke="var(--topology-canvas-node-core-ring)"
          stroke-width="1.5"
        />

        <text
          x="0"
          y="-8"
          text-anchor="middle"
          class="topology-canvas-board__node-label"
        >
          {{ node.label.slice(0, 14) }}
        </text>
        <text
          x="0"
          y="12"
          text-anchor="middle"
          class="topology-canvas-board__node-key"
        >
          {{ node.key }}
        </text>
        <text
          x="0"
          y="30"
          text-anchor="middle"
          class="topology-canvas-board__node-meta"
        >
          {{ node.networks.join(', ') || 'no-network' }}
        </text>

        <g v-if="node.isEntry">
          <circle
            cx="38"
            cy="-36"
            r="10"
            fill="var(--topology-canvas-node-entry)"
          />
          <text
            x="38"
            y="-32"
            text-anchor="middle"
            class="topology-canvas-board__node-badge-text"
          >
            IN
          </text>
        </g>
        <g v-if="node.injectFlag">
          <circle
            cx="-38"
            cy="-36"
            r="10"
            fill="var(--topology-canvas-edge-allow)"
          />
          <text
            x="-38"
            y="-32"
            text-anchor="middle"
            class="topology-canvas-board__node-badge-text"
          >
            F
          </text>
        </g>
      </g>
    </svg>
  </div>
</template>

<style scoped>
.topology-canvas-board__root {
  --topology-canvas-shell: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)),
    color-mix(
      in srgb,
      var(--journal-surface-subtle, var(--color-bg-elevated)) 94%,
      var(--color-bg-base)
    )
  );
  --topology-canvas-surface: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 94%, var(--color-bg-base)),
    color-mix(
      in srgb,
      var(--journal-surface-subtle, var(--color-bg-elevated)) 92%,
      var(--color-bg-base)
    )
  );
  --topology-canvas-border: color-mix(
    in srgb,
    var(--journal-border, var(--color-border-default)) 84%,
    transparent
  );
  --topology-canvas-grid: color-mix(in srgb, var(--color-text-secondary) 14%, transparent);
  --topology-canvas-edge-link: color-mix(in srgb, var(--color-text-secondary) 72%, transparent);
  --topology-canvas-edge-allow: var(--color-success);
  --topology-canvas-edge-deny: var(--color-danger);
  --topology-canvas-node-shell: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 88%,
    var(--color-bg-base)
  );
  --topology-canvas-node-selected: color-mix(
    in srgb,
    var(--color-primary) 20%,
    var(--topology-canvas-node-shell)
  );
  --topology-canvas-node-active: color-mix(
    in srgb,
    var(--color-primary) 74%,
    var(--topology-canvas-node-shell)
  );
  --topology-canvas-node-pending: color-mix(
    in srgb,
    var(--color-warning) 90%,
    var(--topology-canvas-node-shell)
  );
  --topology-canvas-node-entry: color-mix(
    in srgb,
    var(--color-warning) 90%,
    var(--topology-canvas-node-shell)
  );
  --topology-canvas-node-ring: color-mix(in srgb, var(--color-text-secondary) 44%, transparent);
  --topology-canvas-node-core-ring: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --topology-canvas-tier-public: color-mix(in srgb, var(--color-primary) 12%, transparent);
  --topology-canvas-tier-service: color-mix(in srgb, var(--color-text-secondary) 18%, transparent);
  --topology-canvas-tier-internal: color-mix(in srgb, var(--color-danger) 16%, transparent);
  overflow: hidden;
  border: 1px solid var(--topology-canvas-border);
  border-radius: 28px;
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 8%, transparent),
      transparent 20rem
    ),
    var(--topology-canvas-shell);
  padding: 1rem;
  box-shadow: 0 20px 44px var(--color-shadow-soft);
}

.topology-canvas-board__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.topology-canvas-board__eyebrow {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted, var(--color-text-secondary));
}

.topology-canvas-board__title {
  margin-top: 0.45rem;
  font-size: var(--font-size-1-08);
  font-weight: 700;
  color: var(--journal-ink, var(--color-text-primary));
}

.topology-canvas-board__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.topology-canvas-board__legend-pill {
  display: inline-flex;
  align-items: center;
  min-height: 1.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-text-secondary) 22%, transparent);
  background: color-mix(in srgb, var(--color-text-secondary) 8%, transparent);
  padding: 0 0.75rem;
  font-size: var(--font-size-0-76);
  color: var(--journal-ink, var(--color-text-primary));
}

.topology-canvas-board__legend-pill--allow {
  border-color: color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.topology-canvas-board__legend-pill--deny {
  border-color: color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: var(--color-danger);
}

.topology-canvas-board__surface {
  width: 100%;
  height: 600px;
  border: 1px solid var(--topology-canvas-border);
  border-radius: 22px;
  background: var(--topology-canvas-surface);
}

.topology-canvas-board__node-label {
  fill: var(--journal-ink, var(--color-text-primary));
  font-size: 14px;
  font-weight: 600;
}

.topology-canvas-board__node-key {
  fill: var(--journal-muted, var(--color-text-secondary));
  font-size: 10px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.topology-canvas-board__node-meta {
  fill: color-mix(in srgb, var(--journal-muted, var(--color-text-secondary)) 84%, transparent);
  font-size: 9px;
}

.topology-canvas-board__node-badge-text {
  fill: var(--color-bg-base);
  font-size: 9px;
  font-weight: 700;
}

@media (max-width: 767px) {
  .topology-canvas-board__surface {
    height: 520px;
  }
}
</style>

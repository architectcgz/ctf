<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'

import type { AdminImageListItem } from '@/api/contracts'
import SectionCard from '@/components/common/SectionCard.vue'
import TopologyCanvasQuickEditor from './TopologyCanvasQuickEditor.vue'
import TopologyNetworkQuickEditor from './TopologyNetworkQuickEditor.vue'

const TopologyCanvasBoard = defineAsyncComponent(() => import('./TopologyCanvasBoard.vue'))

type TopologyStudioVariant = 'template' | 'challenge'
type CanvasInteractionMode = 'pan' | 'add-node' | 'link' | 'allow' | 'deny'
type EdgeKind = 'link' | 'allow' | 'deny'
type SelectedNodeField = 'name' | 'image_id' | 'tier' | 'inject_flag'

type NodeOption = {
  key: string
  label: string
}

type CanvasPosition = {
  x: number
  y: number
}

type CanvasGraph = {
  nodes: Array<{
    key: string
    label: string
    tier: string
    isEntry: boolean
    injectFlag: boolean
    networks: string[]
    position: CanvasPosition
  }>
  edges: Array<{
    id: string
    from: string
    to: string
    kind: EdgeKind
  }>
}

type EditableNodeDraft = {
  name: string
  image_id: string
  service_port: number | null
  inject_flag: boolean
  tier: 'public' | 'service' | 'internal'
  network_keys: string[]
}

type EditableNetworkDraft = {
  uid: string
  key: string
  name: string
  internal: boolean
}

type NetworkPatch = Partial<Pick<EditableNetworkDraft, 'key' | 'name' | 'internal'>>

const props = withDefaults(
  defineProps<{
    variant?: TopologyStudioVariant
    interactionMode: CanvasInteractionMode
    canvasModeLabel: string
    selectedCanvasSummary: string
    draftValidationIssues: string[]
    canvasGraph: CanvasGraph
    pendingSourceNodeKey: string | null
    selectedNodeKey: string | null
    selectedEdgeId: string | null
    selectedNodeDraft: EditableNodeDraft | null
    hasSelectedEdge: boolean
    nodeOptions: NodeOption[]
    networks: EditableNetworkDraft[]
    images?: AdminImageListItem[]
    selectedEdgeSourceKey: string
    selectedEdgeTargetKey: string
    selectedEdgeKind: EdgeKind
  }>(),
  {
    variant: 'challenge',
    images: () => [],
  }
)

const emit = defineEmits<{
  setInteractionMode: [value: CanvasInteractionMode]
  removeSelectedCanvasItem: []
  selectNode: [nodeKey: string]
  selectEdge: [edgeId: string]
  updatePosition: [payload: { nodeKey: string; position: CanvasPosition }]
  createNodeAt: [position: CanvasPosition]
  createEdge: [payload: { sourceNodeKey: string; targetNodeKey: string; kind: EdgeKind }]
  clearPending: []
  updateSelectedNodeField: [payload: { field: SelectedNodeField; value: string | boolean }]
  updateSelectedNodeServicePort: [value: string]
  toggleSelectedNodeNetwork: [payload: { networkKey: string; checked: boolean }]
  updateSelectedEdgeSourceKey: [value: string]
  updateSelectedEdgeTargetKey: [value: string]
  updateSelectedEdgeKind: [value: string]
  updateNetwork: [payload: { uid: string; patch: NetworkPatch }]
}>()

const isTemplateVariant = computed(() => props.variant === 'template')
const addNodeLabel = computed(() => (isTemplateVariant.value ? '新增节点' : '画布新增节点'))
const allowLabel = computed(() => (isTemplateVariant.value ? 'allow' : 'allow 模式'))
const denyLabel = computed(() => (isTemplateVariant.value ? 'deny' : 'deny 模式'))
const removeSelectionLabel = computed(() =>
  isTemplateVariant.value ? '删除选中' : '删除当前选中'
)
const quickEditorLayoutClass = computed(() =>
  isTemplateVariant.value ? 'mt-4 grid gap-4' : 'mt-4 grid gap-4 xl:grid-cols-[1.08fr_0.92fr]'
)
const emptyMessage = computed(() =>
  isTemplateVariant.value ? '请在画布中选择一个节点或连线进行快速配置' : '请选择一个节点或一条边'
)
const escapeHint = computed(() =>
  isTemplateVariant.value ? '`Esc` 取消 / `Delete` 删除' : '`Esc` 取消连线 / `Delete` 删除选中'
)
</script>

<template>
  <SectionCard title="图形画布" subtitle="拖拽节点调整视图布局，点击节点可快速跳到对应节点编辑卡片。">
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="rounded-xl border px-3 py-2 text-sm font-medium transition"
        :class="
          props.interactionMode === 'pan'
            ? 'border-primary bg-primary/10 text-primary'
            : 'border-border text-text-primary hover:border-primary'
        "
        @click="emit('setInteractionMode', 'pan')"
      >
        浏览
      </button>
      <button
        type="button"
        class="rounded-xl border px-3 py-2 text-sm font-medium transition"
        :class="
          props.interactionMode === 'add-node'
            ? 'border-primary bg-primary/10 text-primary'
            : 'border-border text-text-primary hover:border-primary'
        "
        @click="emit('setInteractionMode', 'add-node')"
      >
        {{ addNodeLabel }}
      </button>
      <button
        type="button"
        class="rounded-xl border px-3 py-2 text-sm font-medium transition"
        :class="
          props.interactionMode === 'link'
            ? 'border-primary bg-primary/10 text-primary'
            : 'border-border text-text-primary hover:border-primary'
        "
        @click="emit('setInteractionMode', 'link')"
      >
        连线模式
      </button>
      <button
        type="button"
        class="rounded-xl border px-3 py-2 text-sm font-medium transition"
        :class="
          props.interactionMode === 'allow'
            ? 'topology-mode-btn--allow-active'
            : 'topology-mode-btn--allow-idle'
        "
        @click="emit('setInteractionMode', 'allow')"
      >
        {{ allowLabel }}
      </button>
      <button
        type="button"
        class="rounded-xl border px-3 py-2 text-sm font-medium transition"
        :class="
          props.interactionMode === 'deny'
            ? 'border-danger bg-danger/10 text-danger'
            : 'border-border text-text-primary hover:border-danger/60'
        "
        @click="emit('setInteractionMode', 'deny')"
      >
        {{ denyLabel }}
      </button>
      <button
        type="button"
        class="rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm font-medium text-danger transition hover:bg-danger/15"
        @click="emit('removeSelectedCanvasItem')"
      >
        {{ removeSelectionLabel }}
      </button>
    </div>

    <div
      class="mb-4 rounded-2xl border border-border bg-elevated px-4 py-3 text-sm text-text-secondary"
      :class="isTemplateVariant ? 'template-canvas-mode-banner' : ''"
    >
      <div class="flex flex-wrap items-center gap-2">
        <span
          class="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs text-primary"
        >
          当前模式：{{ props.canvasModeLabel }}
        </span>
        <span
          class="rounded-full border border-border-subtle bg-surface px-2.5 py-1 text-xs text-text-secondary"
        >
          {{ props.selectedCanvasSummary }}
        </span>
        <span
          class="rounded-full border border-border-subtle bg-surface px-2.5 py-1 text-xs text-text-muted"
        >
          {{ escapeHint }}
        </span>
      </div>
    </div>

    <div
      class="mb-4 rounded-2xl border px-4 py-3 text-sm"
      :class="
        props.draftValidationIssues.length === 0
          ? 'topology-validation-banner--ok'
          : 'topology-validation-banner--warn'
      "
    >
      <div class="font-medium">
        {{ props.draftValidationIssues.length === 0 ? '基础校验已通过' : '基础校验发现问题' }}
      </div>
      <div
        v-if="!isTemplateVariant && props.draftValidationIssues.length === 0"
        class="topology-validation-hint topology-validation-hint--success mt-1 text-xs"
      >
        当前草稿的入口、节点、网络和链路引用关系正常。
      </div>
      <ul v-else-if="props.draftValidationIssues.length > 0" class="mt-2 space-y-1 text-xs">
        <li v-for="issue in props.draftValidationIssues" :key="issue">
          {{ issue }}
        </li>
      </ul>
    </div>

    <TopologyCanvasBoard
      :graph="props.canvasGraph"
      :interaction-mode="props.interactionMode"
      :pending-source-node-key="props.pendingSourceNodeKey"
      :selected-node-key="props.selectedNodeKey"
      :selected-edge-id="props.selectedEdgeId"
      @select-node="emit('selectNode', $event)"
      @select-edge="emit('selectEdge', $event)"
      @create-node-at="emit('createNodeAt', $event)"
      @create-edge="emit('createEdge', $event)"
      @clear-pending="emit('clearPending')"
      @update-position="emit('updatePosition', $event)"
    />

    <div :class="quickEditorLayoutClass">
      <TopologyCanvasQuickEditor
        :variant="isTemplateVariant ? 'template' : 'challenge'"
        :empty-message="emptyMessage"
        :selected-node-draft="props.selectedNodeDraft"
        :has-selected-edge="props.hasSelectedEdge"
        :node-options="props.nodeOptions"
        :networks="props.networks"
        :images="props.images"
        :selected-edge-source-key="props.selectedEdgeSourceKey"
        :selected-edge-target-key="props.selectedEdgeTargetKey"
        :selected-edge-kind="props.selectedEdgeKind"
        @update-selected-node-field="emit('updateSelectedNodeField', $event)"
        @update-selected-node-service-port="emit('updateSelectedNodeServicePort', $event)"
        @toggle-selected-node-network="emit('toggleSelectedNodeNetwork', $event)"
        @update-selected-edge-source-key="emit('updateSelectedEdgeSourceKey', $event)"
        @update-selected-edge-target-key="emit('updateSelectedEdgeTargetKey', $event)"
        @update-selected-edge-kind="emit('updateSelectedEdgeKind', $event)"
      />

      <TopologyNetworkQuickEditor
        v-if="!isTemplateVariant"
        :networks="props.networks"
        @update-network="emit('updateNetwork', $event)"
      />
    </div>
  </SectionCard>
</template>

<style scoped>
.topology-mode-btn--allow-active {
  border-color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.topology-mode-btn--allow-idle {
  border-color: var(--color-border-default);
  color: var(--color-text-primary);
}

.topology-mode-btn--allow-idle:hover {
  border-color: color-mix(in srgb, var(--color-success) 60%, var(--color-border-default));
}

.topology-validation-banner--ok {
  border-color: color-mix(in srgb, var(--color-success) 20%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.topology-validation-banner--warn {
  border-color: color-mix(in srgb, var(--color-warning) 20%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: var(--color-warning);
}

.topology-validation-hint--success {
  color: color-mix(in srgb, var(--color-success) 80%, transparent);
}
</style>

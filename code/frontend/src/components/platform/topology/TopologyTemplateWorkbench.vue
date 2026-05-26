<script setup lang="ts">
import { Layout, Network, RefreshCw, Server, ShieldCheck } from 'lucide-vue-next'

import type { AdminImageListItem, EnvironmentTemplateData } from '@/api/contracts'
import type {
  TopologyLinkDraft,
  TopologyNetworkDraft,
  TopologyNodeDraft,
  TopologyPolicyDraft,
} from '@/features/challenge-topology-studio/model'
import TopologyCanvasWorkspaceSection from './TopologyCanvasWorkspaceSection.vue'
import TopologyConnectivitySections from './TopologyConnectivitySections.vue'
import TopologyEntryNodeSection from './TopologyEntryNodeSection.vue'
import TopologyNetworkSection from './TopologyNetworkSection.vue'
import TopologyNodeSection from './TopologyNodeSection.vue'
import TopologyTemplateSidePanel from './TopologyTemplateSidePanel.vue'

type WorkbenchTab = 'visual' | 'compute' | 'network' | 'policy'
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

const props = defineProps<{
  activeWorkbenchTab: WorkbenchTab
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
  networks: TopologyNetworkDraft[]
  selectedEdgeSourceKey: string
  selectedEdgeTargetKey: string
  selectedEdgeKind: EdgeKind
  entryNodeKey: string
  nodes: TopologyNodeDraft[]
  images: AdminImageListItem[]
  links: TopologyLinkDraft[]
  policies: TopologyPolicyDraft[]
  templateKeyword: string
  templateName: string
  templateDescription: string
  selectedTemplateSummary: string
  selectedTemplateId: string | null
  templates: EnvironmentTemplateData[]
  templateBusy: boolean
}>()

const emit = defineEmits<{
  'update:activeWorkbenchTab': [value: WorkbenchTab]
  refresh: []
  setInteractionMode: [value: CanvasInteractionMode]
  removeSelectedCanvasItem: []
  selectNode: [nodeKey: string]
  selectEdge: [edgeId: string]
  createNodeAt: [position: CanvasPosition]
  createEdge: [payload: { sourceNodeKey: string; targetNodeKey: string; kind: EdgeKind }]
  clearPending: []
  updatePosition: [payload: { nodeKey: string; position: CanvasPosition }]
  updateSelectedNodeField: [payload: { field: SelectedNodeField; value: string | boolean }]
  updateSelectedNodeServicePort: [value: string]
  toggleSelectedNodeNetwork: [payload: { networkKey: string; checked: boolean }]
  updateSelectedEdgeSourceKey: [value: string]
  updateSelectedEdgeTargetKey: [value: string]
  updateSelectedEdgeKind: [value: string]
  updateEntryNodeKey: [value: string]
  addNode: []
  removeNode: [uid: string]
  updateNode: [payload: { uid: string; node: TopologyNodeDraft }]
  addNetwork: []
  removeNetwork: [uid: string]
  updateNetwork: [
    payload: {
      uid: string
      patch: Partial<Pick<TopologyNetworkDraft, 'key' | 'name' | 'cidr' | 'internal'>>
    },
  ]
  addLink: []
  removeLink: [uid: string]
  updateLink: [
    payload: {
      uid: string
      patch: Partial<Pick<TopologyLinkDraft, 'from_node_key' | 'to_node_key'>>
    },
  ]
  addPolicy: []
  removePolicy: [uid: string]
  updatePolicy: [
    payload: {
      uid: string
      patch: Partial<Pick<TopologyPolicyDraft, 'source_node_key' | 'target_node_key' | 'action'>>
    },
  ]
  'update:templateKeyword': [value: string]
  'update:templateName': [value: string]
  'update:templateDescription': [value: string]
  loadTemplate: [template: EnvironmentTemplateData]
  clearTemplateSelection: []
  searchTemplates: []
  resetTemplateForm: [template: EnvironmentTemplateData]
  applyTemplate: [template: EnvironmentTemplateData]
  deleteTemplate: [templateId: string]
  resetTemplateEditor: []
  createTemplate: []
  updateTemplate: []
}>()

const workbenchTabs: Array<{ id: WorkbenchTab; label: string; icon: typeof Layout }> = [
  { id: 'visual', label: '画布', icon: Layout },
  { id: 'compute', label: '节点', icon: Server },
  { id: 'network', label: '网络', icon: Network },
  { id: 'policy', label: '策略', icon: ShieldCheck },
]
</script>

<template>
  <section class="topology-workbench grid gap-6 xl:grid-cols-[1fr_360px]">
    <div class="space-y-6">
      <div class="flex items-center gap-2">
        <div class="template-toolbar-tabs">
          <button
            v-for="tab in workbenchTabs"
            :key="tab.id"
            type="button"
            class="template-toolbar-tab"
            :class="
              props.activeWorkbenchTab === tab.id
                ? 'template-toolbar-tab--active'
                : 'template-toolbar-tab--idle'
            "
            @click="emit('update:activeWorkbenchTab', tab.id)"
          >
            <component :is="tab.icon" class="h-4 w-4" />
            <span class="hidden sm:inline">{{ tab.label }}</span>
          </button>
        </div>
        <button
          type="button"
          class="template-toolbar-refresh"
          title="刷新数据"
          @click="emit('refresh')"
        >
          <RefreshCw class="h-4 w-4" />
        </button>
      </div>

      <div v-if="props.activeWorkbenchTab === 'visual'" class="space-y-6">
        <TopologyCanvasWorkspaceSection
          variant="template"
          :interaction-mode="interactionMode"
          :canvas-mode-label="canvasModeLabel"
          :selected-canvas-summary="selectedCanvasSummary"
          :draft-validation-issues="draftValidationIssues"
          :canvas-graph="canvasGraph"
          :pending-source-node-key="pendingSourceNodeKey"
          :selected-node-key="selectedNodeKey"
          :selected-edge-id="selectedEdgeId"
          :selected-node-draft="selectedNodeDraft"
          :has-selected-edge="hasSelectedEdge"
          :node-options="nodeOptions"
          :networks="networks"
          :selected-edge-source-key="selectedEdgeSourceKey"
          :selected-edge-target-key="selectedEdgeTargetKey"
          :selected-edge-kind="selectedEdgeKind"
          @set-interaction-mode="emit('setInteractionMode', $event)"
          @remove-selected-canvas-item="emit('removeSelectedCanvasItem')"
          @select-node="emit('selectNode', $event)"
          @select-edge="emit('selectEdge', $event)"
          @create-node-at="emit('createNodeAt', $event)"
          @create-edge="emit('createEdge', $event)"
          @clear-pending="emit('clearPending')"
          @update-position="emit('updatePosition', $event)"
          @update-selected-node-field="emit('updateSelectedNodeField', $event)"
          @update-selected-node-service-port="emit('updateSelectedNodeServicePort', $event)"
          @toggle-selected-node-network="emit('toggleSelectedNodeNetwork', $event)"
          @update-selected-edge-source-key="emit('updateSelectedEdgeSourceKey', $event)"
          @update-selected-edge-target-key="emit('updateSelectedEdgeTargetKey', $event)"
          @update-selected-edge-kind="emit('updateSelectedEdgeKind', $event)"
        />
      </div>

      <div v-else-if="props.activeWorkbenchTab === 'compute'" class="space-y-6">
        <TopologyEntryNodeSection
          :entry-node-key="entryNodeKey"
          :node-options="nodeOptions"
          @update-entry-node-key="emit('updateEntryNodeKey', $event)"
        />

        <TopologyNodeSection
          :nodes="nodes"
          :images="images"
          :networks="networks"
          :selected-node-key="selectedNodeKey"
          add-button-class="topology-toolbar-btn topology-toolbar-btn--ghost"
          @add-node="emit('addNode')"
          @remove-node="emit('removeNode', $event)"
          @update-node="emit('updateNode', $event)"
        />
      </div>

      <div v-else-if="props.activeWorkbenchTab === 'network'" class="space-y-6">
        <TopologyNetworkSection
          :networks="networks"
          add-button-class="topology-toolbar-btn topology-toolbar-btn--ghost"
          @add-network="emit('addNetwork')"
          @remove-network="emit('removeNetwork', $event)"
          @update-network="emit('updateNetwork', $event)"
        />
      </div>

      <div v-else class="space-y-6">
        <TopologyConnectivitySections
          :links="links"
          :policies="policies"
          :node-options="nodeOptions"
          add-button-class="topology-toolbar-btn topology-toolbar-btn--ghost"
          @add-link="emit('addLink')"
          @remove-link="emit('removeLink', $event)"
          @update-link="emit('updateLink', $event)"
          @add-policy="emit('addPolicy')"
          @remove-policy="emit('removePolicy', $event)"
          @update-policy="emit('updatePolicy', $event)"
        />
      </div>
    </div>

    <TopologyTemplateSidePanel
      :template-keyword="templateKeyword"
      :template-name="templateName"
      :template-description="templateDescription"
      :is-template-library-mode="true"
      :selected-template-summary="selectedTemplateSummary"
      :selected-template-id="selectedTemplateId"
      :templates="templates"
      :template-busy="templateBusy"
      @update:template-keyword="emit('update:templateKeyword', $event)"
      @update:template-name="emit('update:templateName', $event)"
      @update:template-description="emit('update:templateDescription', $event)"
      @load-template="emit('loadTemplate', $event)"
      @clear-template-selection="emit('clearTemplateSelection')"
      @search-templates="emit('searchTemplates')"
      @reset-template-form="emit('resetTemplateForm', $event)"
      @apply-template="emit('applyTemplate', $event)"
      @delete-template="emit('deleteTemplate', $event)"
      @reset-template-editor="emit('resetTemplateEditor')"
      @create-template="emit('createTemplate')"
      @update-template="emit('updateTemplate')"
    />
  </section>
</template>

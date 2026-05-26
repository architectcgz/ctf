<script setup lang="ts">
import type {
  AdminImageListItem,
  EnvironmentTemplateData,
  TopologyPolicyAction,
  TopologyTier,
} from '@/api/contracts'
import TopologyCanvasWorkspaceSection from './TopologyCanvasWorkspaceSection.vue'
import TopologyChallengeContextRail from './TopologyChallengeContextRail.vue'
import TopologyConnectivitySections from './TopologyConnectivitySections.vue'
import TopologyEntryNodeSection from './TopologyEntryNodeSection.vue'
import TopologyNetworkSection from './TopologyNetworkSection.vue'
import TopologyNodeSection from './TopologyNodeSection.vue'

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
  uid: string
  key: string
  name: string
  image_id: string
  service_port: number | null
  inject_flag: boolean
  tier: TopologyTier
  network_keys: string[]
  env_entries: Array<{
    uid: string
    key: string
    value: string
  }>
  cpu_quota: number | null
  memory_mb: number | null
  pids_limit: number | null
}

type TopologyNetworkDraft = {
  uid: string
  key: string
  name: string
  cidr: string
  internal: boolean
}

type TopologyNodeDraft = EditableNodeDraft

type TopologyLinkDraft = {
  uid: string
  from_node_key: string
  to_node_key: string
}

type TopologyPolicyDraft = {
  uid: string
  source_node_key: string
  target_node_key: string
  action: TopologyPolicyAction
}

type TopologyStatusCard = {
  eyebrow: string
  title: string
  subtitle: string
}

type PackageSourceSummary = {
  title: string
  subtitle: string
  canExport: boolean
}

type PackageBaselineSummary = {
  entryNodeKey: string
  networkCount: number
  nodeCount: number
}

type PackageFile = {
  path: string
  size: number
}

type PackageRevision = {
  id: string
  revision_no: number
  source_type: 'imported' | 'exported'
  package_slug?: string
  topology_source_path?: string
  created_at: string
}

defineProps<{
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
  images: AdminImageListItem[]
  selectedEdgeSourceKey: string
  selectedEdgeTargetKey: string
  selectedEdgeKind: EdgeKind
  entryNodeKey: string
  nodes: TopologyNodeDraft[]
  links: TopologyLinkDraft[]
  policies: TopologyPolicyDraft[]
  saving: boolean
  hasSavedTopology: boolean
  statusCard: TopologyStatusCard
  secondaryCard: TopologyStatusCard
  packageSourceSummary: PackageSourceSummary
  packageBaselineSummary: PackageBaselineSummary | null
  packageFiles: PackageFile[]
  packageRevisionHistory: PackageRevision[]
  exporting: boolean
  selectedTemplateSummary: string
  selectedTemplateId: string | null
  templates: EnvironmentTemplateData[]
  templateKeyword: string
  templateName: string
  templateDescription: string
  templateBusy: boolean
}>()

const emit = defineEmits<{
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
  updateNetwork: [
    payload: {
      uid: string
      patch: Partial<Pick<TopologyNetworkDraft, 'key' | 'name' | 'cidr' | 'internal'>>
    },
  ]
  updateEntryNodeKey: [value: string]
  addNode: []
  removeNode: [uid: string]
  updateNode: [payload: { uid: string; node: TopologyNodeDraft }]
  addNetwork: []
  removeNetwork: [uid: string]
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
  deleteTopology: []
  exportPackage: []
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
</script>

<template>
  <main class="content-pane topology-workspace">
    <div class="topology-primary-column">
      <TopologyCanvasWorkspaceSection
        variant="challenge"
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
        :images="images"
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
        @update-network="emit('updateNetwork', $event)"
      />

      <TopologyEntryNodeSection
        :entry-node-key="entryNodeKey"
        :node-options="nodeOptions"
        :show-delete-action="true"
        :delete-disabled="saving || !hasSavedTopology"
        @update-entry-node-key="emit('updateEntryNodeKey', $event)"
        @delete-topology="emit('deleteTopology')"
      />

      <TopologyNetworkSection
        :networks="networks"
        add-button-class="ui-btn ui-btn--ghost topology-action-btn"
        @add-network="emit('addNetwork')"
        @remove-network="emit('removeNetwork', $event)"
        @update-network="emit('updateNetwork', $event)"
      />

      <TopologyNodeSection
        :nodes="nodes"
        :images="images"
        :networks="networks"
        :selected-node-key="selectedNodeKey"
        add-button-class="ui-btn ui-btn--ghost topology-action-btn"
        @add-node="emit('addNode')"
        @remove-node="emit('removeNode', $event)"
        @update-node="emit('updateNode', $event)"
      />

      <TopologyConnectivitySections
        :links="links"
        :policies="policies"
        :node-options="nodeOptions"
        add-button-class="ui-btn ui-btn--ghost topology-action-btn"
        @add-link="emit('addLink')"
        @remove-link="emit('removeLink', $event)"
        @update-link="emit('updateLink', $event)"
        @add-policy="emit('addPolicy')"
        @remove-policy="emit('removePolicy', $event)"
        @update-policy="emit('updatePolicy', $event)"
      />
    </div>

    <TopologyChallengeContextRail
      :template-keyword="templateKeyword"
      :template-name="templateName"
      :template-description="templateDescription"
      :status-card="statusCard"
      :secondary-card="secondaryCard"
      :package-source-summary="packageSourceSummary"
      :package-baseline-summary="packageBaselineSummary"
      :package-files="packageFiles"
      :package-revision-history="packageRevisionHistory"
      :exporting="exporting"
      :selected-template-summary="selectedTemplateSummary"
      :selected-template-id="selectedTemplateId"
      :templates="templates"
      :template-busy="templateBusy"
      @export-package="emit('exportPackage')"
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
  </main>
</template>

<style scoped>
.topology-primary-column {
  display: grid;
  gap: var(--space-6);
}

:deep(.topology-context-rail) {
  min-width: 0;
  padding-left: var(--space-6);
  border-left: 1px solid var(--topology-divider);
}

:deep(.topology-context-stack) {
  position: sticky;
  top: var(--space-6);
}

.topology-primary-column :deep(.section-card:first-child) {
  padding-top: 0;
  border-top: 0;
}

:deep(.topology-action-btn) {
  --ui-btn-height: 2.45rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-0-84);
  --ui-btn-secondary-border: var(--journal-border);
  --ui-btn-secondary-background: color-mix(
    in srgb,
    var(--journal-surface) 94%,
    var(--color-bg-base)
  );
  --ui-btn-secondary-color: var(--journal-ink);
  --ui-btn-secondary-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-secondary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-secondary-hover-color: var(--journal-accent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-primary-border: transparent;
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-color: var(--color-bg-base);
  --ui-btn-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 88%,
    var(--color-bg-base)
  );
  --ui-btn-primary-hover-shadow: 0 12px 28px
    color-mix(in srgb, var(--journal-accent) 16%, transparent);
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 28%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
  --ui-btn-danger-hover-border: color-mix(in srgb, var(--color-danger) 34%, transparent);
  --ui-btn-danger-hover-background: color-mix(
    in srgb,
    var(--color-danger) 14%,
    var(--journal-surface)
  );
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

:deep(.topology-action-btn:disabled) {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

:deep(.topology-action-btn--icon) {
  min-width: 2.75rem;
  padding-inline: var(--space-3);
}

@media (max-width: 1023px) {
  :deep(.topology-context-rail) {
    padding-left: 0;
    padding-top: var(--space-6);
    border-top: 1px solid var(--topology-divider);
    border-left: 0;
  }

  :deep(.topology-context-stack) {
    position: static;
  }
}
</style>

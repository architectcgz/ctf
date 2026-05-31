<script setup lang="ts">
import { computed, ref } from 'vue'

import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'
import TopologyChallengeWorkbench from './TopologyChallengeWorkbench.vue'
import TopologyChallengeWorkspaceHeader from './TopologyChallengeWorkspaceHeader.vue'
import TopologyTemplateHeroSection from './TopologyTemplateHeroSection.vue'
import TopologyTemplateLibraryHeader from './TopologyTemplateLibraryHeader.vue'
import TopologyTemplateWorkbench from './TopologyTemplateWorkbench.vue'
import {
  useChallengeTopologyStudioPage,
  type TopologyStudioMode,
} from '../model'

const props = withDefaults(
  defineProps<{
    challengeId?: string
    mode?: TopologyStudioMode
  }>(),
  {
    challengeId: '',
    mode: 'challenge',
  }
)

const emit = defineEmits<{
  back: []
}>()

const activeWorkbenchTab = ref<'visual' | 'compute' | 'network' | 'policy'>('visual')

const {
  loading,
  saving,
  exporting,
  templateBusy,
  challenge,
  topology,
  images,
  templates,
  templateKeyword,
  selectedTemplateId,
  templateName,
  templateDescription,
  draft,
  selectedNodeKey,
  selectedEdgeId,
  interactionMode,
  pendingSourceNodeKey,
  nodePositions,
  isTemplateLibraryMode,
  nodeOptions,
  canSaveTemplate,
  selectedTemplate,
  pageHeader,
  loadingText,
  heroEyebrow,
  heroTitle,
  heroDescription,
  statusCard,
  secondaryCard,
  packageBaselineSummary,
  packageFiles,
  packageRevisionHistory,
  packageSourceSummary,
  selectedCanvasSummary,
  draftValidationIssues,
  selectedTemplateSummary,
  topologySummary,
  canvasGraph,
  selectedNodeDraft,
  selectedEdgeMeta,
  selectedLinkDraft,
  selectedPolicyDraft,
  selectedEdgeSourceKey,
  selectedEdgeTargetKey,
  selectedEdgeKind,
  canvasModeLabel,
  updateCanvasQuickNumber,
  toggleSelectedNodeNetwork,
  updateSelectedEdgeSourceKey,
  updateSelectedEdgeTargetKey,
  handleSelectedEdgeKindChange,
  reloadAll,
  handleResetTemplateEditor,
  addNetwork,
  removeNetwork,
  addNode,
  removeNode,
  updateNetworkDraft,
  updateNodeDraft,
  updateSelectedNodeField,
  updateEntryNodeKey,
  updateNodePosition,
  setInteractionMode,
  handleCanvasSelectNode,
  handleCanvasSelectEdge,
  handleCanvasCreateNode,
  handleCanvasCreateEdge,
  removeSelectedCanvasItem,
  addLink,
  addPolicy,
  updateLinkDraft,
  removeLinkDraft,
  updatePolicyDraft,
  removePolicyDraft,
  loadTemplateIntoDraft,
  handleApplyTemplate,
  handleSaveTopology,
  handleExportPackage,
  handleDeleteTopology,
  handleCreateTemplate,
  handleUpdateTemplate,
  handleDeleteTemplate,
  clearTemplateSelection,
  loadTemplates,
  resetTemplateForm,
} = useChallengeTopologyStudioPage({
  challengeId: props.challengeId,
  mode: props.mode,
})

const rootClasses = computed(() => [
  'topology-page',
  isTemplateLibraryMode.value ? 'topology-page--template-library' : 'topology-page--challenge',
  'workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col',
])
</script>

<template>
  <div :class="rootClasses">
    <TopologyTemplateLibraryHeader
      v-if="isTemplateLibraryMode"
      :eyebrow="pageHeader.eyebrow"
      :title="pageHeader.title"
      :description="pageHeader.description"
      @reset="handleResetTemplateEditor"
      @refresh="void reloadAll()"
    />

    <TopologyChallengeWorkspaceHeader
      v-else
      :eyebrow="pageHeader.eyebrow"
      :title="heroTitle"
      :description="heroDescription"
      :summary="topologySummary"
      :exporting="exporting"
      :can-export="packageSourceSummary.canExport"
      :saving="saving"
      @back="emit('back')"
      @refresh="void reloadAll()"
      @export-package="void handleExportPackage()"
      @save="void handleSaveTopology()"
    />

    <div
      v-if="loading && !isTemplateLibraryMode"
      class="content-pane topology-loading-pane"
    >
      <AppLoading>{{ loadingText }}</AppLoading>
    </div>

    <section
      v-else-if="loading && isTemplateLibraryMode"
      class="content-pane template-library-main"
    >
      <div class="flex justify-center py-12">
        <AppLoading>{{ loadingText }}</AppLoading>
      </div>
    </section>

    <template v-else>
      <section
        v-if="isTemplateLibraryMode"
        class="content-pane template-library-main"
      >
        <TopologyTemplateHeroSection
          :hero-eyebrow="heroEyebrow"
          :hero-title="heroTitle"
          :hero-description="heroDescription"
          :topology-summary="topologySummary"
          :status-card="statusCard"
          :secondary-card="secondaryCard"
        />

        <div class="template-library-divider" />

        <TopologyTemplateWorkbench
          v-model:active-workbench-tab="activeWorkbenchTab"
          v-model:template-keyword="templateKeyword"
          v-model:template-name="templateName"
          v-model:template-description="templateDescription"
          :interaction-mode="interactionMode"
          :canvas-mode-label="canvasModeLabel"
          :selected-canvas-summary="selectedCanvasSummary"
          :draft-validation-issues="draftValidationIssues"
          :canvas-graph="canvasGraph"
          :pending-source-node-key="pendingSourceNodeKey"
          :selected-node-key="selectedNodeKey"
          :selected-edge-id="selectedEdgeId"
          :selected-node-draft="selectedNodeDraft"
          :has-selected-edge="Boolean(selectedEdgeMeta)"
          :node-options="nodeOptions"
          :networks="draft.networks"
          :selected-edge-source-key="selectedEdgeSourceKey"
          :selected-edge-target-key="selectedEdgeTargetKey"
          :selected-edge-kind="selectedEdgeKind"
          :entry-node-key="draft.entry_node_key"
          :nodes="draft.nodes"
          :images="images"
          :links="draft.links"
          :policies="draft.policies"
          :selected-template-summary="selectedTemplateSummary"
          :selected-template-id="selectedTemplateId"
          :templates="templates"
          :template-busy="templateBusy"
          @refresh="void reloadAll()"
          @set-interaction-mode="setInteractionMode"
          @remove-selected-canvas-item="removeSelectedCanvasItem"
          @select-node="handleCanvasSelectNode"
          @select-edge="handleCanvasSelectEdge"
          @create-node-at="handleCanvasCreateNode"
          @create-edge="handleCanvasCreateEdge"
          @clear-pending="pendingSourceNodeKey = null"
          @update-position="updateNodePosition"
          @update-selected-node-field="updateSelectedNodeField"
          @update-selected-node-service-port="
            updateCanvasQuickNumber('service_port', $event, selectedNodeDraft)
          "
          @toggle-selected-node-network="
            toggleSelectedNodeNetwork($event.networkKey, $event.checked)
          "
          @update-selected-edge-source-key="updateSelectedEdgeSourceKey"
          @update-selected-edge-target-key="updateSelectedEdgeTargetKey"
          @update-selected-edge-kind="handleSelectedEdgeKindChange"
          @update-entry-node-key="updateEntryNodeKey"
          @add-node="addNode"
          @remove-node="removeNode"
          @update-node="updateNodeDraft"
          @add-network="addNetwork"
          @remove-network="removeNetwork"
          @update-network="updateNetworkDraft"
          @add-link="addLink"
          @remove-link="removeLinkDraft"
          @update-link="updateLinkDraft"
          @add-policy="addPolicy"
          @remove-policy="removePolicyDraft"
          @update-policy="updatePolicyDraft"
          @load-template="loadTemplateIntoDraft"
          @clear-template-selection="clearTemplateSelection"
          @search-templates="void loadTemplates()"
          @reset-template-form="resetTemplateForm"
          @apply-template="(template) => void handleApplyTemplate(template)"
          @delete-template="(templateId) => void handleDeleteTemplate(templateId)"
          @reset-template-editor="handleResetTemplateEditor"
          @create-template="void handleCreateTemplate()"
          @update-template="void handleUpdateTemplate()"
        />
      </section>

      <template v-else>
        <div class="journal-divider" />

        <TopologyChallengeWorkbench
          v-model:template-keyword="templateKeyword"
          v-model:template-name="templateName"
          v-model:template-description="templateDescription"
          :interaction-mode="interactionMode"
          :canvas-mode-label="canvasModeLabel"
          :selected-canvas-summary="selectedCanvasSummary"
          :draft-validation-issues="draftValidationIssues"
          :canvas-graph="canvasGraph"
          :pending-source-node-key="pendingSourceNodeKey"
          :selected-node-key="selectedNodeKey"
          :selected-edge-id="selectedEdgeId"
          :selected-node-draft="selectedNodeDraft"
          :has-selected-edge="Boolean(selectedEdgeMeta)"
          :node-options="nodeOptions"
          :networks="draft.networks"
          :images="images"
          :selected-edge-source-key="selectedEdgeSourceKey"
          :selected-edge-target-key="selectedEdgeTargetKey"
          :selected-edge-kind="selectedEdgeKind"
          :entry-node-key="draft.entry_node_key"
          :nodes="draft.nodes"
          :links="draft.links"
          :policies="draft.policies"
          :saving="saving"
          :has-saved-topology="Boolean(topology)"
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
          @set-interaction-mode="setInteractionMode"
          @remove-selected-canvas-item="removeSelectedCanvasItem"
          @select-node="handleCanvasSelectNode"
          @select-edge="handleCanvasSelectEdge"
          @create-node-at="handleCanvasCreateNode"
          @create-edge="handleCanvasCreateEdge"
          @clear-pending="pendingSourceNodeKey = null"
          @update-position="updateNodePosition"
          @update-selected-node-field="updateSelectedNodeField"
          @update-selected-node-service-port="
            updateCanvasQuickNumber('service_port', $event, selectedNodeDraft)
          "
          @toggle-selected-node-network="
            toggleSelectedNodeNetwork($event.networkKey, $event.checked)
          "
          @update-selected-edge-source-key="updateSelectedEdgeSourceKey"
          @update-selected-edge-target-key="updateSelectedEdgeTargetKey"
          @update-selected-edge-kind="handleSelectedEdgeKindChange"
          @update-network="updateNetworkDraft"
          @update-entry-node-key="updateEntryNodeKey"
          @add-node="addNode"
          @remove-node="removeNode"
          @update-node="updateNodeDraft"
          @add-network="addNetwork"
          @remove-network="removeNetwork"
          @add-link="addLink"
          @remove-link="removeLinkDraft"
          @update-link="updateLinkDraft"
          @add-policy="addPolicy"
          @remove-policy="removePolicyDraft"
          @update-policy="updatePolicyDraft"
          @delete-topology="void handleDeleteTopology()"
          @export-package="void handleExportPackage()"
          @load-template="loadTemplateIntoDraft"
          @clear-template-selection="clearTemplateSelection"
          @search-templates="void loadTemplates()"
          @reset-template-form="resetTemplateForm"
          @apply-template="(template) => void handleApplyTemplate(template)"
          @delete-template="(templateId) => void handleDeleteTemplate(templateId)"
          @reset-template-editor="handleResetTemplateEditor"
          @create-template="void handleCreateTemplate()"
          @update-template="void handleUpdateTemplate()"
        />
      </template>

      <AppEmpty
        v-if="!challenge && !isTemplateLibraryMode"
        title="题目不存在"
        description="无法读取当前题目的基础信息，请返回题目列表后重试。"
        icon="Blocks"
      />
    </template>
  </div>
</template>

<style scoped>
.topology-page--challenge {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 80%, var(--color-bg-base));
  --journal-accent: color-mix(in srgb, var(--color-primary) 88%, var(--journal-ink));
  --topology-panel: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --topology-panel-subtle: color-mix(
    in srgb,
    var(--journal-surface-subtle) 96%,
    var(--color-bg-base)
  );
  --topology-divider: color-mix(in srgb, var(--journal-border) 88%, transparent);
  display: flex;
  flex-direction: column;
  min-height: max(100%, calc(100vh - 5rem));
  padding: var(--space-6) var(--space-7);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 8%, transparent),
      transparent 22rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
    );
}

.topology-page--template-library {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  display: grid;
  gap: var(--space-5);
  min-height: max(100%, calc(100vh - 5rem));
  padding: var(--space-6) var(--space-7);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
}

.topology-loading-pane {
  display: flex;
  justify-content: center;
  padding-block: var(--space-10);
}

.topology-page--template-library .template-library-main {
  border-color: var(--journal-border);
  border-radius: 0;
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 7%, transparent),
      transparent 22rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.topology-page--template-library .template-library-divider {
  margin: var(--space-6) 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

:global([data-theme='dark']) .topology-page--template-library {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary-hover);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

:global([data-theme='dark']) .topology-page--template-library .template-library-main {
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 10%, transparent),
      transparent 18rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
}

@media (max-width: 767px) {
  .topology-page--challenge {
    padding: var(--space-5);
  }
}
</style>

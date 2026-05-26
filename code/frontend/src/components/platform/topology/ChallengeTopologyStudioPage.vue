<script setup lang="ts">
import { computed, ref } from 'vue'

import {
  useChallengeTopologyStudioPage,
  type TopologyStudioMode,
} from '@/features/challenge-topology-studio'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import TopologyChallengeWorkbench from './TopologyChallengeWorkbench.vue'
import TopologyChallengeWorkspaceHeader from './TopologyChallengeWorkspaceHeader.vue'
import TopologyTemplateHeroSection from './TopologyTemplateHeroSection.vue'
import TopologyTemplateLibraryHeader from './TopologyTemplateLibraryHeader.vue'
import TopologyTemplateWorkbench from './TopologyTemplateWorkbench.vue'

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

    <div v-if="loading && !isTemplateLibraryMode" class="content-pane topology-loading-pane">
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
      <section v-if="isTemplateLibraryMode" class="content-pane template-library-main">
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

.topology-page--challenge .workspace-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-page--challenge .topology-topbar-leading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2-5);
}

.topology-page--challenge .workspace-overline,
.topology-page--challenge .topology-page-kicker {
  display: inline-flex;
  align-items: center;
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.topology-page--challenge .topology-topbar-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0 var(--space-3);
  font-size: var(--font-size-0-76);
  font-weight: 600;
  color: var(--journal-accent);
}

.topology-page--challenge .topology-topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.topology-page--challenge .topology-action-btn,
.topology-page--template-library .topology-action-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-0-84);
}

.topology-page--challenge .topology-action-btn {
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

.topology-page--challenge .topology-action-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

.topology-action-btn--icon {
  min-width: 2.75rem;
  padding-inline: var(--space-3);
}

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

.topology-page--challenge .topology-page-heading {
  display: grid;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-page--challenge .topology-page-copy {
  max-width: 48rem;
}

.topology-page--challenge .content-pane.topology-workspace {
  display: grid;
  gap: var(--space-7);
  grid-template-columns: minmax(0, 1fr) minmax(19rem, 22rem);
  align-items: start;
  min-width: 0;
  padding: 0;
}

.topology-page--challenge .topology-primary-column,
.topology-page--challenge .topology-context-stack,
.topology-page--challenge .topology-side-stack {
  display: grid;
  gap: var(--space-6);
}

.topology-page--challenge .topology-context-rail {
  min-width: 0;
  padding-left: var(--space-6);
  border-left: 1px solid var(--topology-divider);
}

.topology-page--challenge .topology-context-stack {
  position: sticky;
  top: var(--space-6);
}

.topology-page--challenge :deep(.section-card) {
  padding: var(--space-5) 0 0;
  border-top: 1px solid var(--topology-divider);
}

.topology-page--challenge .topology-primary-column :deep(.section-card:first-child),
.topology-page--challenge .topology-side-stack :deep(.section-card:first-child) {
  padding-top: 0;
  border-top: 0;
}

.topology-page--challenge :deep(.section-card__header) {
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom-color: var(--topology-divider);
}

.topology-page--challenge :deep(.section-card__header h2) {
  color: var(--journal-ink);
  font-size: var(--font-size-1-08);
}

.topology-page--challenge :deep(.section-card__header p) {
  color: var(--journal-muted);
}

.topology-page--challenge :deep(.section-card__body) {
  padding-left: 0;
}

.topology-page--challenge :deep(.section-card__body > .rounded-2xl),
.topology-page--challenge :deep(.section-card__body > .rounded-xl),
.topology-page--challenge :deep([data-node-editor]),
.topology-page--challenge :deep(.topology-canvas-board__root) {
  border-color: var(--journal-border);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--topology-panel) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--topology-panel-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 14px 30px var(--color-shadow-soft);
}

.topology-page--challenge :deep(input),
.topology-page--challenge :deep(select),
.topology-page--challenge :deep(textarea) {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
}

.topology-page--challenge :deep(input::placeholder),
.topology-page--challenge :deep(textarea::placeholder) {
  color: color-mix(in srgb, var(--journal-muted) 78%, transparent);
}

.topology-page--challenge :deep(option) {
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.topology-page--challenge :deep(input:focus),
.topology-page--challenge :deep(select:focus),
.topology-page--challenge :deep(textarea:focus) {
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
  outline: none;
}

.topology-page--challenge :deep(.topology-canvas-board__surface) {
  border-color: color-mix(in srgb, var(--journal-border) 70%, transparent);
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

@media (max-width: 1023px) {
  .topology-page--challenge .content-pane.topology-workspace {
    grid-template-columns: minmax(0, 1fr);
  }

  .topology-page--challenge .topology-context-rail {
    padding-left: 0;
    padding-top: var(--space-6);
    border-top: 1px solid var(--topology-divider);
    border-left: 0;
  }

  .topology-page--challenge .topology-context-stack {
    position: static;
  }
}

@media (max-width: 767px) {
  .topology-page--challenge {
    padding: var(--space-5);
  }

  .topology-page--challenge .workspace-topbar {
    align-items: flex-start;
    padding-bottom: var(--space-5);
  }

}
</style>

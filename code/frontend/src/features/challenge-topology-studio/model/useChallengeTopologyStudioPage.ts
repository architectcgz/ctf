import { computed, ref } from 'vue'
import type {
  AdminChallengeListItem,
  AdminImageListItem,
  ChallengeTopologyData,
  EnvironmentTemplateData,
} from '@/api/contracts'

import {
  buildTopologyCanvasGraph,
  type CanvasNodePosition,
} from './topologyLayout'
import {
  createEmptyTopologyDraft,
  buildTopologyDraftValidationIssues,
  type TopologyEditorDraft,
} from './topologyDraft'
import { useTopologyCanvasActions } from './useTopologyCanvasActions'
import { useTopologyPersistenceActions } from './useTopologyPersistenceActions'
import { useTopologyEdgeEditing } from './useTopologyEdgeEditing'
import { useTopologyStructureMutations } from './useTopologyStructureMutations'
import type { CanvasInteractionMode } from './topologyTypes'
import { useTopologyDataLoader } from './useTopologyDataLoader'
import { useTopologyInteractionBindings } from './useTopologyInteractionBindings'
import { useTopologyTemplateApply } from './useTopologyTemplateApply'
import { useTopologyTemplateSelection } from './useTopologyTemplateSelection'
import { useTopologyTemplateMutations } from './useTopologyTemplateMutations'
import { useTopologySelectionState } from './useTopologySelectionState'
import { useTopologyStudioPresentation } from './useTopologyStudioPresentation'

export type TopologyStudioMode = 'challenge' | 'template-library'

interface UseChallengeTopologyStudioPageOptions {
  challengeId: string
  mode: TopologyStudioMode
}

export function useChallengeTopologyStudioPage(options: UseChallengeTopologyStudioPageOptions) {
  const loading = ref(true)
  const saving = ref(false)
  const exporting = ref(false)
  const challenge = ref<AdminChallengeListItem | null>(null)
  const topology = ref<ChallengeTopologyData | null>(null)
  const images = ref<AdminImageListItem[]>([])
  const templates = ref<EnvironmentTemplateData[]>([])
  const {
    templateKeyword,
    selectedTemplateId,
    templateName,
    templateDescription,
    canSaveTemplate,
    selectedTemplate,
    selectedTemplateSummary,
    resetTemplateForm,
    clearTemplateSelection,
    reconcileTemplateSelection,
  } = useTopologyTemplateSelection({
    templates,
  })
  const draft = ref<TopologyEditorDraft>(createEmptyTopologyDraft())
  const selectedNodeKey = ref<string | null>(null)
  const selectedEdgeId = ref<string | null>(null)
  const interactionMode = ref<CanvasInteractionMode>('pan')
  const pendingSourceNodeKey = ref<string | null>(null)
  const nodePositions = ref<Record<string, CanvasNodePosition>>({})
  const isTemplateLibraryMode = computed(() => options.mode === 'template-library')
  const {
    selectedNodeDraft,
    selectedEdgeMeta,
    selectedLinkDraft,
    selectedPolicyDraft,
    selectedEdgeSourceKey,
    selectedEdgeTargetKey,
    selectedEdgeKind,
    selectedCanvasSummary,
    syncEntryNode,
    applyTopologyDraft,
    updateCanvasQuickNumber,
    toggleSelectedNodeNetwork,
  } = useTopologySelectionState({
    draft,
    nodePositions,
    selectedNodeKey,
    selectedEdgeId,
    pendingSourceNodeKey,
  })

  const {
    nodeOptions,
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
    topologySummary,
    canvasModeLabel,
  } = useTopologyStudioPresentation({
    challengeId: options.challengeId,
    isTemplateLibraryMode,
    challenge,
    topology,
    selectedTemplate,
    draft,
    interactionMode,
    pendingSourceNodeKey,
  })
  const draftValidationIssues = computed(() => buildTopologyDraftValidationIssues(draft.value))
  const canvasGraph = computed(() => buildTopologyCanvasGraph(draft.value, nodePositions.value))
  const {
    updateSelectedEdgeSourceKey,
    updateSelectedEdgeTargetKey,
    handleSelectedEdgeKindChange,
  } = useTopologyEdgeEditing({
    draft,
    selectedEdgeId,
    selectedEdgeMeta,
    selectedLinkDraft,
    selectedPolicyDraft,
  })

  function applyEmptyTemplateDraft() {
    applyTopologyDraft(createEmptyTopologyDraft())
  }

  const { loadTemplates, reloadAll } = useTopologyDataLoader({
    challengeId: options.challengeId,
    loading,
    isTemplateLibraryMode,
    selectedTemplateId,
    templateKeyword,
    templates,
    challenge,
    topology,
    images,
    applyTopologyDraft,
    applyEmptyTemplateDraft,
    resetTemplateForm,
    reconcileTemplateSelection,
  })

  const { templateBusy, handleCreateTemplate, handleUpdateTemplate, handleDeleteTemplate } =
    useTopologyTemplateMutations({
      canSaveTemplate,
      templateName,
      templateDescription,
      selectedTemplateId,
      templates,
      draft,
      isTemplateLibraryMode,
      applyEmptyTemplateDraft,
      resetTemplateForm,
      loadTemplates,
    })
  const { loadTemplateIntoDraft, handleApplyTemplate } = useTopologyTemplateApply({
    challengeId: options.challengeId,
    templateBusy,
    applyTopologyDraft,
    resetTemplateForm,
    onAfterApplied: async () => {
      await reloadAll()
    },
  })
  const { handleSaveTopology, handleExportPackage, handleDeleteTopology } =
    useTopologyPersistenceActions({
      challengeId: options.challengeId,
      isTemplateLibraryMode,
      draft,
      topology,
      saving,
      exporting,
      applyTopologyDraft,
      applyEmptyTemplateDraft,
      loadTemplates,
      reloadAll,
    })

  function handleResetTemplateEditor() {
    applyTopologyDraft(createEmptyTopologyDraft())
    resetTemplateForm(null)
  }

  const { addNetwork, removeNetwork, addNode, removeNode, addLink, addPolicy } =
    useTopologyStructureMutations({
      draft,
      selectedNodeKey,
      selectedEdgeId,
      syncEntryNode,
    })
  const {
    updateNodePosition,
    setInteractionMode,
    handleCanvasSelectNode,
    handleCanvasSelectEdge,
    handleCanvasCreateNode,
    handleCanvasCreateEdge,
    removeSelectedCanvasItem,
  } = useTopologyCanvasActions({
    draft,
    selectedNodeKey,
    selectedEdgeId,
    interactionMode,
    pendingSourceNodeKey,
    nodePositions,
    removeNode,
  })

  useTopologyInteractionBindings({
    draft,
    selectedNodeKey,
    selectedEdgeId,
    interactionMode,
    pendingSourceNodeKey,
    nodePositions,
    removeSelectedCanvasItem,
    reloadAll,
  })

  return {
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
    updateNodePosition,
    setInteractionMode,
    handleCanvasSelectNode,
    handleCanvasSelectEdge,
    handleCanvasCreateNode,
    handleCanvasCreateEdge,
    removeSelectedCanvasItem,
    addLink,
    addPolicy,
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
  }
}

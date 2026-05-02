import { computed, ref } from 'vue'
import type {
  AdminChallengeListItem,
  AdminImageListItem,
  ChallengeTopologyData,
  ChallengePackageRevisionData,
  EnvironmentTemplateData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

import {
  buildTopologyCanvasGraph,
  normalizeCanvasPositions,
  type CanvasNodePosition,
} from './topologyLayout'
import {
  createDraftFromTopology,
  createEmptyTopologyDraft,
  buildTopologyDraftValidationIssues,
  type TopologyEditorDraft,
  type TopologyLinkDraft,
  type TopologyNodeDraft,
  type TopologyPolicyDraft,
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

export type TopologyStudioMode = 'challenge' | 'template-library'

interface UseChallengeTopologyStudioPageOptions {
  challengeId: string
  mode: TopologyStudioMode
}

export function useChallengeTopologyStudioPage(options: UseChallengeTopologyStudioPageOptions) {
  const toast = useToast()

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

  const nodeOptions = computed(() =>
    draft.value.nodes.map((node) => ({
      key: node.key,
      label: node.name || node.key,
    }))
  )

  const pageHeader = computed(() => ({
    eyebrow: isTemplateLibraryMode.value ? 'Template Library' : 'Topology Studio',
    title: isTemplateLibraryMode.value ? '环境模板库' : '拓扑编排台',
    description: isTemplateLibraryMode.value
      ? '独立管理环境模板，支持列表检索、图形编辑、新建、覆盖和删除。'
      : '按题目维度管理网络分段、节点编排、模板复用和当前已生效的链路策略。',
  }))
  const loadingText = computed(() =>
    isTemplateLibraryMode.value ? '正在同步模板库...' : '正在同步拓扑与模板...'
  )
  const heroEyebrow = computed(() =>
    isTemplateLibraryMode.value ? 'Template Library' : 'Challenge Runtime'
  )
  const heroTitle = computed(() =>
    isTemplateLibraryMode.value
      ? selectedTemplate.value?.name || '环境模板库'
      : challenge.value?.title || `题目 #${options.challengeId}`
  )
  const heroDescription = computed(() =>
    isTemplateLibraryMode.value
      ? '当前页面直接调用环境模板接口，可独立维护模板列表、编辑器草稿与模板写回。'
      : '题目拓扑现在会直接读取题包来源、基线和导出修订。平台允许继续编辑，但会明确标出与题包基线的偏离状态。'
  )
  const statusCard = computed(() => {
    if (isTemplateLibraryMode.value) {
      return {
        eyebrow: '当前选择',
        title: selectedTemplate.value ? '已载入模板' : '空白草稿',
        subtitle: selectedTemplate.value
          ? `模板 ID：${selectedTemplate.value.id}`
          : '当前编辑器草稿尚未绑定到任何模板。',
      }
    }
    return {
      eyebrow: '当前生效',
      title: topology.value ? '已保存' : '未保存',
      subtitle: topology.value
        ? `入口节点：${topology.value.entry_node_key}`
        : '当前编辑器草稿尚未落库。',
    }
  })
  const secondaryCard = computed(() => {
    if (isTemplateLibraryMode.value) {
      return {
        eyebrow: '模板使用',
        title: selectedTemplate.value ? String(selectedTemplate.value.usage_count) : '0',
        subtitle: selectedTemplate.value
          ? `最近更新：${selectedTemplate.value.updated_at}`
          : '选择模板后可查看使用次数和更新时间。',
      }
    }
    return {
      eyebrow: '题包同步',
      title:
        topology.value?.source_type === 'package_import'
          ? topology.value.sync_status === 'drifted'
            ? '已偏离题包'
            : '与题包一致'
          : '平台手工拓扑',
      subtitle:
        topology.value?.source_type === 'package_import'
          ? topology.value.source_path || '题包导入拓扑'
          : topology.value?.template_id
            ? `最近一次按模板 ${topology.value.template_id} 保存`
            : '当前拓扑未绑定题包来源。',
    }
  })
  const packageBaselineSummary = computed(() => {
    const baseline = topology.value?.package_baseline
    if (!baseline) {
      return null
    }
    return {
      entryNodeKey: baseline.entry_node_key,
      networkCount: baseline.networks?.length || 0,
      nodeCount: baseline.nodes.length,
      linkCount: baseline.links?.length || 0,
      policyCount: baseline.policies?.length || 0,
    }
  })
  const packageFiles = computed(() => topology.value?.package_files || [])
  const packageRevisionHistory = computed<ChallengePackageRevisionData[]>(
    () => topology.value?.package_revisions || []
  )
  const packageSourceSummary = computed(() => {
    if (!topology.value?.source_type) {
      return {
        title: '暂无题包来源',
        subtitle: '当前题目拓扑还没有关联到题包导入基线。',
        canExport: false,
      }
    }
    if (topology.value.source_type === 'package_import') {
      return {
        title: topology.value.sync_status === 'drifted' ? '题包基线已漂移' : '题包基线已接入',
        subtitle: topology.value.source_path || '来源于导入题包',
        canExport: true,
      }
    }
    return {
      title: '平台手工拓扑',
      subtitle: topology.value.template_id
        ? `最近一次按模板 ${topology.value.template_id} 保存`
        : '当前题目拓扑尚未来自题包导入。',
      canExport: false,
    }
  })
  const selectedCanvasSummary = computed(() => {
    if (selectedNodeDraft.value) {
      return `已选节点：${selectedNodeDraft.value.name || selectedNodeDraft.value.key} / 网络 ${selectedNodeDraft.value.network_keys.length}`
    }
    if (selectedEdgeMeta.value) {
      return `已选${selectedEdgeMeta.value.kind === 'link' ? '连线' : '策略'}：${selectedEdgeSourceKey.value || '未选源'} -> ${selectedEdgeTargetKey.value || '未选目标'}`
    }
    if (pendingSourceNodeKey.value) {
      return `已选源节点：${pendingSourceNodeKey.value}，等待选择目标节点`
    }
    return '未选中节点或边，可直接点击画布元素进入编辑'
  })
  const draftValidationIssues = computed(() => buildTopologyDraftValidationIssues(draft.value))

  const topologySummary = computed(() => ({
    networks: draft.value.networks.length,
    nodes: draft.value.nodes.length,
    links: draft.value.links.length,
    policies: draft.value.policies.length,
  }))
  const canvasGraph = computed(() => buildTopologyCanvasGraph(draft.value, nodePositions.value))
  const selectedNodeDraft = computed<TopologyNodeDraft | null>(
    () => draft.value.nodes.find((node) => node.key === selectedNodeKey.value) || null
  )
  const selectedEdgeMeta = computed<{
    kind: 'link' | 'policy'
    index: number
    model: TopologyLinkDraft | TopologyPolicyDraft
  } | null>(() => {
    if (!selectedEdgeId.value) {
      return null
    }
    const [prefix, rawIndex] = selectedEdgeId.value.split('-')
    const index = Number(rawIndex)
    if (Number.isNaN(index)) {
      return null
    }
    if (prefix === 'link' && draft.value.links[index]) {
      return { kind: 'link', index, model: draft.value.links[index] }
    }
    if (prefix === 'policy' && draft.value.policies[index]) {
      return { kind: 'policy', index, model: draft.value.policies[index] }
    }
    return null
  })
  const selectedLinkDraft = computed<TopologyLinkDraft | null>(() =>
    selectedEdgeMeta.value?.kind === 'link'
      ? (selectedEdgeMeta.value.model as TopologyLinkDraft)
      : null
  )
  const selectedPolicyDraft = computed<TopologyPolicyDraft | null>(() =>
    selectedEdgeMeta.value?.kind === 'policy'
      ? (selectedEdgeMeta.value.model as TopologyPolicyDraft)
      : null
  )
  const selectedEdgeSourceKey = computed(
    () => selectedLinkDraft.value?.from_node_key || selectedPolicyDraft.value?.source_node_key || ''
  )
  const selectedEdgeTargetKey = computed(
    () => selectedLinkDraft.value?.to_node_key || selectedPolicyDraft.value?.target_node_key || ''
  )
  const selectedEdgeKind = computed<'link' | 'allow' | 'deny'>(() => {
    if (selectedLinkDraft.value) {
      return 'link'
    }
    return selectedPolicyDraft.value?.action || 'deny'
  })
  const canvasModeLabel = computed(() => {
    switch (interactionMode.value) {
      case 'add-node':
        return '点空白处新增节点'
      case 'link':
        return pendingSourceNodeKey.value ? '选择目标节点创建逻辑连线' : '选择源节点创建逻辑连线'
      case 'allow':
        return pendingSourceNodeKey.value
          ? '选择目标节点创建 allow 策略'
          : '选择源节点创建 allow 策略'
      case 'deny':
        return pendingSourceNodeKey.value
          ? '选择目标节点创建 deny 策略'
          : '选择源节点创建 deny 策略'
      default:
        return '拖拽节点调整布局，点击节点聚焦编辑卡片'
    }
  })

  function updateCanvasQuickNumber(
    field: 'service_port',
    value: string,
    node: TopologyNodeDraft | null
  ) {
    if (!node) {
      return
    }
    node[field] = value.trim() === '' ? null : Number(value)
  }

  function toggleSelectedNodeNetwork(networkKey: string, checked: boolean) {
    if (!selectedNodeDraft.value) {
      return
    }
    const next = checked
      ? Array.from(new Set([...selectedNodeDraft.value.network_keys, networkKey]))
      : selectedNodeDraft.value.network_keys.filter((item) => item !== networkKey)
    selectedNodeDraft.value.network_keys =
      next.length > 0 ? next : [draft.value.networks[0]?.key || 'default']
  }
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

  function syncEntryNode() {
    if (!draft.value.nodes.some((node) => node.key === draft.value.entry_node_key)) {
      draft.value.entry_node_key = draft.value.nodes[0]?.key || ''
    }
  }

  function applyTopologyDraft(next: TopologyEditorDraft) {
    draft.value = next
    nodePositions.value = normalizeCanvasPositions(next, nodePositions.value)
    if (!selectedNodeKey.value || !next.nodes.some((node) => node.key === selectedNodeKey.value)) {
      selectedNodeKey.value = next.nodes[0]?.key || null
    }
    syncEntryNode()
  }

  function applyEmptyTemplateDraft() {
    applyTopologyDraft(createEmptyTopologyDraft())
  }

  const { loadTemplates, loadPageData, reloadAll } = useTopologyDataLoader({
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

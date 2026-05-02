import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import {
  deleteChallengeTopology,
  exportChallengePackage,
  getChallengeDetail,
  getChallengeTopology,
  getEnvironmentTemplates,
  getImages,
  saveChallengeTopology,
} from '@/api/admin/authoring'
import type {
  AdminChallengeListItem,
  AdminImageListItem,
  ChallengeTopologyData,
  ChallengePackageRevisionData,
  EnvironmentTemplateData,
} from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import {
  buildTopologyCanvasGraph,
  clampCanvasPosition,
  normalizeCanvasPositions,
  type CanvasNodePosition,
} from './topologyLayout'
import {
  createDraftFromTopology,
  createUniqueNodeDraft,
  createEmptyLinkDraft,
  createEmptyNetworkDraft,
  createEmptyPolicyDraft,
  createEmptyTopologyDraft,
  buildTopologyDraftValidationIssues,
  serializeTopologyDraft,
  type TopologyEditorDraft,
  type TopologyLinkDraft,
  type TopologyNodeDraft,
  type TopologyPolicyDraft,
} from './topologyDraft'
import type { CanvasInteractionMode } from './topologyTypes'
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

  function updateSelectedEdgeKind(value: 'link' | 'allow' | 'deny') {
    const meta = selectedEdgeMeta.value
    if (!meta) {
      return
    }
    if (value === 'link') {
      if (meta.kind === 'link') {
        return
      }
      const policy = meta.model as TopologyPolicyDraft
      draft.value.policies.splice(meta.index, 1)
      draft.value.links.push({
        uid: `link-${Date.now().toString(16)}`,
        from_node_key: policy.source_node_key,
        to_node_key: policy.target_node_key,
      })
    } else {
      if (meta.kind === 'policy') {
        ;(meta.model as TopologyPolicyDraft).action = value
        return
      }
      const link = meta.model as TopologyLinkDraft
      draft.value.links.splice(meta.index, 1)
      draft.value.policies.push({
        uid: `policy-${Date.now().toString(16)}`,
        source_node_key: link.from_node_key,
        target_node_key: link.to_node_key,
        action: value,
      })
    }
    selectedEdgeId.value = null
  }

  function updateSelectedEdgeSourceKey(value: string) {
    if (selectedLinkDraft.value) {
      selectedLinkDraft.value.from_node_key = value
      return
    }
    if (selectedPolicyDraft.value) {
      selectedPolicyDraft.value.source_node_key = value
    }
  }

  function updateSelectedEdgeTargetKey(value: string) {
    if (selectedLinkDraft.value) {
      selectedLinkDraft.value.to_node_key = value
      return
    }
    if (selectedPolicyDraft.value) {
      selectedPolicyDraft.value.target_node_key = value
    }
  }

  function handleSelectedEdgeKindChange(value: string) {
    if (value === 'link' || value === 'allow' || value === 'deny') {
      updateSelectedEdgeKind(value)
    }
  }

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

  async function loadTemplates() {
    templates.value = await getEnvironmentTemplates(templateKeyword.value.trim() || undefined)
    reconcileTemplateSelection()
  }

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

  async function loadPageData() {
    loading.value = true
    try {
      if (isTemplateLibraryMode.value) {
        const imageResult = await getImages({ page: 1, page_size: 200 })
        challenge.value = null
        topology.value = null
        images.value = imageResult.list
        if (!selectedTemplateId.value) {
          applyTopologyDraft(createEmptyTopologyDraft())
          resetTemplateForm(null)
        }
        return
      }

      const [challengeDetail, imageResult, currentTopology] = await Promise.all([
        getChallengeDetail(options.challengeId),
        getImages({ page: 1, page_size: 200 }),
        getChallengeTopology(options.challengeId),
      ])

      challenge.value = challengeDetail
      images.value = imageResult.list
      topology.value = currentTopology
      applyTopologyDraft(createDraftFromTopology(currentTopology))
      resetTemplateForm(
        currentTopology?.template_id
          ? templates.value.find((item) => item.id === currentTopology.template_id) || null
          : null
      )
    } finally {
      loading.value = false
    }
  }

  async function reloadAll() {
    await loadTemplates()
    await loadPageData()
  }

  function handleResetTemplateEditor() {
    applyTopologyDraft(createEmptyTopologyDraft())
    resetTemplateForm(null)
  }

  function addNetwork() {
    draft.value.networks = [
      ...draft.value.networks,
      createEmptyNetworkDraft(draft.value.networks.length + 1),
    ]
  }

  function removeNetwork(uid: string) {
    if (draft.value.networks.length <= 1) {
      toast.warning('至少保留一个网络')
      return
    }

    const removing = draft.value.networks.find((item) => item.uid === uid)
    draft.value.networks = draft.value.networks.filter((item) => item.uid !== uid)
    if (!removing) {
      return
    }

    const fallbackNetworkKey = draft.value.networks[0]?.key
    for (const node of draft.value.nodes) {
      node.network_keys = node.network_keys.filter((key) => key !== removing.key)
      if (node.network_keys.length === 0 && fallbackNetworkKey) {
        node.network_keys = [fallbackNetworkKey]
      }
    }
  }

  function addNode() {
    const next = createUniqueNodeDraft(draft.value)
    draft.value.nodes.push(next)

    if (!draft.value.entry_node_key) {
      draft.value.entry_node_key = next.key
    }
    selectedNodeKey.value = next.key
    selectedEdgeId.value = null
  }

  function removeNode(uid: string) {
    if (draft.value.nodes.length <= 1) {
      toast.warning('至少保留一个节点')
      return
    }

    const removing = draft.value.nodes.find((item) => item.uid === uid)
    draft.value.nodes = draft.value.nodes.filter((item) => item.uid !== uid)
    if (!removing) {
      return
    }

    draft.value.links = draft.value.links.filter(
      (link) => link.from_node_key !== removing.key && link.to_node_key !== removing.key
    )
    draft.value.policies = draft.value.policies.filter(
      (policy) => policy.source_node_key !== removing.key && policy.target_node_key !== removing.key
    )
    syncEntryNode()
  }

  function updateNodePosition(payload: { nodeKey: string; position: CanvasNodePosition }) {
    nodePositions.value = {
      ...nodePositions.value,
      [payload.nodeKey]: clampCanvasPosition(payload.position),
    }
  }

  async function focusNodeEditor(nodeKey: string) {
    selectedNodeKey.value = nodeKey
    selectedEdgeId.value = null
    await nextTick()
    const element = document.querySelector<HTMLElement>(`[data-node-editor="${nodeKey}"]`)
    element?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }

  function setInteractionMode(mode: CanvasInteractionMode) {
    interactionMode.value = mode
    pendingSourceNodeKey.value = null
  }

  function handleCanvasSelectNode(nodeKey: string) {
    if (interactionMode.value === 'pan' || interactionMode.value === 'add-node') {
      void focusNodeEditor(nodeKey)
      return
    }
    pendingSourceNodeKey.value = nodeKey
    selectedNodeKey.value = nodeKey
    selectedEdgeId.value = null
  }

  function handleCanvasSelectEdge(edgeId: string) {
    selectedEdgeId.value = edgeId
    selectedNodeKey.value = null
  }

  function handleCanvasCreateNode(position: CanvasNodePosition) {
    const next = createUniqueNodeDraft(draft.value)

    // Batch position and node update
    nodePositions.value[next.key] = clampCanvasPosition(position)
    draft.value.nodes.push(next)

    selectedNodeKey.value = next.key
    selectedEdgeId.value = null
    interactionMode.value = 'pan'
  }

  function handleCanvasCreateEdge(payload: {
    sourceNodeKey: string
    targetNodeKey: string
    kind: 'link' | 'allow' | 'deny'
  }) {
    if (payload.kind === 'link') {
      draft.value.links = [...draft.value.links, createEmptyLinkDraft()]
      const last = draft.value.links[draft.value.links.length - 1]
      last.from_node_key = payload.sourceNodeKey
      last.to_node_key = payload.targetNodeKey
    } else {
      draft.value.policies = [...draft.value.policies, createEmptyPolicyDraft()]
      const last = draft.value.policies[draft.value.policies.length - 1]
      last.source_node_key = payload.sourceNodeKey
      last.target_node_key = payload.targetNodeKey
      last.action = payload.kind
    }
    pendingSourceNodeKey.value = null
    interactionMode.value = 'pan'
  }

  async function removeSelectedCanvasItem() {
    const confirmed = await confirmDestructiveAction({
      title: '删除选中项',
      message: '确认删除当前选中的节点或连线吗？该操作会直接修改当前草稿。',
      confirmButtonText: '确认删除',
    })
    if (!confirmed) {
      return
    }
    if (selectedEdgeId.value) {
      if (selectedEdgeId.value.startsWith('link-')) {
        const nextLinks = [...draft.value.links]
        const index = Number(selectedEdgeId.value.split('-')[1])
        if (!Number.isNaN(index)) {
          nextLinks.splice(index, 1)
          draft.value.links = nextLinks
        }
      } else if (selectedEdgeId.value.startsWith('policy-')) {
        const nextPolicies = [...draft.value.policies]
        const index = Number(selectedEdgeId.value.split('-')[1])
        if (!Number.isNaN(index)) {
          nextPolicies.splice(index, 1)
          draft.value.policies = nextPolicies
        }
      }
      selectedEdgeId.value = null
      return
    }

    if (!selectedNodeKey.value) {
      toast.warning('请先在画布中选择一个节点或连线')
      return
    }
    const node = draft.value.nodes.find((item) => item.key === selectedNodeKey.value)
    if (!node) {
      toast.warning('当前选中节点已不存在')
      return
    }
    removeNode(node.uid)
  }

  function addLink() {
    draft.value.links = [...draft.value.links, createEmptyLinkDraft()]
  }

  function addPolicy() {
    draft.value.policies = [...draft.value.policies, createEmptyPolicyDraft()]
  }

  async function handleSaveTopology() {
    saving.value = true
    try {
      const saved = await saveChallengeTopology(
        options.challengeId,
        serializeTopologyDraft(draft.value)
      )
      topology.value = saved
      applyTopologyDraft(createDraftFromTopology(saved))
      toast.success('题目拓扑已保存')
      await loadTemplates()
    } finally {
      saving.value = false
    }
  }

  async function handleExportPackage() {
    if (isTemplateLibraryMode.value) {
      return
    }
    exporting.value = true
    try {
      const exported = await exportChallengePackage(options.challengeId)
      toast.success('题目包已导出')
      await reloadAll()
      if (typeof window !== 'undefined' && exported.download_url) {
        window.open(exported.download_url, '_blank', 'noopener')
      }
    } finally {
      exporting.value = false
    }
  }

  async function handleDeleteTopology() {
    if (!topology.value) {
      toast.warning('当前题目还没有已保存的拓扑')
      return
    }
    const confirmed = await confirmDestructiveAction({
      title: '删除题目拓扑',
      message: '确认删除当前题目已保存的拓扑吗？删除后需要重新保存才能恢复。',
      confirmButtonText: '确认删除',
    })
    if (!confirmed) {
      return
    }
    saving.value = true
    try {
      await deleteChallengeTopology(options.challengeId)
      topology.value = null
      applyTopologyDraft(createEmptyTopologyDraft())
      toast.success('题目拓扑已删除')
    } catch (error) {
      const message =
        error instanceof Error && error.message.trim() ? error.message : '删除题目拓扑失败'
      toast.error(message)
    } finally {
      saving.value = false
    }
  }

  function isEditingTarget(target: EventTarget | null): boolean {
    if (!(target instanceof HTMLElement)) {
      return false
    }
    const tag = target.tagName
    return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (isEditingTarget(event.target)) {
      return
    }

    if (event.key === 'Escape') {
      pendingSourceNodeKey.value = null
      selectedEdgeId.value = null
      interactionMode.value = 'pan'
      return
    }

    if (
      (event.key === 'Delete' || event.key === 'Backspace') &&
      (selectedNodeKey.value || selectedEdgeId.value)
    ) {
      event.preventDefault()
      removeSelectedCanvasItem()
    }
  }

  onMounted(async () => {
    window.addEventListener('keydown', handleGlobalKeydown)
    await reloadAll()
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleGlobalKeydown)
  })

  watch(
    () => draft.value.nodes.length,
    () => {
      nodePositions.value = normalizeCanvasPositions(draft.value, nodePositions.value)
      if (
        selectedNodeKey.value &&
        !draft.value.nodes.some((node) => node.key === selectedNodeKey.value)
      ) {
        selectedNodeKey.value = draft.value.nodes[0]?.key || null
      }
    }
  )

  watch(interactionMode, (value) => {
    if (value === 'pan' || value === 'add-node') {
      pendingSourceNodeKey.value = null
    }
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

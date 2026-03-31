import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import {
  createEnvironmentTemplate,
  deleteChallengeTopology,
  deleteEnvironmentTemplate,
  getChallengeDetail,
  getChallengeTopology,
  getEnvironmentTemplates,
  getImages,
  saveChallengeTopology,
  updateEnvironmentTemplate,
} from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminImageListItem,
  ChallengeTopologyData,
  EnvironmentTemplateData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

import {
  buildTopologyCanvasGraph,
  clampCanvasPosition,
  normalizeCanvasPositions,
  type CanvasNodePosition,
} from '@/components/admin/topology/topologyLayout'
import {
  createDraftFromTemplate,
  createDraftFromTopology,
  createEmptyLinkDraft,
  createEmptyNetworkDraft,
  createEmptyNodeDraft,
  createEmptyPolicyDraft,
  createEmptyTopologyDraft,
  serializeEnvironmentTemplateDraft,
  serializeTopologyDraft,
  type TopologyEditorDraft,
  type TopologyLinkDraft,
  type TopologyNodeDraft,
  type TopologyPolicyDraft,
} from '@/components/admin/topology/topologyDraft'
import type { CanvasInteractionMode } from '@/components/admin/topology/TopologyCanvasBoard.vue'

export type TopologyStudioMode = 'challenge' | 'template-library'

interface UseChallengeTopologyStudioPageOptions {
  challengeId: string
  mode: TopologyStudioMode
}

export function useChallengeTopologyStudioPage(options: UseChallengeTopologyStudioPageOptions) {
  const toast = useToast()

  const loading = ref(true)
  const saving = ref(false)
  const templateBusy = ref(false)
  const challenge = ref<AdminChallengeListItem | null>(null)
  const topology = ref<ChallengeTopologyData | null>(null)
  const images = ref<AdminImageListItem[]>([])
  const templates = ref<EnvironmentTemplateData[]>([])
  const templateKeyword = ref('')
  const selectedTemplateId = ref<string | null>(null)
  const templateName = ref('')
  const templateDescription = ref('')
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

  const canSaveTemplate = computed(() => templateName.value.trim().length > 0)
  const selectedTemplate = computed(
    () => templates.value.find((item) => item.id === selectedTemplateId.value) || null
  )
  const pageHeader = computed(() => ({
    eyebrow: isTemplateLibraryMode.value ? 'Template Library' : 'Topology Studio',
    title: isTemplateLibraryMode.value ? '环境模板库' : '拓扑编排台',
    description: isTemplateLibraryMode.value
      ? '独立管理环境模板，支持列表检索、图形编辑、新建、覆盖和删除。'
      : '按挑战维度管理网络分段、节点编排、模板复用和当前已生效的链路策略。',
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
      : challenge.value?.title || `挑战 #${options.challengeId}`
  )
  const heroDescription = computed(() =>
    isTemplateLibraryMode.value
      ? '当前页面直接调用环境模板接口，可独立维护模板列表、编辑器草稿与模板写回。'
      : '当前页面会直接调用拓扑和环境模板接口。前端编辑器当前开放节点级 allow/deny，端口/协议级 ACL 暂未开放。'
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
      eyebrow: '模板绑定',
      title: topology.value?.template_id || '无',
      subtitle: topology.value?.template_id
        ? '当前挑战最近一次是按模板保存的。'
        : '当前拓扑为手工编排或尚未保存。',
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
  const draftValidationIssues = computed(() => {
    const issues: string[] = []
    const trimmedNodeKeys = draft.value.nodes.map((node) => node.key.trim()).filter(Boolean)
    const trimmedNetworkKeys = draft.value.networks
      .map((network) => network.key.trim())
      .filter(Boolean)
    const duplicateNodeKeys = trimmedNodeKeys.filter(
      (key, index) => trimmedNodeKeys.indexOf(key) !== index
    )
    const duplicateNetworkKeys = trimmedNetworkKeys.filter(
      (key, index) => trimmedNetworkKeys.indexOf(key) !== index
    )

    if (duplicateNodeKeys.length > 0) {
      issues.push(`节点 Key 重复：${Array.from(new Set(duplicateNodeKeys)).join('、')}`)
    }
    if (duplicateNetworkKeys.length > 0) {
      issues.push(`网络 Key 重复：${Array.from(new Set(duplicateNetworkKeys)).join('、')}`)
    }
    if (!draft.value.nodes.some((node) => node.key === draft.value.entry_node_key)) {
      issues.push('入口节点未指向现有节点')
    }
    if (draft.value.nodes.some((node) => node.network_keys.length === 0)) {
      issues.push('存在未挂载任何网络的节点')
    }
    if (
      draft.value.links.some(
        (link) =>
          !draft.value.nodes.some((node) => node.key === link.from_node_key) ||
          !draft.value.nodes.some((node) => node.key === link.to_node_key)
      )
    ) {
      issues.push('存在引用不存在节点的逻辑连线')
    }
    if (
      draft.value.policies.some(
        (policy) =>
          !draft.value.nodes.some((node) => node.key === policy.source_node_key) ||
          !draft.value.nodes.some((node) => node.key === policy.target_node_key)
      )
    ) {
      issues.push('存在引用不存在节点的链路策略')
    }

    return issues
  })
  const selectedTemplateSummary = computed(() => {
    if (!selectedTemplate.value) {
      return '尚未选中模板，可从下方模板库载入到当前草稿。'
    }
    return `${selectedTemplate.value.name} · 节点 ${selectedTemplate.value.nodes.length} · 网络 ${selectedTemplate.value.networks?.length || 0} · 使用 ${selectedTemplate.value.usage_count}`
  })

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

  function resetTemplateForm(template?: EnvironmentTemplateData | null) {
    selectedTemplateId.value = template?.id || null
    templateName.value = template?.name || ''
    templateDescription.value = template?.description || ''
  }

  function applyTopologyDraft(next: TopologyEditorDraft) {
    draft.value = next
    nodePositions.value = normalizeCanvasPositions(next, nodePositions.value)
    if (!selectedNodeKey.value || !next.nodes.some((node) => node.key === selectedNodeKey.value)) {
      selectedNodeKey.value = next.nodes[0]?.key || null
    }
    syncEntryNode()
  }

  async function loadTemplates() {
    templates.value = await getEnvironmentTemplates(templateKeyword.value.trim() || undefined)
    if (
      selectedTemplateId.value &&
      !templates.value.some((item) => item.id === selectedTemplateId.value)
    ) {
      resetTemplateForm(null)
    }
  }

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
    const nodeCount = draft.value.nodes.length
    const next = createEmptyNodeDraft(nodeCount + 1)

    const existingKeys = new Set(draft.value.nodes.map((n) => n.key))
    let counter = nodeCount + 1
    while (existingKeys.has(next.key)) {
      counter++
      next.key = `node-${counter}`
      next.name = `节点 ${counter}`
    }

    next.network_keys = [draft.value.networks[0]?.key || 'default']
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
    const nodeCount = draft.value.nodes.length
    const next = createEmptyNodeDraft(nodeCount + 1)

    // Ensure key is truly unique to avoid key collision watch overhead
    const existingKeys = new Set(draft.value.nodes.map((n) => n.key))
    let finalKey = next.key
    let counter = nodeCount + 1
    while (existingKeys.has(finalKey)) {
      counter++
      finalKey = `node-${counter}`
    }
    next.key = finalKey
    next.name = `节点 ${counter}`

    next.network_keys = [draft.value.networks[0]?.key || 'default']

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

  function removeSelectedCanvasItem() {
    if (
      typeof window !== 'undefined' &&
      !window.confirm('确认删除当前选中的节点或连线吗？该操作会直接修改当前草稿。')
    ) {
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

  function loadTemplateIntoDraft(template: EnvironmentTemplateData) {
    applyTopologyDraft(createDraftFromTemplate(template))
    resetTemplateForm(template)
    toast.success('模板已载入编辑器草稿')
  }

  async function handleApplyTemplate(template: EnvironmentTemplateData) {
    if (
      typeof window !== 'undefined' &&
      !window.confirm(`确认将模板“${template.name}”应用到当前挑战吗？已保存拓扑会被模板覆盖。`)
    ) {
      return
    }
    templateBusy.value = true
    try {
      await saveChallengeTopology(options.challengeId, { template_id: Number(template.id) })
      toast.success('模板已应用到挑战')
      await reloadAll()
    } finally {
      templateBusy.value = false
    }
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
      toast.success('挑战拓扑已保存')
      await loadTemplates()
    } finally {
      saving.value = false
    }
  }

  async function handleDeleteTopology() {
    if (!topology.value) {
      toast.warning('当前挑战还没有已保存的拓扑')
      return
    }
    if (
      typeof window !== 'undefined' &&
      !window.confirm('确认删除当前挑战已保存的拓扑吗？删除后需要重新保存才能恢复。')
    ) {
      return
    }
    saving.value = true
    try {
      await deleteChallengeTopology(options.challengeId)
      topology.value = null
      applyTopologyDraft(createEmptyTopologyDraft())
      toast.success('挑战拓扑已删除')
    } finally {
      saving.value = false
    }
  }

  async function handleCreateTemplate() {
    if (!canSaveTemplate.value) {
      toast.error('请填写模板名称')
      return
    }

    templateBusy.value = true
    try {
      const created = await createEnvironmentTemplate(
        serializeEnvironmentTemplateDraft(
          templateName.value,
          templateDescription.value,
          draft.value
        )
      )
      resetTemplateForm(created)
      toast.success('模板已创建')
      await loadTemplates()
    } finally {
      templateBusy.value = false
    }
  }

  async function handleUpdateTemplate() {
    if (!selectedTemplateId.value) {
      toast.warning('请先选择一个模板')
      return
    }
    if (!canSaveTemplate.value) {
      toast.error('请填写模板名称')
      return
    }

    templateBusy.value = true
    try {
      const updated = await updateEnvironmentTemplate(
        selectedTemplateId.value,
        serializeEnvironmentTemplateDraft(
          templateName.value,
          templateDescription.value,
          draft.value
        )
      )
      resetTemplateForm(updated)
      toast.success('模板已更新')
      await loadTemplates()
    } finally {
      templateBusy.value = false
    }
  }

  async function handleDeleteTemplate(templateId: string) {
    const template = templates.value.find((item) => item.id === templateId)
    if (
      typeof window !== 'undefined' &&
      !window.confirm(`确认删除模板“${template?.name || templateId}”吗？该操作不可撤销。`)
    ) {
      return
    }
    templateBusy.value = true
    try {
      await deleteEnvironmentTemplate(templateId)
      if (selectedTemplateId.value === templateId) {
        if (isTemplateLibraryMode.value) {
          applyTopologyDraft(createEmptyTopologyDraft())
        }
        resetTemplateForm(null)
      }
      toast.success('模板已删除')
      await loadTemplates()
    } finally {
      templateBusy.value = false
    }
  }

  function clearTemplateSelection() {
    resetTemplateForm(null)
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
    handleDeleteTopology,
    handleCreateTemplate,
    handleUpdateTemplate,
    handleDeleteTemplate,
    clearTemplateSelection,
    loadTemplates,
    resetTemplateForm,
  }
}

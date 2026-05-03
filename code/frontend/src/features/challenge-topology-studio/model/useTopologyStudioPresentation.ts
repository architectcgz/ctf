import { computed, type ComputedRef, type Ref } from 'vue'

import type {
  AdminChallengeListItem,
  ChallengeTopologyData,
  ChallengePackageRevisionData,
  EnvironmentTemplateData,
} from '@/api/contracts'

import type { TopologyEditorDraft } from './topologyDraft'
import type { CanvasInteractionMode } from './topologyTypes'

interface UseTopologyStudioPresentationOptions {
  challengeId: string
  isTemplateLibraryMode: ComputedRef<boolean>
  challenge: Ref<AdminChallengeListItem | null>
  topology: Ref<ChallengeTopologyData | null>
  selectedTemplate: ComputedRef<EnvironmentTemplateData | null>
  draft: Ref<TopologyEditorDraft>
  interactionMode: Ref<CanvasInteractionMode>
  pendingSourceNodeKey: Ref<string | null>
}

export function useTopologyStudioPresentation(options: UseTopologyStudioPresentationOptions) {
  const {
    challengeId,
    isTemplateLibraryMode,
    challenge,
    topology,
    selectedTemplate,
    draft,
    interactionMode,
    pendingSourceNodeKey,
  } = options

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
      : challenge.value?.title || `题目 #${challengeId}`
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
  const topologySummary = computed(() => ({
    networks: draft.value.networks.length,
    nodes: draft.value.nodes.length,
    links: draft.value.links.length,
    policies: draft.value.policies.length,
  }))
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

  return {
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
  }
}

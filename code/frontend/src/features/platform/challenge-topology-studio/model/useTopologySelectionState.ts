import { computed, type Ref } from 'vue'

import {
  normalizeCanvasPositions,
  type CanvasNodePosition,
} from './topologyLayout'
import type {
  TopologyEditorDraft,
  TopologyLinkDraft,
  TopologyNodeDraft,
  TopologyPolicyDraft,
} from './topologyDraft'

interface UseTopologySelectionStateOptions {
  draft: Ref<TopologyEditorDraft>
  nodePositions: Ref<Record<string, CanvasNodePosition>>
  selectedNodeKey: Ref<string | null>
  selectedEdgeId: Ref<string | null>
  pendingSourceNodeKey: Ref<string | null>
}

export function useTopologySelectionState(options: UseTopologySelectionStateOptions) {
  const {
    draft,
    nodePositions,
    selectedNodeKey,
    selectedEdgeId,
    pendingSourceNodeKey,
  } = options

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

  return {
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
  }
}

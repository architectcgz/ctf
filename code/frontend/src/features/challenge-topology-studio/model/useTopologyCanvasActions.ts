import { nextTick, type Ref } from 'vue'

import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import {
  clampCanvasPosition,
  type CanvasNodePosition,
} from './topologyLayout'
import {
  createUniqueNodeDraft,
  createEmptyLinkDraft,
  createEmptyPolicyDraft,
  type TopologyEditorDraft,
} from './topologyDraft'
import type { CanvasInteractionMode } from './topologyTypes'

interface UseTopologyCanvasActionsOptions {
  draft: Ref<TopologyEditorDraft>
  selectedNodeKey: Ref<string | null>
  selectedEdgeId: Ref<string | null>
  interactionMode: Ref<CanvasInteractionMode>
  pendingSourceNodeKey: Ref<string | null>
  nodePositions: Ref<Record<string, CanvasNodePosition>>
  removeNode: (uid: string) => void
}

export function useTopologyCanvasActions(options: UseTopologyCanvasActionsOptions) {
  const {
    draft,
    selectedNodeKey,
    selectedEdgeId,
    interactionMode,
    pendingSourceNodeKey,
    nodePositions,
    removeNode,
  } = options
  const toast = useToast()

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

  return {
    updateNodePosition,
    focusNodeEditor,
    setInteractionMode,
    handleCanvasSelectNode,
    handleCanvasSelectEdge,
    handleCanvasCreateNode,
    handleCanvasCreateEdge,
    removeSelectedCanvasItem,
  }
}

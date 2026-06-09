import type { ComputedRef, Ref } from 'vue'

import type {
  TopologyEditorDraft,
  TopologyLinkDraft,
  TopologyPolicyDraft,
} from './topologyDraft'

interface SelectedEdgeMeta {
  kind: 'link' | 'policy'
  index: number
  model: TopologyLinkDraft | TopologyPolicyDraft
}

interface UseTopologyEdgeEditingOptions {
  draft: Ref<TopologyEditorDraft>
  selectedEdgeId: Ref<string | null>
  selectedEdgeMeta: ComputedRef<SelectedEdgeMeta | null>
  selectedLinkDraft: ComputedRef<TopologyLinkDraft | null>
  selectedPolicyDraft: ComputedRef<TopologyPolicyDraft | null>
}

export function useTopologyEdgeEditing(options: UseTopologyEdgeEditingOptions) {
  const { draft, selectedEdgeId, selectedEdgeMeta, selectedLinkDraft, selectedPolicyDraft } = options

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

  return {
    updateSelectedEdgeKind,
    updateSelectedEdgeSourceKey,
    updateSelectedEdgeTargetKey,
    handleSelectedEdgeKindChange,
  }
}

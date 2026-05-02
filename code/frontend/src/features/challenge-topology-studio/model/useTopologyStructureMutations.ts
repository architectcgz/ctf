import type { Ref } from 'vue'

import { useToast } from '@/composables/useToast'

import {
  createUniqueNodeDraft,
  createEmptyLinkDraft,
  createEmptyNetworkDraft,
  createEmptyPolicyDraft,
  type TopologyEditorDraft,
} from './topologyDraft'

interface UseTopologyStructureMutationsOptions {
  draft: Ref<TopologyEditorDraft>
  selectedNodeKey: Ref<string | null>
  selectedEdgeId: Ref<string | null>
  syncEntryNode: () => void
}

export function useTopologyStructureMutations(options: UseTopologyStructureMutationsOptions) {
  const { draft, selectedNodeKey, selectedEdgeId, syncEntryNode } = options
  const toast = useToast()

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

  function addLink() {
    draft.value.links = [...draft.value.links, createEmptyLinkDraft()]
  }

  function addPolicy() {
    draft.value.policies = [...draft.value.policies, createEmptyPolicyDraft()]
  }

  return {
    addNetwork,
    removeNetwork,
    addNode,
    removeNode,
    addLink,
    addPolicy,
  }
}

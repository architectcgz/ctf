import type { ComputedRef, Ref } from 'vue'

import type { TopologyTier } from '@/api/contracts'

import { useToast } from '@/composables/useToast'

import {
  createUniqueNodeDraft,
  createEmptyLinkDraft,
  createEmptyNetworkDraft,
  createEmptyPolicyDraft,
  type TopologyLinkDraft,
  type TopologyNetworkDraft,
  type TopologyNodeDraft,
  type TopologyEditorDraft,
  type TopologyPolicyDraft,
} from './topologyDraft'

interface UseTopologyStructureMutationsOptions {
  draft: Ref<TopologyEditorDraft>
  selectedNodeKey: Ref<string | null>
  selectedEdgeId: Ref<string | null>
  selectedNodeDraft: ComputedRef<TopologyNodeDraft | null>
  syncEntryNode: () => void
}

export function useTopologyStructureMutations(options: UseTopologyStructureMutationsOptions) {
  const { draft, selectedNodeKey, selectedEdgeId, selectedNodeDraft, syncEntryNode } = options
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

  function updateNetworkDraft(payload: {
    uid: string
    patch: Partial<Pick<TopologyNetworkDraft, 'key' | 'name' | 'cidr' | 'internal'>>
  }) {
    const network = draft.value.networks.find((item) => item.uid === payload.uid)
    if (!network) return
    Object.assign(network, payload.patch)
  }

  function updateNodeDraft(payload: { uid: string; node: TopologyNodeDraft }) {
    const index = draft.value.nodes.findIndex((item) => item.uid === payload.uid)
    if (index === -1) return
    draft.value.nodes[index] = payload.node
  }

  function updateSelectedNodeField(payload: {
    field: 'name' | 'image_id' | 'tier' | 'inject_flag'
    value: string | boolean
  }) {
    if (!selectedNodeDraft.value) {
      return
    }

    switch (payload.field) {
      case 'inject_flag':
        selectedNodeDraft.value.inject_flag = Boolean(payload.value)
        return
      case 'name':
        selectedNodeDraft.value.name = String(payload.value)
        return
      case 'image_id':
        selectedNodeDraft.value.image_id = String(payload.value)
        return
      case 'tier':
        selectedNodeDraft.value.tier = String(payload.value) as TopologyTier
        return
    }
  }

  function updateEntryNodeKey(value: string) {
    draft.value.entry_node_key = value
  }

  function updateLinkDraft(payload: {
    uid: string
    patch: Partial<Pick<TopologyLinkDraft, 'from_node_key' | 'to_node_key'>>
  }) {
    const link = draft.value.links.find((item) => item.uid === payload.uid)
    if (!link) return
    Object.assign(link, payload.patch)
  }

  function removeLinkDraft(uid: string) {
    draft.value.links = draft.value.links.filter((item) => item.uid !== uid)
  }

  function updatePolicyDraft(payload: {
    uid: string
    patch: Partial<Pick<TopologyPolicyDraft, 'source_node_key' | 'target_node_key' | 'action'>>
  }) {
    const policy = draft.value.policies.find((item) => item.uid === payload.uid)
    if (!policy) return
    Object.assign(policy, payload.patch)
  }

  function removePolicyDraft(uid: string) {
    draft.value.policies = draft.value.policies.filter((item) => item.uid !== uid)
  }

  return {
    addNetwork,
    removeNetwork,
    addNode,
    removeNode,
    addLink,
    addPolicy,
    updateNetworkDraft,
    updateNodeDraft,
    updateSelectedNodeField,
    updateEntryNodeKey,
    updateLinkDraft,
    removeLinkDraft,
    updatePolicyDraft,
    removePolicyDraft,
  }
}

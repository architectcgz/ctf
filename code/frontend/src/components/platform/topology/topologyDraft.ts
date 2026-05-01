import type {
  AdminChallengeTopologyPayload,
  AdminEnvironmentTemplatePayload,
} from '@/api/admin/authoring'
import type {
  ChallengeTopologyData,
  EnvironmentTemplateData,
  TopologyNodeData,
  TopologyPolicyAction,
  TopologyTier,
} from '@/api/contracts'

export interface TopologyEnvEntryDraft {
  uid: string
  key: string
  value: string
}

export interface TopologyNetworkDraft {
  uid: string
  key: string
  name: string
  cidr: string
  internal: boolean
}

export interface TopologyNodeDraft {
  uid: string
  key: string
  name: string
  image_id: string
  service_port: number | null
  inject_flag: boolean
  tier: TopologyTier
  network_keys: string[]
  env_entries: TopologyEnvEntryDraft[]
  cpu_quota: number | null
  memory_mb: number | null
  pids_limit: number | null
}

export interface TopologyLinkDraft {
  uid: string
  from_node_key: string
  to_node_key: string
}

export interface TopologyPolicyDraft {
  uid: string
  source_node_key: string
  target_node_key: string
  action: TopologyPolicyAction
}

export interface TopologyEditorDraft {
  entry_node_key: string
  networks: TopologyNetworkDraft[]
  nodes: TopologyNodeDraft[]
  links: TopologyLinkDraft[]
  policies: TopologyPolicyDraft[]
}

const DEFAULT_NETWORK_KEY = 'default'

function createUid(prefix: string): string {
  return `${prefix}-${Math.random().toString(16).slice(2, 10)}`
}

function toNullableNumber(value: number | null | undefined): number | undefined {
  return value == null || Number.isNaN(value) ? undefined : value
}

function toEnvEntries(env?: Record<string, string>): TopologyEnvEntryDraft[] {
  if (!env) {
    return []
  }
  return Object.entries(env).map(([key, value]) => ({
    uid: createUid('env'),
    key,
    value,
  }))
}

function toEnvMap(entries: TopologyEnvEntryDraft[]): Record<string, string> | undefined {
  const items = entries.reduce<Record<string, string>>((acc, item) => {
    const key = item.key.trim()
    if (!key) {
      return acc
    }
    acc[key] = item.value
    return acc
  }, {})
  return Object.keys(items).length > 0 ? items : undefined
}

function toNodeDraft(node: TopologyNodeData, fallbackNetworkKey: string): TopologyNodeDraft {
  return {
    uid: createUid('node'),
    key: node.key,
    name: node.name,
    image_id: node.image_id || '',
    service_port: node.service_port ?? null,
    inject_flag: Boolean(node.inject_flag),
    tier: node.tier || 'service',
    network_keys: node.network_keys?.length ? [...node.network_keys] : [fallbackNetworkKey],
    env_entries: toEnvEntries(node.env),
    cpu_quota: node.resources?.cpu_quota ?? null,
    memory_mb: node.resources?.memory_mb ?? null,
    pids_limit: node.resources?.pids_limit ?? null,
  }
}

function normalizeNetworks(
  networks?: Array<{ key: string; name: string; cidr?: string; internal?: boolean }>
) {
  if (networks && networks.length > 0) {
    return networks.map((network) => ({
      uid: createUid('network'),
      key: network.key,
      name: network.name,
      cidr: network.cidr || '',
      internal: Boolean(network.internal),
    }))
  }

  return [createEmptyNetworkDraft()]
}

function toDraft(
  source?:
    | Pick<ChallengeTopologyData, 'entry_node_key' | 'networks' | 'nodes' | 'links' | 'policies'>
    | Pick<EnvironmentTemplateData, 'entry_node_key' | 'networks' | 'nodes' | 'links' | 'policies'>
): TopologyEditorDraft {
  if (!source) {
    return createEmptyTopologyDraft()
  }

  const networks = normalizeNetworks(source.networks)
  const fallbackNetworkKey = networks[0]?.key || DEFAULT_NETWORK_KEY
  const nodes = source.nodes.map((node) => toNodeDraft(node, fallbackNetworkKey))

  return {
    entry_node_key: source.entry_node_key || nodes[0]?.key || '',
    networks,
    nodes,
    links: (source.links || []).map((link) => ({
      uid: createUid('link'),
      from_node_key: link.from_node_key,
      to_node_key: link.to_node_key,
    })),
    policies: (source.policies || []).map((policy) => ({
      uid: createUid('policy'),
      source_node_key: policy.source_node_key,
      target_node_key: policy.target_node_key,
      action: policy.action,
    })),
  }
}

function toResources(node: TopologyNodeDraft) {
  const resources = {
    cpu_quota: toNullableNumber(node.cpu_quota),
    memory_mb: toNullableNumber(node.memory_mb),
    pids_limit: toNullableNumber(node.pids_limit),
  }
  return Object.values(resources).some((value) => value !== undefined) ? resources : undefined
}

export function createEmptyNetworkDraft(index = 1): TopologyNetworkDraft {
  const isDefault = index === 1
  return {
    uid: createUid('network'),
    key: isDefault ? DEFAULT_NETWORK_KEY : `net-${index}`,
    name: isDefault ? '默认网络' : `网络 ${index}`,
    cidr: '',
    internal: false,
  }
}

export function createEmptyNodeDraft(index = 1): TopologyNodeDraft {
  return {
    uid: createUid('node'),
    key: `node-${index}`,
    name: `节点 ${index}`,
    image_id: '',
    service_port: index === 1 ? 8080 : null,
    inject_flag: index === 1,
    tier: index === 1 ? 'public' : 'service',
    network_keys: [DEFAULT_NETWORK_KEY],
    env_entries: [],
    cpu_quota: null,
    memory_mb: null,
    pids_limit: null,
  }
}

export function createEmptyLinkDraft(): TopologyLinkDraft {
  return {
    uid: createUid('link'),
    from_node_key: '',
    to_node_key: '',
  }
}

export function createEmptyPolicyDraft(): TopologyPolicyDraft {
  return {
    uid: createUid('policy'),
    source_node_key: '',
    target_node_key: '',
    action: 'deny',
  }
}

export function createEmptyEnvEntryDraft(): TopologyEnvEntryDraft {
  return {
    uid: createUid('env'),
    key: '',
    value: '',
  }
}

export function createEmptyTopologyDraft(): TopologyEditorDraft {
  return {
    entry_node_key: 'node-1',
    networks: [createEmptyNetworkDraft()],
    nodes: [createEmptyNodeDraft()],
    links: [],
    policies: [],
  }
}

export function createDraftFromTopology(
  topology: ChallengeTopologyData | null
): TopologyEditorDraft {
  return toDraft(topology || undefined)
}

export function createDraftFromTemplate(template: EnvironmentTemplateData): TopologyEditorDraft {
  return toDraft(template)
}

export function serializeTopologyDraft(draft: TopologyEditorDraft): AdminChallengeTopologyPayload {
  return {
    entry_node_key: draft.entry_node_key.trim(),
    networks: draft.networks.map((network) => ({
      key: network.key.trim(),
      name: network.name.trim(),
      cidr: network.cidr.trim() || undefined,
      internal: network.internal,
    })),
    nodes: draft.nodes.map((node) => ({
      key: node.key.trim(),
      name: node.name.trim(),
      image_id: node.image_id ? Number(node.image_id) : undefined,
      service_port: toNullableNumber(node.service_port),
      inject_flag: node.inject_flag,
      tier: node.tier,
      network_keys: node.network_keys,
      env: toEnvMap(node.env_entries),
      resources: toResources(node),
    })),
    links: draft.links.map((link) => ({
      from_node_key: link.from_node_key.trim(),
      to_node_key: link.to_node_key.trim(),
    })),
    policies: draft.policies.map((policy) => ({
      source_node_key: policy.source_node_key.trim(),
      target_node_key: policy.target_node_key.trim(),
      action: policy.action,
    })),
  }
}

export function serializeEnvironmentTemplateDraft(
  name: string,
  description: string,
  draft: TopologyEditorDraft
): AdminEnvironmentTemplatePayload {
  return {
    name: name.trim(),
    description: description.trim(),
    entry_node_key: draft.entry_node_key.trim(),
    networks: serializeTopologyDraft(draft).networks,
    nodes: serializeTopologyDraft(draft).nodes || [],
    links: serializeTopologyDraft(draft).links,
    policies: serializeTopologyDraft(draft).policies,
  }
}

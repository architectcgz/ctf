import { ApiError, request } from '../request'

import type {
  AdminChallengeHint,
  AdminChallengeImportImageDelivery,
  AdminChallengeImportCommitData,
  AdminChallengeImportPreview,
  AdminChallengeImportTopologyData,
  AdminChallengeListItem,
  AdminChallengePublishRequestData,
  AdminChallengeWriteupData,
  AdminImageListItem,
  ChallengePackageExportData,
  ChallengePackageFileData,
  ChallengePackageRevisionData,
  ChallengeTopologyData,
  EnvironmentTemplateData,
  PageResult,
  TopologyLinkData,
  TopologyNetworkData,
  TopologyNodeData,
  TopologySpecData,
  TopologyTrafficPolicyData,
  WriteupVisibility,
} from '../contracts'

interface RawAdminChallengeItem {
  id: string | number
  title: string
  description?: string
  category: AdminChallengeListItem['category']
  difficulty: AdminChallengeListItem['difficulty']
  points: number
  instance_sharing?: AdminChallengeListItem['instance_sharing']
  created_by?: string | number | null
  image_id?: string | number | null
  attachment_url?: string
  hints?: Array<{
    id: string | number
    level: number
    title?: string
    content: string
  }>
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
}

interface RawAdminChallengePublishRequestData {
  id: string | number
  challenge_id: string | number
  status: AdminChallengePublishRequestData['status'] | 'pending' | 'passed'
  requested_by?: string | number | null
  request_source?: string | null
  active?: boolean
  failure_summary?: string | null
  started_at?: string | null
  finished_at?: string | null
  published_at?: string | null
  result?: {
    challenge_id: string | number
    precheck: {
      passed: boolean
      started_at: string
      ended_at: string
      steps: Array<{
        name: string
        passed: boolean
        message: string
      }>
    }
    runtime: {
      passed: boolean
      started_at: string
      ended_at: string
      access_url?: string
      container_count: number
      network_count: number
      steps: Array<{
        name: string
        passed: boolean
        message: string
      }>
    }
  } | null
  created_at: string
  updated_at: string
}

interface RawChallengeImportPreview {
  id: string | number
  file_name: string
  slug: string
  title: string
  description: string
  category: AdminChallengeListItem['category']
  difficulty: AdminChallengeListItem['difficulty']
  points: number
  attachments?: Array<{
    name: string
    path: string
  }>
  hints?: Array<{
    id?: string | number
    level: number
    title?: string
    content: string
  }>
  flag: {
    type: 'static' | 'dynamic' | 'regex' | 'manual_review'
    prefix?: string
  }
  runtime: {
    type?: string
    image_ref?: string
  }
  image_delivery?: AdminChallengeImportImageDelivery | null
  extensions: {
    topology: {
      source?: string
      enabled: boolean
    }
  }
  topology?: {
    source?: string
    entry_node_key: string
    networks?: RawTopologyNetworkData[]
    nodes: Array<{
      key: string
      name: string
      image_ref?: string
      service_port?: number
      inject_flag?: boolean
      tier?: TopologyNodeData['tier']
      network_keys?: string[]
      env?: Record<string, string>
    }>
    links?: RawTopologyLinkData[]
    policies?: RawTopologyTrafficPolicyData[]
  }
  package_files?: Array<{
    path: string
    size: number
  }>
  warnings?: string[]
  created_at: string
}

interface RawTopologyNetworkData {
  key: string
  name: string
  cidr?: string
  internal?: boolean
}

interface RawTopologyNodeResourcesData {
  cpu_quota?: number
  memory_mb?: number
  pids_limit?: number
}

interface RawChallengePackageFileData {
  path: string
  size: number
}

interface RawTopologyNodeData {
  key: string
  name: string
  image_id?: string | number | null
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyNodeData['tier']
  network_keys?: string[]
  env?: Record<string, string>
  resources?: RawTopologyNodeResourcesData
}

interface RawTopologyLinkData {
  from_node_key: string
  to_node_key: string
}

interface RawTopologyTrafficPolicyData {
  source_node_key: string
  target_node_key: string
  action: TopologyTrafficPolicyData['action']
  protocol?: TopologyTrafficPolicyData['protocol']
  ports?: number[]
}

interface RawChallengeTopologyData {
  id: string | number
  challenge_id: string | number
  template_id?: string | number | null
  entry_node_key: string
  networks?: RawTopologyNetworkData[]
  nodes: RawTopologyNodeData[]
  links?: RawTopologyLinkData[]
  policies?: RawTopologyTrafficPolicyData[]
  source_type?: ChallengeTopologyData['source_type']
  source_path?: string
  sync_status?: ChallengeTopologyData['sync_status']
  package_revision_id?: string | number | null
  last_export_revision_id?: string | number | null
  package_baseline?: {
    entry_node_key: string
    networks?: RawTopologyNetworkData[]
    nodes: RawTopologyNodeData[]
    links?: RawTopologyLinkData[]
    policies?: RawTopologyTrafficPolicyData[]
  }
  package_files?: Array<{
    path: string
    size: number
  }>
  package_revisions?: Array<{
    id: string | number
    revision_no: number
    source_type: ChallengePackageRevisionData['source_type']
    parent_revision_id?: string | number | null
    package_slug?: string
    archive_path?: string
    source_dir?: string
    topology_source_path?: string
    created_by?: string | number | null
    created_at: string
    updated_at: string
  }>
  created_at: string
  updated_at: string
}

interface RawChallengePackageExportData {
  challenge_id: string | number
  revision_id: string | number
  archive_path: string
  source_dir: string
  file_name: string
  download_url?: string
  created_at: string
}

interface RawEnvironmentTemplateData {
  id: string | number
  name: string
  description: string
  entry_node_key: string
  networks?: RawTopologyNetworkData[]
  nodes: RawTopologyNodeData[]
  links?: RawTopologyLinkData[]
  policies?: RawTopologyTrafficPolicyData[]
  usage_count: number
  created_at: string
  updated_at: string
}

interface RawChallengeFlagConfig {
  flag_type: 'static' | 'dynamic' | 'regex' | 'manual_review'
  flag_regex?: string
  flag_prefix?: string
  configured: boolean
}

interface RawAdminChallengeWriteupData {
  id: string | number
  challenge_id: string | number
  title: string
  content: string
  visibility: WriteupVisibility
  created_by?: string | number | null
  is_recommended?: boolean
  recommended_at?: string | null
  recommended_by?: string | number | null
  created_at: string
  updated_at: string
}

interface RawImageItem {
  id: string | number
  name: string
  tag: string
  description?: string
  size?: number
  status: AdminImageListItem['status']
  source_type?: AdminImageListItem['source_type']
  digest?: string
  build_job_id?: string | number | null
  last_error?: string
  verified_at?: string | null
  created_at: string
  updated_at: string
}

export interface AdminChallengePayload {
  title: string
  description?: string
  category: AdminChallengeListItem['category']
  difficulty: Extract<
    AdminChallengeListItem['difficulty'],
    'beginner' | 'easy' | 'medium' | 'hard' | 'insane'
  >
  points: number
  image_id: number
  attachment_url?: string
  hints?: AdminChallengeHint[]
}

export interface AdminChallengeFlagPayload {
  flag_type: 'static' | 'dynamic' | 'regex' | 'manual_review'
  flag?: string
  flag_regex?: string
  flag_prefix?: string
}

export interface AdminTopologyNodeResourcesPayload {
  cpu_quota?: number
  memory_mb?: number
  pids_limit?: number
}

export interface AdminTopologyNetworkPayload {
  key: string
  name: string
  cidr?: string
  internal?: boolean
}

export interface AdminTopologyNodePayload {
  key: string
  name: string
  image_id?: number
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyNodeData['tier']
  network_keys?: string[]
  env?: Record<string, string>
  resources?: AdminTopologyNodeResourcesPayload
}

export interface AdminTopologyLinkPayload {
  from_node_key: string
  to_node_key: string
}

export interface AdminTopologyPolicyPayload {
  source_node_key: string
  target_node_key: string
  action: TopologyTrafficPolicyData['action']
  protocol?: TopologyTrafficPolicyData['protocol']
  ports?: number[]
}

export interface AdminChallengeTopologyPayload {
  template_id?: number
  entry_node_key?: string
  networks?: AdminTopologyNetworkPayload[]
  nodes?: AdminTopologyNodePayload[]
  links?: AdminTopologyLinkPayload[]
  policies?: AdminTopologyPolicyPayload[]
}

export interface AdminEnvironmentTemplatePayload {
  name: string
  description?: string
  entry_node_key: string
  networks?: AdminTopologyNetworkPayload[]
  nodes: AdminTopologyNodePayload[]
  links?: AdminTopologyLinkPayload[]
  policies?: AdminTopologyPolicyPayload[]
}

export interface AdminChallengeWriteupPayload {
  title: string
  content: string
  visibility: WriteupVisibility
}

export interface AdminImagePayload {
  name: string
  tag: string
  description?: string
}

function isNotFoundError(error: unknown): boolean {
  return (
    (error instanceof ApiError && error.status === 404) ||
    ((error as { name?: string; status?: number } | undefined)?.name === 'ApiError' &&
      (error as { status?: number }).status === 404)
  )
}

function normalizeAdminChallengeWriteup(
  item: RawAdminChallengeWriteupData
): AdminChallengeWriteupData {
  return {
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    title: item.title,
    content: item.content,
    visibility: item.visibility,
    created_by: item.created_by == null ? undefined : String(item.created_by),
    is_recommended: item.is_recommended ?? false,
    recommended_at: item.recommended_at || undefined,
    recommended_by: item.recommended_by == null ? undefined : String(item.recommended_by),
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeChallenge(
  item: RawAdminChallengeItem,
  flagConfig?: RawChallengeFlagConfig
): AdminChallengeListItem {
  const hints: AdminChallengeHint[] | undefined = item.hints?.map((hint) => ({
    id: String(hint.id),
    level: hint.level,
    title: hint.title,
    content: hint.content,
  }))

  return {
    id: String(item.id),
    title: item.title,
    description: item.description,
    category: item.category,
    difficulty: item.difficulty,
    points: item.points,
    status: item.status,
    instance_sharing: item.instance_sharing ?? 'per_user',
    created_by: item.created_by == null ? undefined : String(item.created_by),
    image_id: item.image_id == null ? undefined : String(item.image_id),
    attachment_url: item.attachment_url,
    hints,
    created_at: item.created_at,
    updated_at: item.updated_at,
    flag_config: flagConfig
      ? {
          configured: flagConfig.configured,
          flag_type: flagConfig.flag_type,
          flag_regex: flagConfig.flag_regex,
          flag_prefix: flagConfig.flag_prefix,
        }
      : undefined,
  }
}

function normalizeChallengePublishRequest(
  item: RawAdminChallengePublishRequestData
): AdminChallengePublishRequestData {
  const status =
    item.status === 'pending' ? 'queued' : item.status === 'passed' ? 'succeeded' : item.status

  return {
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    status,
    active: item.active ?? (status === 'queued' || status === 'running'),
    requested_by: item.requested_by == null ? undefined : String(item.requested_by),
    request_source: item.request_source ?? undefined,
    failure_summary: item.failure_summary ?? undefined,
    started_at: item.started_at ?? undefined,
    finished_at: item.finished_at ?? undefined,
    published_at: item.published_at ?? undefined,
    result:
      item.result == null
        ? undefined
        : {
            challenge_id: String(item.result.challenge_id),
            precheck: {
              passed: item.result.precheck.passed,
              started_at: item.result.precheck.started_at,
              ended_at: item.result.precheck.ended_at,
              steps: item.result.precheck.steps.map((step) => ({
                name: step.name,
                passed: step.passed,
                message: step.message,
              })),
            },
            runtime: {
              passed: item.result.runtime.passed,
              started_at: item.result.runtime.started_at,
              ended_at: item.result.runtime.ended_at,
              access_url: item.result.runtime.access_url,
              container_count: item.result.runtime.container_count,
              network_count: item.result.runtime.network_count,
              steps: item.result.runtime.steps.map((step) => ({
                name: step.name,
                passed: step.passed,
                message: step.message,
              })),
            },
          },
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeChallengeImportPreview(
  item: RawChallengeImportPreview
): AdminChallengeImportPreview {
  return {
    id: String(item.id),
    file_name: item.file_name,
    slug: item.slug,
    title: item.title,
    description: item.description,
    category: item.category,
    difficulty: item.difficulty,
    points: item.points,
    attachments: item.attachments,
    hints: item.hints?.map((hint) => ({
      id: hint.id == null ? undefined : String(hint.id),
      level: hint.level,
      title: hint.title,
      content: hint.content,
    })),
    flag: item.flag,
    runtime: item.runtime,
    image_delivery: item.image_delivery ?? undefined,
    extensions: item.extensions,
    topology: item.topology ? normalizeChallengeImportTopology(item.topology) : undefined,
    package_files: item.package_files?.map(normalizeChallengePackageFile),
    warnings: item.warnings,
    created_at: item.created_at,
  }
}

function normalizeChallengeImportTopology(
  item: NonNullable<RawChallengeImportPreview['topology']>
): AdminChallengeImportTopologyData {
  return {
    source: item.source,
    entry_node_key: item.entry_node_key,
    networks: item.networks?.map(normalizeTopologyNetwork),
    nodes: item.nodes.map((node) => ({
      key: node.key,
      name: node.name,
      image_ref: node.image_ref,
      service_port: node.service_port,
      inject_flag: node.inject_flag,
      tier: node.tier,
      network_keys: node.network_keys,
      env: node.env,
    })),
    links: item.links?.map(normalizeTopologyLink),
    policies: item.policies?.map(normalizeTopologyPolicy),
  }
}

function normalizeChallengePackageFile(
  item: RawChallengePackageFileData
): ChallengePackageFileData {
  return {
    path: item.path,
    size: item.size,
  }
}

function normalizeTopologyNetwork(item: RawTopologyNetworkData): TopologyNetworkData {
  return {
    key: item.key,
    name: item.name,
    cidr: item.cidr,
    internal: item.internal,
  }
}

function normalizeTopologyNode(item: RawTopologyNodeData): TopologyNodeData {
  return {
    key: item.key,
    name: item.name,
    image_id: item.image_id == null ? undefined : String(item.image_id),
    service_port: item.service_port,
    inject_flag: item.inject_flag,
    tier: item.tier,
    network_keys: item.network_keys,
    env: item.env,
    resources: item.resources,
  }
}

function normalizeTopologyLink(item: RawTopologyLinkData): TopologyLinkData {
  return {
    from_node_key: item.from_node_key,
    to_node_key: item.to_node_key,
  }
}

function normalizeTopologyPolicy(item: RawTopologyTrafficPolicyData): TopologyTrafficPolicyData {
  return {
    source_node_key: item.source_node_key,
    target_node_key: item.target_node_key,
    action: item.action,
    protocol: item.protocol,
    ports: item.ports,
  }
}

function normalizeChallengeTopology(item: RawChallengeTopologyData): ChallengeTopologyData {
  return {
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    template_id: item.template_id == null ? undefined : String(item.template_id),
    entry_node_key: item.entry_node_key,
    networks: item.networks?.map(normalizeTopologyNetwork),
    nodes: item.nodes.map(normalizeTopologyNode),
    links: item.links?.map(normalizeTopologyLink),
    policies: item.policies?.map(normalizeTopologyPolicy),
    source_type: item.source_type,
    source_path: item.source_path,
    sync_status: item.sync_status,
    package_revision_id:
      item.package_revision_id == null ? undefined : String(item.package_revision_id),
    last_export_revision_id:
      item.last_export_revision_id == null ? undefined : String(item.last_export_revision_id),
    package_baseline: item.package_baseline
      ? normalizeTopologySpec(item.package_baseline)
      : undefined,
    package_files: item.package_files?.map(normalizeChallengePackageFile),
    package_revisions: item.package_revisions?.map(normalizeChallengePackageRevision),
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeTopologySpec(
  item: NonNullable<RawChallengeTopologyData['package_baseline']>
): TopologySpecData {
  return {
    entry_node_key: item.entry_node_key,
    networks: item.networks?.map(normalizeTopologyNetwork),
    nodes: item.nodes.map(normalizeTopologyNode),
    links: item.links?.map(normalizeTopologyLink),
    policies: item.policies?.map(normalizeTopologyPolicy),
  }
}

function normalizeChallengePackageRevision(
  item: NonNullable<RawChallengeTopologyData['package_revisions']>[number]
): ChallengePackageRevisionData {
  return {
    id: String(item.id),
    revision_no: item.revision_no,
    source_type: item.source_type,
    parent_revision_id:
      item.parent_revision_id == null ? undefined : String(item.parent_revision_id),
    package_slug: item.package_slug,
    archive_path: item.archive_path,
    source_dir: item.source_dir,
    topology_source_path: item.topology_source_path,
    created_by: item.created_by == null ? undefined : String(item.created_by),
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeChallengePackageExport(
  item: RawChallengePackageExportData
): ChallengePackageExportData {
  return {
    challenge_id: String(item.challenge_id),
    revision_id: String(item.revision_id),
    archive_path: item.archive_path,
    source_dir: item.source_dir,
    file_name: item.file_name,
    download_url: item.download_url,
    created_at: item.created_at,
  }
}

function normalizeEnvironmentTemplate(item: RawEnvironmentTemplateData): EnvironmentTemplateData {
  return {
    id: String(item.id),
    name: item.name,
    description: item.description,
    entry_node_key: item.entry_node_key,
    networks: item.networks?.map(normalizeTopologyNetwork),
    nodes: item.nodes.map(normalizeTopologyNode),
    links: item.links?.map(normalizeTopologyLink),
    policies: item.policies?.map(normalizeTopologyPolicy),
    usage_count: item.usage_count,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeImage(item: RawImageItem): AdminImageListItem {
  return {
    id: String(item.id),
    name: item.name,
    tag: item.tag,
    description: item.description,
    status: item.status,
    source_type: item.source_type,
    digest: item.digest,
    build_job_id: item.build_job_id == null ? undefined : String(item.build_job_id),
    last_error: item.last_error,
    verified_at: item.verified_at ?? undefined,
    size_bytes: item.size,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

export async function getChallenges(
  params?: Record<string, unknown>,
  options?: { signal?: AbortSignal }
): Promise<PageResult<AdminChallengeListItem>> {
  const response = await request<PageResult<RawAdminChallengeItem>>({
    method: 'GET',
    url: '/authoring/challenges',
    params,
    signal: options?.signal,
  })
  return {
    ...response,
    list: response.list.map((item) => normalizeChallenge(item)),
  }
}

export async function getChallengeDetail(id: string): Promise<AdminChallengeListItem> {
  const [challenge, flagConfig] = await Promise.all([
    request<RawAdminChallengeItem>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}`,
    }),
    request<RawChallengeFlagConfig>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/flag`,
    }).catch(() => undefined),
  ])

  return normalizeChallenge(challenge, flagConfig)
}

export async function previewChallengeImport(file: File): Promise<AdminChallengeImportPreview> {
  const form = new FormData()
  form.append('file', file)

  const response = await request<RawChallengeImportPreview>({
    method: 'POST',
    url: '/authoring/challenge-imports',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return normalizeChallengeImportPreview(response)
}

export async function listChallengeImports(): Promise<AdminChallengeImportPreview[]> {
  const response = await request<RawChallengeImportPreview[]>({
    method: 'GET',
    url: '/authoring/challenge-imports',
  })
  return Array.isArray(response) ? response.map(normalizeChallengeImportPreview) : []
}

export async function getChallengeImport(id: string): Promise<AdminChallengeImportPreview> {
  const response = await request<RawChallengeImportPreview>({
    method: 'GET',
    url: `/authoring/challenge-imports/${encodeURIComponent(id)}`,
  })
  return normalizeChallengeImportPreview(response)
}

export async function commitChallengeImport(id: string): Promise<AdminChallengeImportCommitData> {
  const response = await request<{ challenge: RawAdminChallengeItem }>({
    method: 'POST',
    url: `/authoring/challenge-imports/${encodeURIComponent(id)}/commit`,
  })

  return {
    challenge: normalizeChallenge(response.challenge),
  }
}

export async function createChallenge(data: AdminChallengePayload) {
  const response = await request<RawAdminChallengeItem>({
    method: 'POST',
    url: '/authoring/challenges',
    data,
  })
  return {
    challenge: normalizeChallenge(response),
  }
}

export async function updateChallenge(id: string, data: Partial<AdminChallengePayload>) {
  await request<void>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}`,
    data,
  })
}

export async function configureChallengeFlag(id: string, data: AdminChallengeFlagPayload) {
  return request<{ message: string }>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}/flag`,
    data,
  })
}

export async function getChallengeFlagConfig(id: string) {
  return request<RawChallengeFlagConfig>({
    method: 'GET',
    url: `/authoring/challenges/${encodeURIComponent(id)}/flag`,
  })
}

export async function createChallengePublishRequest(
  id: string
): Promise<AdminChallengePublishRequestData> {
  const response = await request<RawAdminChallengePublishRequestData>({
    method: 'POST',
    url: `/authoring/challenges/${encodeURIComponent(id)}/publish-requests`,
  })

  return normalizeChallengePublishRequest(response)
}

export async function getLatestChallengePublishRequest(
  id: string
): Promise<AdminChallengePublishRequestData | null> {
  try {
    const response = await request<RawAdminChallengePublishRequestData>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/publish-requests/latest`,
    })

    return normalizeChallengePublishRequest(response)
  } catch (error) {
    if (isNotFoundError(error)) {
      return null
    }
    throw error
  }
}

export async function deleteChallenge(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}`,
  })
}

export async function getChallengeWriteup(id: string): Promise<AdminChallengeWriteupData | null> {
  try {
    const response = await request<RawAdminChallengeWriteupData>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/writeup`,
    })
    return normalizeAdminChallengeWriteup(response)
  } catch (error) {
    if (isNotFoundError(error)) {
      return null
    }
    throw error
  }
}

export async function saveChallengeWriteup(id: string, data: AdminChallengeWriteupPayload) {
  const response = await request<RawAdminChallengeWriteupData>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup`,
    data,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function deleteChallengeWriteup(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup`,
  })
}

export async function recommendChallengeWriteup(id: string): Promise<AdminChallengeWriteupData> {
  const response = await request<RawAdminChallengeWriteupData>({
    method: 'POST',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup/recommend`,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function unrecommendChallengeWriteup(id: string): Promise<AdminChallengeWriteupData> {
  const response = await request<RawAdminChallengeWriteupData>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup/recommend`,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function getChallengeTopology(id: string): Promise<ChallengeTopologyData | null> {
  try {
    const response = await request<RawChallengeTopologyData>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/topology`,
    })
    return normalizeChallengeTopology(response)
  } catch (error) {
    if (isNotFoundError(error)) {
      return null
    }
    throw error
  }
}

export async function saveChallengeTopology(id: string, data: AdminChallengeTopologyPayload) {
  const response = await request<RawChallengeTopologyData>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}/topology`,
    data,
  })
  return normalizeChallengeTopology(response)
}

export async function exportChallengePackage(id: string): Promise<ChallengePackageExportData> {
  const response = await request<RawChallengePackageExportData>({
    method: 'POST',
    url: `/authoring/challenges/${encodeURIComponent(id)}/package-export`,
  })
  return normalizeChallengePackageExport(response)
}

export async function deleteChallengeTopology(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}/topology`,
  })
}

export async function getEnvironmentTemplates(
  keyword?: string
): Promise<EnvironmentTemplateData[]> {
  const response = await request<RawEnvironmentTemplateData[]>({
    method: 'GET',
    url: '/authoring/environment-templates',
    params: { keyword },
  })
  return response.map(normalizeEnvironmentTemplate)
}

export async function getEnvironmentTemplate(id: string): Promise<EnvironmentTemplateData> {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'GET',
    url: `/authoring/environment-templates/${encodeURIComponent(id)}`,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function createEnvironmentTemplate(data: AdminEnvironmentTemplatePayload) {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'POST',
    url: '/authoring/environment-templates',
    data,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function updateEnvironmentTemplate(id: string, data: AdminEnvironmentTemplatePayload) {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'PUT',
    url: `/authoring/environment-templates/${encodeURIComponent(id)}`,
    data,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function deleteEnvironmentTemplate(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/environment-templates/${encodeURIComponent(id)}`,
  })
}

export async function getImages(params?: Record<string, unknown>) {
  const response = await request<PageResult<RawImageItem>>({
    method: 'GET',
    url: '/authoring/images',
    params,
  })
  return {
    ...response,
    list: response.list.map(normalizeImage),
  }
}

export async function createImage(data: AdminImagePayload) {
  const response = await request<RawImageItem>({ method: 'POST', url: '/authoring/images', data })
  return {
    image: normalizeImage(response),
  }
}

export async function deleteImage(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/images/${encodeURIComponent(id)}`,
  })
}

import { request } from '../request'

import type {
  AWDChallengeStatus,
  AWDCheckerType,
  AWDDeploymentMode,
  AWDReadinessStatus,
  AWDServiceType,
  AdminAwdChallengeData,
  AdminAwdChallengeImportCommitData,
  AdminAwdChallengeImportPreview,
  AdminChallengeImportImageDelivery,
  ChallengeCategory,
  PageResult,
} from '../contracts'

export interface AdminAwdChallengeListParams {
  page?: number
  page_size?: number
  keyword?: string
  service_type?: AWDServiceType
  deployment_mode?: AWDDeploymentMode
  readiness_status?: AWDReadinessStatus
  status?: AWDChallengeStatus
}

export interface AdminAwdChallengeCreatePayload {
  name: string
  slug: string
  category: AdminAwdChallengeData['category']
  difficulty: AdminAwdChallengeData['difficulty']
  description?: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
}

export interface AdminAwdChallengeUpdatePayload {
  name?: string
  slug?: string
  category?: AdminAwdChallengeData['category']
  difficulty?: AdminAwdChallengeData['difficulty']
  description?: string
  service_type?: AWDServiceType
  deployment_mode?: AWDDeploymentMode
  status?: AWDChallengeStatus
}

interface RawAdminAwdChallengeData {
  id: string | number
  name: string
  slug: string
  category: ChallengeCategory
  difficulty: AdminAwdChallengeData['difficulty']
  description: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
  version: string
  status: AWDChallengeStatus
  readiness_status: AWDReadinessStatus
  checker_type?: AWDCheckerType | null
  checker_config?: Record<string, unknown> | null
  flag_mode?: string | null
  flag_config?: Record<string, unknown> | null
  defense_entry_mode?: string | null
  access_config?: Record<string, unknown> | null
  runtime_config?: Record<string, unknown> | null
  created_by?: string | number | null
  last_verified_at?: string | null
  created_at: string
  updated_at: string
}

interface RawAdminAwdChallengePageData {
  items: RawAdminAwdChallengeData[]
  total: number
  page: number
  size: number
}

interface RawAdminAwdChallengeImportPreview {
  id: string | number
  file_name: string
  slug: string
  title: string
  category: ChallengeCategory
  difficulty: AdminAwdChallengeImportPreview['difficulty']
  description: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
  version: string
  checker_type: AWDCheckerType
  checker_config?: Record<string, unknown> | null
  flag_mode?: string | null
  flag_config?: Record<string, unknown> | null
  defense_entry_mode?: string | null
  access_config?: Record<string, unknown> | null
  runtime_config?: Record<string, unknown> | null
  image_delivery?: AdminChallengeImportImageDelivery | null
  warnings?: string[] | null
  created_at: string
}

interface RawAdminAwdChallengeImportCommitData {
  challenge: RawAdminAwdChallengeData
}

function normalizeAdminAwdChallenge(item: RawAdminAwdChallengeData): AdminAwdChallengeData {
  const normalized: AdminAwdChallengeData = {
    id: String(item.id),
    name: item.name,
    slug: item.slug,
    category: item.category,
    difficulty: item.difficulty,
    description: item.description,
    service_type: item.service_type,
    deployment_mode: item.deployment_mode,
    version: item.version,
    status: item.status,
    readiness_status: item.readiness_status,
    created_by: item.created_by == null ? undefined : String(item.created_by),
    last_verified_at: item.last_verified_at || undefined,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }

  if (item.checker_type) {
    normalized.checker_type = item.checker_type
  }
  if (item.checker_config && Object.keys(item.checker_config).length > 0) {
    normalized.checker_config = item.checker_config
  }
  if (item.flag_mode) {
    normalized.flag_mode = item.flag_mode
  }
  if (item.flag_config && Object.keys(item.flag_config).length > 0) {
    normalized.flag_config = item.flag_config
  }
  if (item.defense_entry_mode) {
    normalized.defense_entry_mode = item.defense_entry_mode
  }
  if (item.access_config && Object.keys(item.access_config).length > 0) {
    normalized.access_config = item.access_config
  }
  if (item.runtime_config && Object.keys(item.runtime_config).length > 0) {
    normalized.runtime_config = item.runtime_config
  }

  return normalized
}

function normalizeAdminAwdChallengeImportPreview(
  item: RawAdminAwdChallengeImportPreview
): AdminAwdChallengeImportPreview {
  const normalized: AdminAwdChallengeImportPreview = {
    id: String(item.id),
    file_name: item.file_name,
    slug: item.slug,
    title: item.title,
    category: item.category,
    difficulty: item.difficulty,
    description: item.description,
    service_type: item.service_type,
    deployment_mode: item.deployment_mode,
    version: item.version,
    checker_type: item.checker_type,
    created_at: item.created_at,
  }

  if (item.checker_config && Object.keys(item.checker_config).length > 0) {
    normalized.checker_config = item.checker_config
  }
  if (item.flag_mode) {
    normalized.flag_mode = item.flag_mode
  }
  if (item.flag_config && Object.keys(item.flag_config).length > 0) {
    normalized.flag_config = item.flag_config
  }
  if (item.defense_entry_mode) {
    normalized.defense_entry_mode = item.defense_entry_mode
  }
  if (item.access_config && Object.keys(item.access_config).length > 0) {
    normalized.access_config = item.access_config
  }
  if (item.runtime_config && Object.keys(item.runtime_config).length > 0) {
    normalized.runtime_config = item.runtime_config
  }
  if (item.image_delivery) {
    normalized.image_delivery = item.image_delivery
  }
  if (item.warnings && item.warnings.length > 0) {
    normalized.warnings = item.warnings
  }

  return normalized
}

export async function listAdminAwdChallenges(
  params?: AdminAwdChallengeListParams
): Promise<PageResult<AdminAwdChallengeData>> {
  const response = await request<RawAdminAwdChallengePageData>({
    method: 'GET',
    url: '/authoring/awd-challenges',
    params,
  })

  return {
    list: response.items.map(normalizeAdminAwdChallenge),
    total: response.total,
    page: response.page,
    page_size: response.size,
  }
}

export async function getAdminAwdChallenge(id: string): Promise<AdminAwdChallengeData> {
  const response = await request<RawAdminAwdChallengeData>({
    method: 'GET',
    url: `/authoring/awd-challenges/${encodeURIComponent(id)}`,
  })
  return normalizeAdminAwdChallenge(response)
}

export async function createAdminAwdChallenge(
  data: AdminAwdChallengeCreatePayload
): Promise<AdminAwdChallengeData> {
  const response = await request<RawAdminAwdChallengeData>({
    method: 'POST',
    url: '/authoring/awd-challenges',
    data,
  })
  return normalizeAdminAwdChallenge(response)
}

export async function updateAdminAwdChallenge(
  id: string,
  data: AdminAwdChallengeUpdatePayload
): Promise<AdminAwdChallengeData> {
  const response = await request<RawAdminAwdChallengeData>({
    method: 'PUT',
    url: `/authoring/awd-challenges/${encodeURIComponent(id)}`,
    data,
  })
  return normalizeAdminAwdChallenge(response)
}

export async function deleteAdminAwdChallenge(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/awd-challenges/${encodeURIComponent(id)}`,
  })
}

export async function previewAdminAwdChallengeImport(
  file: File
): Promise<AdminAwdChallengeImportPreview> {
  const formData = new FormData()
  formData.append('file', file)

  const response = await request<RawAdminAwdChallengeImportPreview>({
    method: 'POST',
    url: '/authoring/awd-challenge-imports',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return normalizeAdminAwdChallengeImportPreview(response)
}

export async function listAdminAwdChallengeImports(): Promise<AdminAwdChallengeImportPreview[]> {
  const response = await request<RawAdminAwdChallengeImportPreview[]>({
    method: 'GET',
    url: '/authoring/awd-challenge-imports',
  })
  return Array.isArray(response) ? response.map(normalizeAdminAwdChallengeImportPreview) : []
}

export async function commitAdminAwdChallengeImport(
  id: string
): Promise<AdminAwdChallengeImportCommitData> {
  const response = await request<RawAdminAwdChallengeImportCommitData>({
    method: 'POST',
    url: `/authoring/awd-challenge-imports/${encodeURIComponent(id)}/commit`,
  })

  return {
    challenge: normalizeAdminAwdChallenge(response.challenge),
  }
}

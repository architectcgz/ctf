import { request } from '../request'

import type {
  AdminContestAWDServiceCreatePayload,
  AdminContestAWDServiceUpdatePayload,
  AdminAWDCheckerPreviewPayload,
  AdminAWDTrafficEventsParams,
  RawAdminContestAWDServiceItem,
  RawAdminContestAWDInstanceOrchestration,
  RawAdminContestAWDInstanceItem,
  RawAdminContestAWDInstancePrewarm,
  RawAWDTrafficSummaryData,
  RawAWDTrafficEventItem,
  RawAWDCheckerPreviewData,
  AdminContestAWDServiceData,
  AdminContestAWDInstanceOrchestrationData,
  AdminContestAWDInstanceItemData,
  AdminContestAWDInstancePrewarmData,
  AWDTrafficSummaryData,
  AWDTrafficEventPageData,
  AWDCheckerPreviewData,
  PageResult,
} from './contest-support'

// Re-export types for barrel consumers
export type {
  AdminContestAWDServiceCreatePayload,
  AdminContestAWDServiceUpdatePayload,
  AdminAWDCheckerPreviewPayload,
  AdminAWDTrafficEventsParams,
} from './contest-support'

import {
  normalizeAdminContestAWDService,
  normalizeAdminContestAWDInstanceOrchestration,
  normalizeAdminContestAWDInstanceItem,
  normalizeAdminContestAWDInstancePrewarm,
  normalizeAWDTrafficSummary,
  normalizeAWDTrafficEvent,
  normalizeAWDCheckerPreview,
} from './contest-support'

export async function listContestAWDServices(
  contestId: string
): Promise<AdminContestAWDServiceData[]> {
  const response = await request<RawAdminContestAWDServiceItem[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/services`,
  })
  return response.map(normalizeAdminContestAWDService)
}

export async function createContestAWDService(
  contestId: string,
  data: AdminContestAWDServiceCreatePayload
): Promise<AdminContestAWDServiceData> {
  const response = await request<RawAdminContestAWDServiceItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/services`,
    data,
  })
  return normalizeAdminContestAWDService(response)
}

export async function updateContestAWDService(
  contestId: string,
  serviceId: string,
  data: AdminContestAWDServiceUpdatePayload
): Promise<void> {
  await request<void>({
    method: 'PUT',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}`,
    data,
  })
}

export async function deleteContestAWDService(
  contestId: string,
  serviceId: string
): Promise<void> {
  await request<void>({
    method: 'DELETE',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}`,
  })
}

export async function getContestAWDInstanceOrchestration(
  contestId: string
): Promise<AdminContestAWDInstanceOrchestrationData> {
  const response = await request<RawAdminContestAWDInstanceOrchestration>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/instances`,
  })
  return normalizeAdminContestAWDInstanceOrchestration(response)
}

export async function startContestAWDTeamServiceInstance(
  contestId: string,
  data: { team_id: string | number; service_id: string | number }
): Promise<AdminContestAWDInstanceItemData> {
  const response = await request<RawAdminContestAWDInstanceItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/instances`,
    data: {
      team_id: Number(data.team_id),
      service_id: Number(data.service_id),
    },
  })
  return normalizeAdminContestAWDInstanceItem(response)
}

export async function prewarmContestAWDInstances(
  contestId: string,
  data?: { team_id?: string | number }
): Promise<AdminContestAWDInstancePrewarmData> {
  const response = await request<RawAdminContestAWDInstancePrewarm>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/instances/prewarm`,
    data:
      data?.team_id == null
        ? {}
        : {
            team_id: Number(data.team_id),
          },
  })
  return normalizeAdminContestAWDInstancePrewarm(response)
}

export async function getContestAWDRoundTrafficSummary(
  contestId: string,
  roundId: string
): Promise<AWDTrafficSummaryData> {
  const response = await request<RawAWDTrafficSummaryData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/traffic/summary`,
  })
  return normalizeAWDTrafficSummary(response)
}

export async function listContestAWDRoundTrafficEvents(
  contestId: string,
  roundId: string,
  params?: AdminAWDTrafficEventsParams
): Promise<AWDTrafficEventPageData> {
  const response = await request<PageResult<RawAWDTrafficEventItem>>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/traffic/events`,
    params,
  })
  return {
    ...response,
    list: response.list.map(normalizeAWDTrafficEvent),
  }
}

export async function runContestAWDCheckerPreview(
  contestId: string,
  data: AdminAWDCheckerPreviewPayload
): Promise<AWDCheckerPreviewData> {
  const response = await request<RawAWDCheckerPreviewData>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/checker-preview`,
    data,
    timeout: 30000,
  })
  return normalizeAWDCheckerPreview(response)
}

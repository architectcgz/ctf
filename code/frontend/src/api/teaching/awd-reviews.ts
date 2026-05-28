import { request } from '../request'

import type {
  AwdReviewArchiveData,
  AwdReviewContestItemData,
  AwdReviewAttackItemData,
  AwdReviewRoundItemData,
  AwdReviewSelectedRoundData,
  AwdReviewServiceItemData,
  AwdReviewTeamItemData,
  AwdReviewTrafficItemData,
  PageResult,
  ReportExportData,
} from '../contracts'

interface RawAwdReviewContestItem extends Omit<AwdReviewContestItemData, 'id'> {
  id: string | number
}

interface RawAwdReviewRoundItem
  extends Omit<AwdReviewRoundItemData, 'id' | 'contest_id'> {
  id: string | number
  contest_id: string | number
}

interface RawAwdReviewTeamItem
  extends Omit<AwdReviewTeamItemData, 'team_id' | 'captain_id'> {
  team_id: string | number
  captain_id: string | number
}

interface RawAwdReviewServiceItem
  extends Omit<
    AwdReviewServiceItemData,
    'id' | 'round_id' | 'team_id' | 'service_id' | 'challenge_id'
  > {
  id: string | number
  round_id: string | number
  team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawAwdReviewAttackItem
  extends Omit<
    AwdReviewAttackItemData,
    'id' | 'round_id' | 'attacker_team_id' | 'victim_team_id' | 'service_id' | 'challenge_id'
  > {
  id: string | number
  round_id: string | number
  attacker_team_id: string | number
  victim_team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawAwdReviewTrafficItem
  extends Omit<
    AwdReviewTrafficItemData,
    | 'id'
    | 'contest_id'
    | 'round_id'
    | 'attacker_team_id'
    | 'victim_team_id'
    | 'service_id'
    | 'challenge_id'
  > {
  id: string | number
  contest_id: string | number
  round_id: string | number
  attacker_team_id: string | number
  victim_team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawAwdReviewSelectedRound
  extends Omit<AwdReviewSelectedRoundData, 'round' | 'teams' | 'services' | 'attacks' | 'traffic'> {
  round: RawAwdReviewRoundItem
  teams: RawAwdReviewTeamItem[]
  services: RawAwdReviewServiceItem[]
  attacks: RawAwdReviewAttackItem[]
  traffic: RawAwdReviewTrafficItem[]
}

interface RawAwdReviewArchiveResponse
  extends Omit<AwdReviewArchiveData, 'contest' | 'rounds' | 'selected_round' | 'scope'> {
  scope: {
    snapshot_type: string
    requested_by: number
    requested_id: string | number
    requested_role?: string
  }
  contest: RawAwdReviewContestItem
  rounds: RawAwdReviewRoundItem[]
  selected_round?: RawAwdReviewSelectedRound
}

interface RawAwdReviewContestPageResponse
  extends PageResult<RawAwdReviewContestItem> {
  summary?: {
    running_count?: number
    export_ready_count?: number
  }
}

export interface AwdReviewContestPageData
  extends PageResult<AwdReviewContestItemData> {
  summary: {
    running_count: number
    export_ready_count: number
  }
}

function normalizeReportExportData(
  payload: ReportExportData & { report_id: string | number }
): ReportExportData {
  return {
    ...payload,
    report_id: String(payload.report_id),
  }
}

function normalizeAwdReviewContest(
  item: RawAwdReviewContestItem
): AwdReviewContestItemData {
  return {
    ...item,
    id: String(item.id),
  }
}

function normalizeAwdReviewRound(
  item: RawAwdReviewRoundItem
): AwdReviewRoundItemData {
  return {
    ...item,
    id: String(item.id),
    contest_id: String(item.contest_id),
  }
}

function normalizeAwdReviewTeam(
  item: RawAwdReviewTeamItem
): AwdReviewTeamItemData {
  return {
    ...item,
    team_id: String(item.team_id),
    captain_id: String(item.captain_id),
  }
}

function normalizeAwdReviewService(
  item: RawAwdReviewServiceItem
): AwdReviewServiceItemData {
  return {
    ...item,
    id: String(item.id),
    round_id: String(item.round_id),
    team_id: String(item.team_id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeAwdReviewAttack(
  item: RawAwdReviewAttackItem
): AwdReviewAttackItemData {
  return {
    ...item,
    id: String(item.id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    victim_team_id: String(item.victim_team_id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeAwdReviewTraffic(
  item: RawAwdReviewTrafficItem
): AwdReviewTrafficItemData {
  return {
    ...item,
    id: String(item.id),
    contest_id: String(item.contest_id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    victim_team_id: String(item.victim_team_id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeAwdReviewSelectedRound(
  item: RawAwdReviewSelectedRound
): AwdReviewSelectedRoundData {
  return {
    round: normalizeAwdReviewRound(item.round),
    teams: item.teams.map(normalizeAwdReviewTeam),
    services: item.services.map(normalizeAwdReviewService),
    attacks: item.attacks.map(normalizeAwdReviewAttack),
    traffic: item.traffic.map(normalizeAwdReviewTraffic),
  }
}

export async function listTeacherAWDReviews(
  params?: {
    status?: AwdReviewContestItemData['status']
    keyword?: string
    page?: number
    page_size?: number
  },
  options?: { signal?: AbortSignal }
): Promise<AwdReviewContestPageData> {
  const payload = await request<RawAwdReviewContestPageResponse>({
    method: 'GET',
    url: '/teacher/awd/reviews',
    params: {
      status: params?.status,
      keyword: params?.keyword,
      page: params?.page,
      page_size: params?.page_size,
    },
    signal: options?.signal,
  })

  return {
    ...payload,
    list: payload.list.map(normalizeAwdReviewContest),
    summary: {
      running_count: payload.summary?.running_count ?? 0,
      export_ready_count: payload.summary?.export_ready_count ?? 0,
    },
  }
}

export async function getTeacherAWDReview(
  contestId: string,
  params?: {
    round?: number
    team_id?: string
  }
): Promise<AwdReviewArchiveData> {
  const payload = await request<RawAwdReviewArchiveResponse>({
    method: 'GET',
    url: `/teacher/awd/reviews/${encodeURIComponent(contestId)}`,
    params: {
      round: params?.round,
      team_id: params?.team_id,
    },
  })

  return {
    ...payload,
    scope: {
      ...payload.scope,
      requested_id: String(payload.scope.requested_id),
    },
    contest: normalizeAwdReviewContest(payload.contest),
    rounds: payload.rounds.map(normalizeAwdReviewRound),
    selected_round: payload.selected_round
      ? normalizeAwdReviewSelectedRound(payload.selected_round)
      : undefined,
  }
}

export async function exportTeacherAWDReviewArchive(
  contestId: string,
  data?: { round_number?: number }
): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({
    method: 'POST',
    url: `/teacher/awd/reviews/${encodeURIComponent(contestId)}/export/archive`,
    data,
  })

  return normalizeReportExportData(payload)
}

export async function exportTeacherAWDReviewReport(
  contestId: string,
  data?: { round_number?: number }
): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({
    method: 'POST',
    url: `/teacher/awd/reviews/${encodeURIComponent(contestId)}/export/report`,
    data,
  })

  return normalizeReportExportData(payload)
}

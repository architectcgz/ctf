import { request } from '../request'

import type {
  ReportExportData,
  TeacherAWDReviewArchiveData,
  TeacherAWDReviewAttackItemData,
  TeacherAWDReviewContestItemData,
  TeacherAWDReviewRoundItemData,
  TeacherAWDReviewSelectedRoundData,
  TeacherAWDReviewServiceItemData,
  TeacherAWDReviewTeamItemData,
  TeacherAWDReviewTrafficItemData,
} from '../contracts'

interface RawTeacherAWDReviewContestItem extends Omit<TeacherAWDReviewContestItemData, 'id'> {
  id: string | number
}

interface RawTeacherAWDReviewRoundItem extends Omit<
  TeacherAWDReviewRoundItemData,
  'id' | 'contest_id'
> {
  id: string | number
  contest_id: string | number
}

interface RawTeacherAWDReviewTeamItem extends Omit<
  TeacherAWDReviewTeamItemData,
  'team_id' | 'captain_id'
> {
  team_id: string | number
  captain_id: string | number
}

interface RawTeacherAWDReviewServiceItem extends Omit<
  TeacherAWDReviewServiceItemData,
  'id' | 'round_id' | 'team_id' | 'service_id' | 'challenge_id'
> {
  id: string | number
  round_id: string | number
  team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawTeacherAWDReviewAttackItem extends Omit<
  TeacherAWDReviewAttackItemData,
  'id' | 'round_id' | 'attacker_team_id' | 'victim_team_id' | 'service_id' | 'challenge_id'
> {
  id: string | number
  round_id: string | number
  attacker_team_id: string | number
  victim_team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawTeacherAWDReviewTrafficItem extends Omit<
  TeacherAWDReviewTrafficItemData,
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

interface RawTeacherAWDReviewSelectedRound extends Omit<
  TeacherAWDReviewSelectedRoundData,
  'round' | 'teams' | 'services' | 'attacks' | 'traffic'
> {
  round: RawTeacherAWDReviewRoundItem
  teams: RawTeacherAWDReviewTeamItem[]
  services: RawTeacherAWDReviewServiceItem[]
  attacks: RawTeacherAWDReviewAttackItem[]
  traffic: RawTeacherAWDReviewTrafficItem[]
}

interface RawTeacherAWDReviewArchiveResponse extends Omit<
  TeacherAWDReviewArchiveData,
  'contest' | 'rounds' | 'selected_round' | 'scope'
> {
  scope: {
    snapshot_type: string
    requested_by: number
    requested_id: string | number
    requested_role?: string
  }
  contest: RawTeacherAWDReviewContestItem
  rounds: RawTeacherAWDReviewRoundItem[]
  selected_round?: RawTeacherAWDReviewSelectedRound
}

function normalizeReportExportData(
  payload: ReportExportData & { report_id: string | number }
): ReportExportData {
  return {
    ...payload,
    report_id: String(payload.report_id),
  }
}

function normalizeTeacherAWDReviewContest(
  item: RawTeacherAWDReviewContestItem
): TeacherAWDReviewContestItemData {
  return {
    ...item,
    id: String(item.id),
  }
}

function normalizeTeacherAWDReviewRound(
  item: RawTeacherAWDReviewRoundItem
): TeacherAWDReviewRoundItemData {
  return {
    ...item,
    id: String(item.id),
    contest_id: String(item.contest_id),
  }
}

function normalizeTeacherAWDReviewTeam(
  item: RawTeacherAWDReviewTeamItem
): TeacherAWDReviewTeamItemData {
  return {
    ...item,
    team_id: String(item.team_id),
    captain_id: String(item.captain_id),
  }
}

function normalizeTeacherAWDReviewService(
  item: RawTeacherAWDReviewServiceItem
): TeacherAWDReviewServiceItemData {
  return {
    ...item,
    id: String(item.id),
    round_id: String(item.round_id),
    team_id: String(item.team_id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeTeacherAWDReviewAttack(
  item: RawTeacherAWDReviewAttackItem
): TeacherAWDReviewAttackItemData {
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

function normalizeTeacherAWDReviewTraffic(
  item: RawTeacherAWDReviewTrafficItem
): TeacherAWDReviewTrafficItemData {
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

function normalizeTeacherAWDSelectedRound(
  item: RawTeacherAWDReviewSelectedRound
): TeacherAWDReviewSelectedRoundData {
  return {
    round: normalizeTeacherAWDReviewRound(item.round),
    teams: item.teams.map(normalizeTeacherAWDReviewTeam),
    services: item.services.map(normalizeTeacherAWDReviewService),
    attacks: item.attacks.map(normalizeTeacherAWDReviewAttack),
    traffic: item.traffic.map(normalizeTeacherAWDReviewTraffic),
  }
}

export async function listTeacherAWDReviews(params?: {
  status?: TeacherAWDReviewContestItemData['status']
  keyword?: string
}): Promise<TeacherAWDReviewContestItemData[]> {
  const payload = await request<{ contests: RawTeacherAWDReviewContestItem[] }>({
    method: 'GET',
    url: '/teacher/awd/reviews',
    params: {
      status: params?.status,
      keyword: params?.keyword,
    },
  })

  return payload.contests.map(normalizeTeacherAWDReviewContest)
}

export async function getTeacherAWDReview(
  contestId: string,
  params?: {
    round?: number
    team_id?: string
  }
): Promise<TeacherAWDReviewArchiveData> {
  const payload = await request<RawTeacherAWDReviewArchiveResponse>({
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
    contest: normalizeTeacherAWDReviewContest(payload.contest),
    rounds: payload.rounds.map(normalizeTeacherAWDReviewRound),
    selected_round: payload.selected_round
      ? normalizeTeacherAWDSelectedRound(payload.selected_round)
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

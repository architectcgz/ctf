import { request } from './request'
import { normalizeInstanceData, type RawInstanceData } from './instance'

import type {
  AWDAttackLogData,
  AWDDefenseCommandData,
  AWDDefenseFileData,
  AWDDefenseFileSaveData,
  AWDDefenseSSHAccessData,
  AWDRoundData,
  ContestAnnouncement,
  ContestAWDWorkspaceData,
  ContestAWDWorkspaceRecentEventData,
  ContestAWDWorkspaceServiceData,
  ContestAWDWorkspaceTargetServiceData,
  ContestAWDWorkspaceTargetTeamData,
  ContestAWDWorkspaceTeamData,
  ContestChallengeItem,
  ContestDetailData,
  ContestListItem,
  ContestMyProgressData,
  ContestScoreboardData,
  InstanceData,
  PageResult,
  SubmitFlagData,
  TeamData,
} from './contracts'

type RawContestStatus = 'draft' | 'registration' | 'running' | 'frozen' | 'ended'

interface RawContestItem {
  id: string | number
  title: string
  description?: string
  mode: ContestDetailData['mode']
  status: RawContestStatus
  start_time: string
  end_time: string
  freeze_time?: string | null
}

interface RawTeamMember {
  user_id: string | number
  username: string
  joined_at: string
}

interface RawTeamData {
  id: string | number
  name: string
  invite_code?: string
  captain_user_id: string | number
  members: RawTeamMember[]
}

interface RawTeamCreateResp {
  id: string | number
  name: string
  contest_id: string | number
  captain_id: string | number
  invite_code?: string
  max_members: number
  member_count: number
  created_at: string
}

interface RawContestChallengeItem extends Omit<
  ContestChallengeItem,
  'id' | 'challenge_id' | 'awd_service_id'
> {
  id: string | number
  challenge_id: string | number
  awd_service_id?: string | number | null
}

interface RawAWDRoundData extends Omit<AWDRoundData, 'id' | 'contest_id'> {
  id: string | number
  contest_id: string | number
}

interface RawAWDAttackLogData extends Omit<
  AWDAttackLogData,
  'id' | 'round_id' | 'attacker_team_id' | 'victim_team_id' | 'service_id' | 'challenge_id'
> {
  id: string | number
  round_id: string | number
  attacker_team_id: string | number
  victim_team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawContestAWDWorkspaceTeamData extends Omit<ContestAWDWorkspaceTeamData, 'team_id'> {
  team_id: string | number
}

interface RawContestAWDWorkspaceServiceData extends Omit<
  ContestAWDWorkspaceServiceData,
  'service_id' | 'challenge_id' | 'instance_id'
> {
  service_id?: string | number
  challenge_id: string | number
  instance_id?: string | number
}

interface RawContestAWDWorkspaceTargetServiceData extends Omit<
  ContestAWDWorkspaceTargetServiceData,
  'service_id' | 'challenge_id' | 'reachable'
> {
  service_id?: string | number
  challenge_id: string | number
  reachable?: boolean
  access_url?: string
}

interface RawContestAWDWorkspaceTargetTeamData extends Omit<
  ContestAWDWorkspaceTargetTeamData,
  'team_id' | 'services'
> {
  team_id: string | number
  services: RawContestAWDWorkspaceTargetServiceData[]
}

interface RawContestAWDWorkspaceRecentEventData extends Omit<
  ContestAWDWorkspaceRecentEventData,
  'id' | 'service_id' | 'challenge_id' | 'peer_team_id'
> {
  id: string | number
  service_id?: string | number
  challenge_id: string | number
  peer_team_id: string | number
}

interface RawContestAWDWorkspaceData extends Omit<
  ContestAWDWorkspaceData,
  'contest_id' | 'current_round' | 'my_team' | 'services' | 'targets' | 'recent_events'
> {
  contest_id: string | number
  current_round?: RawAWDRoundData
  my_team?: RawContestAWDWorkspaceTeamData | null
  services: RawContestAWDWorkspaceServiceData[]
  targets: RawContestAWDWorkspaceTargetTeamData[]
  recent_events: RawContestAWDWorkspaceRecentEventData[]
}

export type GetContestsData = PageResult<ContestListItem>

function normalizeContestStatus(status: RawContestStatus): ContestListItem['status'] {
  if (status === 'registration') {
    return 'registering'
  }
  return status
}

function normalizeContest(item: RawContestItem): ContestDetailData {
  return {
    id: String(item.id),
    title: item.title,
    description: item.description,
    mode: item.mode,
    status: normalizeContestStatus(item.status),
    starts_at: item.start_time,
    ends_at: item.end_time,
    scoreboard_frozen: Boolean(item.freeze_time),
  }
}

function normalizeContestChallenge(item: RawContestChallengeItem): ContestChallengeItem {
  return {
    ...item,
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    awd_service_id: item.awd_service_id == null ? undefined : String(item.awd_service_id),
  }
}

function normalizeTeam(payload: RawTeamData | null): TeamData | null {
  if (!payload) {
    return null
  }

  return {
    id: String(payload.id),
    name: payload.name,
    invite_code: payload.invite_code,
    captain_user_id: String(payload.captain_user_id),
    members: payload.members.map((member) => ({
      user_id: String(member.user_id),
      username: member.username,
      joined_at: member.joined_at,
    })),
  }
}

function normalizeAWDRound(item: RawAWDRoundData): AWDRoundData {
  return {
    ...item,
    id: String(item.id),
    contest_id: String(item.contest_id),
  }
}

function normalizeAWDAttackLog(item: RawAWDAttackLogData): AWDAttackLogData {
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

function normalizeContestAWDWorkspaceTeam(
  item: RawContestAWDWorkspaceTeamData
): ContestAWDWorkspaceTeamData {
  return {
    ...item,
    team_id: String(item.team_id),
  }
}

function normalizeContestAWDWorkspaceService(
  item: RawContestAWDWorkspaceServiceData
): ContestAWDWorkspaceServiceData {
  return {
    ...item,
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
    instance_id: item.instance_id == null ? undefined : String(item.instance_id),
  }
}

function normalizeContestAWDWorkspaceTargetService(
  item: RawContestAWDWorkspaceTargetServiceData
): ContestAWDWorkspaceTargetServiceData {
  return {
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
    reachable: item.reachable ?? Boolean(item.access_url),
  }
}

function normalizeContestAWDWorkspaceTargetTeam(
  item: RawContestAWDWorkspaceTargetTeamData
): ContestAWDWorkspaceTargetTeamData {
  return {
    ...item,
    team_id: String(item.team_id),
    services: item.services.map(normalizeContestAWDWorkspaceTargetService),
  }
}

function normalizeContestAWDWorkspaceEvent(
  item: RawContestAWDWorkspaceRecentEventData
): ContestAWDWorkspaceRecentEventData {
  return {
    ...item,
    id: String(item.id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
    peer_team_id: String(item.peer_team_id),
  }
}

export async function getContests(params?: Record<string, unknown>): Promise<GetContestsData> {
  const response = await request<PageResult<RawContestItem>>({
    method: 'GET',
    url: '/contests',
    params,
  })
  return {
    ...response,
    list: response.list.map(normalizeContest),
  }
}

export async function getContestDetail(id: string): Promise<ContestDetailData> {
  const response = await request<RawContestItem>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}`,
  })
  return normalizeContest(response)
}

export async function registerContest(id: string) {
  return request<void>({ method: 'POST', url: `/contests/${encodeURIComponent(id)}/register` })
}

export async function getContestChallenges(id: string): Promise<ContestChallengeItem[]> {
  const response = await request<RawContestChallengeItem[]>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/challenges`,
  })
  return response.map(normalizeContestChallenge)
}

export async function submitContestFlag(
  contestId: string,
  contestChallengeId: string,
  flag: string
): Promise<SubmitFlagData> {
  return request<SubmitFlagData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/challenges/${encodeURIComponent(contestChallengeId)}/submissions`,
    data: { flag },
  })
}

export async function getScoreboard(
  id: string,
  params?: Record<string, unknown>
): Promise<ContestScoreboardData> {
  return request<ContestScoreboardData>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/scoreboard`,
    params,
  })
}

export async function getAnnouncements(id: string): Promise<ContestAnnouncement[]> {
  return request<ContestAnnouncement[]>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/announcements`,
  })
}

export async function createTeam(id: string, data: Record<string, unknown>): Promise<void> {
  await request<RawTeamCreateResp>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(id)}/teams`,
    data,
  })
}

export async function joinTeam(contestId: string, teamId: string) {
  return request<void>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/join`,
  })
}

export async function getMyProgress(id: string): Promise<ContestMyProgressData> {
  return request<ContestMyProgressData>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/my-progress`,
  })
}

export async function getMyTeam(contestId: string): Promise<TeamData | null> {
  const response = await request<RawTeamData | null>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(contestId)}/my-team`,
  })
  return normalizeTeam(response)
}

export async function getContestAWDWorkspace(contestId: string): Promise<ContestAWDWorkspaceData> {
  const response = await request<RawContestAWDWorkspaceData>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(contestId)}/awd/workspace`,
  })

  return {
    contest_id: String(response.contest_id),
    current_round: response.current_round ? normalizeAWDRound(response.current_round) : undefined,
    my_team: response.my_team
      ? normalizeContestAWDWorkspaceTeam(response.my_team)
      : response.my_team,
    services: response.services.map(normalizeContestAWDWorkspaceService),
    targets: response.targets.map(normalizeContestAWDWorkspaceTargetTeam),
    recent_events: response.recent_events.map(normalizeContestAWDWorkspaceEvent),
  }
}

export async function startContestAWDServiceInstance(
  contestId: string,
  serviceId: string
): Promise<InstanceData> {
  const payload = await request<RawInstanceData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/instances`,
    suppressErrorToast: true,
  })
  return normalizeInstanceData(payload)
}

export async function submitContestAWDAttack(
  contestId: string,
  serviceId: string,
  data: { victim_team_id: number; flag: string }
): Promise<AWDAttackLogData> {
  const response = await request<RawAWDAttackLogData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/submissions`,
    data,
  })
  return normalizeAWDAttackLog(response)
}

export async function requestContestAWDTargetAccess(
  contestId: string,
  serviceId: string,
  victimTeamId: string
): Promise<{ access_url: string }> {
  return request<{ access_url: string }>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/targets/${encodeURIComponent(victimTeamId)}/access`,
  })
}

export async function requestContestAWDDefenseSSH(
  contestId: string,
  serviceId: string
): Promise<AWDDefenseSSHAccessData> {
  return request<AWDDefenseSSHAccessData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/defense/ssh`,
  })
}

export async function readContestAWDDefenseFile(
  contestId: string,
  serviceId: string,
  path: string
): Promise<AWDDefenseFileData> {
  return request<AWDDefenseFileData>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/defense/files`,
    params: { path },
  })
}

export async function saveContestAWDDefenseFile(
  contestId: string,
  serviceId: string,
  data: { path: string; content: string; backup: boolean }
): Promise<AWDDefenseFileSaveData> {
  return request<AWDDefenseFileSaveData>({
    method: 'PUT',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/defense/files`,
    data,
  })
}

export async function runContestAWDDefenseCommand(
  contestId: string,
  serviceId: string,
  command: string
): Promise<AWDDefenseCommandData> {
  return request<AWDDefenseCommandData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/awd/services/${encodeURIComponent(serviceId)}/defense/commands`,
    data: { command },
  })
}

export async function kickTeamMember(contestId: string, teamId: string, userId: string) {
  return request<void>({
    method: 'DELETE',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/members/${encodeURIComponent(userId)}`,
  })
}

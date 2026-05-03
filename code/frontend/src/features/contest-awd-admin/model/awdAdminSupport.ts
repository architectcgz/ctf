import { ApiError } from '@/api/request'
import type {
  AWDReadinessAction,
  AWDReadinessData,
  AWDRoundData,
  AWDTrafficStatusGroup,
  AdminContestAWDInstanceOrchestrationData,
} from '@/api/contracts'

const ERR_AWD_READINESS_BLOCKED = 14025
const AWD_TRAFFIC_DEFAULT_PAGE_SIZE = 20

export interface AWDTrafficFilterState {
  attacker_team_id: string
  victim_team_id: string
  service_id: string
  awd_challenge_id: string
  status_group: 'all' | AWDTrafficStatusGroup
  path_keyword: string
  page: number
  page_size: number
}

export interface AWDReadinessOverrideDialogState {
  open: boolean
  action: Extract<AWDReadinessAction, 'create_round' | 'run_current_round_check'> | null
  title: string
  readiness: AWDReadinessData | null
  confirmLoading: boolean
  pendingRoundPayload?: {
    round_number: number
    status?: AWDRoundData['status']
    attack_score?: number
    defense_score?: number
  }
}

export function createEmptyInstanceOrchestration(): AdminContestAWDInstanceOrchestrationData {
  return {
    contest_id: '',
    teams: [],
    services: [],
    instances: [],
  }
}

export function createDefaultTrafficFilters(): AWDTrafficFilterState {
  return {
    attacker_team_id: '',
    victim_team_id: '',
    service_id: '',
    awd_challenge_id: '',
    status_group: 'all',
    path_keyword: '',
    page: 1,
    page_size: AWD_TRAFFIC_DEFAULT_PAGE_SIZE,
  }
}

export function createDefaultOverrideDialogState(): AWDReadinessOverrideDialogState {
  return {
    open: false,
    action: null,
    title: '',
    readiness: null,
    confirmLoading: false,
  }
}

function getSelectedRoundStorageKey(contestId: string): string {
  return `ctf_admin_awd_selected_round:${contestId}`
}

export function loadStoredSelectedRoundId(contestId: string): string | null {
  if (typeof window === 'undefined') {
    return null
  }
  const value = window.sessionStorage.getItem(getSelectedRoundStorageKey(contestId))
  return value?.trim() || null
}

export function persistSelectedRoundId(contestId: string, roundId: string | null): void {
  if (typeof window === 'undefined') {
    return
  }
  const storageKey = getSelectedRoundStorageKey(contestId)
  if (roundId) {
    window.sessionStorage.setItem(storageKey, roundId)
    return
  }
  window.sessionStorage.removeItem(storageKey)
}

export function pickRoundId(
  rounds: AWDRoundData[],
  currentRoundId: string | null,
  preferredRoundId?: string
): string | null {
  if (preferredRoundId && rounds.some((item) => item.id === preferredRoundId)) {
    return preferredRoundId
  }
  if (currentRoundId && rounds.some((item) => item.id === currentRoundId)) {
    return currentRoundId
  }
  const runningRound = rounds.find((item) => item.status === 'running')
  return runningRound?.id || rounds[rounds.length - 1]?.id || null
}

export function isAWDReadinessBlockedError(error: unknown): boolean {
  return error instanceof ApiError && error.code === ERR_AWD_READINESS_BLOCKED
}

export function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

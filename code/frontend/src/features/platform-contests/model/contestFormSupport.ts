import type { AdminContestUpdatePayload } from '@/api/admin/contests'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'

export type PlatformContestStatus = Extract<
  ContestStatus,
  'draft' | 'registering' | 'running' | 'frozen' | 'ended'
>

export interface ContestFormDraft {
  title: string
  description: string
  mode: 'jeopardy' | 'awd'
  starts_at: string
  ends_at: string
  status: PlatformContestStatus
}

export interface ContestFieldLocks {
  mode: boolean
  starts_at: boolean
  ends_at: boolean
}

const STATUS_TRANSITIONS: Record<PlatformContestStatus, PlatformContestStatus[]> = {
  draft: ['draft', 'registering'],
  registering: ['registering', 'draft', 'running'],
  running: ['running', 'frozen', 'ended'],
  frozen: ['frozen', 'ended'],
  ended: ['ended'],
}

export function toDateTimeLocal(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const offset = date.getTimezoneOffset()
  const localDate = new Date(date.getTime() - offset * 60_000)
  return localDate.toISOString().slice(0, 16)
}

export function toISOString(value: string): string {
  return new Date(value).toISOString()
}

export function createEmptyDraft(): ContestFormDraft {
  const startAt = new Date()
  startAt.setMinutes(0, 0, 0)
  startAt.setHours(startAt.getHours() + 1)

  const endAt = new Date(startAt)
  endAt.setHours(endAt.getHours() + 2)

  return {
    title: '',
    description: '',
    mode: 'jeopardy',
    starts_at: toDateTimeLocal(startAt.toISOString()),
    ends_at: toDateTimeLocal(endAt.toISOString()),
    status: 'draft',
  }
}

export function createDraftFromContest(contest: ContestDetailData): ContestFormDraft {
  return {
    title: contest.title,
    description: contest.description || '',
    mode: contest.mode === 'awd' ? 'awd' : 'jeopardy',
    starts_at: toDateTimeLocal(contest.starts_at),
    ends_at: toDateTimeLocal(contest.ends_at),
    status: normalizeEditableStatus(contest.status),
  }
}

export function normalizeEditableStatus(status: ContestStatus): PlatformContestStatus {
  if (
    status === 'draft' ||
    status === 'registering' ||
    status === 'running' ||
    status === 'frozen' ||
    status === 'ended'
  ) {
    return status
  }
  return 'draft'
}

export function createFieldLocks(status: PlatformContestStatus | null): ContestFieldLocks {
  if (!status) {
    return {
      mode: false,
      starts_at: false,
      ends_at: false,
    }
  }

  return {
    mode: status !== 'draft',
    starts_at: status === 'registering' || status === 'running' || status === 'ended',
    ends_at: status === 'running' || status === 'ended',
  }
}

export function createContestStatusOptions(status: PlatformContestStatus | null) {
  if (!status) {
    return []
  }
  return STATUS_TRANSITIONS[status].map((nextStatus) => ({ label: nextStatus, value: nextStatus }))
}

export function shouldConfirmContestTermination(
  currentStatus: PlatformContestStatus | null,
  targetStatus: PlatformContestStatus
): boolean {
  return targetStatus === 'ended' && (currentStatus === 'running' || currentStatus === 'frozen')
}

export function buildContestUpdatePayload(
  draft: ContestFormDraft,
  fieldLocks: ContestFieldLocks
): AdminContestUpdatePayload {
  const payload: AdminContestUpdatePayload = {
    title: draft.title.trim(),
    description: draft.description.trim(),
    mode: draft.mode,
    status: draft.status,
  }

  if (!fieldLocks.starts_at) {
    payload.starts_at = toISOString(draft.starts_at)
  }
  if (!fieldLocks.ends_at) {
    payload.ends_at = toISOString(draft.ends_at)
  }

  return payload
}

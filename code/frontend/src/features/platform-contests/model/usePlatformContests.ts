import { computed, ref, watch } from 'vue'

import {
  getContests,
  type AdminContestUpdatePayload,
} from '@/api/admin/contests'
import { ApiError } from '@/api/request'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import {
  useAwdStartOverrideFlow,
} from './useAwdStartOverrideFlow'
import { useContestDialogState } from './useContestDialogState'
import { useContestSaveFlow } from './useContestSaveFlow'

export type PlatformContestStatus = Extract<
  ContestStatus,
  'draft' | 'registering' | 'running' | 'frozen' | 'ended'
>
type StatusFilter = 'all' | PlatformContestStatus
const ERR_AWD_READINESS_BLOCKED = 14025

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

function toDateTimeLocal(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const offset = date.getTimezoneOffset()
  const localDate = new Date(date.getTime() - offset * 60_000)
  return localDate.toISOString().slice(0, 16)
}

function toISOString(value: string): string {
  return new Date(value).toISOString()
}

function createEmptyDraft(): ContestFormDraft {
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

export async function confirmContestTermination(contestTitle: string): Promise<boolean> {
  const normalizedTitle = contestTitle.trim()
  return confirmDestructiveAction({
    title: '确认结束赛事',
    message: normalizedTitle
      ? `结束赛事“${normalizedTitle}”后，学生端将立即停止参赛入口，当前进行中的作答与攻防操作也会一并终止。确认继续吗？`
      : '结束当前赛事后，学生端将立即停止参赛入口，当前进行中的作答与攻防操作也会一并终止。确认继续吗？',
    confirmButtonText: '确认结束',
    cancelButtonText: '继续编辑',
  })
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
function shouldGateAWDContestStart(
  mode: ContestFormDraft['mode'],
  targetStatus: PlatformContestStatus
): boolean {
  return mode === 'awd' && targetStatus === 'running'
}

function isAWDReadinessBlockedError(error: unknown): boolean {
  return error instanceof ApiError && error.code === ERR_AWD_READINESS_BLOCKED
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

export function usePlatformContests() {
  const toast = useToast()
  const statusFilter = ref<StatusFilter>('all')
  const {
    dialogOpen,
    editingContestId,
    editingBaseStatus,
    formDraft,
    prepareCreateContest: prepareCreateContestState,
    openCreateDialog: openCreateDialogState,
    openEditDialog: openEditDialogState,
    closeDialog: closeDialogState,
  } = useContestDialogState({
    createEmptyDraft,
    createDraftFromContest,
    normalizeEditableStatus,
  })

  const pagination = usePagination<ContestDetailData>(({ page, page_size }) =>
    getContests({
      page,
      page_size,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
    })
  )

  const mode = computed<'create' | 'edit'>(() => (editingContestId.value ? 'edit' : 'create'))
  const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
  const statusOptions = computed(() => createContestStatusOptions(editingBaseStatus.value))

  watch(statusFilter, async () => {
    await pagination.changePage(1)
  })

  async function finalizeContestUpdateSuccess() {
    toast.success('竞赛已更新')
    closeDialog()
    await pagination.refresh()
  }
  const {
    awdStartOverrideDialogState,
    openAWDStartOverrideDialog,
    closeAWDStartOverrideDialog,
    confirmAWDStartOverride,
  } = useAwdStartOverrideFlow({
    editingContestId,
    onContestUpdated: finalizeContestUpdateSuccess,
    isBlockedError: isAWDReadinessBlockedError,
    humanizeRequestError,
    notifyError: (message) => {
      toast.error(message)
    },
  })

  function prepareCreateContest() {
    prepareCreateContestState()
    closeAWDStartOverrideDialog()
  }

  function openCreateDialog() {
    openCreateDialogState()
    closeAWDStartOverrideDialog()
  }

  function openEditDialog(contest: ContestDetailData) {
    openEditDialogState(contest)
    closeAWDStartOverrideDialog()
  }

  function closeDialog() {
    closeDialogState()
    closeAWDStartOverrideDialog()
  }

  const { saving, saveContest } = useContestSaveFlow({
    editingContestId,
    editingBaseStatus,
    fieldLocks,
    closeDialog,
    refreshContests: pagination.refresh,
    openAWDStartOverrideDialog,
    shouldConfirmContestTermination,
    confirmContestTermination,
    buildContestUpdatePayload,
    shouldGateAWDContestStart,
    isAWDReadinessBlockedError,
    toISOString,
    humanizeRequestError,
    notifySuccess: (message) => {
      toast.success(message)
    },
    notifyError: (message) => {
      toast.error(message)
    },
  })

  return {
    ...pagination,
    statusFilter,
    dialogOpen,
    mode,
    saving,
    formDraft,
    fieldLocks,
    statusOptions,
    awdStartOverrideDialogState,
    prepareCreateContest,
    openCreateDialog,
    openEditDialog,
    closeDialog,
    closeAWDStartOverrideDialog,
    confirmAWDStartOverride,
    saveContest,
  }
}

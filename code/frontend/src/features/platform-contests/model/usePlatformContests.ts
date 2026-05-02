import { computed, ref, watch } from 'vue'

import {
  createContest,
  getContests,
  updateContest,
  type AdminContestCreatePayload,
  type AdminContestUpdatePayload,
} from '@/api/admin/contests'
import { ApiError } from '@/api/request'
import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import {
  createDefaultAWDStartOverrideDialogState,
  useAwdStartOverrideFlow,
} from './useAwdStartOverrideFlow'

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
  mode: ContestDetailData['mode'] | null,
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
  const dialogOpen = ref(false)
  const saving = ref(false)
  const editingContestId = ref<string | null>(null)
  const editingBaseStatus = ref<PlatformContestStatus | null>(null)
  const formDraft = ref<ContestFormDraft>(createEmptyDraft())

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

  function prepareCreateContest() {
    editingContestId.value = null
    editingBaseStatus.value = null
    formDraft.value = createEmptyDraft()
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
    dialogOpen.value = false
  }

  function openCreateDialog() {
    prepareCreateContest()
    dialogOpen.value = true
  }

  function openEditDialog(contest: ContestDetailData) {
    editingContestId.value = contest.id
    editingBaseStatus.value = normalizeEditableStatus(contest.status)
    formDraft.value = createDraftFromContest(contest)
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
    dialogOpen.value = true
  }

  function closeDialog() {
    dialogOpen.value = false
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
  }

  async function finalizeContestUpdateSuccess() {
    toast.success('竞赛已更新')
    dialogOpen.value = false
    closeAWDStartOverrideDialog()
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

  async function saveContest(draft: ContestFormDraft): Promise<'create' | 'edit' | null> {
    const title = draft.title.trim()
    const description = draft.description.trim()

    if (!title) {
      toast.error('请填写竞赛标题')
      return null
    }

    saving.value = true
    try {
      if (editingContestId.value) {
        if (shouldConfirmContestTermination(editingBaseStatus.value, draft.status)) {
          const confirmed = await confirmContestTermination(title)
          if (!confirmed) {
            return null
          }
        }

        const payload = buildContestUpdatePayload(draft, fieldLocks.value)
        if (shouldGateAWDContestStart(draft.mode, draft.status)) {
          try {
            await updateContest(editingContestId.value, payload)
            await finalizeContestUpdateSuccess()
            return 'edit'
          } catch (error) {
            if (isAWDReadinessBlockedError(error)) {
              await openAWDStartOverrideDialog(payload)
              return null
            }
            toast.error(humanizeRequestError(error, '竞赛更新失败'))
          }
          return null
        }

        await updateContest(editingContestId.value, payload)
        await finalizeContestUpdateSuccess()
        return 'edit'
      } else {
        const payload: AdminContestCreatePayload = {
          title,
          description,
          mode: draft.mode,
          starts_at: toISOString(draft.starts_at),
          ends_at: toISOString(draft.ends_at),
        }
        await createContest(payload)
        toast.success('竞赛已创建')
        dialogOpen.value = false
        await pagination.refresh()
        return 'create'
      }
    } catch (error) {
      if (!(error instanceof ApiError)) {
        toast.error(
          humanizeRequestError(error, editingContestId.value ? '竞赛更新失败' : '竞赛创建失败')
        )
      }
      return null
    } finally {
      saving.value = false
    }
  }

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

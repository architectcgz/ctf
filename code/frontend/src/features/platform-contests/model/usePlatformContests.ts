import { computed, ref, watch } from 'vue'

import {
  getContests,
} from '@/api/admin/contests'
import { ApiError } from '@/api/request'
import type { ContestDetailData } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import {
  useAwdStartOverrideFlow,
} from './useAwdStartOverrideFlow'
import {
  buildContestUpdatePayload,
  createContestStatusOptions,
  createDraftFromContest,
  createEmptyDraft,
  createFieldLocks,
  normalizeEditableStatus,
  shouldConfirmContestTermination,
  toISOString,
  type PlatformContestStatus,
} from './contestFormSupport'
import { useContestDialogState } from './useContestDialogState'
import { useContestSaveFlow } from './useContestSaveFlow'

type StatusFilter = 'all' | PlatformContestStatus
const ERR_AWD_READINESS_BLOCKED = 14025

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

function shouldGateAWDContestStart(
  mode: 'jeopardy' | 'awd',
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

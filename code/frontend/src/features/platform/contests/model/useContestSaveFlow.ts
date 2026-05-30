import { ref, type ComputedRef, type Ref } from 'vue'

import {
  createContest,
  updateContest,
  type AdminContestCreatePayload,
  type AdminContestUpdatePayload,
} from '@/api/admin/contests'
import { ApiError } from '@/api/request'

import type {
  ContestFieldLocks,
  ContestFormDraft,
  PlatformContestStatus,
} from './contestFormSupport'

interface UseContestSaveFlowOptions {
  editingContestId: Ref<string | null>
  editingBaseStatus: Ref<PlatformContestStatus | null>
  fieldLocks: ComputedRef<ContestFieldLocks>
  closeDialog: () => void
  refreshContests: () => Promise<void>
  openAWDStartOverrideDialog: (payload: AdminContestUpdatePayload) => Promise<void>
  shouldConfirmContestTermination: (
    currentStatus: PlatformContestStatus | null,
    targetStatus: PlatformContestStatus
  ) => boolean
  confirmContestTermination: (contestTitle: string) => Promise<boolean>
  buildContestUpdatePayload: (
    draft: ContestFormDraft,
    fieldLocks: ContestFieldLocks
  ) => AdminContestUpdatePayload
  shouldGateAWDContestStart: (
    mode: ContestFormDraft['mode'],
    targetStatus: PlatformContestStatus
  ) => boolean
  isAWDReadinessBlockedError: (error: unknown) => boolean
  toISOString: (value: string) => string
  humanizeRequestError: (error: unknown, fallback: string) => string
  notifySuccess: (message: string) => void
  notifyError: (message: string) => void
}

export function useContestSaveFlow(options: UseContestSaveFlowOptions) {
  const {
    editingContestId,
    editingBaseStatus,
    fieldLocks,
    closeDialog,
    refreshContests,
    openAWDStartOverrideDialog,
    shouldConfirmContestTermination,
    confirmContestTermination,
    buildContestUpdatePayload,
    shouldGateAWDContestStart,
    isAWDReadinessBlockedError,
    toISOString,
    humanizeRequestError,
    notifySuccess,
    notifyError,
  } = options

  const saving = ref(false)

  async function saveContest(draft: ContestFormDraft): Promise<'create' | 'edit' | null> {
    const title = draft.title.trim()
    const description = draft.description.trim()

    if (!title) {
      notifyError('请填写竞赛标题')
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
            notifySuccess('竞赛已更新')
            closeDialog()
            await refreshContests()
            return 'edit'
          } catch (error) {
            if (isAWDReadinessBlockedError(error)) {
              await openAWDStartOverrideDialog(payload)
              return null
            }
            notifyError(humanizeRequestError(error, '竞赛更新失败'))
          }
          return null
        }

        await updateContest(editingContestId.value, payload)
        notifySuccess('竞赛已更新')
        closeDialog()
        await refreshContests()
        return 'edit'
      }

      const payload: AdminContestCreatePayload = {
        title,
        description,
        mode: draft.mode,
        starts_at: toISOString(draft.starts_at),
        ends_at: toISOString(draft.ends_at),
      }
      await createContest(payload)
      notifySuccess('竞赛已创建')
      closeDialog()
      await refreshContests()
      return 'create'
    } catch (error) {
      if (!(error instanceof ApiError)) {
        notifyError(humanizeRequestError(error, editingContestId.value ? '竞赛更新失败' : '竞赛创建失败'))
      }
      return null
    } finally {
      saving.value = false
    }
  }

  return {
    saving,
    saveContest,
  }
}

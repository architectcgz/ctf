import { ref, type Ref } from 'vue'

import { getContestAWDReadiness, updateContest, type AdminContestUpdatePayload } from '@/api/admin/contests'
import type { AWDReadinessData } from '@/api/contracts'

export interface AWDStartOverrideDialogState {
  open: boolean
  title: string
  readiness: AWDReadinessData | null
  confirmLoading: boolean
  pendingPayload: AdminContestUpdatePayload | null
}

interface UseAwdStartOverrideFlowOptions {
  editingContestId: Ref<string | null>
  onContestUpdated: () => Promise<void>
  isBlockedError: (error: unknown) => boolean
  humanizeRequestError: (error: unknown, fallback: string) => string
  notifyError: (message: string) => void
}

export function createDefaultAWDStartOverrideDialogState(): AWDStartOverrideDialogState {
  return {
    open: false,
    title: '',
    readiness: null,
    confirmLoading: false,
    pendingPayload: null,
  }
}

export function useAwdStartOverrideFlow(options: UseAwdStartOverrideFlowOptions) {
  const {
    editingContestId,
    onContestUpdated,
    isBlockedError,
    humanizeRequestError,
    notifyError,
  } = options

  const awdStartOverrideDialogState = ref<AWDStartOverrideDialogState>(
    createDefaultAWDStartOverrideDialogState()
  )

  async function openAWDStartOverrideDialog(payload: AdminContestUpdatePayload) {
    if (!editingContestId.value) {
      return
    }
    const readiness = await getContestAWDReadiness(editingContestId.value)
    awdStartOverrideDialogState.value = {
      open: true,
      title: '启动赛事',
      readiness,
      confirmLoading: false,
      pendingPayload: payload,
    }
  }

  function closeAWDStartOverrideDialog() {
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
  }

  async function confirmAWDStartOverride(reason: string) {
    const contestId = editingContestId.value
    const payload = awdStartOverrideDialogState.value.pendingPayload
    const normalizedReason = reason.trim()
    if (!contestId || !payload || !normalizedReason) {
      return
    }

    awdStartOverrideDialogState.value = {
      ...awdStartOverrideDialogState.value,
      confirmLoading: true,
    }

    try {
      await updateContest(contestId, {
        ...payload,
        force_override: true,
        override_reason: normalizedReason,
      })
      await onContestUpdated()
    } catch (error) {
      if (isBlockedError(error)) {
        await openAWDStartOverrideDialog(payload)
        return
      }
      notifyError(humanizeRequestError(error, '竞赛更新失败'))
    } finally {
      if (awdStartOverrideDialogState.value.open) {
        awdStartOverrideDialogState.value = {
          ...awdStartOverrideDialogState.value,
          confirmLoading: false,
        }
      }
    }
  }

  return {
    awdStartOverrideDialogState,
    openAWDStartOverrideDialog,
    closeAWDStartOverrideDialog,
    confirmAWDStartOverride,
  }
}

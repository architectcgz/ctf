import { ref, type Ref } from 'vue'

import {
  createContestAWDRound,
  getContestAWDReadiness,
  runContestAWDCurrentRoundCheck,
} from '@/api/admin/contests'
import type {
  AWDReadinessAction,
  AWDReadinessData,
  AWDRoundData,
  ContestDetailData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

import {
  createDefaultOverrideDialogState,
  humanizeRequestError,
  isAWDReadinessBlockedError,
  type AWDReadinessOverrideDialogState,
} from './awdAdminSupport'

interface UseAwdReadinessDecisionOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  onAfterOverride: (preferredRoundId?: string) => Promise<void>
}

export function useAwdReadinessDecision(options: UseAwdReadinessDecisionOptions) {
  const { selectedContest, onAfterOverride } = options
  const toast = useToast()

  const readiness = ref<AWDReadinessData | null>(null)
  const loadingReadiness = ref(false)
  const overrideDialogState = ref<AWDReadinessOverrideDialogState>(
    createDefaultOverrideDialogState()
  )

  function resetOverrideDialog() {
    overrideDialogState.value = createDefaultOverrideDialogState()
  }

  async function refreshReadiness() {
    if (!selectedContest.value || selectedContest.value.mode !== 'awd') {
      readiness.value = null
      resetOverrideDialog()
      return null
    }

    loadingReadiness.value = true
    try {
      const nextReadiness = await getContestAWDReadiness(selectedContest.value.id)
      readiness.value = nextReadiness
      return nextReadiness
    } finally {
      loadingReadiness.value = false
    }
  }

  async function openOverrideDialog(
    action: Extract<AWDReadinessAction, 'create_round' | 'run_current_round_check'>,
    title: string,
    pendingRoundPayload?: AWDReadinessOverrideDialogState['pendingRoundPayload']
  ) {
    const snapshot = await refreshReadiness()
    overrideDialogState.value = {
      open: true,
      action,
      title,
      readiness: snapshot || readiness.value,
      confirmLoading: false,
      pendingRoundPayload,
    }
  }

  function closeOverrideDialog() {
    resetOverrideDialog()
  }

  async function confirmOverrideAction(reason: string) {
    if (!selectedContest.value) {
      return
    }

    const normalizedReason = reason.trim()
    const currentAction = overrideDialogState.value.action
    const currentTitle = overrideDialogState.value.title
    const pendingRoundPayload = overrideDialogState.value.pendingRoundPayload
    if (!normalizedReason || !currentAction) {
      return
    }

    overrideDialogState.value = {
      ...overrideDialogState.value,
      confirmLoading: true,
    }

    try {
      if (currentAction === 'create_round' && pendingRoundPayload) {
        const round = await createContestAWDRound(selectedContest.value.id, {
          ...pendingRoundPayload,
          force_override: true,
          override_reason: normalizedReason,
        })
        toast.success(`第 ${round.round_number} 轮已创建`)
        resetOverrideDialog()
        await onAfterOverride(round.id)
        return
      }

      const result = await runContestAWDCurrentRoundCheck(selectedContest.value.id, {
        force_override: true,
        override_reason: normalizedReason,
      })
      toast.success(`第 ${result.round.round_number} 轮服务巡检已执行`)
      resetOverrideDialog()
      await onAfterOverride(result.round.id)
    } catch (error) {
      if (isAWDReadinessBlockedError(error)) {
        await openOverrideDialog(currentAction, currentTitle || '强制继续', pendingRoundPayload)
        return
      }
      toast.error(humanizeRequestError(error, '执行强制放行失败'))
    } finally {
      if (overrideDialogState.value.open) {
        overrideDialogState.value = {
          ...overrideDialogState.value,
          confirmLoading: false,
        }
      }
    }
  }

  return {
    readiness,
    loadingReadiness,
    overrideDialogState,
    refreshReadiness,
    openOverrideDialog,
    closeOverrideDialog,
    confirmOverrideAction,
  }
}

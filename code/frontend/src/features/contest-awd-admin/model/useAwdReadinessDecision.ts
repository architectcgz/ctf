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
import { useToast } from '@/shared/model/common/useToast'

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

type AwdOverrideAction = Extract<AWDReadinessAction, 'create_round' | 'run_current_round_check'>

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

  function setOverrideDialogState(nextState: Partial<AWDReadinessOverrideDialogState>) {
    overrideDialogState.value = {
      ...overrideDialogState.value,
      ...nextState,
    }
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

  async function loadOverrideReadinessSnapshot() {
    try {
      return await refreshReadiness()
    } catch (error) {
      toast.error(humanizeRequestError(error, '读取开赛就绪摘要失败'))
      return readiness.value
    }
  }

  async function executeOverrideAction(params: {
    contestId: string
    action: AwdOverrideAction
    reason: string
    pendingRoundPayload?: AWDReadinessOverrideDialogState['pendingRoundPayload']
  }) {
    const { contestId, action, reason, pendingRoundPayload } = params
    if (action === 'create_round' && pendingRoundPayload) {
      const round = await createContestAWDRound(contestId, {
        ...pendingRoundPayload,
        force_override: true,
        override_reason: reason,
      })
      toast.success(`第 ${round.round_number} 轮已创建`)
      return round.id
    }

    const result = await runContestAWDCurrentRoundCheck(contestId, {
      force_override: true,
      override_reason: reason,
    })
    toast.success(`第 ${result.round.round_number} 轮服务巡检已执行`)
    return result.round.id
  }

  async function openOverrideDialog(
    action: AwdOverrideAction,
    title: string,
    pendingRoundPayload?: AWDReadinessOverrideDialogState['pendingRoundPayload']
  ) {
    const snapshot = await loadOverrideReadinessSnapshot()
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

    setOverrideDialogState({ confirmLoading: true })

    try {
      const preferredRoundId = await executeOverrideAction({
        contestId: selectedContest.value.id,
        action: currentAction,
        reason: normalizedReason,
        pendingRoundPayload,
      })
      resetOverrideDialog()
      await onAfterOverride(preferredRoundId)
    } catch (error) {
      if (isAWDReadinessBlockedError(error)) {
        await openOverrideDialog(currentAction, currentTitle || '强制继续', pendingRoundPayload)
        return
      }
      toast.error(humanizeRequestError(error, '执行强制放行失败'))
    } finally {
      if (overrideDialogState.value.open) {
        setOverrideDialogState({ confirmLoading: false })
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

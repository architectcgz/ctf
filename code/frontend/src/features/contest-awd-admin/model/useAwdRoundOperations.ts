import { computed, ref, type ComputedRef, type Ref } from 'vue'

import {
  createContestAWDAttackLog,
  createContestAWDRound,
  createContestAWDServiceCheck,
  runContestAWDCurrentRoundCheck,
  runContestAWDRoundCheck,
} from '@/api/admin/contests'
import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDTeamServiceData,
  ContestDetailData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

import { humanizeRequestError, isAWDReadinessBlockedError } from './awdAdminSupport'

interface UseAwdRoundOperationsOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  selectedRoundId: Ref<string | null>
  selectedRound: ComputedRef<AWDRoundData | null>
  refresh: (preferredRoundId?: string) => Promise<void>
  refreshRoundDetail: (roundId?: string | null) => Promise<void>
  openOverrideDialog: (
    action: 'create_round' | 'run_current_round_check',
    label: string,
    payload?: {
      round_number: number
      status?: AWDRoundData['status']
      attack_score?: number
      defense_score?: number
    }
  ) => Promise<void>
}

export function useAwdRoundOperations(options: UseAwdRoundOperationsOptions) {
  const {
    selectedContest,
    selectedRoundId,
    selectedRound,
    refresh,
    refreshRoundDetail,
    openOverrideDialog,
  } = options
  const toast = useToast()

  const checking = ref(false)
  const creatingRound = ref(false)
  const savingServiceCheck = ref(false)
  const savingAttackLog = ref(false)

  const canOperateRound = computed(() => Boolean(selectedContest.value && selectedRoundId.value))

  async function runSelectedRoundCheck() {
    if (!selectedContest.value) {
      return
    }

    const activeRoundId = selectedRoundId.value
    const shouldRunCurrentRound = selectedRound.value?.status === 'running' || !activeRoundId
    checking.value = true
    try {
      const result = shouldRunCurrentRound
        ? await runContestAWDCurrentRoundCheck(selectedContest.value.id)
        : await runContestAWDRoundCheck(selectedContest.value.id, activeRoundId)
      toast.success(`第 ${result.round.round_number} 轮服务巡检已执行`)
      await refresh(result.round.id)
    } catch (error) {
      if (shouldRunCurrentRound && isAWDReadinessBlockedError(error)) {
        await openOverrideDialog('run_current_round_check', '立即巡检当前轮')
        return
      }
      toast.error(humanizeRequestError(error, '执行巡检失败'))
    } finally {
      checking.value = false
    }
  }

  async function createRound(payload: {
    round_number: number
    status?: AWDRoundData['status']
    attack_score?: number
    defense_score?: number
  }) {
    if (!selectedContest.value) {
      return
    }

    creatingRound.value = true
    try {
      const round = await createContestAWDRound(selectedContest.value.id, payload)
      toast.success(`第 ${round.round_number} 轮已创建`)
      await refresh(round.id)
      return round
    } catch (error) {
      if (isAWDReadinessBlockedError(error)) {
        await openOverrideDialog('create_round', '创建轮次', payload)
        return
      }
      toast.error(humanizeRequestError(error, '创建轮次失败'))
    } finally {
      creatingRound.value = false
    }
  }

  async function createServiceCheck(payload: {
    team_id: number
    service_id: number
    service_status: AWDTeamServiceData['service_status']
    check_result?: Record<string, unknown>
  }) {
    if (!selectedContest.value || !canOperateRound.value || !selectedRoundId.value) {
      return
    }

    savingServiceCheck.value = true
    try {
      await createContestAWDServiceCheck(selectedContest.value.id, selectedRoundId.value, payload)
      toast.success('服务检查结果已记录')
      await refreshRoundDetail(selectedRoundId.value)
    } finally {
      savingServiceCheck.value = false
    }
  }

  async function createAttackLog(payload: {
    attacker_team_id: number
    victim_team_id: number
    service_id: number
    attack_type: AWDAttackLogData['attack_type']
    submitted_flag?: string
    is_success: boolean
  }) {
    if (!selectedContest.value || !canOperateRound.value || !selectedRoundId.value) {
      return
    }

    savingAttackLog.value = true
    try {
      await createContestAWDAttackLog(selectedContest.value.id, selectedRoundId.value, payload)
      toast.success('攻击日志已记录')
      await refreshRoundDetail(selectedRoundId.value)
    } finally {
      savingAttackLog.value = false
    }
  }

  return {
    checking,
    creatingRound,
    savingServiceCheck,
    savingAttackLog,
    runSelectedRoundCheck,
    createRound,
    createServiceCheck,
    createAttackLog,
  }
}

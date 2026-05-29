import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

import type {
  AwdCreateAttackLogPayload,
  AwdCreateRoundPayload,
  AwdCreateServiceCheckPayload,
} from './awdOperationsDialogContracts'
import { useAwdOperationsDialogState } from './useAwdOperationsDialogState'

describe('useAwdOperationsDialogState', () => {
  it('应只在运行态允许打开 dialog', () => {
    const runtimeStageReady = ref(false)
    const state = useAwdOperationsDialogState({
      runtimeStageReady,
      rounds: ref([]),
      createRound: vi.fn().mockResolvedValue(undefined),
      createServiceCheck: vi.fn().mockResolvedValue(undefined),
      createAttackLog: vi.fn().mockResolvedValue(undefined),
      closeOverrideDialog: vi.fn(),
    })

    state.openRoundDialog()
    state.openServiceCheckDialog()
    state.openAttackLogDialog()

    expect(state.roundDialogOpen.value).toBe(false)
    expect(state.serviceCheckDialogOpen.value).toBe(false)
    expect(state.attackLogDialogOpen.value).toBe(false)

    runtimeStageReady.value = true

    state.openRoundDialog()
    state.openServiceCheckDialog()
    state.openAttackLogDialog()

    expect(state.roundDialogOpen.value).toBe(true)
    expect(state.serviceCheckDialogOpen.value).toBe(true)
    expect(state.attackLogDialogOpen.value).toBe(true)
  })

  it('应在保存成功后关闭对应 dialog，并保留 override close guard', async () => {
    const createRound = vi.fn().mockResolvedValue(undefined)
    const createServiceCheck = vi.fn().mockResolvedValue(undefined)
    const createAttackLog = vi.fn().mockResolvedValue(undefined)
    const closeOverrideDialog = vi.fn()
    const state = useAwdOperationsDialogState({
      runtimeStageReady: ref(true),
      rounds: ref([{ round_number: 2 }]),
      createRound,
      createServiceCheck,
      createAttackLog,
      closeOverrideDialog,
    })

    state.updateRoundDialogOpen(true)
    state.updateServiceCheckDialogOpen(true)
    state.updateAttackLogDialogOpen(true)

    const createRoundPayload: AwdCreateRoundPayload = {
      round_number: 3,
      status: 'pending',
      attack_score: 10,
      defense_score: 20,
    }
    const createServiceCheckPayload: AwdCreateServiceCheckPayload = {
      team_id: 1,
      service_id: 2,
      service_status: 'up',
      check_result: { latency_ms: 38 },
    }
    const createAttackLogPayload: AwdCreateAttackLogPayload = {
      attacker_team_id: 1,
      victim_team_id: 2,
      service_id: 3,
      attack_type: 'flag_capture',
      submitted_flag: 'flag{demo}',
      is_success: true,
    }

    await state.handleCreateRound(createRoundPayload)
    await state.handleCreateServiceCheck(createServiceCheckPayload)
    await state.handleCreateAttackLog(createAttackLogPayload)

    expect(state.nextRoundNumber.value).toBe(3)
    expect(state.roundDialogOpen.value).toBe(false)
    expect(state.serviceCheckDialogOpen.value).toBe(false)
    expect(state.attackLogDialogOpen.value).toBe(false)
    expect(createRound).toHaveBeenCalledTimes(1)
    expect(createServiceCheck).toHaveBeenCalledTimes(1)
    expect(createAttackLog).toHaveBeenCalledTimes(1)

    state.handleOverrideDialogOpenChange(true)
    expect(closeOverrideDialog).not.toHaveBeenCalled()
    state.handleOverrideDialogOpenChange(false)
    expect(closeOverrideDialog).toHaveBeenCalledTimes(1)
  })
})

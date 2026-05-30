import { describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'

import { useAwdOperationsDialogState } from './useAwdOperationsDialogState'

describe('useAwdOperationsDialogState', () => {
  it('应只在运行态允许打开 dialog', () => {
    const selectedContest = ref<'registering' | 'running'>('registering')
    const runtimeStageReady = computed(() => selectedContest.value === 'running')
    const state = useAwdOperationsDialogState({
      runtimeStageReady,
      rounds: ref([]),
      createRound: vi.fn().mockResolvedValue(undefined),
      createServiceCheck: vi.fn().mockResolvedValue(undefined),
      createAttackLog: vi.fn().mockResolvedValue(undefined),
      closeOverrideDialog: vi.fn(),
    })

    state.roundDialog.requestOpen()
    state.serviceCheckDialog.requestOpen()
    state.attackLogDialog.requestOpen()

    expect(state.roundDialog.open.value).toBe(false)
    expect(state.serviceCheckDialog.open.value).toBe(false)
    expect(state.attackLogDialog.open.value).toBe(false)

    selectedContest.value = 'running'

    state.roundDialog.requestOpen()
    state.serviceCheckDialog.requestOpen()
    state.attackLogDialog.requestOpen()

    expect(state.roundDialog.open.value).toBe(true)
    expect(state.serviceCheckDialog.open.value).toBe(true)
    expect(state.attackLogDialog.open.value).toBe(true)
  })

  it('应在保存成功后关闭对应 dialog，并保留 override close guard', async () => {
    const createRound = vi.fn().mockResolvedValue(undefined)
    const createServiceCheck = vi.fn().mockResolvedValue(undefined)
    const createAttackLog = vi.fn().mockResolvedValue(undefined)
    const closeOverrideDialog = vi.fn()
    const state = useAwdOperationsDialogState({
      runtimeStageReady: computed(() => true),
      rounds: ref([{ round_number: 2 }]),
      createRound,
      createServiceCheck,
      createAttackLog,
      closeOverrideDialog,
    })

    state.roundDialog.setOpen(true)
    state.serviceCheckDialog.setOpen(true)
    state.attackLogDialog.setOpen(true)

    await state.roundDialog.submitAndClose({
      round_number: 3,
      status: 'pending',
      attack_score: 10,
      defense_score: 20,
    })
    await state.serviceCheckDialog.submitAndClose({
      team_id: 1,
      service_id: 2,
      service_status: 'up',
      check_result: { latency_ms: 38 },
    })
    await state.attackLogDialog.submitAndClose({
      attacker_team_id: 1,
      victim_team_id: 2,
      service_id: 3,
      attack_type: 'flag_capture',
      submitted_flag: 'flag{demo}',
      is_success: true,
    })

    expect(state.nextRoundNumber.value).toBe(3)
    expect(state.roundDialog.open.value).toBe(false)
    expect(state.serviceCheckDialog.open.value).toBe(false)
    expect(state.attackLogDialog.open.value).toBe(false)
    expect(createRound).toHaveBeenCalledTimes(1)
    expect(createServiceCheck).toHaveBeenCalledTimes(1)
    expect(createAttackLog).toHaveBeenCalledTimes(1)

    state.handleOverrideDialogOpenChange(true)
    expect(closeOverrideDialog).not.toHaveBeenCalled()
    state.handleOverrideDialogOpenChange(false)
    expect(closeOverrideDialog).toHaveBeenCalledTimes(1)
  })
})

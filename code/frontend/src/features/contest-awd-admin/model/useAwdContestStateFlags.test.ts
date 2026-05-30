import { describe, expect, it } from 'vitest'
import { ref } from 'vue'

import type { AWDRoundData, ContestDetailData } from '@/api/contracts'

import { useAwdContestStateFlags } from './useAwdContestStateFlags'

function buildContest(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'awd-1',
    title: '2026 AWD 联赛',
    description: '攻防赛',
    mode: 'awd',
    status: 'running',
    starts_at: '2026-03-18T09:00:00.000Z',
    ends_at: '2026-03-18T18:00:00.000Z',
    ...overrides,
  }
}

function buildRound(overrides: Partial<AWDRoundData> = {}): AWDRoundData {
  return {
    id: 'round-1',
    contest_id: 'awd-1',
    round_number: 1,
    status: 'running',
    attack_score: 50,
    defense_score: 50,
    created_at: '2026-03-18T09:00:00.000Z',
    updated_at: '2026-03-18T09:00:00.000Z',
    ...overrides,
  }
}

describe('useAwdContestStateFlags', () => {
  it('应统一推导 runtime stage、round operation 与 auto refresh 规则', () => {
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const selectedRoundId = ref<string | null>('round-1')
    const selectedRound = ref<AWDRoundData | null>(buildRound())

    const state = useAwdContestStateFlags({
      selectedContest,
      selectedRoundId,
      selectedRound,
    })

    expect(state.hasSelectedContest.value).toBe(true)
    expect(state.runtimeStageReady.value).toBe(true)
    expect(state.canOperateSelectedRound.value).toBe(true)
    expect(state.shouldUseCurrentRoundCheck.value).toBe(true)
    expect(state.shouldAutoRefresh.value).toBe(true)

    selectedContest.value = buildContest({ status: 'registering' })
    expect(state.runtimeStageReady.value).toBe(false)
    expect(state.shouldAutoRefresh.value).toBe(false)

    selectedContest.value = buildContest({ status: 'running' })
    selectedRound.value = buildRound({ status: 'pending' })
    expect(state.shouldUseCurrentRoundCheck.value).toBe(false)
    expect(state.shouldAutoRefresh.value).toBe(false)

    selectedRoundId.value = null
    expect(state.canOperateSelectedRound.value).toBe(false)
    expect(state.shouldUseCurrentRoundCheck.value).toBe(true)
  })
})

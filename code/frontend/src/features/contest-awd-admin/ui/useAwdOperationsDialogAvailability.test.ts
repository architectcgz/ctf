import { describe, expect, it } from 'vitest'
import { ref } from 'vue'

import type { AdminContestChallengeViewData, AdminContestTeamData } from '@/api/contracts'

import { useAwdOperationsDialogAvailability } from './useAwdOperationsDialogAvailability'

function buildTeam(overrides: Partial<AdminContestTeamData> = {}): AdminContestTeamData {
  return {
    id: '1',
    contest_id: 'awd-1',
    name: 'Red',
    captain_id: '10',
    max_members: 5,
    member_count: 1,
    created_at: '2026-03-18T09:00:00.000Z',
    ...overrides,
  }
}

function buildChallengeLink(
  overrides: Partial<AdminContestChallengeViewData> = {}
): AdminContestChallengeViewData {
  return {
    id: 'link-1',
    contest_id: 'awd-1',
    challenge_id: 'challenge-1',
    title: 'Web Checker',
    category: 'web',
    difficulty: 'easy',
    points: 120,
    order: 1,
    is_visible: true,
    created_at: '2026-03-18T09:00:00.000Z',
    ...overrides,
  }
}

describe('useAwdOperationsDialogAvailability', () => {
  it('应根据队伍与题目数量推导 service/attack hint', () => {
    const teams = ref<AdminContestTeamData[]>([])
    const challengeLinks = ref<AdminContestChallengeViewData[]>([])
    const state = useAwdOperationsDialogAvailability({
      teams,
      challengeLinks,
    })

    expect(state.canRecordServiceChecks.value).toBe(false)
    expect(state.canRecordAttackLogs.value).toBe(false)
    expect(state.serviceCheckHint.value).toContain('队伍和题目')
    expect(state.attackLogHint.value).toContain('2 支队伍')

    teams.value = [buildTeam(), buildTeam({ id: '2', name: 'Blue', captain_id: '20' })]
    challengeLinks.value = [buildChallengeLink()]

    expect(state.canRecordServiceChecks.value).toBe(true)
    expect(state.canRecordAttackLogs.value).toBe(true)
    expect(state.serviceCheckHint.value).toBe('')
    expect(state.attackLogHint.value).toBe('')
  })
})

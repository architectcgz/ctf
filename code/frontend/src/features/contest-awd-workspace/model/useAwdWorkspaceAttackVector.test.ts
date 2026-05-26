import { computed, nextTick, ref } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import type { ContestAWDWorkspaceTargetTeamData, ContestChallengeItem } from '@/api/contracts'
import { useAwdWorkspaceAttackVector } from './useAwdWorkspaceAttackVector'

function challenge(serviceId: string, title = serviceId): ContestChallengeItem {
  return {
    id: serviceId,
    challenge_id: serviceId,
    awd_challenge_id: `${serviceId}-challenge`,
    awd_service_id: serviceId,
    title,
    category: 'misc',
    difficulty: 'easy',
    points: 100,
    solved_count: 0,
    is_solved: false,
  }
}

function target(teamId: string, teamName: string, serviceIds: string[]): ContestAWDWorkspaceTargetTeamData {
  return {
    team_id: teamId,
    team_name: teamName,
    services: serviceIds.map((serviceId) => ({
      service_id: serviceId,
      awd_challenge_id: `${serviceId}-challenge`,
      reachable: true,
    })),
  }
}

describe('useAwdWorkspaceAttackVector', () => {
  it('默认选择当前第一个可用攻击题目', () => {
    const challenges = ref([challenge('service-a', 'Alpha'), challenge('service-b', 'Beta')])
    const targets = ref<ContestAWDWorkspaceTargetTeamData[]>([])

    const state = useAwdWorkspaceAttackVector({
      challenges: computed(() => challenges.value),
      targets: computed(() => targets.value),
      submitAttack: vi.fn(),
    })

    expect(state.activeChallengeKey.value).toBe('service-a')
    expect(state.attackToolbarChallengeOptions.value).toEqual([
      { key: 'service-a', title: 'Alpha' },
      { key: 'service-b', title: 'Beta' },
    ])
  })

  it('切换关键字后只保留匹配队伍', async () => {
    const challenges = ref([challenge('service-a', 'Alpha')])
    const targets = ref([
      target('1', 'Blue Team', ['service-a']),
      target('2', 'Red Squad', ['service-a']),
    ])

    const state = useAwdWorkspaceAttackVector({
      challenges: computed(() => challenges.value),
      targets: computed(() => targets.value),
      submitAttack: vi.fn(),
    })

    state.targetKeyword.value = 'red'
    await nextTick()

    expect(state.filteredTargets.value).toHaveLength(1)
    expect(state.filteredTargets.value[0].team_name).toBe('Red Squad')
    expect(state.filteredTargets.value[0].active_service?.service_id).toBe('service-a')
  })

  it('提交成功后清空对应 flag 输入', async () => {
    const submitAttack = vi.fn().mockResolvedValue({ id: 'attack-1' })

    const state = useAwdWorkspaceAttackVector({
      challenges: computed(() => [challenge('service-a', 'Alpha')]),
      targets: computed(() => [target('7', 'Blue Team', ['service-a'])]),
      submitAttack,
    })

    state.flagInputs.value['service-a:7'] = '  FLAG{demo}  '
    await state.handleSubmit('service-a', '7')

    expect(submitAttack).toHaveBeenCalledWith('service-a', 7, '  FLAG{demo}  ')
    expect(state.flagInputs.value['service-a:7']).toBe('')
  })
})

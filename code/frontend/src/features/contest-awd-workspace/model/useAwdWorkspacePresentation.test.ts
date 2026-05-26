import { computed, ref } from 'vue'
import { describe, expect, it } from 'vitest'

import type { ContestChallengeItem } from '@/api/contracts'
import { useAwdWorkspacePresentation } from './useAwdWorkspacePresentation'

function challenge(serviceId: string, challengeId: string, title: string): ContestChallengeItem {
  return {
    id: challengeId,
    challenge_id: challengeId,
    awd_challenge_id: `${challengeId}-awd`,
    awd_service_id: serviceId,
    title,
    category: 'misc',
    difficulty: 'easy',
    points: 100,
    solved_count: 0,
    is_solved: false,
  }
}

describe('useAwdWorkspacePresentation', () => {
  it('事件标题应优先按 service_id 命中挑战标题', () => {
    const challenges = ref([
      challenge('7009', 'legacy-101', 'Bank Portal'),
      challenge('7010', 'legacy-102', 'Patch Relay'),
    ])

    const state = useAwdWorkspacePresentation({
      challenges: computed(() => challenges.value),
    })

    expect(
      state.getChallengeTitleForEvent({
        service_id: '7009',
        awd_challenge_id: 'legacy-unknown',
      })
    ).toBe('Bank Portal')
  })

  it('攻击结果文案应按 awd_challenge_id 映射', () => {
    const challenges = ref([challenge('7009', 'legacy-101', 'Bank Portal')])

    const state = useAwdWorkspacePresentation({
      challenges: computed(() => challenges.value),
    })

    expect(
      state.formatAttackResultToast({
        awd_challenge_id: 'legacy-101-awd',
        is_success: true,
        score_gained: 60,
      })
    ).toBe('Bank Portal: 攻击成功，+60 分')
  })

  it('缺少 awd_challenge_id 的挑战不应参与 AWD 文案映射', () => {
    const challenges = ref<ContestChallengeItem[]>([
      {
        id: 'legacy-101',
        challenge_id: 'legacy-101',
        awd_service_id: '7009',
        title: 'Legacy Portal',
        category: 'misc',
        difficulty: 'easy',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])

    const state = useAwdWorkspacePresentation({
      challenges: computed(() => challenges.value),
    })

    expect(
      state.getChallengeTitleForEvent({
        service_id: '7009',
        awd_challenge_id: 'awd-unknown',
      })
    ).toBe('awd-unknown')
  })

  it('事件与服务展示文案应保持当前标签', () => {
    const state = useAwdWorkspacePresentation({
      challenges: computed(() => [] as ContestChallengeItem[]),
    })

    expect(state.eventDirectionLabel('attack_out')).toBe('对外攻击')
    expect(state.eventDirectionLabel('attack_in')).toBe('受到攻击')
    expect(state.eventResultLabel(false)).toBe('失败')
    expect(state.formatServiceRef('7009')).toBe('服务 #7009')
    expect(state.formatServiceRef()).toBe('服务 #--')
  })
})

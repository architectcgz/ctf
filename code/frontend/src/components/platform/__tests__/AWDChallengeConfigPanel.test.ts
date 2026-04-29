import { describe, expect, it } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'

import AWDChallengeConfigPanel from '../contest/AWDChallengeConfigPanel.vue'
import type { AdminContestChallengeViewData } from '@/api/contracts'

function buildChallenge(
  overrides: Partial<AdminContestChallengeViewData> = {},
): AdminContestChallengeViewData {
  return {
    id: 'link-1',
    contest_id: 'contest-1',
    challenge_id: '101',
    title: 'Web 入门',
    category: 'web',
    difficulty: 'easy',
    points: 120,
    order: 1,
    is_visible: true,
    awd_service_id: 'service-1',
    awd_challenge_id: 'template-1',
    awd_checker_type: 'http_standard',
    awd_checker_config: {},
    awd_sla_score: 1,
    awd_defense_score: 2,
    awd_checker_validation_state: 'pending',
    awd_checker_last_preview_at: undefined,
    awd_checker_last_preview_result: undefined,
    created_at: '2026-03-10T00:00:00.000Z',
    ...overrides,
  }
}

describe('AWDChallengeConfigPanel', () => {
  it('应该把题目标题链接到管理员题目预览页', () => {
    const wrapper = mount(AWDChallengeConfigPanel, {
      props: {
        challengeLinks: [buildChallenge()],
      },
      global: {
        stubs: {
          AppEmpty: {
            props: ['title', 'description'],
            template: '<div>{{ title }}{{ description }}</div>',
          },
          RouterLink: RouterLinkStub,
        },
      },
    })

    const titleLink = wrapper.findAllComponents(RouterLinkStub).find((link) => link.text() === 'Web 入门')

    expect(titleLink?.props('to')).toEqual({
      name: 'PlatformChallengeDetail',
      params: { id: '101' },
    })
  })
})

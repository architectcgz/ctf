import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import ChallengeManageRoutePage from '@/pages/platform/challenges/ChallengeManageRoutePage.vue'
import challengeManageRouteSource from '@/pages/platform/challenges/ChallengeManageRoutePage.vue?raw'

describe('ChallengeManageRoutePage', () => {
  it('应通过 pages 层组合题目管理 feature 页面', () => {
    const wrapper = mount(ChallengeManageRoutePage, {
      global: {
        stubs: {
          ChallengeManagePage: {
            template: '<div data-testid="challenge-manage-page">题目管理</div>',
          },
        },
      },
    })

    expect(wrapper.find('[data-testid="challenge-manage-page"]').exists()).toBe(true)
    expect(challengeManageRouteSource).toContain(
      "import { ChallengeManagePage } from '@/features/platform/challenges'"
    )
    expect(challengeManageRouteSource).not.toContain('useRoute')
    expect(challengeManageRouteSource).not.toContain('useRouter')
    expect(challengeManageRouteSource).not.toContain('@/api/')
  })
})

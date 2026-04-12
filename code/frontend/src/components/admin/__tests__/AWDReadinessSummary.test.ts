import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AWDReadinessSummary from '../contest/AWDReadinessSummary.vue'

function buildReadiness(overrides: Record<string, unknown> = {}) {
  return {
    contest_id: 'awd-1',
    ready: false,
    total_challenges: 0,
    passed_challenges: 0,
    pending_challenges: 0,
    failed_challenges: 0,
    stale_challenges: 0,
    missing_checker_challenges: 0,
    blocking_count: 0,
    global_blocking_reasons: ['no_challenges'],
    blocking_actions: ['start_contest'],
    items: [],
    ...overrides,
  }
}

describe('AWDReadinessSummary', () => {
  it('零题目时不应展示题目侧已满足开赛要求的正向空态文案', () => {
    const wrapper = mount(AWDReadinessSummary, {
      props: {
        loading: false,
        readiness: buildReadiness(),
      },
      global: {
        stubs: {
          AppEmpty: {
            props: ['title', 'description'],
            template: '<div class="app-empty-stub">{{ title }}|{{ description }}</div>',
          },
        },
      },
    })

    expect(wrapper.text()).toContain('系统级阻塞仍会拦截开赛关键动作')
    expect(wrapper.text()).not.toContain('题目侧的 checker 校验已经满足开赛关键动作要求')
  })
})

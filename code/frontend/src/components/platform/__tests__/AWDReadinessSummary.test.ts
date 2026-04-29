import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AWDReadinessChecklist from '../contest/AWDReadinessChecklist.vue'
import AWDReadinessSummary from '../contest/AWDReadinessSummary.vue'
import type { AWDReadinessData } from '@/api/contracts'

function buildReadiness(overrides: Partial<AWDReadinessData> = {}): AWDReadinessData {
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
  it('应该显示可开赛 / 待修复 / 不可开赛结论', () => {
    const readyWrapper = mount(AWDReadinessSummary, {
      props: {
        loading: false,
        readiness: buildReadiness({
          ready: true,
          total_challenges: 2,
          passed_challenges: 2,
          blocking_count: 0,
          global_blocking_reasons: [],
          blocking_actions: [],
        }),
      },
    })

    const needsFixWrapper = mount(AWDReadinessSummary, {
      props: {
        loading: false,
        readiness: buildReadiness({
          total_challenges: 1,
          blocking_count: 1,
          global_blocking_reasons: [],
          items: [
            {
              awd_challenge_id: '101',
              title: 'Challenge 101',
              checker_type: 'http_standard',
              validation_state: 'failed',
              last_preview_at: '2026-04-12T08:00:00.000Z',
              last_access_url: 'http://checker.internal/flag',
              blocking_reason: 'last_preview_failed',
            },
          ],
        }),
      },
    })

    const blockedWrapper = mount(AWDReadinessSummary, {
      props: {
        loading: false,
        readiness: buildReadiness(),
      },
    })

    expect(readyWrapper.text()).toContain('可开赛')
    expect(needsFixWrapper.text()).toContain('待修复')
    expect(needsFixWrapper.text()).not.toContain('可强制开赛')
    expect(blockedWrapper.text()).toContain('不可开赛')
  })

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

    expect(wrapper.text()).toContain('请先处理上方系统级阻塞项')
    expect(wrapper.text()).not.toContain('题目侧的 checker 校验已经满足开赛关键动作要求')
  })

  it('明细组件只展示统计与阻塞列表，不包含完整摘要的就绪标题与决策卡', () => {
    const wrapper = mount(AWDReadinessChecklist, {
      props: {
        readiness: buildReadiness({
          ready: true,
          total_challenges: 2,
          passed_challenges: 2,
          global_blocking_reasons: [],
          blocking_actions: [],
        }),
      },
    })

    expect(wrapper.text()).not.toContain('开赛就绪摘要')
    expect(wrapper.text()).not.toContain('就绪决策')
    expect(wrapper.text()).toContain('最近通过')
    expect(wrapper.text()).toContain('阻塞短名单')
  })
})

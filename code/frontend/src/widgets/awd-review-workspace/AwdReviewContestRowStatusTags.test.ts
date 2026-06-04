import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AwdReviewContestRowStatusTags from './AwdReviewContestRowStatusTags.vue'

describe('AwdReviewContestRowStatusTags', () => {
  it('导出可用时应显示可导出标签', () => {
    const wrapper = mount(AwdReviewContestRowStatusTags, {
      props: {
        statusLabel: '进行中',
        exportReady: true,
      },
    })

    expect(wrapper.text()).toContain('进行中')
    expect(wrapper.text()).toContain('可导出')
    expect(wrapper.text()).not.toContain('实时复盘')
  })

  it('导出不可用时应显示实时复盘标签', () => {
    const wrapper = mount(AwdReviewContestRowStatusTags, {
      props: {
        statusLabel: '已结束',
        exportReady: false,
      },
    })

    expect(wrapper.text()).toContain('已结束')
    expect(wrapper.text()).toContain('实时复盘')
    expect(wrapper.text()).not.toContain('可导出')
  })
})

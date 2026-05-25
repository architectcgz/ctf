import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AwdReviewStatusChip from './AwdReviewStatusChip.vue'

describe('AwdReviewStatusChip', () => {
  it('应渲染状态文案和 running 样式类', () => {
    const wrapper = mount(AwdReviewStatusChip, {
      props: {
        status: 'running',
        label: '进行中',
      },
    })

    expect(wrapper.text()).toContain('进行中')
    expect(wrapper.classes()).toContain('awd-review-status-chip')
    expect(wrapper.classes()).toContain('awd-review-status-chip--running')
  })

  it('空状态应回退到 idle 类名', () => {
    const wrapper = mount(AwdReviewStatusChip, {
      props: {
        status: '',
        label: '未开始',
      },
    })

    expect(wrapper.classes()).toContain('awd-review-status-chip--idle')
  })
})

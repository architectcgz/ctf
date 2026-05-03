import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherAWDReviewContestRowMetrics from './TeacherAWDReviewContestRowMetrics.vue'

describe('TeacherAWDReviewContestRowMetrics', () => {
  it('应渲染主副文案', () => {
    const wrapper = mount(TeacherAWDReviewContestRowMetrics, {
      props: {
        primary: '第 2 轮',
        secondary: '共 6 轮',
      },
    })

    expect(wrapper.text()).toContain('第 2 轮')
    expect(wrapper.text()).toContain('共 6 轮')
  })
})

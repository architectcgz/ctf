import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherAWDReviewWorkspaceState from './TeacherAWDReviewWorkspaceState.vue'

describe('TeacherAWDReviewWorkspaceState', () => {
  it('错误态应透传重试事件', async () => {
    const wrapper = mount(TeacherAWDReviewWorkspaceState, {
      props: {
        loading: false,
        error: '加载失败',
        hasReview: false,
      },
    })

    await wrapper.get('button.teacher-btn--primary').trigger('click')

    expect(wrapper.emitted('loadReview')).toBeTruthy()
  })

  it('有数据时应渲染默认插槽', () => {
    const wrapper = mount(TeacherAWDReviewWorkspaceState, {
      props: {
        loading: false,
        error: null,
        hasReview: true,
      },
      slots: {
        default: '<div class="awd-review-content">loaded</div>',
      },
    })

    expect(wrapper.find('.awd-review-content').exists()).toBe(true)
  })
})

import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherReviewArchiveState from './TeacherReviewArchiveState.vue'

describe('TeacherReviewArchiveState', () => {
  it('错误态应透传 reload 事件', async () => {
    const wrapper = mount(TeacherReviewArchiveState, {
      props: {
        loading: false,
        error: '加载失败',
        hasArchive: false,
      },
    })

    await wrapper.get('.ui-btn.ui-btn--primary').trigger('click')

    expect(wrapper.emitted('reload')).toBeTruthy()
  })

  it('有数据时应渲染默认插槽', () => {
    const wrapper = mount(TeacherReviewArchiveState, {
      props: {
        loading: false,
        error: null,
        hasArchive: true,
      },
      slots: {
        default: '<div class="archive-content">loaded</div>',
      },
    })

    expect(wrapper.find('.archive-content').exists()).toBe(true)
  })
})

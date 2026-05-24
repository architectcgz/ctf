import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherAWDReviewWorkspaceActions from './TeacherAWDReviewWorkspaceActions.vue'

describe('TeacherAWDReviewWorkspaceActions', () => {
  it('应透传返回与导出事件', async () => {
    const wrapper = mount(TeacherAWDReviewWorkspaceActions, {
      props: {
        loading: false,
        hasReview: true,
        exporting: null,
        canExportReport: true,
      },
    })

    const [backButton, archiveButton, reportButton] = wrapper.findAll('button')

    await backButton.trigger('click')
    await archiveButton.trigger('click')
    await reportButton.trigger('click')

    expect(wrapper.emitted('openIndex')).toBeTruthy()
    expect(wrapper.emitted('exportArchive')).toBeTruthy()
    expect(wrapper.emitted('exportReport')).toBeTruthy()
  })

  it('应根据状态禁用导出按钮', () => {
    const wrapper = mount(TeacherAWDReviewWorkspaceActions, {
      props: {
        loading: true,
        hasReview: false,
        exporting: 'report',
        canExportReport: false,
      },
    })

    expect(wrapper.get('[data-testid="awd-review-export-archive"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="awd-review-export-report"]').attributes('disabled')).toBeDefined()
  })
})

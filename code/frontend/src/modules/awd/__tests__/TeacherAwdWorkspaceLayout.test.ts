import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { TEACHER_AWD_PAGES } from '@/modules/awd/navigation'
import TeacherAwdWorkspaceLayout from '@/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue'

describe('TeacherAwdWorkspaceLayout', () => {
  it('renders the teacher shell with page summary and slot content', () => {
    const wrapper = mount(TeacherAwdWorkspaceLayout, {
      props: {
        contestTitle: '春季 AWD 复盘',
        pageTitle: '教学总览',
        pageDescription: '查看整场复盘摘要与轮次线索。',
        pages: TEACHER_AWD_PAGES,
        currentPage: 'overview',
        heroMetrics: [{ label: '轮次数', value: '30' }],
      },
      slots: {
        default: '<div data-testid="teacher-awd-slot">teacher body</div>',
      },
    })

    expect(wrapper.text()).toContain('春季 AWD 复盘')
    expect(wrapper.find('[data-testid="teacher-awd-slot"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="awd-page-nav-item"]')).toHaveLength(5)
  })
})

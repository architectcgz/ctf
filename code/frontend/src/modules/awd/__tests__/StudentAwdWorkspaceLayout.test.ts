import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { STUDENT_AWD_PAGES } from '@/modules/awd/navigation'
import StudentAwdWorkspaceLayout from '@/modules/awd/layouts/StudentAwdWorkspaceLayout.vue'

describe('StudentAwdWorkspaceLayout', () => {
  it('renders hero, left navigation, and slot content', () => {
    const wrapper = mount(StudentAwdWorkspaceLayout, {
      props: {
        contestTitle: '2026 春季 AWD',
        pageTitle: '战场总览',
        pageDescription: '查看当前轮次与战场动态。',
        pages: STUDENT_AWD_PAGES,
        currentPage: 'overview',
        heroMetrics: [{ label: '当前轮', value: 'R12' }],
      },
      slots: {
        default: '<div data-testid="student-awd-slot">student body</div>',
      },
    })

    expect(wrapper.text()).toContain('2026 春季 AWD')
    expect(wrapper.text()).toContain('战场总览')
    expect(wrapper.find('[data-testid="student-awd-slot"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="awd-page-nav-item"]')).toHaveLength(5)
  })
})

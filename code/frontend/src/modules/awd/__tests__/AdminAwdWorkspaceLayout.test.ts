import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { ADMIN_AWD_PAGES } from '@/modules/awd/navigation'
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

describe('AdminAwdWorkspaceLayout', () => {
  it('renders the admin shell with 9 page navigation items', () => {
    const wrapper = mount(AdminAwdWorkspaceLayout, {
      props: {
        contestTitle: '春季攻防演练',
        pageTitle: 'AWD 总览',
        pageDescription: '查看赛事健康度与异常信号。',
        pages: ADMIN_AWD_PAGES,
        currentPage: 'overview',
        heroMetrics: [{ label: '在线服务', value: '43/48' }],
      },
      slots: {
        default: '<div data-testid="admin-awd-slot">admin body</div>',
      },
    })

    expect(wrapper.text()).toContain('春季攻防演练')
    expect(wrapper.find('[data-testid="admin-awd-slot"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="awd-page-nav-item"]')).toHaveLength(9)
  })
})

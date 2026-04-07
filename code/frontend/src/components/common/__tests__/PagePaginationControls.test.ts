import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import PagePaginationControls from '../PagePaginationControls.vue'

describe('PagePaginationControls', () => {
  it('应在提交合法页码时发出 changePage', async () => {
    const wrapper = mount(PagePaginationControls, {
      props: {
        page: 2,
        totalPages: 5,
        total: 100,
        totalLabel: '共 100 条',
        showJump: true,
      },
    })

    await wrapper.get('input').setValue('5')
    await wrapper.get('form').trigger('submit')

    expect(wrapper.emitted('changePage')).toEqual([[5]])
  })

  it('关闭 jump 时不应渲染跳页表单，但仍应显示页码状态', () => {
    const wrapper = mount(PagePaginationControls, {
      props: {
        page: 1,
        totalPages: 2,
        total: 21,
        totalLabel: '共 21 条',
      },
    })

    expect(wrapper.text()).toContain('共 21 条')
    expect(wrapper.text()).toContain('1 / 2')
    expect(wrapper.find('form').exists()).toBe(false)
  })
})

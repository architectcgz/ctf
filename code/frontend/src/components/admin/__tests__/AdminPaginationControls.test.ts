import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AdminPaginationControls from '../AdminPaginationControls.vue'

describe('AdminPaginationControls', () => {
  it('应在提交合法页码时发出 changePage', async () => {
    const wrapper = mount(AdminPaginationControls, {
      props: {
        page: 2,
        totalPages: 5,
        total: 100,
        totalLabel: '共 100 条',
      },
    })

    await wrapper.get('input').setValue('5')
    await wrapper.get('form').trigger('submit')

    expect(wrapper.emitted('changePage')).toEqual([[5]])
  })

  it('输入非法页码时不应发出 changePage', async () => {
    const wrapper = mount(AdminPaginationControls, {
      props: {
        page: 2,
        totalPages: 5,
        total: 100,
        totalLabel: '共 100 条',
      },
    })

    await wrapper.get('input').setValue('9')
    await wrapper.get('form').trigger('submit')

    expect(wrapper.emitted('changePage')).toBeUndefined()
  })
})

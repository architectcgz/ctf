import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import WorkspaceDirectoryPagination from '../WorkspaceDirectoryPagination.vue'

describe('WorkspaceDirectoryPagination', () => {
  it('应复用分页控件并透传页码切换事件', async () => {
    const wrapper = mount(WorkspaceDirectoryPagination, {
      props: {
        page: 2,
        totalPages: 5,
        total: 18,
        totalLabel: '共 18 条',
      },
    })

    expect(wrapper.text()).toContain('共 18 条')
    expect(wrapper.text()).toContain('2 / 5')

    await wrapper.get('button:last-of-type').trigger('click')
    expect(wrapper.emitted('changePage')?.[0]).toEqual([3])
  })

  it('应将目录分页的数量单位补全为完整摘要', () => {
    const wrapper = mount(WorkspaceDirectoryPagination, {
      props: {
        page: 1,
        totalPages: 2,
        total: 26,
        totalLabel: '个班级',
      },
    })

    expect(wrapper.text()).toContain('共 26 个班级')
  })
})

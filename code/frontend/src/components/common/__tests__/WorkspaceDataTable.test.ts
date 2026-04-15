import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import WorkspaceDataTable from '../WorkspaceDataTable.vue'

describe('WorkspaceDataTable', () => {
  it('应该按列配置渲染表头、行数据与自定义单元格插槽', () => {
    const wrapper = mount(WorkspaceDataTable, {
      props: {
        columns: [
          { key: 'title', label: '题目名称', widthClass: 'w-[40%]' },
          { key: 'category', label: '分类', align: 'center' },
          { key: 'actions', label: '操作', align: 'right' },
        ],
        rows: [
          { id: 'row-1', title: 'SQL Diary', category: 'web' },
        ],
        rowKey: 'id',
      },
      slots: {
        'cell-category': '<span class="category-slot">Web</span>',
        'cell-actions': '<button class="action-slot">查看</button>',
      },
    })

    const headers = wrapper.findAll('th')
    expect(headers).toHaveLength(3)
    expect(headers[0]?.text()).toBe('题目名称')
    expect(headers[1]?.classes()).toContain('workspace-data-table__cell--center')
    expect(wrapper.text()).toContain('SQL Diary')
    expect(wrapper.find('.category-slot').exists()).toBe(true)
    expect(wrapper.find('.action-slot').exists()).toBe(true)
  })
})

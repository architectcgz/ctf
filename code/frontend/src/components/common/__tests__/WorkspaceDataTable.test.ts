import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import WorkspaceDataTable from '../WorkspaceDataTable.vue'
import workspaceDataTableSource from '../WorkspaceDataTable.vue?raw'

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

  it('夜间模式下表头与行分隔线应使用柔和 token，而不是继续写死浅色边框', () => {
    expect(workspaceDataTableSource).toContain('--workspace-table-line')
    expect(workspaceDataTableSource).toContain(":global([data-theme='dark']) .workspace-data-table-shell")
    expect(workspaceDataTableSource).not.toContain('border-bottom: 1px solid #f1f5f9;')
    expect(workspaceDataTableSource).not.toContain('border-bottom: 1px solid #e2e8f0;')
  })


  it('浅色主题回退也应继续复用共享 token，而不是写死 slate 色板', () => {
    expect(workspaceDataTableSource).toContain('var(--color-border-default)')
    expect(workspaceDataTableSource).toContain('var(--color-text-primary)')
    expect(workspaceDataTableSource).toContain('var(--color-text-secondary)')
    expect(workspaceDataTableSource).toContain('var(--color-text-muted)')
    expect(workspaceDataTableSource).not.toContain('#e8eef5')
    expect(workspaceDataTableSource).not.toContain('#dbe4ee')
    expect(workspaceDataTableSource).not.toContain('#0f172a')
    expect(workspaceDataTableSource).not.toContain('#475569')
    expect(workspaceDataTableSource).not.toContain('#94a3b8')
  })
})

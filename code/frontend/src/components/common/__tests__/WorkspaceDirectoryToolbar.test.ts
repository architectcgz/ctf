import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import WorkspaceDirectoryToolbar from '../WorkspaceDirectoryToolbar.vue'

describe('WorkspaceDirectoryToolbar', () => {
  it('应支持搜索输入、排序选择与筛选面板展示', async () => {
    const wrapper = mount(WorkspaceDirectoryToolbar, {
      attachTo: document.body,
      props: {
        modelValue: '',
        total: 18,
        selectedSortLabel: '最近更新',
        sortOptions: [
          { key: 'updated', label: '最近更新' },
          { key: 'points', label: '分值由高到低' },
        ],
      },
      slots: {
        'filter-panel': `
          <label class="challenge-filter-field">
            <span class="challenge-filter-label">题目分类</span>
            <select class="challenge-filter-select">
              <option value="">全部分类</option>
            </select>
          </label>
        `,
      },
    })

    await wrapper.get('.workspace-directory-toolbar__search-input').setValue('web')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['web'])

    await wrapper.get('.workspace-directory-toolbar__sort-button').trigger('click')
    expect(wrapper.find('.workspace-directory-toolbar__sort-menu').exists()).toBe(true)
    await wrapper.get('.workspace-directory-toolbar__menu-item:last-child').trigger('click')
    expect(wrapper.emitted('selectSort')?.[0]?.[0]).toMatchObject({
      key: 'points',
      label: '分值由高到低',
    })

    await wrapper.get('.workspace-directory-toolbar__filter-toggle').trigger('click')
    expect(wrapper.find('.workspace-directory-toolbar__filter-panel').exists()).toBe(true)
    expect(wrapper.find('.challenge-filter-select').exists()).toBe(true)

    await wrapper.get('.workspace-directory-toolbar__filter-reset').trigger('click')
    expect(wrapper.emitted('resetFilters')).toHaveLength(1)

    wrapper.unmount()
  })
})

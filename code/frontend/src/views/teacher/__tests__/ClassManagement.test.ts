import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import ClassManagement from '../ClassManagement.vue'
import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('ClassManagement', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 3 },
    ])
  })

  it('应该展示班级列表并支持进入班级学生页', async () => {
    const wrapper = mount(ClassManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.find('#class-manage-tab-overview').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#class-manage-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#class-manage-directory').attributes('aria-hidden')).toBe('true')
    expect(wrapper.findAll('.teacher-summary-item')).toHaveLength(3)

    await wrapper.get('#class-manage-tab-directory').trigger('click')
    await flushPromises()

    expect(wrapper.find('#class-manage-tab-directory').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#class-manage-directory').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(2)
    expect(wrapper.find('.teacher-directory-head').text()).toContain('班级编号')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('班级名称')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('学生数')
    expect(wrapper.find('.teacher-directory-head').text()).not.toContain('标签')
    expect(wrapper.find('.teacher-directory-head').text()).not.toContain('数据')

    const headChildren = Array.from(wrapper.find('.teacher-directory-head').element.children).map((element) =>
      element.className.toString()
    )
    expect(headChildren[0]).toContain('teacher-directory-head-cell-class-code')
    expect(headChildren[1]).toContain('teacher-directory-head-cell-class-name')
    expect(headChildren[2]).toContain('teacher-directory-head-cell-student-count')

    const rows = wrapper.findAll('.teacher-directory-row')
    const firstRowChildren = Array.from(rows[0].element.children).map((element) =>
      element.className.toString()
    )
    expect(firstRowChildren[0]).toContain('teacher-directory-cell-class-code')
    expect(firstRowChildren[1]).toContain('teacher-directory-cell-class-name')
    expect(firstRowChildren[2]).toContain('teacher-directory-cell-student-count')
    expect(rows[0].find('.teacher-directory-cell-class-code').text()).toContain('CL-01')
    expect(rows[0].find('.teacher-directory-cell-class-name').text()).toBe('Class A')
    expect(rows[0].find('.teacher-directory-cell-student-count').text()).toBe('2')
    expect(rows[0].find('.teacher-directory-row-title').attributes('title')).toBe('Class A')
    expect(rows[0].find('.teacher-directory-row-tags').exists()).toBe(false)
    expect(rows[0].find('.teacher-directory-row-metrics').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Large')
    expect(wrapper.text()).not.toContain('Standard')
    expect(wrapper.text()).not.toContain('Compact')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('Class B')

    await wrapper.findAll('button').find((node) => node.text().includes('进入'))?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherClassStudents',
      params: { className: 'Class A' },
    })
  })

  it('应该支持按班级编号或班级名称筛选', async () => {
    const wrapper = mount(ClassManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    await wrapper.get('#class-manage-tab-directory').trigger('click')
    await flushPromises()

    expect(wrapper.find('.teacher-controls').exists()).toBe(true)
    expect(wrapper.find('.teacher-controls-title').text()).toContain('班级筛选')
    expect(wrapper.find('.teacher-filter-control').exists()).toBe(true)
    const searchInput = wrapper.find('input[placeholder="搜索班级编号或名称"]')
    expect(searchInput.exists()).toBe(true)

    await searchInput.setValue('CL-02')
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.text()).toContain('Class B')
    expect(wrapper.text()).not.toContain('Class A')

    await searchInput.setValue('Class A')
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).not.toContain('Class B')

    await searchInput.setValue('unknown')
    expect(wrapper.text()).toContain('没有匹配班级')
  })

  it('应该为班级名称保留单行省略和完整悬浮提示', () => {
    expect(classManagementSource).toContain('role="tablist"')
    expect(classManagementSource).toContain('class-manage-tab-overview')
    expect(classManagementSource).toContain('class-manage-tab-directory')
    expect(classManagementSource).toContain('class-manage-overview')
    expect(classManagementSource).toContain('class-manage-directory')
    expect(classManagementSource).toMatch(/class="teacher-directory-row-title"[\s\S]*:title="item\.name"/s)
    expect(classManagementSource).toMatch(/\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
  })
})

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import ClassManagement from '../ClassManagement.vue'
import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import { useAuthStore } from '@/stores/auth'

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
  const reportDialogStub = {
    name: 'TeacherClassReportExportDialog',
    props: ['modelValue', 'defaultClassName'],
    template:
      '<div data-testid="class-report-dialog" :data-open="String(modelValue)" :data-default-class-name="defaultClassName || \'\'" />',
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClasses.mockResolvedValue({
      list: [
        { name: 'Class A', student_count: 2 },
        { name: 'Class B', student_count: 3 },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        class_name: 'Class A',
      })
  })

  it('应该展示班级列表并支持进入班级学生页', async () => {
    const wrapper = mount(ClassManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.find('.top-tabs').exists()).toBe(false)
    expect(wrapper.find('#class-manage-tab-overview').exists()).toBe(false)
    expect(wrapper.find('#class-manage-tab-directory').exists()).toBe(false)
    expect(wrapper.findAll('.metric-panel-card')).toHaveLength(2)
    expect(wrapper.text()).not.toContain('已就绪')
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(2)
    expect(wrapper.find('.teacher-directory-pagination').exists()).toBe(true)
    expect(wrapper.find('.teacher-directory-head').text()).toContain('班级编号')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('班级名称')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('学生数')
    expect(wrapper.find('.teacher-directory-head').text()).not.toContain('标签')
    expect(wrapper.find('.teacher-directory-head').text()).not.toContain('数据')

    const headChildren = Array.from(wrapper.find('.teacher-directory-head').element.children).map(
      (element) => element.className.toString()
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

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('进入'))
      ?.trigger('click')

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
        stubs: {
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('.workspace-directory-section.teacher-directory-section').exists()).toBe(
      true
    )
    expect(wrapper.find('.list-heading').exists()).toBe(true)
    expect(wrapper.find('.teacher-filter-control').exists()).toBe(true)
    const searchInput = wrapper.find('input[placeholder="搜索班级编号或名称"]')
    expect(searchInput.exists()).toBe(true)
    expect(wrapper.text()).not.toContain('班级筛选')
    expect(wrapper.text()).not.toContain('支持按班级编号或班级名称快速定位班级入口。')

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

  it('应该支持切换班级目录分页', async () => {
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClasses
      .mockResolvedValueOnce({
        list: Array.from({ length: 20 }, (_, index) => ({
          name: `Class ${index + 1}`,
          student_count: index + 1,
        })),
        total: 21,
        page: 1,
        page_size: 20,
      })
      .mockResolvedValueOnce({
        list: [{ name: 'Class 21', student_count: 21 }],
        total: 21,
        page: 2,
        page_size: 20,
      })

    const wrapper = mount(ClassManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('共 21 个班级')
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('1 / 2')

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')
    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenNthCalledWith(2, { page: 2, page_size: 20 })
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.text()).toContain('Class 21')
    expect(wrapper.text()).not.toContain('Class 20')
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('2 / 2')
  })

  it('应该为班级名称保留单行省略和完整悬浮提示', () => {
    expect(classManagementSource).not.toContain('role="tablist"')
    expect(classManagementSource).not.toContain('class-manage-tab-overview')
    expect(classManagementSource).not.toContain('class-manage-tab-directory')
    expect(classManagementSource).not.toContain('class-manage-overview')
    expect(classManagementSource).not.toContain('class-manage-directory')
    expect(classManagementSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(classManagementSource).toContain('class="list-heading"')
    expect(classManagementSource).not.toContain('teacher-controls-title')
    expect(classManagementSource).not.toContain('teacher-controls-copy')
    expect(classManagementSource).not.toContain('班级筛选')
    expect(classManagementSource).not.toContain('支持按班级编号或班级名称快速定位班级入口。')
    expect(classManagementSource).toMatch(
      /class="teacher-directory-row-title"[\s\S]*:title="item\.name"/s
    )
    expect(classManagementSource).toMatch(
      /\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
  })

  it('班级管理概况卡片应复用教学概览的默认 metric-panel 外观', () => {
    expect(classManagementSource).toContain('class="teacher-summary metric-panel-default-surface"')
    expect(classManagementSource).not.toContain('teacher-summary--overview-metrics')
    expect(classManagementSource).not.toContain('当前状态')
    expect(classManagementSource).not.toContain('已就绪')
  })

  it('班级管理概览头部应与学生管理页使用同一套 teacher header 结构', () => {
    expect(classManagementSource).toContain('<header class="teacher-topbar">')
    expect(classManagementSource).toContain('<div class="teacher-heading">')
    expect(classManagementSource).toContain('<h1 class="teacher-title">班级管理</h1>')
    expect(classManagementSource).not.toContain(
      '<div class="teacher-heading workspace-tab-heading__main">'
    )
    expect(classManagementSource).not.toContain(
      '<h1 class="teacher-title workspace-tab-heading__title">班级管理</h1>'
    )
  })

  it('点击导出班级报告时应打开上下文对话框', async () => {
    const wrapper = mount(ClassManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('导出班级报告'))
      ?.trigger('click')
    await flushPromises()

    const dialog = wrapper.get('[data-testid="class-report-dialog"]')
    expect(dialog.attributes('data-open')).toBe('true')
    expect(dialog.attributes('data-default-class-name')).toBe('Class A')
    expect(pushMock).not.toHaveBeenCalledWith({ name: 'TeacherAWDReviewIndex' })
  })
})

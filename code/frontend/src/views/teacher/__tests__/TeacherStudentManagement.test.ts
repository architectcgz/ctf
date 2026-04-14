import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import TeacherStudentManagement from '../TeacherStudentManagement.vue'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('TeacherStudentManagement', () => {
  const reportDialogStub = {
    name: 'TeacherClassReportExportDialog',
    props: ['modelValue', 'defaultClassName'],
    template:
      '<div data-testid="class-report-dialog" :data-open="String(modelValue)" :data-default-class-name="defaultClassName || \'\'" />',
  }

  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClassStudents.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockImplementation(async (_className, params) => {
      if (params?.keyword === 'alice') {
        return [
          {
            id: 'stu-1',
            username: 'alice',
            name: 'Alice Zhang',
            student_no: '2024001',
            recent_event_count: 0,
          },
        ]
      }
      if (params?.keyword === 'Alice') {
        return [
          {
            id: 'stu-1',
            username: 'alice',
            name: 'Alice Zhang',
            student_no: '2024001',
            recent_event_count: 0,
          },
        ]
      }
      return [
        {
          id: 'stu-1',
          username: 'alice',
          name: 'Alice Zhang',
          student_no: '2024001',
          recent_event_count: 0,
        },
        { id: 'stu-2', username: 'bob', recent_event_count: 2, solved_count: 1 },
      ]
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        class_name: 'Class A',
      },
      'token'
    )
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应该支持搜索学生并进入学员分析页', async () => {
    const wrapper = mount(TeacherStudentManagement, {
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

    expect(wrapper.text()).toContain('学生管理')
    expect(wrapper.find('.workspace-directory-section.teacher-directory-section').exists()).toBe(
      true
    )
    expect(wrapper.find('.list-heading').exists()).toBe(true)
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(2)
    expect(wrapper.find('.teacher-directory-head').text()).toContain('学号')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('学生名称')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('昵称')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('薄弱项')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('做题数')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('得分数')
    expect(wrapper.find('.teacher-directory-head').text()).not.toContain('数据')
    const headChildren = Array.from(wrapper.find('.teacher-directory-head').element.children).map(
      (element) => element.className.toString()
    )
    expect(headChildren[0]).toContain('teacher-directory-head-cell-student-no')
    expect(headChildren[1]).toContain('teacher-directory-head-cell-name')
    expect(headChildren[2]).toContain('teacher-directory-head-cell-alias')

    const rows = wrapper.findAll('.teacher-directory-row')
    const firstRowChildren = Array.from(rows[0].element.children).map((element) =>
      element.className.toString()
    )
    expect(firstRowChildren[0]).toContain('teacher-directory-cell-student-no')
    expect(firstRowChildren[1]).toContain('teacher-directory-cell-name')
    expect(firstRowChildren[2]).toContain('teacher-directory-cell-alias')
    expect(rows[0].find('.teacher-directory-cell-student-no').text()).toContain('2024001')
    expect(rows[0].find('.teacher-directory-cell-name').text()).toContain('Alice Zhang')
    expect(rows[0].find('.teacher-directory-cell-alias').text()).toContain('alice')
    expect(rows[0].find('.teacher-directory-row-title').attributes('title')).toBe('Alice Zhang')
    expect(rows[0].find('.teacher-directory-row-points').attributes('title')).toBe('alice')
    expect(rows[0].find('.teacher-directory-row-tags').text()).toContain('暂无薄弱项')
    expect(rows[0].find('.teacher-directory-row-solved').text()).toBe('0')
    expect(rows[0].find('.teacher-directory-row-score').text()).toBe('0')
    expect(rows[0].find('.teacher-directory-row-tags').text()).not.toContain('Student')
    expect(rows[1].find('.teacher-directory-cell-student-no').text()).toContain('未设置学号')
    expect(rows[1].find('.teacher-directory-cell-name').text()).toContain('未设置姓名')
    expect(rows[1].find('.teacher-directory-cell-alias').text()).toContain('bob')
    expect(rows[1].find('.teacher-directory-row-tags').text()).toContain('暂无薄弱项')
    expect(rows[1].find('.teacher-directory-row-solved').text()).toBe('1')
    expect(rows[1].find('.teacher-directory-row-score').text()).toBe('0')
    expect(rows[1].find('.teacher-directory-row-status').exists()).toBe(false)
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('bob')
    expect(wrapper.text()).not.toContain('学生筛选')

    const searchInput = wrapper.find('input[placeholder="搜索姓名或用户名"]')
    await searchInput.setValue('Alice')
    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledTimes(1)
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenLastCalledWith('Class A', {
      keyword: 'Alice',
      student_no: undefined,
    })
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')

    wrapper.findComponent({ name: 'StudentManagementPage' }).vm.$emit('openStudent', 'stu-1')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })

  it('学生管理概况卡片应复用 admin dashboard 的共享数值卡片结构', () => {
    expect(studentManagementSource).toContain('class="teacher-summary metric-panel-default-surface"')
    expect(studentManagementSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid"'
    )
    expect(studentManagementSource).toContain('class="progress-card metric-panel-card"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-item progress-card metric-panel-card"'
    )
    expect(studentManagementSource).toContain('class="progress-card-label metric-panel-label"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-label progress-card-label metric-panel-label"'
    )
    expect(studentManagementSource).toContain('class="progress-card-value metric-panel-value"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-value progress-card-value metric-panel-value"'
    )
    expect(studentManagementSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-helper progress-card-hint metric-panel-helper"'
    )
  })

  it('应该忽略过期搜索请求的返回结果', async () => {
    const slowRequest = deferred<
      Array<{
        id: string
        username: string
        name?: string
        student_no?: string
        recent_event_count?: number
        solved_count?: number
      }>
    >()
    const fastRequest = deferred<
      Array<{
        id: string
        username: string
        name?: string
        student_no?: string
        recent_event_count?: number
        solved_count?: number
      }>
    >()

    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassStudents
      .mockResolvedValueOnce([
        {
          id: 'stu-1',
          username: 'alice',
          name: 'Alice Zhang',
          student_no: '2024001',
          recent_event_count: 0,
        },
        { id: 'stu-2', username: 'bob', recent_event_count: 2, solved_count: 1 },
      ])
      .mockImplementationOnce(() => slowRequest.promise)
      .mockImplementationOnce(() => fastRequest.promise)

    const wrapper = mount(TeacherStudentManagement, {
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

    const searchInput = wrapper.find('input[placeholder="搜索姓名或用户名"]')
    await searchInput.setValue('A')
    vi.advanceTimersByTime(250)
    await flushPromises()
    await searchInput.setValue('Ali')
    vi.advanceTimersByTime(250)

    fastRequest.resolve([
      {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice Zhang',
        student_no: '2024001',
        recent_event_count: 0,
      },
    ])
    await flushPromises()

    slowRequest.resolve([{ id: 'stu-2', username: 'bob', recent_event_count: 2 }])
    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')
  })

  it('班级筛选为空时应该显示全部学生并保留学生所属班级跳转', async () => {
    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 1 },
      { name: 'Class B', student_count: 2 },
    ])
    teacherApiMocks.getClassStudents.mockImplementation(async (className, params) => {
      if (className === 'Class A') {
        if (params?.keyword === 'Carol') {
          return []
        }
        return [
          {
            id: 'stu-1',
            username: 'alice',
            name: 'Alice Zhang',
            student_no: '2024001',
            recent_event_count: 0,
          },
        ]
      }

      if (params?.keyword === 'Carol') {
        return [
          {
            id: 'stu-3',
            username: 'carol',
            name: 'Carol Chen',
            student_no: '2024003',
            recent_event_count: 1,
          },
        ]
      }

      return [
        {
          id: 'stu-2',
          username: 'bob',
          name: 'Bob Li',
          student_no: '2024002',
          recent_event_count: 2,
        },
        {
          id: 'stu-3',
          username: 'carol',
          name: 'Carol Chen',
          student_no: '2024003',
          recent_event_count: 1,
        },
      ]
    })

    const wrapper = mount(TeacherStudentManagement, {
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

    const classSelect = wrapper.find('select')
    const options = classSelect.findAll('option').map((option) => ({
      value: option.element.getAttribute('value'),
      text: option.text(),
    }))

    expect(options).toEqual([
      { value: '', text: '全部班级' },
      { value: 'Class A', text: 'Class A · 1' },
      { value: 'Class B', text: 'Class B · 2' },
    ])

    await classSelect.setValue('')
    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenNthCalledWith(2, 'Class A', {
      keyword: undefined,
      student_no: undefined,
    })
    expect(teacherApiMocks.getClassStudents).toHaveBeenNthCalledWith(3, 'Class B', {
      keyword: undefined,
      student_no: undefined,
    })
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(3)
    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).toContain('Bob Li')
    expect(wrapper.text()).toContain('Carol Chen')

    const searchInput = wrapper.find('input[placeholder="搜索姓名或用户名"]')
    await searchInput.setValue('Carol')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenNthCalledWith(4, 'Class A', {
      keyword: 'Carol',
      student_no: undefined,
    })
    expect(teacherApiMocks.getClassStudents).toHaveBeenNthCalledWith(5, 'Class B', {
      keyword: 'Carol',
      student_no: undefined,
    })
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.text()).toContain('Carol Chen')
    expect(wrapper.text()).not.toContain('Alice Zhang')

    const carolRow = wrapper
      .findAll('.teacher-directory-row')
      .find((row) => row.text().includes('Carol Chen'))
    expect(carolRow).toBeDefined()
    await carolRow!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentAnalysis',
      params: { className: 'Class B', studentId: 'stu-3' },
    })
  })

  it('默认班级不是可访问班级时应该回退为全部班级并显示全部学生', async () => {
    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 1 },
      { name: 'Class B', student_count: 1 },
    ])
    teacherApiMocks.getClassStudents.mockImplementation(async (className) => {
      if (className === 'Class A') {
        return [
          {
            id: 'stu-1',
            username: 'alice',
            name: 'Alice Zhang',
            student_no: '2024001',
          },
        ]
      }

      return [
        {
          id: 'stu-2',
          username: 'bob',
          name: 'Bob Li',
          student_no: '2024002',
        },
      ]
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        class_name: 'Missing Class',
      },
      'token'
    )

    const wrapper = mount(TeacherStudentManagement, {
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

    expect(wrapper.find('select').element.value).toBe('')
    expect(teacherApiMocks.getClassStudents).toHaveBeenNthCalledWith(1, 'Class A', {
      keyword: undefined,
      student_no: undefined,
    })
    expect(teacherApiMocks.getClassStudents).toHaveBeenNthCalledWith(2, 'Class B', {
      keyword: undefined,
      student_no: undefined,
    })
    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).toContain('Bob Li')
  })

  it('应该支持学生目录分页切换', async () => {
    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 21 }])
    teacherApiMocks.getClassStudents.mockResolvedValue(
      Array.from({ length: 21 }, (_, index) => ({
        id: `stu-${index + 1}`,
        username: `student-${index + 1}`,
        name: `Student ${index + 1}`,
        student_no: `2024${String(index + 1).padStart(3, '0')}`,
      }))
    )

    const wrapper = mount(TeacherStudentManagement, {
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

    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(20)
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('共 21 名学生')
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('1 / 2')
    expect(wrapper.text()).toContain('Student 20')
    expect(wrapper.text()).not.toContain('Student 21')

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')
    await flushPromises()

    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('2 / 2')
    expect(wrapper.text()).toContain('Student 21')
    expect(wrapper.text()).not.toContain('Student 20')
  })

  it('应该为学生列表姓名和昵称保留单行省略与完整提示', () => {
    expect(studentManagementSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(studentManagementSource).toContain('class="list-heading"')
    expect(studentManagementSource).not.toContain('teacher-controls-title')
    expect(studentManagementSource).not.toContain('学生筛选')
    expect(studentManagementSource).toContain('<span>做题数</span>')
    expect(studentManagementSource).toContain('<span>得分数</span>')
    expect(studentManagementSource).toContain('class="teacher-directory-row-solved"')
    expect(studentManagementSource).toContain('class="teacher-directory-row-score"')
    expect(studentManagementSource).not.toContain('<span>数据</span>')
    expect(studentManagementSource).not.toContain('class="teacher-directory-row-metrics"')
    expect(studentManagementSource).toMatch(
      /class="teacher-directory-row-title"[\s\S]*:title="student\.name \|\| '未设置姓名'"/s
    )
    expect(studentManagementSource).toMatch(
      /class="teacher-directory-row-points"[\s\S]*:title="student\.username"/s
    )
    expect(studentManagementSource).toMatch(
      /\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(studentManagementSource).toMatch(
      /\.teacher-directory-row-points\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
  })

  it('点击导出班级报告时应打开当前筛选班级的上下文对话框', async () => {
    const wrapper = mount(TeacherStudentManagement, {
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
      .find((button) => button.text().includes('导出班级报告'))
      ?.trigger('click')
    await flushPromises()

    const dialog = wrapper.get('[data-testid="class-report-dialog"]')
    expect(dialog.attributes('data-open')).toBe('true')
    expect(dialog.attributes('data-default-class-name')).toBe('Class A')
    expect(pushMock).not.toHaveBeenCalledWith({ name: 'TeacherAWDReviewIndex' })
  })
})

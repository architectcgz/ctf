import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

import TeacherClassStudents from '../TeacherClassStudents.vue'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'

const pushMock = vi.fn()
const routeMock = {
  params: {
    className: 'Class A',
  },
  query: {
    panel: 'students',
  },
}

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
  getClassReview: vi.fn(),
  getClassSummary: vi.fn(),
  getClassTrend: vi.fn(),
  getStudentRecommendations: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
    useRoute: () => routeMock,
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

describe('TeacherClassStudents', () => {
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
    routeMock.params.className = 'Class A'
    routeMock.query.panel = 'students'
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassReview.mockReset()
    teacherApiMocks.getClassSummary.mockReset()
    teacherApiMocks.getClassTrend.mockReset()
    teacherApiMocks.getStudentRecommendations.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockResolvedValue([
      {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice Zhang',
        solved_count: 3,
        total_score: 280,
        recent_event_count: 0,
        weak_dimension: 'crypto',
      },
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 1,
        total_score: 100,
        recent_event_count: 2,
        weak_dimension: 'pwn',
      },
    ])
    teacherApiMocks.getStudentRecommendations.mockResolvedValue([
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        reason: '针对薄弱维度：密码',
      },
    ])
    teacherApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          key: 'activity',
          title: '班级活跃度需要补强',
          detail: 'Class A 近 7 天活跃率为 50%，适合通过定向训练把低活跃学生重新拉回训练节奏。',
          accent: 'warning',
        },
        {
          key: 'weak_dimension',
          title: '优先补薄弱维度',
          detail: 'crypto 是当前最集中的薄弱项，涉及 1 名学生，建议本周统一布置该维度基础题。',
          accent: 'primary',
          students: [{ id: 'stu-1', username: 'alice' }],
          recommendation: {
            challenge_id: '12',
            title: 'crypto-lab',
            category: 'crypto',
            difficulty: 'medium',
            reason: '针对薄弱维度：密码',
          },
        },
        {
          key: 'focus_students',
          title: '先跟进重点学生',
          detail: '建议教师先跟进 alice，并优先布置推荐题做补强训练。',
          accent: 'primary',
          students: [{ id: 'stu-1', username: 'alice' }],
        },
      ],
    })
    teacherApiMocks.getClassSummary.mockResolvedValue({
      class_name: 'Class A',
      student_count: 2,
      average_solved: 2,
      active_student_count: 1,
      active_rate: 50,
      recent_event_count: 6,
    })
    teacherApiMocks.getClassTrend.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 1, event_count: 3, solve_count: 1 },
      ],
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

  it('应该展示班级学生列表并支持进入学员分析页', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('学生列表')
    expect(wrapper.find('.teacher-topbar').exists()).toBe(true)
    expect(wrapper.find('.teacher-summary').exists()).toBe(true)
    expect(wrapper.find('.teacher-controls').exists()).toBe(true)
    expect(wrapper.find('[role="tablist"]').exists()).toBe(true)
    expect(wrapper.find('#class-tab-overview').exists()).toBe(true)
    expect(wrapper.find('#class-tab-trend').exists()).toBe(true)
    expect(wrapper.find('#class-tab-students').exists()).toBe(true)
    expect(wrapper.find('#class-tab-review').exists()).toBe(true)
    expect(wrapper.find('#class-tab-insight').exists()).toBe(true)
    expect(wrapper.find('#class-tab-action').exists()).toBe(true)
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('bob')
    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('280')
    expect(wrapper.text()).toContain('crypto')
    expect(wrapper.text()).toContain('教学复盘结论')
    expect(wrapper.text()).toContain('班级活跃度需要补强')
    expect(wrapper.text()).toContain('优先补薄弱维度')
    expect(wrapper.text()).toContain('先跟进重点学生')
    expect(wrapper.text()).toContain('班级 Top 学生')
    expect(wrapper.text()).toContain('薄弱维度分布')
    expect(wrapper.text()).toContain('优先介入学生')
    expect(wrapper.text()).toContain('建议训练题')
    expect(wrapper.text()).toContain('推荐训练题')
    expect(wrapper.text()).toContain('crypto-lab')
    await wrapper.find('#class-tab-students').trigger('click')
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(2)
    expect(wrapper.find('.teacher-directory-head').text()).toContain('学号')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('学生名称')
    expect(wrapper.find('.teacher-directory-head').text()).toContain('昵称')
    expect(wrapper.find('.teacher-directory-row-title').attributes('title')).toBe('Alice Zhang')
    expect(wrapper.find('.teacher-directory-row-points').attributes('title')).toBe('alice')
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledWith('Class A')
    expect(teacherApiMocks.getStudentRecommendations).toHaveBeenCalledWith('stu-1')

    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openStudent', 'stu-1')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })

  it('班级详情页应采用与教学概览一致的顶部 tabs 壳层结构，并去掉页面内重复顶栏', () => {
    expect(classStudentsPageSource).toMatch(/class="[^"]*\bworkspace-shell\b[^"]*"/)
    expect(classStudentsPageSource).toMatch(/class="[^"]*\bteacher-management-shell\b[^"]*"/)
    expect(classStudentsPageSource).not.toContain('class="workspace-topbar"')
    expect(classStudentsPageSource).toContain('class="top-tabs"')
    expect(classStudentsPageSource).toContain('class="content-pane"')
    expect(classStudentsPageSource).toMatch(
      /<div class="[^"]*\bworkspace-shell\b[^"]*">[\s\S]*<nav class="top-tabs"[\s\S]*<main class="content-pane">/s
    )
    expect(classStudentsPageSource).toMatch(
      /class="teacher-directory-row-title"[\s\S]*:title="student\.name \|\| '未设置姓名'"/s
    )
    expect(classStudentsPageSource).toMatch(
      /class="teacher-directory-row-points"[\s\S]*:title="student\.username"/s
    )
    expect(classStudentsPageSource).toMatch(
      /\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(classStudentsPageSource).toMatch(
      /\.teacher-directory-row-points\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
  })

  it('应该保留已解码的班级名并使用原值请求学生列表', async () => {
    routeMock.params.className = '100% 班级'

    mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('100% 班级', {
      student_no: undefined,
    })
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledWith('100% 班级')
    expect(teacherApiMocks.getClassSummary).toHaveBeenCalledWith('100% 班级')
    expect(teacherApiMocks.getClassTrend).toHaveBeenCalledWith('100% 班级')
  })

  it('管理员从班级详情返回班级管理时应回到后台班级页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        class_name: 'Class A',
      })

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openClassManagement')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformClassManagement' })
  })

  it('选择班级下拉框后应跳转到对应班级页面并保持 panel 查询参数', async () => {
    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 1 },
    ])

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()
    await wrapper.find('#class-tab-students').trigger('click')

    const classSelect = wrapper.find('select[aria-label="选择班级"]')
    expect(classSelect.exists()).toBe(true)

    await classSelect.setValue('Class B')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherClassStudents',
      params: { className: 'Class B' },
      query: { panel: 'students' },
    })
  })

  it('应该忽略过期学号搜索请求的返回结果', async () => {
    const slowRequest = deferred<
      Array<{
        id: string
        username: string
        name?: string
        solved_count?: number
        total_score?: number
        recent_event_count?: number
        weak_dimension?: string
      }>
    >()
    const fastRequest = deferred<
      Array<{
        id: string
        username: string
        name?: string
        solved_count?: number
        total_score?: number
        recent_event_count?: number
        weak_dimension?: string
      }>
    >()

    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassReview.mockReset()
    teacherApiMocks.getStudentRecommendations.mockReset()
    teacherApiMocks.getClassSummary.mockReset()
    teacherApiMocks.getClassTrend.mockReset()
    teacherApiMocks.getClassStudents
      .mockResolvedValueOnce([
        {
          id: 'stu-1',
          username: 'alice',
          name: 'Alice Zhang',
          solved_count: 3,
          total_score: 280,
          recent_event_count: 0,
          weak_dimension: 'crypto',
        },
        {
          id: 'stu-2',
          username: 'bob',
          solved_count: 1,
          total_score: 100,
          recent_event_count: 2,
          weak_dimension: 'pwn',
        },
      ])
      .mockImplementationOnce(() => slowRequest.promise)
      .mockImplementationOnce(() => fastRequest.promise)
    teacherApiMocks.getClassSummary.mockResolvedValue({
      class_name: 'Class A',
      student_count: 2,
      average_solved: 2,
      active_student_count: 1,
      active_rate: 50,
      recent_event_count: 6,
    })
    teacherApiMocks.getClassTrend.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 1, event_count: 3, solve_count: 1 },
      ],
    })
    teacherApiMocks.getStudentRecommendations.mockResolvedValue([
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        reason: '针对薄弱维度：密码',
      },
    ])
    teacherApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          key: 'activity',
          title: '班级活跃度需要补强',
          detail: 'Class A 近 7 天活跃率为 50%，适合通过定向训练把低活跃学生重新拉回训练节奏。',
          accent: 'warning',
        },
      ],
    })

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await wrapper.find('#class-tab-students').trigger('click')

    const studentNoInput = wrapper.find('input[placeholder="输入学号精确查询"]')
    await studentNoInput.setValue('20260001')
    await studentNoInput.setValue('20260002')

    fastRequest.resolve([
      {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice Zhang',
        solved_count: 3,
        total_score: 280,
        recent_event_count: 0,
        weak_dimension: 'crypto',
      },
    ])
    await flushPromises()

    slowRequest.resolve([
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 1,
        total_score: 100,
        recent_event_count: 2,
        weak_dimension: 'pwn',
      },
    ])
    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')
  })

  it('按学号筛选时应该只刷新学生列表，不重复请求班级概览数据', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    await wrapper.find('#class-tab-students').trigger('click')

    expect(wrapper.text()).not.toContain('学生筛选')

    const studentNoInput = wrapper.find('input[placeholder="输入学号精确查询"]')
    expect(studentNoInput.exists()).toBe(true)

    await studentNoInput.setValue('20260001')
    await flushPromises()

    expect(wrapper.find('.teacher-filter-reset').exists()).toBe(true)
    expect(wrapper.find('.teacher-filter-clear').exists()).toBe(true)

    await wrapper.find('.teacher-filter-clear').trigger('click')
    await flushPromises()

    expect((studentNoInput.element as HTMLInputElement).value).toBe('')
    expect(wrapper.find('.teacher-filter-reset').exists()).toBe(false)

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('Class A', {
      student_no: '20260001',
    })
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getClassSummary).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getClassTrend).toHaveBeenCalledTimes(1)
  })

  it('点击导出班级报告时应打开当前班级上下文对话框', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
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

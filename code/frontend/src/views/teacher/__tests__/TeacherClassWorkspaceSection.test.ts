import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

import TeacherClassWorkspaceSection from '../TeacherClassWorkspaceSection.vue'
import classWorkspaceSectionPageSource from '@/components/teacher/class-management/ClassWorkspaceSectionPage.vue?raw'

const pushMock = vi.fn()
const routeMock = {
  name: 'TeacherClassTrend',
  params: {
    className: 'Class A',
  },
  query: {
    panel: 'review',
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

describe('TeacherClassWorkspaceSection', () => {
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
    routeMock.name = 'TeacherClassTrend'
    routeMock.params.className = 'Class A'
    routeMock.query.panel = 'review'

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
    teacherApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          key: 'activity',
          title: '班级活跃度需要补强',
          detail: 'Class A 近 7 天活跃率为 50%。',
          accent: 'warning',
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
    teacherApiMocks.getStudentRecommendations.mockResolvedValue([
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        reason: '针对薄弱维度：密码',
      },
    ])

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

  it('应在独立页面承载班级工作区视角，而不是把学生列表继续堆在同一页', async () => {
    const wrapper = mount(TeacherClassWorkspaceSection, {
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

    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('班级训练趋势')
    expect(wrapper.text()).toContain('班级近 7 天训练趋势')
    expect(wrapper.text()).not.toContain('学生列表')

    wrapper
      .findComponent({ name: 'ClassWorkspaceSectionPage' })
      .vm.$emit('openClassOverview')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherClassStudents',
      params: { className: 'Class A' },
    })
  })

  it('切换班级时应保留当前工作区子页面语义，并清理旧 panel 查询参数', async () => {
    routeMock.name = 'TeacherClassReview'

    const wrapper = mount(TeacherClassWorkspaceSection, {
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

    wrapper.findComponent({ name: 'ClassWorkspaceSectionPage' }).vm.$emit('selectClass', 'Class B')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherClassReview',
      params: { className: 'Class B' },
      query: {},
    })
  })

  it('管理员进入独立子页面时，返回总览应使用后台路由', async () => {
    routeMock.name = 'AdminClassInsights'

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        class_name: 'Class A',
      },
      'token'
    )

    const wrapper = mount(TeacherClassWorkspaceSection, {
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

    wrapper
      .findComponent({ name: 'ClassWorkspaceSectionPage' })
      .vm.$emit('openClassOverview')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'AdminClassStudents',
      params: { className: 'Class A' },
    })
  })

  it('独立工作区页面应复用共享 shell，并通过动态组件承载当前视角内容', () => {
    expect(classWorkspaceSectionPageSource).toContain('class="workspace-shell teacher-management-shell teacher-surface"')
    expect(classWorkspaceSectionPageSource).not.toContain('class="top-tabs"')
    expect(classWorkspaceSectionPageSource).toContain(':is="activeSection.component"')
    expect(classWorkspaceSectionPageSource).toContain('class="workspace-subpanel workspace-subpanel--flat"')
    expect(classWorkspaceSectionPageSource).not.toContain('<TeacherClassTrendPanel')
    expect(classWorkspaceSectionPageSource).not.toContain('<TeacherClassReviewPanel')
    expect(classWorkspaceSectionPageSource).not.toContain('<TeacherClassInsightsPanel')
    expect(classWorkspaceSectionPageSource).not.toContain('<TeacherInterventionPanel')
  })
})

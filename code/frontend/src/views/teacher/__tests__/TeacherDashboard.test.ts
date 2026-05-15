import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherDashboard from '../TeacherDashboard.vue'
import teacherDashboardSource from '../TeacherDashboard.vue?raw'
import teacherDashboardPageSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()
const teacherSurfacePattern =
  /--journal-ink:\s*var\(--color-text-primary\);[\s\S]*--journal-surface:\s*color-mix\(in srgb, var\(--color-bg-surface\) 88%, var\(--color-bg-base\)\);/s
const forbiddenTeacherSurfaceLiterals = ['rgba(255, 255, 255, 0.98)', '#ffffff', '#f8fafc']

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherOverview: vi.fn(),
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
  getClassReview: vi.fn(),
  getClassSummary: vi.fn(),
  getClassTrend: vi.fn(),
  getStudentRecommendations: vi.fn(),
  getStudentProgress: vi.fn(),
  getStudentSkillProfile: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

async function mountDashboard(path = '/academy/overview') {
  window.history.replaceState(window.history.state, '', path)

  const wrapper = mount(TeacherDashboard, {
    global: {
      stubs: {
        LineChart: true,
        SkillRadar: true,
      },
    },
  })

  await flushPromises()
  await flushPromises()
  return wrapper
}

describe('TeacherDashboard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()

    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getTeacherOverview.mockResolvedValue({
      summary: {
        class_count: 2,
        student_count: 5,
        active_student_count: 3,
        active_rate: 60,
        average_solved: 3.4,
        recent_event_count: 12,
        risk_student_count: 1,
      },
      trend: {
        points: [
          { date: '2026-03-05', active_student_count: 2, event_count: 5, solve_count: 2 },
          { date: '2026-03-06', active_student_count: 3, event_count: 7, solve_count: 4 },
        ],
      },
      focus_classes: [
        {
          class_name: 'Class A',
          student_count: 3,
          active_rate: 50,
          recent_event_count: 5,
          risk_student_count: 1,
          dominant_weak_dimension: 'crypto',
        },
      ],
      focus_students: [
        {
          id: 'stu-1',
          username: 'alice',
          name: 'Alice',
          class_name: 'Class A',
          solved_count: 4,
          total_score: 320,
          recent_event_count: 0,
          weak_dimension: 'crypto',
        },
      ],
      spotlight_student: {
        id: 'stu-2',
        username: 'bob',
        name: 'Bob',
        class_name: 'Class B',
        solved_count: 6,
        total_score: 430,
        recent_event_count: 4,
        weak_dimension: 'web',
      },
      weak_dimensions: [
        { dimension: 'crypto', student_count: 2 },
        { dimension: 'pwn', student_count: 1 },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    })
  })

  it('应该展示教师概览且不再加载班级详情接口', async () => {
    const wrapper = await mountDashboard()

    expect(teacherApiMocks.getTeacherOverview).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getClasses).not.toHaveBeenCalled()
    expect(teacherApiMocks.getClassStudents).not.toHaveBeenCalled()
    expect(teacherApiMocks.getClassReview).not.toHaveBeenCalled()
    expect(teacherApiMocks.getClassSummary).not.toHaveBeenCalled()
    expect(teacherApiMocks.getClassTrend).not.toHaveBeenCalled()
    expect(teacherApiMocks.getStudentRecommendations).not.toHaveBeenCalled()
    expect(teacherApiMocks.getStudentProgress).not.toHaveBeenCalled()
    expect(teacherApiMocks.getStudentSkillProfile).not.toHaveBeenCalled()

    expect(wrapper.text()).toContain('教学介入台')
    expect(wrapper.text()).toContain('当前覆盖 2 个班级')
    expect(wrapper.text()).toContain('2 个班级纳入总览')
    expect(wrapper.text()).toContain('Class A 复盘摘要')
    expect(wrapper.text()).toContain('crypto 方向仍是当前薄弱维度')
    expect(wrapper.text()).toContain('Bob 当前保持领先')
    expect(wrapper.text()).toContain('Alice')
    expect(wrapper.text()).toContain('薄弱项 crypto')
    expect(wrapper.find('#overview .teacher-dashboard-overview-head').exists()).toBe(true)
    expect(wrapper.find('#overview .workspace-overline').text()).toBe('Teaching Overview')

    expect(
      wrapper.findAll('#overview .teacher-overview-card.progress-card.metric-panel-card')
    ).toHaveLength(4)
    expect(wrapper.findAll('#portrait .summary-note.progress-card.metric-panel-card')).toHaveLength(
      3
    )
    expect(wrapper.findAll('#trend .focus-class-row')).toHaveLength(1)
    expect(wrapper.findAll('#review .review-highlight-item')).toHaveLength(1)
    expect(wrapper.findAll('#intervention .intervention-target-row')).toHaveLength(1)
  })

  it('路由页应仅负责组合 overview owner，不直接依赖教师接口实现', () => {
    expect(teacherDashboardSource).toContain('useTeacherOverviewPage')
    expect(teacherDashboardSource).not.toContain("from '@/api/teacher'")
    expect(teacherDashboardPageSource).toContain('useTeacherOverviewWorkspace')
    expect(teacherDashboardPageSource).not.toContain('TeacherClassTrendPanel')
    expect(teacherDashboardPageSource).not.toContain('TeacherClassReviewPanel')
    expect(teacherDashboardPageSource).not.toContain('TeacherInterventionPanel')
  })

  it('教师概览夜间模式样式应继续基于主题变量且不回流亮色硬编码', () => {
    expect(teacherDashboardPageSource).toMatch(teacherSurfacePattern)
    for (const literal of forbiddenTeacherSurfaceLiterals) {
      expect(
        teacherDashboardPageSource,
        `TeacherDashboardPage.vue contains forbidden literal: ${literal}`
      ).not.toContain(literal)
    }
  })

  it('教师概览应复用全局 workspace tab 与 metric-panel 样式栈', async () => {
    expect(teacherDashboardPageSource).toContain(
      'class="workspace-shell teacher-management-shell teacher-surface teacher-dashboard-shell flex min-h-full flex-1 flex-col"'
    )
    expect(teacherDashboardPageSource).toContain('class="workspace-tabbar top-tabs"')
    expect(teacherDashboardPageSource).toContain('class="workspace-tab top-tab"')
    expect(teacherDashboardPageSource).toContain(
      'class="workspace-panel-header teacher-dashboard-overview-head"'
    )
    expect(teacherDashboardPageSource).toContain(
      'class="workspace-panel-header__meta meta-strip"'
    )
    expect(teacherDashboardPageSource).toContain(
      'class="workspace-panel-header__summary teacher-overview-summary progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(teacherDashboardPageSource).toContain(
      'class="summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(teacherDashboardPageSource).toContain('--workspace-brand: var(--journal-accent);')
    expect(teacherDashboardPageSource).toContain(
      '--metric-panel-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(teacherDashboardPageSource).not.toContain('overview-pulse-panel')
    expect(teacherDashboardPageSource).toContain('class="workspace-overline"')
    expect(teacherDashboardPageSource).not.toContain('openReportExport')

    const wrapper = await mountDashboard()
    const summary = wrapper.get('#overview .teacher-overview-summary')
    const portraitSummary = wrapper.get('#portrait .summary-grid')

    expect(summary.classes()).toContain('progress-strip')
    expect(summary.classes()).toContain('metric-panel-grid')
    expect(summary.classes()).toContain('metric-panel-default-surface')
    expect(portraitSummary.classes()).toContain('progress-strip')
    expect(portraitSummary.classes()).toContain('metric-panel-grid')
    expect(portraitSummary.classes()).toContain('metric-panel-default-surface')
    expect(wrapper.text()).not.toContain('Quick Actions')
    expect(wrapper.text()).not.toContain('导出报告')
    expect(wrapper.text()).not.toContain('班级管理')
  })

  it('管理员从教师概览进入班级管理时应回到后台班级页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: 'Class A',
    })

    const wrapper = await mountDashboard()

    wrapper.findComponent({ name: 'TeacherDashboardPage' }).vm.$emit('openClassManagement')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformClassManagement' })
  })

  it('带 panel 查询参数进入时应激活对应教师概览 tab', async () => {
    const wrapper = await mountDashboard('/academy/overview?panel=portrait')

    expect(window.location.search).toBe('?panel=portrait')
    expect(wrapper.find('#dashboard-tab-portrait').classes()).toContain('active')
    expect(wrapper.find('#portrait').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#overview').attributes('aria-hidden')).toBe('true')
  })
})

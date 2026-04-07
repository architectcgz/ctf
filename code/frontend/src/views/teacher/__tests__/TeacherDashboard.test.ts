import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherDashboard from '../TeacherDashboard.vue'
import teacherDashboardPageSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import teacherClassReviewPanelSource from '@/components/teacher/TeacherClassReviewPanel.vue?raw'
import teacherInterventionPanelSource from '@/components/teacher/TeacherInterventionPanel.vue?raw'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()
const teacherSurfacePattern =
  /--journal-ink:\s*var\(--color-text-primary\);[\s\S]*--journal-surface:\s*color-mix\(in srgb, var\(--color-bg-surface\) 88%, var\(--color-bg-base\)\);/s
const forbiddenTeacherSurfaceLiterals = ['rgba(255, 255, 255, 0.98)', '#ffffff', '#f8fafc']
const teacherSurfaceSources = [
  ['TeacherDashboardPage.vue', teacherDashboardPageSource],
  ['TeacherClassReviewPanel.vue', teacherClassReviewPanelSource],
  ['TeacherInterventionPanel.vue', teacherInterventionPanelSource],
] as const

const teacherApiMocks = vi.hoisted(() => ({
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

describe('TeacherDashboard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()

    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockResolvedValue([
      {
        id: 'stu-1',
        username: 'alice',
        solved_count: 4,
        total_score: 320,
        recent_event_count: 0,
        weak_dimension: 'crypto',
      },
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 2,
        total_score: 180,
        recent_event_count: 3,
        weak_dimension: 'pwn',
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
      average_solved: 3.5,
      active_student_count: 2,
      active_rate: 100,
      recent_event_count: 8,
    })
    teacherApiMocks.getClassTrend.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 2, event_count: 4, solve_count: 2 },
      ],
    })
    teacherApiMocks.getStudentProgress.mockResolvedValue({
      total_challenges: 6,
      solved_challenges: 3,
      by_category: { web: { total: 3, solved: 2 } },
      by_difficulty: { easy: { total: 2, solved: 1 } },
    })
    teacherApiMocks.getStudentSkillProfile.mockResolvedValue({
      dimensions: [
        { key: 'web', name: 'Web', value: 75 },
        { key: 'crypto', name: '密码', value: 40 },
      ],
      updated_at: '2026-03-07T12:00:00Z',
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

  it('应该展示教师概览且不加载学员详情接口', async () => {
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

    expect(wrapper.text()).toContain('教学介入台')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('3.5')
    expect(wrapper.text()).toContain('100%')
    expect(wrapper.text()).toContain('教学复盘结论')
    expect(wrapper.text()).toContain('班级活跃度需要补强')
    expect(wrapper.text()).toContain('优先补薄弱维度')
    expect(wrapper.text()).toContain('先跟进重点学生')
    expect(wrapper.text()).toContain('班级 Top 学生')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('薄弱维度分布')
    expect(wrapper.text()).toContain('crypto')
    expect(wrapper.text()).toContain('优先介入学生')
    expect(wrapper.text()).toContain('近 7 天无训练动作')
    expect(wrapper.text()).toContain('建议训练题')
    expect(wrapper.text()).toContain('crypto-lab')
    expect(wrapper.text()).toContain('推荐训练题')
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledWith('Class A')
    expect(teacherApiMocks.getClassTrend).toHaveBeenCalledWith('Class A')
    expect(teacherApiMocks.getStudentRecommendations).toHaveBeenCalledWith('stu-1')
    expect(teacherApiMocks.getStudentProgress).not.toHaveBeenCalled()
    expect(teacherApiMocks.getStudentSkillProfile).not.toHaveBeenCalled()
  })

  it('教师概览夜间模式样式应基于主题变量而不是亮色硬编码', () => {
    expect(teacherDashboardPageSource).toMatch(teacherSurfacePattern)
    for (const [sourceName, source] of teacherSurfaceSources) {
      for (const forbiddenTeacherSurfaceLiteral of forbiddenTeacherSurfaceLiterals) {
        expect(source, `${sourceName} contains forbidden literal: ${forbiddenTeacherSurfaceLiteral}`).not.toContain(
          forbiddenTeacherSurfaceLiteral
        )
      }
    }
  })

  it('教师概览应采用 workspace tabs 结构而不是单一仪表盘堆叠', () => {
    expect(teacherDashboardPageSource).toContain('class="workspace-shell"')
    expect(teacherDashboardPageSource).toContain('role="tablist"')
    expect(teacherDashboardPageSource).toContain('top-tab-overview')
    expect(teacherDashboardPageSource).toContain('top-tab-portrait')
    expect(teacherDashboardPageSource).toContain('top-tab-trend')
    expect(teacherDashboardPageSource).toContain('top-tab-insight')
    expect(teacherDashboardPageSource).toContain('top-tab-advice')
    expect(teacherDashboardPageSource).toContain('top-tab-action')
  })

  it('教师概览的小标题应隔离 overline 样式以避免继承装饰横线', () => {
    expect(teacherDashboardPageSource).toContain('workspace-overline')
    expect(teacherDashboardPageSource).not.toContain('class="overline"')
  })

  it('教师概览趋势页不应保留冗余说明文案', async () => {
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

    expect(wrapper.text()).not.toContain(
      '把训练事件、成功解题和活跃学生放在同一条时间轴上观察，能更快判断课堂节奏是否需要调整。'
    )
    expect(wrapper.text()).not.toContain('把训练事件、成功解题和活跃学生放在同一条时间轴上观察。')
  })

  it('教师概览不应渲染设计介绍式文案', async () => {
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

    const redundantCopy = [
      '班级训练进度与能力画像工作台。',
      '当前班级的薄弱维度按学生集中度排序，优先从人数最多的方向开始补强更容易看到整体提升。',
      '聚焦风险学生、头部样本和班级主要断层，帮助教师先决定“先拉谁、补什么、怎么讲”。',
      '这里优先显示可直接转成课堂动作的建议，再补充复盘结论，避免教师在多个面板之间来回跳转。',
      '把高优先级学生和建议训练题放在同一块工作台里，便于教师直接落实到后续训练安排。',
    ]

    for (const copy of redundantCopy) {
      expect(wrapper.text()).not.toContain(copy)
    }
  })

  it('教师概览页不应再渲染快捷操作区块', async () => {
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

    expect(wrapper.find('.overview-quick-actions').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Quick Actions')
    expect(wrapper.text()).not.toContain('班级管理')
    expect(wrapper.text()).not.toContain('导出报告')
    expect(wrapper.text()).not.toContain('展开能力画像')
    expect(wrapper.text()).not.toContain('查看介入建议')
  })
})

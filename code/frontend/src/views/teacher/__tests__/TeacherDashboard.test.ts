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
    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.text()).toContain('导出报告')
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
})

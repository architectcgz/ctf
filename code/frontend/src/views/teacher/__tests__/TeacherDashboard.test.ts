import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherDashboard from '../TeacherDashboard.vue'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
  getStudentProgress: vi.fn(),
  getStudentSkillProfile: vi.fn(),
  getStudentRecommendations: vi.fn(),
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
      { id: 'stu-1', username: 'alice' },
      { id: 'stu-2', username: 'bob' },
    ])
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
      { challenge_id: '12', title: 'crypto-lab', category: 'crypto', difficulty: 'medium', reason: '针对薄弱维度：密码' },
    ])

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    }, 'token')
  })

  it('应该展示教师概览与选中学员信息', async () => {
    const wrapper = mount(TeacherDashboard, {
      global: {
        stubs: {
          SkillRadar: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('教学概览')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('crypto-lab')
  })
})

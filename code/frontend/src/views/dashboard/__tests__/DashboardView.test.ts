import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import DashboardView from '../DashboardView.vue'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()
const replaceMock = vi.fn()

const assessmentApiMocks = vi.hoisted(() => ({
  getMyProgress: vi.fn(),
  getMyTimeline: vi.fn(),
  getRecommendations: vi.fn(),
  getSkillProfile: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: pushMock,
      replace: replaceMock,
    }),
  }
})

vi.mock('@/api/assessment', () => assessmentApiMocks)

describe('DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    replaceMock.mockReset()

    assessmentApiMocks.getMyProgress.mockReset()
    assessmentApiMocks.getMyTimeline.mockReset()
    assessmentApiMocks.getRecommendations.mockReset()
    assessmentApiMocks.getSkillProfile.mockReset()

    assessmentApiMocks.getMyProgress.mockResolvedValue({
      total_score: 320,
      total_solved: 5,
      rank: 7,
      category_stats: [
        { category: 'web', solved: 3, total: 5 },
        { category: 'crypto', solved: 2, total: 4 },
      ],
      difficulty_stats: [
        { difficulty: 'easy', solved: 3, total: 4 },
        { difficulty: 'medium', solved: 2, total: 5 },
      ],
    })
    assessmentApiMocks.getMyTimeline.mockResolvedValue([
      {
        id: 'solve-1',
        type: 'solve',
        title: 'web-basic',
        created_at: '2026-03-07T10:00:00Z',
        points: 100,
        meta: { raw_type: 'flag_submit' },
      },
    ])
    assessmentApiMocks.getRecommendations.mockResolvedValue([
      { challenge_id: '12', title: 'crypto-lab', category: 'crypto', difficulty: 'medium', reason: '补强密码维度' },
    ])
    assessmentApiMocks.getSkillProfile.mockResolvedValue({
      dimensions: [
        { key: 'web', name: 'Web', value: 80 },
        { key: 'crypto', name: '密码', value: 45 },
      ],
    })
  })

  it('应该展示学生仪表盘内容', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
    }, 'token')

    const wrapper = mount(DashboardView)

    await flushPromises()

    expect(wrapper.text()).toContain('alice 的训练仪表盘')
    expect(wrapper.text()).toContain('320')
    expect(wrapper.text()).toContain('#7')
    expect(wrapper.text()).toContain('crypto-lab')
    expect(wrapper.text()).toContain('待加强：密码')
  })

  it('应该把教师用户重定向到教师首页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    }, 'token')

    mount(DashboardView)
    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({ name: 'TeacherDashboard' })
  })
})

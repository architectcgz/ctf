import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import SkillProfile from '../SkillProfile.vue'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()

const assessmentApiMocks = vi.hoisted(() => ({
  getSkillProfile: vi.fn(),
  getRecommendations: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  getClassStudents: vi.fn(),
  getStudentRecommendations: vi.fn(),
  getStudentSkillProfile: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: pushMock,
    }),
  }
})

vi.mock('@/api/assessment', () => assessmentApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)

describe('SkillProfile', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()

    assessmentApiMocks.getSkillProfile.mockReset()
    assessmentApiMocks.getRecommendations.mockReset()
    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getStudentRecommendations.mockReset()
    teacherApiMocks.getStudentSkillProfile.mockReset()

    assessmentApiMocks.getSkillProfile.mockResolvedValue({
      updated_at: '2026-03-14T10:00:00Z',
      dimensions: [
        { key: 'web', name: 'Web', value: 72 },
        { key: 'crypto', name: '密码', value: 45 },
      ],
    })
    assessmentApiMocks.getRecommendations.mockResolvedValue([
      {
        challenge_id: 'chal-1',
        title: '密码学入门',
        category: 'crypto',
        difficulty: 'easy',
        reason: '优先补强密码分析能力',
      },
    ])
    teacherApiMocks.getClassStudents.mockResolvedValue([])
    teacherApiMocks.getStudentRecommendations.mockResolvedValue([])
    teacherApiMocks.getStudentSkillProfile.mockResolvedValue(null)
  })

  it('应该渲染能力画像与推荐靶场', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'student-1',
        username: 'alice',
        role: 'student',
        class_name: 'Class A',
      },
      'token'
    )

    const wrapper = mount(SkillProfile, {
      global: {
        stubs: {
          RadarChart: {
            template: '<div data-test="radar-chart">Radar</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('能力维度分析')
    expect(wrapper.text()).toContain('薄弱项提示')
    expect(wrapper.text()).toContain('密码学入门')
    expect(wrapper.find('[data-test="radar-chart"]').exists()).toBe(true)
  })
})

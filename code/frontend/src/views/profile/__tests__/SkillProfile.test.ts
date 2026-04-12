import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import SkillProfile from '../SkillProfile.vue'
import skillProfileSource from '../SkillProfile.vue?raw'
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
    expect(wrapper.find('[role="tablist"]').exists()).toBe(true)
    expect(wrapper.find('#skill-profile-tab-analysis').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-analysis').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#skill-profile-panel-weakness').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-recommendations').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('.skill-overview-head').exists()).toBe(true)
    expect(wrapper.find('.skill-overview-head').text()).toContain('能力画像')
    expect(wrapper.find('.skill-overview-head').text()).toContain('查看当前能力维度表现，并根据薄弱项获取推荐靶场。')
    expect(wrapper.find('.skill-overview-actions').exists()).toBe(true)
    expect(wrapper.text()).toContain('能力维度分析')
    expect(wrapper.find('[data-test="radar-chart"]').exists()).toBe(true)

    await wrapper.get('#skill-profile-tab-weakness').trigger('click')
    await flushPromises()

    expect(wrapper.find('#skill-profile-tab-weakness').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-analysis').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-weakness').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#skill-profile-panel-weakness .skill-overview-head').exists()).toBe(false)
    expect(wrapper.find('#skill-profile-panel-weakness .skill-overview-actions').exists()).toBe(false)
    expect(wrapper.text()).toContain('薄弱项提示')

    await wrapper.get('#skill-profile-tab-recommendations').trigger('click')
    await flushPromises()

    expect(wrapper.find('#skill-profile-tab-recommendations').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-recommendations').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#skill-profile-panel-recommendations .skill-overview-head').exists()).toBe(false)
    expect(wrapper.find('#skill-profile-panel-recommendations .skill-overview-actions').exists()).toBe(false)
    expect(wrapper.text()).toContain('密码学入门')
  })

  it('应该将标签栏放在内容区前部，保持与学生仪表盘一致的层级位置', () => {
    expect(skillProfileSource).toContain('class="top-tabs"')
    expect(skillProfileSource.indexOf('class="top-tabs"')).toBeGreaterThan(
      skillProfileSource.indexOf('<div class="journal-eyebrow">Skill Profile</div>')
    )
    expect(skillProfileSource.indexOf('class="top-tabs"')).toBeLessThan(
      skillProfileSource.indexOf('<h1 class="journal-page-title workspace-page-title')
    )
    expect(skillProfileSource.indexOf('class="top-tabs"')).toBeLessThan(
      skillProfileSource.indexOf('class="skill-teacher-panel')
    )
    expect(skillProfileSource.indexOf('class="top-tabs"')).toBeLessThan(
      skillProfileSource.indexOf('class="skill-board')
    )
    expect(skillProfileSource).toMatch(
      /id="skill-profile-panel-analysis"[\s\S]*class="skill-overview-head"[\s\S]*<h1 class="journal-page-title workspace-page-title[\s\S]*<p class="skill-overview-copy workspace-page-copy[\s\S]*class="skill-overview-actions"/s
    )
    expect(skillProfileSource).toContain('class="skill-board px-1 md:px-2"')
    expect(skillProfileSource).not.toContain('class="skill-board mt-6')
    expect(skillProfileSource).not.toMatch(/\.skill-board\s*\{[\s\S]*border-top:\s*1px solid var\(--journal-divider\);/s)
    expect(skillProfileSource).not.toMatch(
      /\.skill-section \+ \.skill-section\s*\{[\s\S]*border-top:\s*1px solid var\(--journal-divider\);/s
    )
  })
})

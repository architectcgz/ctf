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
    assessmentApiMocks.getRecommendations.mockResolvedValue({
      weak_dimensions: [
        {
          dimension: 'crypto',
          label: '密码',
          severity: 'warning',
          confidence: 0.88,
          evidence: '当前密码维度已经形成低分与足量训练样本的组合信号。',
        },
      ],
      challenges: [
        {
          challenge_id: 'chal-1',
          title: '密码学入门',
          category: 'crypto',
          difficulty: 'easy',
          summary: '优先补强密码分析能力',
          evidence: '当前密码维度已经形成低分与足量训练样本的组合信号。',
        },
      ],
    })
    teacherApiMocks.getClassStudents.mockResolvedValue([])
    teacherApiMocks.getStudentRecommendations.mockResolvedValue({
      weak_dimensions: [],
      challenges: [],
    })
    teacherApiMocks.getStudentSkillProfile.mockResolvedValue(null)
  })

  it('应该渲染能力画像与推荐靶场', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
    })

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
    expect(wrapper.find('#skill-profile-panel-analysis').classes()).toContain('active')
    expect(wrapper.find('#skill-profile-panel-weakness').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-weakness').classes()).not.toContain('active')
    expect(wrapper.find('#skill-profile-panel-recommendations').attributes('aria-hidden')).toBe(
      'true'
    )
    expect(wrapper.find('#skill-profile-panel-recommendations').classes()).not.toContain('active')
    expect(wrapper.find('.skill-overview-head').exists()).toBe(true)
    expect(wrapper.find('.skill-overview-head').text()).toContain('能力画像')
    expect(wrapper.find('.skill-overview-head').text()).toContain(
      '查看当前能力维度表现，并根据薄弱项获取推荐靶场。'
    )
    expect(wrapper.find('.skill-overview-actions').exists()).toBe(true)
    expect(wrapper.text()).toContain('能力维度分析')
    expect(wrapper.find('[data-test="radar-chart"]').exists()).toBe(true)

    await wrapper.get('#skill-profile-tab-weakness').trigger('click')
    await flushPromises()

    expect(wrapper.find('#skill-profile-tab-weakness').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-analysis').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#skill-profile-panel-analysis').classes()).not.toContain('active')
    expect(wrapper.find('#skill-profile-panel-weakness').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#skill-profile-panel-weakness').classes()).toContain('active')
    expect(wrapper.find('#skill-profile-panel-weakness .skill-overview-head').exists()).toBe(false)
    expect(wrapper.find('#skill-profile-panel-weakness .skill-overview-actions').exists()).toBe(
      false
    )
    expect(wrapper.text()).toContain('薄弱项提示')

    await wrapper.get('#skill-profile-tab-recommendations').trigger('click')
    await flushPromises()

    expect(wrapper.find('#skill-profile-tab-recommendations').attributes('aria-selected')).toBe(
      'true'
    )
    expect(wrapper.find('#skill-profile-panel-recommendations').attributes('aria-hidden')).toBe(
      'false'
    )
    expect(wrapper.find('#skill-profile-panel-recommendations').classes()).toContain('active')
    expect(wrapper.find('#skill-profile-panel-analysis').classes()).not.toContain('active')
    expect(wrapper.find('#skill-profile-panel-recommendations .skill-overview-head').exists()).toBe(
      false
    )
    expect(
      wrapper.find('#skill-profile-panel-recommendations .skill-overview-actions').exists()
    ).toBe(false)
    expect(wrapper.text()).toContain('密码学入门')
  })

  it('当 advice 未给出薄弱维度时，不应再按画像低分自行判定弱项', async () => {
    assessmentApiMocks.getRecommendations.mockResolvedValue({
      weak_dimensions: [],
      challenges: [],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
    })

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
    await wrapper.get('#skill-profile-tab-weakness').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('暂时没有明显短板')
    expect(wrapper.text()).not.toContain('建议加强密码')
  })

  it('应该将页面顶部标签栏放在内容区外，保持与学生仪表盘一致的层级位置', () => {
    expect(skillProfileSource).toContain('class="skill-profile-page"')
    expect(skillProfileSource).toContain('class="workspace-tabbar top-tabs"')
    expect(skillProfileSource).toContain('class="workspace-tab top-tab"')
    expect(skillProfileSource).not.toContain('class="skill-profile-tabs-head"')
    expect(skillProfileSource).not.toContain('--page-top-tabs-gap: var(--space-7);')
    expect(skillProfileSource).not.toContain('--page-top-tabs-padding: 0 var(--space-7);')
    expect(skillProfileSource).not.toContain('--page-top-tab-min-height: 52px;')
    expect(skillProfileSource).toMatch(
      /class="workspace-tabbar top-tabs"[\s\S]*<\/nav>\s*<main class="content-pane">\s*<div class="skill-profile-content">/s
    )
    expect(skillProfileSource.indexOf('class="workspace-tabbar top-tabs"')).toBeLessThan(
      skillProfileSource.indexOf('<h1 class="journal-page-title workspace-page-title')
    )
    expect(skillProfileSource.indexOf('class="workspace-tabbar top-tabs"')).toBeLessThan(
      skillProfileSource.indexOf('class="skill-teacher-panel')
    )
    expect(skillProfileSource.indexOf('class="workspace-tabbar top-tabs"')).toBeLessThan(
      skillProfileSource.indexOf('class="skill-board')
    )
    expect(skillProfileSource).toMatch(
      /id="skill-profile-panel-analysis"[\s\S]*class="skill-overview-head"[\s\S]*<h1 class="journal-page-title workspace-page-title[\s\S]*<p class="skill-overview-copy workspace-page-copy[\s\S]*class="skill-overview-actions"/s
    )
    expect(skillProfileSource).toContain('class="skill-board px-1 md:px-2"')
    expect(skillProfileSource).not.toContain('class="skill-board mt-6')
    expect(skillProfileSource).not.toMatch(
      /\.skill-board\s*\{[^}]*border-top:\s*1px solid var\(--journal-divider\);/s
    )
    expect(skillProfileSource).not.toMatch(
      /\.skill-section \+ \.skill-section\s*\{[\s\S]*border-top:\s*1px solid var\(--journal-divider\);/s
    )
  })

  it('能力画像页面顶部标签栏应复用共享顶部 tab 边距，不应在 content pane 内局部抵消 padding', () => {
    expect(skillProfileSource).toContain('class="skill-profile-content"')
    expect(skillProfileSource).toContain('gap: var(--workspace-tabs-panel-gap);')
    expect(skillProfileSource).not.toContain('--page-top-tabs-margin: 0;')
    expect(skillProfileSource).not.toContain('--page-top-tabs-padding: 0;')
    expect(skillProfileSource).not.toContain('margin-top: var(--workspace-tabs-panel-gap);')
    expect(skillProfileSource).not.toContain('--page-top-tabs-padding: 0 var(--space-7);')
  })

  it('不应渲染能力画像页级眉标', () => {
    expect(skillProfileSource).not.toContain('<div class="workspace-overline">Skill Profile</div>')
    expect(skillProfileSource).not.toContain('<div class="journal-eyebrow">Skill Profile</div>')
    expect(skillProfileSource).not.toContain('journal-eyebrow-text')
  })

  it('应该把能力画像内容区的 soft eyebrow 收敛为局部 section kicker', () => {
    expect(skillProfileSource).toMatch(
      /<div class="skill-section-kicker">\s*Teacher View\s*<\/div>/s
    )
    expect(skillProfileSource).toMatch(
      /<div class="skill-section-kicker">\s*Radar Analysis\s*<\/div>/s
    )
    expect(skillProfileSource).toMatch(
      /<div class="skill-section-kicker">\s*Weak Points\s*<\/div>/s
    )
    expect(skillProfileSource).toMatch(
      /<div class="skill-section-kicker">\s*Recommendations\s*<\/div>/s
    )
    expect(skillProfileSource).not.toContain(
      '<div class="journal-eyebrow journal-eyebrow-soft">Teacher View</div>'
    )
    expect(skillProfileSource).not.toContain(
      '<div class="journal-eyebrow journal-eyebrow-soft">Radar Analysis</div>'
    )
    expect(skillProfileSource).not.toContain(
      '<div class="journal-eyebrow journal-eyebrow-soft">Weak Points</div>'
    )
    expect(skillProfileSource).not.toContain(
      '<div class="journal-eyebrow journal-eyebrow-soft">Recommendations</div>'
    )
  })

  it('教师视角学员选择框应接入共享 ui-control 原语', () => {
    expect(skillProfileSource).toMatch(/class="ui-control-wrap(?:\s+[^\"]+)?"/)
    expect(skillProfileSource).toContain('class="ui-control"')
    expect(skillProfileSource).not.toMatch(/^\.skill-student-select\s*\{/m)
    expect(skillProfileSource).not.toMatch(/^\.skill-student-select:focus\s*\{/m)
    expect(skillProfileSource).not.toMatch(/^\.skill-student-select:focus-visible\s*\{/m)
  })

  it('应该把能力画像页残留的图表高度、骨架圆角和小字号收敛为语义类', () => {
    expect(skillProfileSource).not.toContain('rounded-[24px]')
    expect(skillProfileSource).not.toContain('h-[30rem]')
    expect(skillProfileSource).not.toContain('md:h-[34rem]')
    expect(skillProfileSource).not.toContain('xl:h-[38rem]')
    expect(skillProfileSource).not.toContain('text-[1.05rem]')
    expect(skillProfileSource).not.toContain('text-[0.8rem]')
    expect(skillProfileSource).not.toContain('text-[1.9rem]')
    expect(skillProfileSource).not.toContain('md:text-[2.1rem]')
    expect(skillProfileSource).not.toContain('text-[11px]')
    expect(skillProfileSource).toContain('skill-loading-card')
    expect(skillProfileSource).toContain('skill-radar-height')
    expect(skillProfileSource).toContain('skill-dimension-legend__name')
    expect(skillProfileSource).toContain('skill-dimension-legend__hint')
    expect(skillProfileSource).toContain('skill-dimension-legend__score')
    expect(skillProfileSource).toContain('skill-difficulty-pill')
  })

  it('应该把能力画像页错误态、弱项提示和推荐区的文字色收敛为语义类', () => {
    expect(skillProfileSource).not.toContain('text-[var(--color-danger)]')
    expect(skillProfileSource).not.toContain('text-[var(--journal-ink)]')
    expect(skillProfileSource).not.toContain('text-[var(--journal-muted)]')
    expect(skillProfileSource).not.toContain('text-[var(--journal-accent)]')
    expect(skillProfileSource).not.toContain('text-[var(--journal-accent-strong)]')
    expect(skillProfileSource).toContain('skill-error-icon')
    expect(skillProfileSource).toContain('skill-error-copy')
    expect(skillProfileSource).toContain('skill-page-title')
    expect(skillProfileSource).toContain('skill-dimension-legend__total')
    expect(skillProfileSource).toContain('skill-weak-title')
    expect(skillProfileSource).toContain('skill-weak-title__icon')
    expect(skillProfileSource).toContain('skill-weak-dimension')
    expect(skillProfileSource).toContain('skill-section-copy')
    expect(skillProfileSource).toContain('skill-recommend-feedback')
    expect(skillProfileSource).toContain('skill-recommend-title')
    expect(skillProfileSource).toContain('skill-recommend-reason')
    expect(skillProfileSource).toContain('skill-recommend-arrow')
  })
})

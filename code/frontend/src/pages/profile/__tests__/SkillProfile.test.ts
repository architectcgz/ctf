import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia, type Pinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import SkillProfile from '@/pages/profile/SkillProfileRoutePage.vue'
import appRouteLinkSource from '@/shared/ui/navigation/AppRouteLink.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'
import skillProfileWorkspaceShellSource from '@/features/skill-profile/ui/SkillProfileWorkspaceShell.vue?raw'
import skillProfilePageModelSource from '@/features/skill-profile/model/useSkillProfilePage.ts?raw'
import skillProfilePanelRouteSource from '@/features/skill-profile/model/skillProfilePanelRoute.ts?raw'
import { useAuthStore } from '@/stores/auth'

const assessmentApiMocks = vi.hoisted(() => ({
  getSkillProfile: vi.fn(),
  getRecommendations: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  getClassStudents: vi.fn(),
  getStudentRecommendations: vi.fn(),
  getStudentSkillProfile: vi.fn(),
}))

vi.mock('@/api/assessment', () => assessmentApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)

let pinia: Pinia

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/profile', component: SkillProfile },
      {
        path: '/challenges',
        name: 'Challenges',
        component: { template: '<div>challenge list</div>' },
      },
      {
        path: '/challenges/:id',
        name: 'ChallengeDetail',
        component: { template: '<div>challenge detail</div>' },
      },
    ],
  })
}

async function mountPage(path = '/profile') {
  window.history.replaceState(window.history.state, '', path)
  const router = createTestRouter()
  await router.push(path)
  await router.isReady()

  const wrapper = mount(SkillProfile, {
    global: {
      plugins: [pinia, router],
      stubs: {
        RadarChart: {
          template: '<div data-test="radar-chart">Radar</div>',
        },
      },
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('SkillProfile', () => {
  const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    localStorage.clear()

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

  it('应该渲染六维画像与推荐靶场', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
    })

    const { wrapper } = await mountPage()

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
    expect(wrapper.find('.skill-overview-head').text()).toContain('六维学习画像')
    expect(wrapper.find('.skill-overview-head').text()).toContain(
      '查看当前六个能力维度的训练分布，并根据薄弱维度获取推荐靶场。'
    )
    expect(wrapper.find('.skill-overview-actions').exists()).toBe(true)
    expect(wrapper.text()).toContain('六维分布分析')
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
    expect(wrapper.text()).toContain('薄弱维度提示')

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

    const { wrapper } = await mountPage()
    await wrapper.get('#skill-profile-tab-weakness').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('暂时没有明显短板')
    expect(wrapper.text()).not.toContain('建议加强密码')
  })

  it('应通过 AppRouteLink 消费做题与推荐题目路由目标', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
    })

    expect(appRouteLinkSource).toContain("from 'vue-router'")
    expect(skillProfileWorkspaceShellSource).toContain("from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(skillProfileWorkspaceShellSource).toContain('<AppRouteLink')
    expect(skillProfilePageModelSource).toContain('skillProfileChallengesRoute')
    expect(skillProfilePageModelSource).toContain('buildChallengeRoute')
    expect(skillProfilePageModelSource).toContain(
      "import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(skillProfilePageModelSource).toContain('resolveSkillProfilePanel(query.value.panel)')
    expect(skillProfilePageModelSource).toContain(
      'await replaceQuery(buildSkillProfilePanelQuery(query.value, panel))'
    )
    expect(skillProfilePanelRouteSource).toContain('resolveSkillProfilePanel')
    expect(skillProfilePanelRouteSource).toContain('buildSkillProfilePanelQuery')
    expect(skillProfilePageModelSource).not.toContain("from 'vue-router'")
    expect(skillProfileSource).toContain("from '@/shared/lib/keyboard/useTabKeyboardNavigation'")
    expect(skillProfileSource).not.toContain("from '@/shared/model/navigation/useUrlSyncedTabs'")

    const { wrapper, router } = await mountPage()

    const challengesLink = wrapper
      .findAll('a')
      .find((node) => node.text().includes('去做题'))
    expect(challengesLink).toBeDefined()

    await challengesLink!.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.name).toBe('Challenges')

    await router.push('/profile?panel=recommendations')
    window.history.replaceState(window.history.state, '', '/profile?panel=recommendations')
    const { wrapper: recommendationWrapper, router: recommendationRouter } =
      await mountPage('/profile?panel=recommendations')

    expect(
      recommendationWrapper.find('#skill-profile-tab-recommendations').attributes('aria-selected')
    ).toBe('true')
    expect(
      recommendationWrapper.find('#skill-profile-panel-recommendations').attributes('aria-hidden')
    ).toBe('false')
    expect(
      recommendationWrapper.find('#skill-profile-panel-analysis').attributes('aria-hidden')
    ).toBe('true')

    const recommendationLink = recommendationWrapper
      .findAll('a')
      .find((node) => node.text().includes('密码学入门'))
    expect(recommendationLink).toBeDefined()

    await recommendationLink!.trigger('click')
    await flushPromises()
    expect(recommendationRouter.currentRoute.value.name).toBe('ChallengeDetail')
    expect(recommendationRouter.currentRoute.value.params.id).toBe('chal-1')
  })

  it('点击能力画像标签时应回写 panel 查询参数', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
    })

    const { wrapper, router } = await mountPage('/profile')

    await wrapper.get('#skill-profile-tab-weakness').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.query.panel).toBe('weakness')
    expect(wrapper.find('#skill-profile-panel-weakness').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#skill-profile-panel-analysis').attributes('aria-hidden')).toBe('true')
  })

  it('教师视角学员选择框应复用 user entity 的展示名 owner', async () => {
    teacherApiMocks.getClassStudents.mockResolvedValue([
      { id: 'stu-1', username: 'alice', name: 'Alice Zhang' },
      { id: 'stu-2', username: 'bob' },
    ])

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    })

    const { wrapper } = await mountPage()
    const options = wrapper.findAll('#skill-student-select option').map((option) => option.text())

    expect(options).toEqual(['我的六维画像', 'Alice Zhang (alice)', 'bob'])
    expect(skillProfileWorkspaceShellSource).toContain("from '@/entities/user'")
    expect(skillProfileWorkspaceShellSource).toContain('getUserDisplayLabel')
    expect(skillProfileWorkspaceShellSource).not.toContain('{{ student.name || student.username }}')
    expect(skillProfileWorkspaceShellSource).not.toContain('formatStudentOptionLabel')
  })
})

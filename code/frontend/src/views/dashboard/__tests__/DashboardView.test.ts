import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import DashboardView from '../DashboardView.vue'
import dashboardViewSource from '../DashboardView.vue?raw'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = {
  params: {} as Record<string, string>,
  query: {} as Record<string, string>,
}

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
    useRoute: () => routeState,
  }
})

vi.mock('@/api/assessment', () => assessmentApiMocks)

function mountDashboard() {
  return mount(DashboardView, {
    global: {
      stubs: {
        RouterLink: {
          template: '<a><slot /></a>',
        },
        RadarChart: {
          template: '<div data-test="radar-chart">Radar</div>',
        },
      },
    },
  })
}

describe('DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.params = {}
    routeState.query = {}

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
        id: 'read-1',
        type: 'challenge_detail_view',
        title: 'web-basic',
        detail: '查看题目详情，开始分析题面与环境线索',
        created_at: '2026-03-07T09:00:00Z',
        meta: { raw_type: 'challenge_detail_view' },
      },
      {
        id: 'hint-1',
        type: 'hint',
        title: 'web-basic',
        detail: '解锁第 1 级提示：先看回显',
        created_at: '2026-03-07T09:30:00Z',
        meta: { raw_type: 'hint_unlock' },
      },
      {
        id: 'access-1',
        type: 'instance_access',
        title: 'web-basic',
        detail: '访问攻击目标，开始与靶机进行实际交互',
        created_at: '2026-03-07T09:40:00Z',
        meta: { raw_type: 'instance_access' },
      },
      {
        id: 'extend-1',
        type: 'instance_extend',
        title: 'web-basic',
        detail: '延长实例有效期，继续当前利用过程',
        created_at: '2026-03-07T09:45:00Z',
        meta: { raw_type: 'instance_extend' },
      },
      {
        id: 'solve-1',
        type: 'solve',
        title: 'web-basic',
        detail: '第 2 次提交命中 Flag，获得 100 分',
        created_at: '2026-03-07T10:00:00Z',
        points: 100,
        meta: { raw_type: 'flag_submit' },
      },
    ])
    assessmentApiMocks.getRecommendations.mockResolvedValue([
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        reason: '补强密码维度',
      },
      {
        challenge_id: '24',
        title: 'web-xss',
        category: 'web',
        difficulty: 'easy',
        reason: '保持 Web 练习节奏',
      },
      {
        challenge_id: '36',
        title: 'pwn-intro',
        category: 'pwn',
        difficulty: 'easy',
        reason: '补齐基础利用动作',
      },
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
    authStore.setAuth(
      {
        id: 'student-1',
        username: 'alice',
        role: 'student',
        class_name: 'Class A',
      },
      'token'
    )

    const wrapper = mountDashboard()

    await flushPromises()

    expect(wrapper.text()).toContain('alice 的训练总览')
    expect(wrapper.text()).toContain('320')
    expect(wrapper.text()).toContain('#7')

    const tabTexts = wrapper.findAll('[role="tab"]').map((tab) => tab.text())
    expect(tabTexts).toEqual(['训练总览', '训练队列', '分类补强', '训练记录', '强度推进'])
  })

  it('应该把竞技表现统计区域渲染为共享摘要卡片', async () => {
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

    const wrapper = mountDashboard()

    await flushPromises()

    const summary = wrapper.get('.story-metric-grid')

    expect(summary.classes()).toContain('progress-strip')
    expect(summary.classes()).toContain('metric-panel-grid')
    expect(summary.classes()).toContain('metric-panel-default-surface')
    expect(summary.findAll('.journal-metric.progress-card.metric-panel-card')).toHaveLength(4)
    expect(summary.findAll('.progress-card-label.metric-panel-label')).toHaveLength(4)
    expect(summary.findAll('.progress-card-value.metric-panel-value')).toHaveLength(4)
    expect(summary.findAll('.progress-card-hint.metric-panel-helper')).toHaveLength(4)
    expect(summary.text()).toContain('总得分')
    expect(summary.text()).toContain('当前累计获得的训练积分')
    expect(summary.text()).toContain('完成率')
    expect(summary.text()).toContain('按当前分类题量计算的覆盖比例')
  })

  it('应该把当前排名区域渲染为独立卡片', async () => {
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

    const wrapper = mountDashboard()

    await flushPromises()

    const rankSummary = wrapper.get('.journal-rank-summary')

    expect(rankSummary.classes()).toContain('progress-card')
    expect(rankSummary.classes()).toContain('metric-panel-card')
    expect(rankSummary.classes()).toContain('metric-panel-default-surface')
    expect(rankSummary.get('.progress-card-label.metric-panel-label').text()).toContain('当前排名')
    expect(rankSummary.get('.progress-card-value.metric-panel-value').text()).toBe('#7')
    expect(rankSummary.get('.progress-card-hint.metric-panel-helper').text()).toContain('积分排名')
  })

  it('应该在 recommendation 子菜单下激活并显示训练建议面板', async () => {
    routeState.query = { panel: 'recommendation' }

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

    const wrapper = mountDashboard()

    await flushPromises()

    const recommendationPanel = wrapper.get('#dashboard-panel-recommendation')
    const overviewPanel = wrapper.get('#dashboard-panel-overview')
    const recommendationStyle = recommendationPanel.attributes('style') ?? ''
    const overviewStyle = overviewPanel.attributes('style') ?? ''

    expect(recommendationPanel.classes()).toContain('active')
    expect(recommendationPanel.attributes('aria-hidden')).toBe('false')
    expect(recommendationStyle).not.toContain('display: none;')
    expect(recommendationPanel.isVisible()).toBe(true)

    expect(overviewPanel.classes()).not.toContain('active')
    expect(overviewPanel.attributes('aria-hidden')).toBe('true')
    expect(overviewStyle).toContain('display: none;')
    expect(overviewPanel.isVisible()).toBe(false)

    expect(recommendationPanel.text()).toContain('现在先练这几道')
    expect(recommendationPanel.text()).toContain('当前目标难度')
    expect(recommendationPanel.text()).toContain('浏览全部题目')
    expect(recommendationPanel.text()).toContain('crypto-lab')
    expect(recommendationPanel.text()).toContain('web-xss')
    expect(recommendationPanel.text()).toContain('pwn-intro')
    expect(recommendationPanel.findAll('.recommend-item')).toHaveLength(3)
  })

  it('应该在 recommendation 空状态下保留唯一的浏览全部题目主 CTA', async () => {
    routeState.query = { panel: 'recommendation' }
    assessmentApiMocks.getRecommendations.mockResolvedValue([])

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

    const wrapper = mountDashboard()

    await flushPromises()

    const recommendationPanel = wrapper.get('#dashboard-panel-recommendation')
    const browseButtons = recommendationPanel
      .findAll('button')
      .filter((button) => button.text().trim() === '浏览全部题目')

    expect(recommendationPanel.attributes('aria-hidden')).toBe('false')
    expect(recommendationPanel.isVisible()).toBe(true)
    expect(recommendationPanel.text()).toContain('当前没有推荐题目，可以先去题目列表探索新的方向。')
    expect(recommendationPanel.findAll('.recommend-item')).toHaveLength(0)
    expect(browseButtons).toHaveLength(1)
    expect(browseButtons[0].classes()).toContain('journal-btn-primary')
  })

  it('应该在 category 子菜单下展示行动优先的分类列表并支持跳转到对应分类题目', async () => {
    routeState.query = { panel: 'category' }

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

    const wrapper = mountDashboard()

    await flushPromises()

    const categoryPanel = wrapper.get('#dashboard-panel-category')
    const cryptoAction = categoryPanel.get('[data-test="category-action-crypto"]')

    expect(categoryPanel.classes()).toContain('active')
    expect(categoryPanel.attributes('aria-hidden')).toBe('false')
    expect(categoryPanel.isVisible()).toBe(true)
    expect(categoryPanel.text()).toContain('优先补这个分类')
    expect(categoryPanel.text()).toContain('crypto')
    expect(categoryPanel.findAll('.category-action-item')).toHaveLength(2)

    await cryptoAction.get('button').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'Challenges',
      query: { category: 'crypto' },
    })
  })

  it('应该在 category 空状态下避免展示虚构分类名，并保留合理的训练入口', async () => {
    routeState.query = { panel: 'category' }
    assessmentApiMocks.getMyProgress.mockResolvedValue({
      total_score: 320,
      total_solved: 0,
      rank: 7,
      category_stats: [],
      difficulty_stats: [
        { difficulty: 'easy', solved: 0, total: 4 },
        { difficulty: 'medium', solved: 0, total: 5 },
      ],
    })

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

    const wrapper = mountDashboard()

    await flushPromises()

    const categoryPanel = wrapper.get('#dashboard-panel-category')
    const primaryButton = categoryPanel
      .findAll('button')
      .find((button) => button.text().trim() === '去训练')

    expect(categoryPanel.attributes('aria-hidden')).toBe('false')
    expect(categoryPanel.isVisible()).toBe(true)
    expect(categoryPanel.text()).not.toContain('新的分类')
    expect(categoryPanel.text()).toContain('先开始积累分类覆盖面')
    expect(categoryPanel.text()).toContain('当前还没有分类统计数据，先完成几道题再回来查看。')
    expect(categoryPanel.findAll('.category-action-item')).toHaveLength(0)
    expect(primaryButton).toBeTruthy()

    await primaryButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'Challenges' })
  })

  it('应该在 difficulty 子菜单下展示强度推进工作区并支持跳转到对应难度题目', async () => {
    routeState.query = { panel: 'difficulty' }

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

    const wrapper = mountDashboard()

    await flushPromises()

    const difficultyPanel = wrapper.get('#dashboard-panel-difficulty')
    const mediumAction = difficultyPanel.get('[data-test="difficulty-action-medium"]')

    expect(difficultyPanel.classes()).toContain('active')
    expect(difficultyPanel.attributes('aria-hidden')).toBe('false')
    expect(difficultyPanel.isVisible()).toBe(true)
    expect(difficultyPanel.text()).toContain('先推这一档强度')
    expect(difficultyPanel.text()).toContain('中等')
    expect(difficultyPanel.findAll('.difficulty-action-item')).toHaveLength(2)
    expect(mediumAction.classes()).toContain('difficulty-action-item--primary')

    await mediumAction.get('button').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'Challenges',
      query: { difficulty: 'medium' },
    })
  })

  it('应该在 difficulty 空状态下避免展示虚构难度名，并保留合理的训练入口', async () => {
    routeState.query = { panel: 'difficulty' }
    assessmentApiMocks.getMyProgress.mockResolvedValue({
      total_score: 320,
      total_solved: 0,
      rank: 7,
      category_stats: [],
      difficulty_stats: [],
    })

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

    const wrapper = mountDashboard()

    await flushPromises()

    const difficultyPanel = wrapper.get('#dashboard-panel-difficulty')
    const primaryButton = difficultyPanel
      .findAll('button')
      .find((button) => button.text().trim() === '去训练')

    expect(difficultyPanel.attributes('aria-hidden')).toBe('false')
    expect(difficultyPanel.isVisible()).toBe(true)
    expect(difficultyPanel.text()).toContain('先开始建立强度节奏')
    expect(difficultyPanel.text()).toContain('当前还没有难度统计数据，先完成几道题再回来查看。')
    expect(difficultyPanel.text()).not.toContain('待选择')
    expect(difficultyPanel.findAll('.difficulty-action-item')).toHaveLength(0)
    expect(primaryButton).toBeTruthy()

    await primaryButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'Challenges' })
  })

  it('应该在 difficulty 同完成率时优先更低难度，并让标题主 CTA 与主推行保持一致', async () => {
    routeState.query = { panel: 'difficulty' }
    assessmentApiMocks.getMyProgress.mockResolvedValue({
      total_score: 320,
      total_solved: 0,
      rank: 7,
      category_stats: [{ category: 'web', solved: 0, total: 2 }],
      difficulty_stats: [
        { difficulty: 'beginner', solved: 0, total: 1 },
        { difficulty: 'medium', solved: 0, total: 5 },
      ],
    })

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

    const wrapper = mountDashboard()

    await flushPromises()

    const difficultyPanel = wrapper.get('#dashboard-panel-difficulty')
    const primaryButton = difficultyPanel
      .findAll('button')
      .find((button) => button.text().trim() === '先做入门')
    const beginnerAction = difficultyPanel.get('[data-test="difficulty-action-beginner"]')
    const mediumAction = difficultyPanel.get('[data-test="difficulty-action-medium"]')

    expect(difficultyPanel.text()).toContain('先推这一档强度：入门')
    expect(primaryButton).toBeTruthy()
    expect(beginnerAction.classes()).toContain('difficulty-action-item--primary')
    expect(mediumAction.classes()).not.toContain('difficulty-action-item--primary')

    await primaryButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'Challenges',
      query: { difficulty: 'beginner' },
    })
  })

  it('应该在带 variant 参数时继续展示当前首页风格', async () => {
    routeState.params = { variant: '2' }

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

    const wrapper = mountDashboard()

    await flushPromises()

    expect(wrapper.text()).toContain('alice 的训练总览')
  })

  it('应该在 timeline 子菜单下展示训练记录', async () => {
    routeState.query = { panel: 'timeline' }

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

    const wrapper = mountDashboard()

    await flushPromises()

    const summary = wrapper.get('.timeline-metric-grid')

    expect(summary.classes()).toContain('progress-strip')
    expect(summary.classes()).toContain('metric-panel-grid')
    expect(summary.classes()).toContain('metric-panel-default-surface')
    expect(summary.findAll('.timeline-metric-card.progress-card.metric-panel-card')).toHaveLength(4)
    expect(summary.findAll('.progress-card-label.metric-panel-label')).toHaveLength(4)
    expect(summary.findAll('.progress-card-value.metric-panel-value')).toHaveLength(4)
    expect(summary.findAll('.progress-card-hint.metric-panel-helper')).toHaveLength(4)
    expect(wrapper.text()).toContain('训练记录')
    expect(wrapper.text()).toContain('成功解题')
    expect(wrapper.text()).toContain('累计命中 Flag 的训练次数')
    expect(wrapper.text()).toContain('提交次数')
    expect(wrapper.text()).toContain('最近训练周期内的提交总量')
    expect(wrapper.text()).toContain('实例操作')
    expect(wrapper.text()).toContain('启动、访问和续期等实例相关动作')
    expect(wrapper.text()).toContain('总记录')
    expect(wrapper.text()).toContain('当前时间线中收录的训练事件数量')
    expect(wrapper.text()).toContain('查看题目详情')
    expect(wrapper.text()).toContain('访问攻击目标')
    expect(wrapper.text()).toContain('解锁第 1 级提示')
  })

  it('应该把教师用户重定向到教师首页', async () => {
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

    mountDashboard()
    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({ name: 'TeacherDashboard' })
  })

  it('应该移除仪表盘页级 shell 上遗留的 journal-eyebrow-text 修饰类', () => {
    expect(dashboardViewSource).toContain(
      'class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"'
    )
    expect(dashboardViewSource).not.toContain('journal-eyebrow-text')
  })
})

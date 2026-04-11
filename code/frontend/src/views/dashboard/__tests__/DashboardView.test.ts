import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import DashboardView from '../DashboardView.vue'
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

    expect(wrapper.text()).toContain('Training Journal')
    expect(wrapper.text()).toContain('alice 的训练总览')
    expect(wrapper.text()).toContain('320')
    expect(wrapper.text()).toContain('#7')
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

  it('应该在 recommendation 子菜单下展示训练建议', async () => {
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

    expect(wrapper.text()).toContain('Priority Focus')
    expect(wrapper.text()).toContain('补短板计划')
    expect(wrapper.text()).toContain('crypto-lab')
    expect(wrapper.text()).toContain('推荐摘要')
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

    expect(wrapper.text()).toContain('Training Journal')
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
})

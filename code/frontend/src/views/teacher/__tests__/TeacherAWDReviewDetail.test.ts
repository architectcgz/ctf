import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherAWDReviewDetail from '../TeacherAWDReviewDetail.vue'
import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import { useAuthStore } from '@/stores/auth'

const routeMock = {
  params: {
    contestId: 'contest-1',
  },
  query: {} as Record<string, string>,
}

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherAWDReview: vi.fn(),
  exportTeacherAWDReviewArchive: vi.fn(),
  exportTeacherAWDReviewReport: vi.fn(),
}))

const pushMock = vi.fn()
const replaceMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('TeacherAWDReviewDetail', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    routeMock.params.contestId = 'contest-1'
    routeMock.query = {}
    pushMock.mockReset()
    replaceMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getTeacherAWDReview.mockResolvedValue({
      generated_at: '2026-04-12T10:00:00Z',
      scope: {
        snapshot_type: 'live',
        requested_by: 5,
        requested_id: 'contest-1',
      },
      contest: {
        id: 'contest-1',
        title: '春季 AWD 联训',
        mode: 'awd',
        status: 'running',
        current_round: 2,
        round_count: 4,
        team_count: 6,
        export_ready: false,
      },
      overview: {
        round_count: 4,
        team_count: 6,
        service_count: 12,
        attack_count: 8,
        traffic_count: 20,
      },
      rounds: [
        {
          id: 'round-1',
          contest_id: 'contest-1',
          round_number: 1,
          status: 'finished',
          service_count: 6,
          attack_count: 3,
          traffic_count: 8,
        },
        {
          id: 'round-2',
          contest_id: 'contest-1',
          round_number: 2,
          status: 'running',
          service_count: 6,
          attack_count: 5,
          traffic_count: 12,
        },
      ],
      selected_round: {
        round: {
          id: 'round-2',
          contest_id: 'contest-1',
          round_number: 2,
          status: 'running',
          service_count: 6,
          attack_count: 5,
          traffic_count: 12,
        },
        teams: [
          {
            team_id: 'team-1',
            team_name: 'Blue Team',
            captain_id: 'stu-1',
            total_score: 320,
            member_count: 4,
          },
        ],
        services: [],
        attacks: [],
        traffic: [],
      },
    })
    teacherApiMocks.exportTeacherAWDReviewArchive.mockResolvedValue({
      report_id: '31',
      status: 'processing',
    })
    teacherApiMocks.exportTeacherAWDReviewReport.mockResolvedValue({
      report_id: '32',
      status: 'processing',
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
      },
      'token'
    )
  })

  it('默认显示整场总览，并在进行中赛事上禁用教师报告导出', async () => {
    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(teacherApiMocks.getTeacherAWDReview).toHaveBeenCalledWith('contest-1', {
      round: undefined,
      team_id: undefined,
    })
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(
      wrapper.find('[data-testid="awd-review-export-report"]').attributes('disabled')
    ).toBeDefined()
  })

  it('query 带 round 时应切到单轮摘要，并在赛后开放两个导出按钮', async () => {
    routeMock.query = { round: '2' }
    teacherApiMocks.getTeacherAWDReview.mockResolvedValueOnce({
      generated_at: '2026-04-12T10:00:00Z',
      scope: {
        snapshot_type: 'final',
        requested_by: 5,
        requested_id: 'contest-1',
      },
      contest: {
        id: 'contest-1',
        title: '期末 AWD 复盘',
        mode: 'awd',
        status: 'ended',
        current_round: 4,
        round_count: 4,
        team_count: 6,
        export_ready: true,
      },
      overview: {
        round_count: 4,
        team_count: 6,
        service_count: 12,
        attack_count: 8,
        traffic_count: 20,
      },
      rounds: [],
      selected_round: {
        round: {
          id: 'round-2',
          contest_id: 'contest-1',
          round_number: 2,
          status: 'finished',
          service_count: 6,
          attack_count: 5,
          traffic_count: 12,
        },
        teams: [],
        services: [],
        attacks: [],
        traffic: [],
      },
    })

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(teacherApiMocks.getTeacherAWDReview).toHaveBeenCalledWith('contest-1', {
      round: 2,
      team_id: undefined,
    })
    expect(wrapper.text()).toContain('第 2 轮')
    expect(
      wrapper.find('[data-testid="awd-review-export-archive"]').attributes('disabled')
    ).toBeUndefined()
    expect(
      wrapper.find('[data-testid="awd-review-export-report"]').attributes('disabled')
    ).toBeUndefined()
  })

  it('轮次切换区应并入目录段结构，不再保留独立筛选卡片壳', () => {
    expect(awdReviewDetailSource).toContain(
      'class="workspace-directory-section teacher-directory-section awd-review-round-section"'
    )
    expect(awdReviewDetailSource).toContain('class="list-heading"')
    expect(awdReviewDetailSource).not.toContain('teacher-controls-title')
    expect(awdReviewDetailSource).not.toContain('teacher-controls-copy')
    expect(awdReviewDetailSource).not.toContain(
      '默认展示整场总览；切换到单轮后，可继续按队伍查看该轮服务、攻击和流量证据。'
    )
  })

  it('详情页加载骨架应通过语义类承接，不再直接写圆角和背景混色', () => {
    expect(awdReviewDetailSource).toContain('awd-review-loading-card')
    expect(awdReviewDetailSource).not.toContain('rounded-[22px]')
    expect(awdReviewDetailSource).not.toContain(
      'bg-[color-mix(in_srgb,var(--journal-surface-subtle)_92%,transparent)]'
    )
  })

  it('管理员在 AWD 复盘详情里返回目录和切换轮次时应使用后台路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
      },
      'token'
    )

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('返回目录'))?.trigger('click')
    await wrapper.findAll('.awd-review-round-pill')[1]?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'AdminAWDReviewIndex' })
    expect(replaceMock).toHaveBeenCalledWith({
      name: 'AdminAWDReviewDetail',
      params: {
        contestId: 'contest-1',
      },
      query: {
        round: '1',
      },
    })
  })
})

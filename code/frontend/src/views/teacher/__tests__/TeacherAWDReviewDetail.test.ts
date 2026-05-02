import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import TeacherAWDReviewDetail from '../TeacherAWDReviewDetail.vue'
import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'

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

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

function collectTestIdTexts(testId: string): string[] {
  return [...document.body.querySelectorAll(`[data-testid="${testId}"]`)]
    .map((node) => node.textContent?.trim() || '')
    .filter(Boolean)
}

describe('TeacherAWDReviewDetail', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    routeMock.params.contestId = 'contest-1'
    routeMock.query = {}
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
  })

  it('不应继续保留亮色硬编码 surface、状态色或 text-slate 工具类', () => {
    expect(awdReviewDetailSource).not.toContain('#fdfdfd')
    expect(awdReviewDetailSource).not.toContain('#3182ce')
    expect(awdReviewDetailSource).not.toContain('rgba(0,0,0,0.02)')
    expect(awdReviewDetailSource).not.toContain('text-slate-500')
    expect(awdReviewDetailSource).not.toContain('text-slate-900')
    expect(awdReviewDetailSource).not.toContain('text-emerald-600')
    expect(awdReviewDetailSource).not.toContain('text-red-600')
    expect(awdReviewDetailSource).not.toContain('text-blue-600')
    expect(awdReviewDetailSource).not.toContain('bg-emerald-400')
  })

  it('页面应通过 feature model 获取详情状态，不再直接耦合 teacher api', () => {
    expect(awdReviewDetailSource).toContain("useTeacherAwdReviewDetail } from '@/features/teacher-awd-review'")
    expect(awdReviewDetailSource).not.toContain("from '@/api/teacher'")
    expect(awdReviewDetailSource).not.toContain('const activeContestTitle = computed')
    expect(awdReviewDetailSource).not.toContain('const summaryStats = computed')
    expect(awdReviewDetailSource).not.toContain('function contestStatusLabel')
    expect(awdReviewDetailSource).not.toContain('function formatServiceRef')
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

  it('单轮复盘应在主视图和队伍抽屉中保留运行态 service 标识', async () => {
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
          service_count: 1,
          attack_count: 1,
          traffic_count: 1,
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
        services: [
          {
            id: 'service-row-1',
            round_id: 'round-2',
            team_id: 'team-1',
            team_name: 'Blue Team',
            service_id: '7009',
            challenge_id: 'challenge-1',
            challenge_title: 'Bank Portal',
            service_status: 'up',
            attack_received: 0,
            sla_score: 18,
            defense_score: 40,
            attack_score: 0,
            updated_at: '2026-04-12T10:02:00Z',
          },
        ],
        attacks: [
          {
            id: 'attack-row-1',
            round_id: 'round-2',
            attacker_team_id: 'team-2',
            attacker_team_name: 'Red Team',
            victim_team_id: 'team-1',
            victim_team_name: 'Blue Team',
            service_id: '7009',
            challenge_id: 'challenge-1',
            challenge_title: 'Bank Portal',
            attack_type: 'flag_capture',
            source: 'submission',
            is_success: true,
            score_gained: 60,
            created_at: '2026-04-12T10:03:00Z',
          },
        ],
        traffic: [
          {
            id: 'traffic-row-1',
            contest_id: 'contest-1',
            round_id: 'round-2',
            attacker_team_id: 'team-2',
            attacker_team_name: 'Red Team',
            victim_team_id: 'team-1',
            victim_team_name: 'Blue Team',
            service_id: '7009',
            challenge_id: 'challenge-1',
            challenge_title: 'Bank Portal',
            method: 'POST',
            path: '/flag',
            status_code: 200,
            source: 'runtime_proxy',
            created_at: '2026-04-12T10:04:00Z',
          },
        ],
      },
    })

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(
      wrapper
        .findAll('[data-testid="awd-review-service-id"]')
        .map((node) => node.text())
    ).toContain('Service #7009')
    expect(
      wrapper
        .findAll('[data-testid="awd-review-attack-service-id"]')
        .map((node) => node.text())
    ).toContain('Service #7009')
    expect(
      wrapper
        .findAll('[data-testid="awd-review-traffic-service-id"]')
        .map((node) => node.text())
    ).toContain('Service #7009')

    await wrapper.find('.teacher-directory-row').trigger('click')
    await flushPromises()

    expect(collectTestIdTexts('awd-review-drawer-service-id')).toContain('Service #7009')
    expect(collectTestIdTexts('awd-review-drawer-attack-service-id')).toContain('Service #7009')
    expect(collectTestIdTexts('awd-review-drawer-traffic-service-id')).toContain('Service #7009')
  })

  it('页面应通过 widget 组合复盘工作区，不直接承载复盘区块模板', () => {
    expect(awdReviewDetailSource).toContain(
      "import { TeacherAWDReviewWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('teacher-controls-title')
    expect(awdReviewDetailSource).not.toContain('teacher-controls-copy')
    expect(awdReviewDetailSource).not.toContain(
      '默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。'
    )
  })

  it('导出归档失败时不应抛到全局错误页', async () => {
    teacherApiMocks.exportTeacherAWDReviewArchive.mockRejectedValue(new Error('导出失败'))

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    const exportButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('归档导出'))

    expect(exportButton).toBeTruthy()

    await expect(exportButton!.trigger('click')).resolves.toBeUndefined()
    await flushPromises()

    expect(teacherApiMocks.exportTeacherAWDReviewArchive).toHaveBeenCalledWith(
      'contest-1',
      undefined
    )
  })
})

import { computed, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'

import { useAwdReviewDetailData } from './useAwdReviewDetailData'

const awdReviewApiMocks = vi.hoisted(() => ({
  getAwdReviewByRole: vi.fn(),
}))

vi.mock('@/api/awd-reviews', () => ({
  getAwdReviewByRole: awdReviewApiMocks.getAwdReviewByRole,
}))

describe('useAwdReviewDetailData', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    useAuthStore().user = { id: 'teacher-1', role: 'teacher' } as never
    awdReviewApiMocks.getAwdReviewByRole.mockReset()
  })

  it('应按当前 role、contestId 和 round 加载 AWD 复盘详情', async () => {
    const contestId = ref('contest-1')
    const selectedRoundNumber = ref<number | undefined>(2)
    awdReviewApiMocks.getAwdReviewByRole.mockResolvedValue({
      contest: {
        id: 'contest-1',
        title: '期末 AWD 复盘',
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
          round_number: 2,
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

    const composable = useAwdReviewDetailData({
      contestId: computed(() => contestId.value),
      selectedRoundNumber: computed(() => selectedRoundNumber.value),
    })

    await composable.loadReview()

    expect(awdReviewApiMocks.getAwdReviewByRole).toHaveBeenCalledWith('teacher', 'contest-1', {
      round: 2,
      team_id: undefined,
    })
    expect(composable.activeContestTitle.value).toBe('期末 AWD 复盘')
    expect(composable.activeSummaryTitle.value).toBe('第 2 轮')
    expect(composable.summaryStats.value.teamCount).toBe(1)
    expect(composable.canExportReport.value).toBe(true)
  })

  it('队伍切换后若下一次详情里已不存在该 team，应回收抽屉状态', async () => {
    const contestId = ref('contest-1')
    const selectedRoundNumber = ref<number | undefined>(1)
    awdReviewApiMocks.getAwdReviewByRole
      .mockResolvedValueOnce({
        contest: {
          id: 'contest-1',
          title: '春季 AWD 联训',
          export_ready: false,
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
            round_number: 1,
            service_count: 6,
            attack_count: 3,
            traffic_count: 8,
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
      .mockResolvedValueOnce({
        contest: {
          id: 'contest-1',
          title: '春季 AWD 联训',
          export_ready: false,
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
            round_number: 1,
            service_count: 6,
            attack_count: 3,
            traffic_count: 8,
          },
          teams: [],
          services: [],
          attacks: [],
          traffic: [],
        },
      })

    const composable = useAwdReviewDetailData({
      contestId: computed(() => contestId.value),
      selectedRoundNumber: computed(() => selectedRoundNumber.value),
    })

    await composable.loadReview()
    composable.openTeam({
      team_id: 'team-1',
      team_name: 'Blue Team',
      captain_id: 'stu-1',
      total_score: 320,
      member_count: 4,
    })

    await composable.loadReview()

    expect(composable.selectedTeamId.value).toBeNull()
    expect(composable.selectedTeam.value).toBeNull()
  })

  it('加载失败时应回填统一错误状态并清空详情', async () => {
    const contestId = ref('contest-1')
    const selectedRoundNumber = ref<number | undefined>(undefined)
    awdReviewApiMocks.getAwdReviewByRole.mockRejectedValue(new Error('boom'))

    const composable = useAwdReviewDetailData({
      contestId: computed(() => contestId.value),
      selectedRoundNumber: computed(() => selectedRoundNumber.value),
    })

    await composable.loadReview()

    expect(composable.error.value).toBe('加载 AWD 复盘详情失败，请稍后重试')
    expect(composable.review.value).toBeNull()
  })
})

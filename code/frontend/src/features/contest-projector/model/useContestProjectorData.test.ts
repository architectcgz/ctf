import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'

import { useContestProjectorData } from '@/features/contest-projector'
import type { AWDRoundData, ContestDetailData } from '@/api/contracts'

const adminApiMocks = vi.hoisted(() => ({
  getAdminContestLiveScoreboard: vi.fn(),
  getContestAWDRoundSummary: vi.fn(),
  getContestAWDRoundTrafficSummary: vi.fn(),
  getContests: vi.fn(),
  listContestAWDRoundAttacks: vi.fn(),
  listContestAWDRoundServices: vi.fn(),
  listContestAWDRounds: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
}))

vi.mock('@/api/admin/contests', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

function buildContest(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'contest-1',
    title: 'AWD Drill',
    description: 'AWD',
    mode: 'awd',
    status: 'running',
    starts_at: '2026-04-12T09:00:00.000Z',
    ends_at: '2026-04-12T12:00:00.000Z',
    scoreboard_frozen: false,
    ...overrides,
  }
}

function buildRound(overrides: Partial<AWDRoundData> = {}): AWDRoundData {
  return {
    id: 'round-1',
    contest_id: 'contest-1',
    round_number: 1,
    status: 'finished',
    attack_score: 50,
    defense_score: 50,
    created_at: '2026-04-12T09:00:00.000Z',
    updated_at: '2026-04-12T09:00:00.000Z',
    ...overrides,
  }
}

function mountHarness() {
  const state = {
    projector: null as ReturnType<typeof useContestProjectorData> | null,
  }
  const Harness = defineComponent({
    setup() {
      state.projector = useContestProjectorData()
      return () => null
    },
  })

  const wrapper = mount(Harness)
  if (!state.projector) {
    throw new Error('projector composable was not mounted')
  }
  return {
    wrapper,
    projector: state.projector,
  }
}

describe('useContestProjectorData', () => {
  beforeEach(() => {
    vi.useRealTimers()
    toastMocks.error.mockReset()
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    adminApiMocks.getContests.mockResolvedValue({
      list: [buildContest()],
      total: 1,
      page: 1,
      page_size: 100,
    })
    adminApiMocks.getAdminContestLiveScoreboard.mockResolvedValue({
      contest: buildContest(),
      scoreboard: {
        list: [],
        total: 0,
        page: 1,
        page_size: 20,
      },
    })
    adminApiMocks.listContestAWDRounds.mockResolvedValue([
      buildRound({ id: 'round-1', round_number: 1, status: 'finished' }),
      buildRound({ id: 'round-2', round_number: 2, status: 'running' }),
    ])
    adminApiMocks.listContestAWDRoundServices.mockResolvedValue([])
    adminApiMocks.listContestAWDRoundAttacks.mockResolvedValue([])
    adminApiMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: buildRound({ id: 'round-2' }),
      items: [],
    })
    adminApiMocks.getContestAWDRoundTrafficSummary.mockResolvedValue({
      contest_id: 'contest-1',
      round_id: 'round-2',
      total_request_count: 0,
      active_attacker_team_count: 0,
      victim_team_count: 0,
      error_request_count: 0,
      unique_path_count: 0,
      latest_event_at: undefined,
      top_attackers: [],
      top_victims: [],
      top_challenges: [],
      top_error_paths: [],
      trend_buckets: [],
    })
  })

  it('默认跟随当前运行轮次', async () => {
    const { wrapper, projector } = mountHarness()

    await projector.loadContests()
    await flushPromises()

    expect(projector.roundAutoFollow.value).toBe(true)
    expect(projector.selectedRoundId.value).toBe('round-2')
    expect(adminApiMocks.listContestAWDRoundServices).toHaveBeenLastCalledWith(
      'contest-1',
      'round-2'
    )

    wrapper.unmount()
  })

  it('手动选择历史轮次后刷新不自动跳回运行轮次', async () => {
    const { wrapper, projector } = mountHarness()

    await projector.loadContests()
    await projector.selectRound('round-1')
    await projector.loadScoreboard()
    await flushPromises()

    expect(projector.roundAutoFollow.value).toBe(false)
    expect(projector.selectedRoundId.value).toBe('round-1')
    expect(adminApiMocks.listContestAWDRoundServices).toHaveBeenLastCalledWith(
      'contest-1',
      'round-1'
    )

    wrapper.unmount()
  })

  it('可以从手动轮次恢复实时跟随', async () => {
    const { wrapper, projector } = mountHarness()

    await projector.loadContests()
    await projector.selectRound('round-1')
    await projector.followCurrentRound()
    await flushPromises()

    expect(projector.roundAutoFollow.value).toBe(true)
    expect(projector.selectedRoundId.value).toBe('round-2')
    expect(adminApiMocks.listContestAWDRoundServices).toHaveBeenLastCalledWith(
      'contest-1',
      'round-2'
    )

    wrapper.unmount()
  })
})

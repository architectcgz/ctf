import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, ref } from 'vue'

import { useContestAWDWorkspace } from '@/composables/useContestAWDWorkspace'

const contestApiMocks = vi.hoisted(() => ({
  getContestAWDWorkspace: vi.fn(),
  getScoreboard: vi.fn(),
  startContestAWDServiceInstance: vi.fn(),
  submitContestAWDAttack: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('@/api/contest', () => contestApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('useContestAWDWorkspace', () => {
  beforeEach(() => {
    vi.useRealTimers()
    contestApiMocks.getContestAWDWorkspace.mockReset()
    contestApiMocks.getScoreboard.mockReset()
    contestApiMocks.startContestAWDServiceInstance.mockReset()
    contestApiMocks.submitContestAWDAttack.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()

    contestApiMocks.getContestAWDWorkspace.mockResolvedValue({
      contest_id: '1',
      current_round: {
        id: '41',
        contest_id: '1',
        round_number: 2,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2026-04-12T08:00:00Z',
        updated_at: '2026-04-12T08:01:00Z',
      },
      my_team: {
        team_id: '13',
        team_name: 'Red',
      },
      services: [],
      targets: [],
      recent_events: [],
    })
    contestApiMocks.getScoreboard.mockResolvedValue({
      contest: {
        id: '1',
        title: 'AWD 联赛',
        status: 'running',
        started_at: '2026-04-12T08:00:00Z',
        ends_at: '2026-04-12T10:00:00Z',
      },
      scoreboard: {
        list: [],
        total: 0,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
  })

  it('运行中的比赛应每 15 秒自动刷新，并在结束后停止', async () => {
    vi.useFakeTimers()

    const contestStatus = ref('running')

    const wrapper = mount(
      defineComponent({
        setup() {
          useContestAWDWorkspace({
            contestId: computed(() => '1'),
            contestStatus,
          } as any)
          return () => null
        },
      })
    )

    await flushPromises()

    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.getScoreboard).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(15_000)
    await flushPromises()

    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledTimes(2)
    expect(contestApiMocks.getScoreboard).toHaveBeenCalledTimes(2)

    contestStatus.value = 'ended'
    await flushPromises()

    await vi.advanceTimersByTimeAsync(15_000)
    await flushPromises()

    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledTimes(2)
    expect(contestApiMocks.getScoreboard).toHaveBeenCalledTimes(2)

    wrapper.unmount()
  })

  it('提交攻击后应允许外部格式化 toast 文案', async () => {
    contestApiMocks.submitContestAWDAttack.mockResolvedValueOnce({
      id: '88',
      round_id: '41',
      attacker_team_id: '13',
      attacker_team: 'Red',
      victim_team_id: '14',
      victim_team: 'Blue',
      service_id: '7009',
      challenge_id: 'legacy-101',
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2026-04-12T08:03:00Z',
    })

    let submitAttack!: (serviceId: string, victimTeamId: number, flag: string) => Promise<unknown>

    mount(
      defineComponent({
        setup() {
          const workspace = useContestAWDWorkspace({
            contestId: computed(() => '1'),
            contestStatus: computed(() => 'running'),
            formatAttackResultToast: (result: any) =>
              result.service_id === '7009'
                ? `Bank Portal 攻击成功，+${result.score_gained} 分`
                : '',
          } as any)
          submitAttack = workspace.submitAttack
          return () => null
        },
      })
    )

    await flushPromises()
    await submitAttack('7009', 14, 'flag{demo}')
    await flushPromises()

    expect(toastMocks.success).toHaveBeenCalledWith('Bank Portal 攻击成功，+60 分')
  })

  it('启动 AWD 服务时应优先调用 service_id 实例接口', async () => {
    contestApiMocks.startContestAWDServiceInstance.mockResolvedValueOnce({
      id: '900',
      challenge_id: 'legacy-101',
      status: 'running',
      share_scope: 'per_team',
      access_url: 'http://red.internal',
      flag_type: 'dynamic',
      expires_at: '2026-04-12T12:00:00Z',
      remaining_extends: 1,
      created_at: '2026-04-12T09:02:00Z',
    })

    let startService!: (serviceId: string, challengeId?: string) => Promise<void>

    mount(
      defineComponent({
        setup() {
          const workspace = useContestAWDWorkspace({
            contestId: computed(() => '1'),
            contestStatus: computed(() => 'running'),
          } as any)
          startService = workspace.startService
          return () => null
        },
      })
    )

    await flushPromises()
    await startService('7009', 'legacy-101')
    await flushPromises()

    expect(contestApiMocks.startContestAWDServiceInstance).toHaveBeenCalledWith('1', '7009')
  })
})

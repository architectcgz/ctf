import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, ref } from 'vue'

import { useContestAWDWorkspace } from '@/composables/useContestAWDWorkspace'

const contestApiMocks = vi.hoisted(() => ({
  getContestAWDWorkspace: vi.fn(),
  getScoreboard: vi.fn(),
  requestContestAWDTargetAccess: vi.fn(),
  startContestAWDServiceInstance: vi.fn(),
  submitContestAWDAttack: vi.fn(),
}))

const instanceApiMocks = vi.hoisted(() => ({
  requestInstanceAccess: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('@/api/contest', () => contestApiMocks)
vi.mock('@/api/instance', () => instanceApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('useContestAWDWorkspace', () => {
  beforeEach(() => {
    vi.useRealTimers()
    contestApiMocks.getContestAWDWorkspace.mockReset()
    contestApiMocks.getScoreboard.mockReset()
    contestApiMocks.requestContestAWDTargetAccess.mockReset()
    contestApiMocks.startContestAWDServiceInstance.mockReset()
    contestApiMocks.submitContestAWDAttack.mockReset()
    instanceApiMocks.requestInstanceAccess.mockReset()
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
    contestApiMocks.requestContestAWDTargetAccess.mockResolvedValue({
      access_url: '/api/v1/contests/1/awd/services/7009/targets/14/proxy/',
    })
    instanceApiMocks.requestInstanceAccess.mockResolvedValue({
      access_url: '/api/v1/instances/900/proxy/',
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

  it('攻击提交进行中重复触发时只应发起一次请求', async () => {
    let resolveAttack:
      | ((value: {
          id: string
          round_id: string
          attacker_team_id: string
          attacker_team: string
          victim_team_id: string
          victim_team: string
          service_id: string
          challenge_id: string
          attack_type: string
          source: string
          submitted_flag: string
          is_success: boolean
          score_gained: number
          created_at: string
        }) => void)
      | null = null

    contestApiMocks.submitContestAWDAttack.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveAttack = resolve
        })
    )

    let submitAttack!: (serviceId: string, victimTeamId: number, flag: string) => Promise<unknown>

    mount(
      defineComponent({
        setup() {
          const workspace = useContestAWDWorkspace({
            contestId: computed(() => '1'),
            contestStatus: computed(() => 'running'),
          } as any)
          submitAttack = workspace.submitAttack
          return () => null
        },
      })
    )

    await flushPromises()

    const firstAttempt = submitAttack('7009', 14, 'flag{demo}')
    const secondAttempt = submitAttack('7009', 14, 'flag{demo}')

    expect(contestApiMocks.submitContestAWDAttack).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.submitContestAWDAttack).toHaveBeenCalledWith('1', '7009', {
      victim_team_id: 14,
      flag: 'flag{demo}',
    })

    if (!resolveAttack) {
      throw new Error('attack promise resolver was not captured')
    }

    const finishAttack = resolveAttack as (value: {
      id: string
      round_id: string
      attacker_team_id: string
      attacker_team: string
      victim_team_id: string
      victim_team: string
      service_id: string
      challenge_id: string
      attack_type: string
      source: string
      submitted_flag: string
      is_success: boolean
      score_gained: number
      created_at: string
    }) => void

    finishAttack({
      id: '89',
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
      created_at: '2026-04-12T08:04:00Z',
    })

    await expect(secondAttempt).resolves.toBeNull()
    await expect(firstAttempt).resolves.toMatchObject({
      service_id: '7009',
      victim_team_id: '14',
      submitted_flag: 'flag{demo}',
    })
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

  it('打开跨队攻击入口时应请求目标代理 access 并防止重复点击', async () => {
    let resolveAccess: ((value: { access_url: string }) => void) | null = null
    const openMock = vi.spyOn(window, 'open').mockImplementation(() => null)

    contestApiMocks.requestContestAWDTargetAccess.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveAccess = resolve
        })
    )

    let openTarget!: (serviceId: string, victimTeamId: string) => Promise<string | null>
    let openingTargetKey!: { value: string }

    mount(
      defineComponent({
        setup() {
          const workspace = useContestAWDWorkspace({
            contestId: computed(() => '1'),
            contestStatus: computed(() => 'running'),
          } as any)
          openTarget = workspace.openTarget
          openingTargetKey = workspace.openingTargetKey
          return () => null
        },
      })
    )

    await flushPromises()

    const firstAttempt = openTarget('7009', '14')
    const secondAttempt = openTarget('7009', '14')

    expect(openingTargetKey.value).toBe('7009:14')
    expect(contestApiMocks.requestContestAWDTargetAccess).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.requestContestAWDTargetAccess).toHaveBeenCalledWith('1', '7009', '14')

    if (!resolveAccess) {
      throw new Error('target access promise resolver was not captured')
    }
    const finishAccess = resolveAccess as (value: { access_url: string }) => void
    finishAccess({
      access_url: '/api/v1/contests/1/awd/services/7009/targets/14/proxy/',
    })

    await expect(secondAttempt).resolves.toBeNull()
    await expect(firstAttempt).resolves.toBe(
      '/api/v1/contests/1/awd/services/7009/targets/14/proxy/'
    )
    expect(openMock).toHaveBeenCalledWith(
      '/api/v1/contests/1/awd/services/7009/targets/14/proxy/',
      '_blank',
      'noopener,noreferrer'
    )
    expect(openingTargetKey.value).toBe('')

    openMock.mockRestore()
  })

  it('打开本队服务时应请求实例代理 access 并防止重复点击', async () => {
    let resolveAccess: ((value: { access_url: string }) => void) | null = null
    const openMock = vi.spyOn(window, 'open').mockImplementation(() => null)

    instanceApiMocks.requestInstanceAccess.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveAccess = resolve
        })
    )

    let openService!: (instanceId: string) => Promise<string | null>
    let openingServiceKey!: { value: string }

    mount(
      defineComponent({
        setup() {
          const workspace = useContestAWDWorkspace({
            contestId: computed(() => '1'),
            contestStatus: computed(() => 'running'),
          } as any)
          openService = workspace.openService
          openingServiceKey = workspace.openingServiceKey
          return () => null
        },
      })
    )

    await flushPromises()

    const firstAttempt = openService('900')
    const secondAttempt = openService('900')

    expect(openingServiceKey.value).toBe('900')
    expect(instanceApiMocks.requestInstanceAccess).toHaveBeenCalledTimes(1)
    expect(instanceApiMocks.requestInstanceAccess).toHaveBeenCalledWith('900')

    if (!resolveAccess) {
      throw new Error('instance access promise resolver was not captured')
    }
    const finishAccess = resolveAccess as (value: { access_url: string }) => void
    finishAccess({
      access_url: '/api/v1/instances/900/proxy/',
    })

    await expect(secondAttempt).resolves.toBeNull()
    await expect(firstAttempt).resolves.toBe('/api/v1/instances/900/proxy/')
    expect(openMock).toHaveBeenCalledWith(
      '/api/v1/instances/900/proxy/',
      '_blank',
      'noopener,noreferrer'
    )
    expect(openingServiceKey.value).toBe('')

    openMock.mockRestore()
  })
})

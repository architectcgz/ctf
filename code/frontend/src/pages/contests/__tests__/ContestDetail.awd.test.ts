import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import {
  contestApiMocks,
  contestDetailPageSource,
  contestDetailRoutePageSource,
  contestDetailSource,
  contestDetailWorkspaceSource,
  contestPresentationSource,
  contestTeamDialogsSource,
  contestTeamPanelSource,
  contestTeamWorkspaceSectionSource,
  destructiveConfirmMock,
  resetContestDetailTestHarness,
  routeQueryTransportSource,
  router,
  teamPresentationSource,
  webSocketMocks,
} from './ContestDetail.test-harness'
import ContestDetail from '@/pages/contests/ContestDetailRoutePage.vue'

describe('ContestDetail', () => {
  beforeEach(async () => {
    await resetContestDetailTestHarness()
  })

  it('运行中的 AWD 战场应显示防守告警、自动刷新提示，并支持筛选目标', async () => {
    vi.useFakeTimers()

    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_challenge_id: 'awd-101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValue({
      contest_id: '1',
      current_round: {
        id: '41',
        contest_id: '1',
        round_number: 2,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2024-03-15T09:00:00Z',
        updated_at: '2024-03-15T09:01:00Z',
      },
      my_team: {
        team_id: '13',
        team_name: 'Red',
      },
      services: [
        {
          service_id: '7009',
          awd_challenge_id: 'awd-101',
          access_url: 'http://red.internal',
          service_status: 'compromised',
          checker_type: 'http_standard',
          attack_received: 2,
          sla_score: 0,
          defense_score: 0,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7009',
              awd_challenge_id: 'awd-101',
              reachable: true,
            },
          ],
        },
        {
          team_id: '15',
          team_name: 'Green',
          services: [
            {
              service_id: '7009',
              awd_challenge_id: 'awd-101',
              reachable: false,
            },
          ],
        },
      ],
      recent_events: [],
    })
    contestApiMocks.getScoreboard.mockResolvedValue({
      contest: {
        id: '1',
        title: '2026 春季校园 AWD 联赛',
        status: 'running',
        started_at: '2024-03-15T09:00:00Z',
        ends_at: '2024-03-15T21:00:00Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: '13',
            team_name: 'Red',
            score: 158,
            solved_count: 0,
            last_submission_at: '2024-03-15T09:03:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('我的防守')
    expect(wrapper.text()).toContain('Service A')
    expect(wrapper.text()).toContain('BLUE')
    expect(wrapper.text()).toContain('GREEN')

    await wrapper.get('#awd-target-search').setValue('Blue')
    await flushPromises()

    expect(wrapper.text()).toContain('BLUE')
    expect(wrapper.text()).not.toContain('GREEN')

    await wrapper.get('#awd-target-search').setValue('Yellow')
    await flushPromises()

    expect(wrapper.text()).toContain('当前题目下没有匹配的目标队伍。')

    await vi.advanceTimersByTimeAsync(15_000)
    await flushPromises()

    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledTimes(2)
    expect(contestApiMocks.getScoreboard).toHaveBeenCalledTimes(2)
  })

  it('学生 AWD 工作台应展示运行态 service 标识', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_challenge_id: 'awd-101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
      contest_id: '1',
      current_round: {
        id: '41',
        contest_id: '1',
        round_number: 2,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2024-03-15T09:00:00Z',
        updated_at: '2024-03-15T09:01:00Z',
      },
      my_team: {
        team_id: '13',
        team_name: 'Red',
      },
      services: [
        {
          service_id: '7009',
          awd_challenge_id: 'awd-101',
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [
        {
          team_id: '14',
          team_name: 'Blue',
          services: [
            {
              service_id: '7010',
              awd_challenge_id: 'awd-101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [
        {
          id: 'attack-1',
          service_id: '7010',
          awd_challenge_id: 'awd-101',
          direction: 'attack_out',
          peer_team_id: '14',
          peer_team_name: 'Blue',
          is_success: true,
          score_gained: 60,
          created_at: '2024-03-15T09:03:00Z',
        },
      ],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('服务 #7009')
    expect(wrapper.text()).toContain('服务 #7010')
  })

  it('学生 AWD 战场不再暴露独立防守工作台入口按钮', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '201',
        challenge_id: '101',
        awd_challenge_id: 'awd-101',
        awd_service_id: '7009',
        title: 'Service A',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.getContestAWDWorkspace.mockResolvedValueOnce({
      contest_id: '1',
      current_round: {
        id: '41',
        contest_id: '1',
        round_number: 2,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2024-03-15T09:00:00Z',
        updated_at: '2024-03-15T09:01:00Z',
      },
      my_team: {
        team_id: '13',
        team_name: 'Red',
      },
      services: [
        {
          service_id: '7009',
          awd_challenge_id: 'awd-101',
          instance_id: '9001',
          access_url: 'http://red.internal',
          instance_status: 'running',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
      ],
      targets: [],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    const defenseButton = wrapper.findAll('button').find((node) => node.text().trim() === '防守')
    expect(defenseButton).toBeUndefined()
    expect(router.currentRoute.value.fullPath).toBe('/contests/1?panel=challenges')
  })
})

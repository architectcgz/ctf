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

  it('AWD 赛事应切换到战场页签并加载学生工作台', async () => {
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
              service_id: '7009',
              awd_challenge_id: 'awd-101',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })
    contestApiMocks.getScoreboard.mockResolvedValueOnce({
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

    expect(wrapper.text()).toContain('攻防战场')
    expect(wrapper.text()).toContain('攻击向量')
    expect(wrapper.text()).toContain('目标题目')
    expect(wrapper.text()).toContain('BLUE')
    expect(contestApiMocks.getContestAWDWorkspace).toHaveBeenCalledWith('1')
  })

  it('AWD 赛事在未入队时应显示先加入队伍的提示', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '2026 春季校园 AWD 联赛',
      description: '测试描述',
      status: 'running',
      mode: 'awd',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
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
      my_team: null,
      services: [],
      targets: [],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('先加入队伍')
    expect(wrapper.text()).toContain('需要先加入队伍后才能进入 AWD 战场。')
    expect(wrapper.text()).toContain('攻防战场')
  })
})

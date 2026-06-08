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

  it('学生 AWD 工作台应允许用 awd service 标识切换攻击题目', async () => {
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
      {
        id: '202',
        challenge_id: '102',
        awd_challenge_id: 'awd-102',
        awd_service_id: '7010',
        title: 'Service B',
        category: 'pwn',
        difficulty: 'hard',
        points: 200,
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
          access_url: 'http://red-a.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:00Z',
        },
        {
          service_id: '7010',
          awd_challenge_id: 'awd-102',
          access_url: 'http://red-b.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2024-03-15T09:02:30Z',
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
            {
              service_id: '7010',
              awd_challenge_id: 'awd-102',
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [],
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('服务 #7009')

    await wrapper.get('#awd-target-challenge').setValue('7010')
    await flushPromises()

    expect(wrapper.text()).toContain('服务 #7010')
    expect(wrapper.text()).not.toContain('http://blue-a.internal')
    expect(wrapper.text()).not.toContain('http://blue-b.internal')
  })

  it('学生 AWD 工作台应通过跨队攻击代理打开目标服务', async () => {
    const openMock = vi.spyOn(window, 'open').mockImplementation(() => null)
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
        title: 'Bank Portal',
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

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    await wrapper.get('[data-testid="awd-open-target-7009-14"]').trigger('click')
    await flushPromises()

    expect(contestApiMocks.requestContestAWDTargetAccess).toHaveBeenCalledWith('1', '7009', '14')
    expect(openMock).toHaveBeenCalledWith(
      '/api/v1/contests/1/awd/services/7009/targets/14/proxy/',
      '_blank',
      'noopener,noreferrer'
    )

    openMock.mockRestore()
  })
})

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

  it('学生 AWD 最近反馈应优先按 service 标识回填题目标题', async () => {
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
      targets: [],
      recent_events: [
        {
          id: 'attack-1',
          service_id: '7009',
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

    expect(
      wrapper.findAll('[data-testid="awd-feedback-challenge-title"]').map((node) => node.text())
    ).toContain('Bank Portal')
  })

  it('学生 AWD 提交结果提示应优先按 service 标识回填题目标题', async () => {
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
    contestApiMocks.getScoreboard.mockResolvedValue({
      contest: {
        id: '1',
        title: '2026 春季校园 AWD 联赛',
        status: 'running',
        started_at: '2024-03-15T09:00:00Z',
        ends_at: '2024-03-15T21:00:00Z',
      },
      scoreboard: {
        list: [],
        total: 0,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
    contestApiMocks.submitContestAWDAttack.mockResolvedValueOnce({
      id: '88',
      round_id: '41',
      attacker_team_id: '13',
      attacker_team: 'Red',
      victim_team_id: '14',
      victim_team: 'Blue',
      service_id: '7009',
      awd_challenge_id: 'awd-101',
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2024-03-15T09:03:00Z',
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    await wrapper.find('input[placeholder="输入获取到的 Flag..."]').setValue('flag{demo}')
    await wrapper
      .findAll('button')
      .find((node) => node.text().trim() === '提交')
      ?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Bank Portal: 攻击成功，+60 分')
  })
})

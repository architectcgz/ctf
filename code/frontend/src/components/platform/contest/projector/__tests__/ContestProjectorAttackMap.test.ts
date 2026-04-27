import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ContestProjectorAttackMap from '@/components/platform/contest/projector/ContestProjectorAttackMap.vue'
import type { AWDTeamServiceData, ScoreboardRow } from '@/api/contracts'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorServiceMatrixRow,
} from '@/components/platform/contest/projector/contestProjectorTypes'

class ResizeObserverStub {
  observe = vi.fn()
  disconnect = vi.fn()
}

function buildEdge(overrides: Partial<ContestProjectorAttackEdge> = {}): ContestProjectorAttackEdge {
  return {
    id: 'blue->red',
    attacker_team_id: 'blue',
    attacker_team: 'Blue Team',
    victim_team_id: 'red',
    victim_team: 'Red Team',
    latest_service_id: 'service-1',
    latest_challenge_id: 'challenge-1',
    latest_target_key: 'red:service:service-1',
    success: 1,
    failed: 1,
    total: 2,
    score: 30,
    latest_at: '2026-04-27T15:49:02.000Z',
    latest_service_label: 'Supply Ticket',
    successRate: 50,
    reciprocalSuccess: 1,
    ...overrides,
  }
}

function buildService(overrides: Partial<AWDTeamServiceData> = {}): AWDTeamServiceData {
  return {
    id: 'service-state-1',
    round_id: 'round-1',
    team_id: 'red',
    team_name: 'Red Team',
    service_id: 'service-1',
    service_name: 'Supply Ticket',
    challenge_id: 'challenge-1',
    challenge_title: 'Supply Ticket Challenge',
    service_status: 'up',
    check_result: {},
    attack_received: 0,
    sla_score: 0,
    defense_score: 0,
    attack_score: 0,
    updated_at: '2026-04-27T15:49:02.000Z',
    ...overrides,
  }
}

function buildRows(): ContestProjectorServiceMatrixRow[] {
  return [
    {
      team_id: 'blue',
      team_name: 'Blue Team',
      services: [
        buildService({
          id: 'service-state-2',
          team_id: 'blue',
          team_name: 'Blue Team',
          service_id: 'service-2',
          service_name: 'IoT Hub',
          challenge_id: 'challenge-2',
        }),
      ],
    },
    {
      team_id: 'red',
      team_name: 'Red Team',
      services: [buildService()],
    },
  ]
}

function buildScoreboardRows(): ScoreboardRow[] {
  return [
    {
      rank: 1,
      team_id: 'blue',
      team_name: 'Blue Team',
      score: 3520,
      solved_count: 2,
      last_submission_at: '2026-04-27T15:49:02.000Z',
    },
    {
      rank: 2,
      team_id: 'red',
      team_name: 'Red Team',
      score: 2980,
      solved_count: 1,
      last_submission_at: '2026-04-27T15:48:02.000Z',
    },
  ]
}

describe('ContestProjectorAttackMap', () => {
  beforeEach(() => {
    vi.stubGlobal('ResizeObserver', ResizeObserverStub)
  })

  it('显示攻击方、目标方和互攻状态', () => {
    const wrapper = mount(ContestProjectorAttackMap, {
      props: {
        rows: buildRows(),
        edges: [buildEdge()],
        scoreboardRows: buildScoreboardRows(),
      },
    })

    expect(wrapper.text()).toContain('Blue Team')
    expect(wrapper.text()).toContain('Red Team')
    expect(wrapper.text()).toContain('图例说明')
    expect(wrapper.text()).toContain('实时攻击地图')
    expect(wrapper.text()).toContain('比赛状态')
    expect(wrapper.text()).toContain('团队排名')
    expect(wrapper.text()).toContain('成功 1')
    expect(wrapper.text()).toContain('失败 1')
    expect(wrapper.text()).toContain('Supply Ticket')
    expect(wrapper.text()).toContain('互攻')
  })

  it('没有攻击关系时显示空状态', () => {
    const wrapper = mount(ContestProjectorAttackMap, {
      props: {
        rows: [],
        edges: [],
        scoreboardRows: [],
      },
    })

    expect(wrapper.text()).toContain('暂无目标服务')
  })
})

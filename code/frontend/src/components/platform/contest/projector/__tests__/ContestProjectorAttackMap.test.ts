import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import ContestProjectorAttackMap from '@/components/platform/contest/projector/ContestProjectorAttackMap.vue'
import type { AWDAttackLogData, AWDTeamServiceData, ScoreboardRow } from '@/api/contracts'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorServiceMatrixRow,
} from '@/components/platform/contest/projector/contestProjectorTypes'

const attackDetailOverlayCss = readFileSync(
  join(
    process.cwd(),
    'src/components/platform/contest/projector/ContestProjectorAttackDetailOverlay.css'
  ),
  'utf-8'
)

class ResizeObserverStub {
  observe = vi.fn()
  disconnect = vi.fn()
}

function buildEdge(
  overrides: Partial<ContestProjectorAttackEdge> = {}
): ContestProjectorAttackEdge {
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

function buildAttackEvent(overrides: Partial<AWDAttackLogData> = {}): AWDAttackLogData {
  return {
    id: 'attack-1',
    round_id: 'round-1',
    attacker_team_id: 'blue',
    attacker_team: 'Blue Team',
    victim_team_id: 'red',
    victim_team: 'Red Team',
    service_id: 'service-1',
    challenge_id: 'challenge-1',
    attack_type: 'flag_capture',
    source: 'submission',
    is_success: true,
    score_gained: 30,
    created_at: '2026-04-27T15:49:02.000Z',
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

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('显示攻击方、目标方和互攻状态', () => {
    const wrapper = mount(ContestProjectorAttackMap, {
      props: {
        rows: buildRows(),
        edges: [buildEdge()],
        scoreboardRows: buildScoreboardRows(),
        firstBlood: buildAttackEvent(),
        latestAttackEvents: [
          buildAttackEvent(),
          buildAttackEvent({
            id: 'attack-2',
            is_success: false,
            score_gained: 0,
            created_at: '2026-04-27T15:48:02.000Z',
          }),
        ],
      },
    })

    expect(wrapper.text()).toContain('Blue Team')
    expect(wrapper.text()).toContain('Red Team')
    expect(wrapper.text()).toContain('图例说明')
    expect(wrapper.text()).toContain('实时攻击地图')
    expect(wrapper.text()).toContain('团队排名')
    expect(wrapper.text()).toContain('首血')
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
        firstBlood: null,
        latestAttackEvents: [],
      },
    })

    expect(wrapper.text()).toContain('暂无攻击事件')
  })

  it('点击团队排名时打开完整排名详情', async () => {
    const wrapper = mount(ContestProjectorAttackMap, {
      attachTo: document.body,
      props: {
        rows: buildRows(),
        edges: [buildEdge()],
        scoreboardRows: buildScoreboardRows(),
        firstBlood: buildAttackEvent(),
        latestAttackEvents: [buildAttackEvent()],
      },
    })

    await wrapper.find('.rank-block').trigger('click')

    expect(document.body.textContent).toContain('完整团队排名')
    expect(document.body.textContent).toContain('解题 2')
    expect(document.body.textContent).toContain('受损 0')
  })

  it('详情弹出框应使用实体 surface 背景，避免排名和服务列表透出底层画面', () => {
    expect(attackDetailOverlayCss).toContain('background: var(--projector-detail-panel-surface);')
    expect(attackDetailOverlayCss).toContain('background: var(--projector-detail-item-surface);')
    expect(attackDetailOverlayCss).not.toContain('var(--color-bg-elevated) 58%, transparent')
    expect(attackDetailOverlayCss).not.toContain('var(--color-bg-elevated) 50%, transparent')
    expect(attackDetailOverlayCss).not.toContain('var(--color-bg-elevated) 52%, transparent')
  })
})

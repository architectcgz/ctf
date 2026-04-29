import { describe, expect, it } from 'vitest'
import { ref } from 'vue'

import type { AWDAttackLogData, AWDTrafficSummaryData, AWDTeamServiceData, ScoreboardRow } from '@/api/contracts'
import { useContestProjectorDerived } from '@/composables/useContestProjectorDerived'

function buildAttack(overrides: Partial<AWDAttackLogData> = {}): AWDAttackLogData {
  return {
    id: 'attack-1',
    round_id: 'round-1',
    attacker_team_id: 'blue',
    attacker_team: 'Blue Team',
    victim_team_id: 'red',
    victim_team: 'Red Team',
    service_id: 'service-1',
    awd_challenge_id: 'challenge-1',
    attack_type: 'flag_capture',
    source: 'submission',
    is_success: true,
    score_gained: 30,
    created_at: '2026-04-27T15:49:02.000Z',
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
    awd_challenge_id: 'challenge-1',
    awd_challenge_title: 'Supply Ticket Challenge',
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

describe('useContestProjectorDerived', () => {
  it('按攻击方向聚合队伍互攻关系', () => {
    const derived = useContestProjectorDerived({
      scoreboardRows: ref<ScoreboardRow[]>([]),
      services: ref<AWDTeamServiceData[]>([
        buildService(),
        buildService({
          id: 'service-state-2',
          team_id: 'blue',
          team_name: 'Blue Team',
          service_id: 'service-2',
          service_name: 'IoT Hub',
          awd_challenge_id: 'challenge-2',
        }),
      ]),
      attacks: ref<AWDAttackLogData[]>([
        buildAttack(),
        buildAttack({
          id: 'attack-2',
          is_success: false,
          score_gained: 0,
          created_at: '2026-04-27T15:50:02.000Z',
        }),
        buildAttack({
          id: 'attack-3',
          attacker_team_id: 'red',
          attacker_team: 'Red Team',
          victim_team_id: 'blue',
          victim_team: 'Blue Team',
          service_id: 'service-2',
          awd_challenge_id: 'challenge-2',
          created_at: '2026-04-27T15:51:02.000Z',
        }),
      ]),
      trafficSummary: ref<AWDTrafficSummaryData | null>(null),
    })

    expect(derived.attackEdges.value).toHaveLength(2)
    expect(derived.attackEdges.value[0]).toMatchObject({
      attacker_team: 'Red Team',
      victim_team: 'Blue Team',
      success: 1,
      failed: 0,
      score: 30,
      latest_service_id: 'service-2',
      latest_challenge_id: 'challenge-2',
      latest_target_key: 'blue:service:service-2',
      latest_service_label: 'IoT Hub',
      reciprocalSuccess: 1,
    })
    expect(derived.attackEdges.value[1]).toMatchObject({
      attacker_team: 'Blue Team',
      victim_team: 'Red Team',
      success: 1,
      failed: 1,
      total: 2,
      latest_service_id: 'service-1',
      latest_challenge_id: 'challenge-1',
      latest_target_key: 'red:service:service-1',
      latest_service_label: 'Supply Ticket',
      successRate: 50,
      reciprocalSuccess: 1,
    })
  })
})

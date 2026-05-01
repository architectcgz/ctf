import { computed, ref } from 'vue'
import { describe, expect, it } from 'vitest'

import { useAwdInspectorDerivedData } from '@/features/awd-inspector'

describe('useAwdInspectorDerivedData', () => {
  it('服务告警样本应优先保留运行态 service_id', () => {
    const services = ref([
      {
        id: 'service-row-1',
        round_id: 'round-2',
        team_id: 'team-1',
        team_name: 'Blue Team',
        service_id: '7009',
        awd_challenge_id: 'challenge-1',
        service_status: 'down' as const,
        checker_type: 'http_standard' as const,
        check_result: {
          error_code: 'connect_timeout',
          check_source: 'manual_current_round',
        },
        attack_received: 0,
        sla_score: 0,
        defense_score: 0,
        attack_score: 0,
        updated_at: '2026-04-12T10:02:00Z',
      },
    ])

    const { serviceAlerts } = useAwdInspectorDerivedData({
      services,
      attacks: ref([]),
      trafficSummary: ref(null),
      trafficEvents: ref([]),
      serviceTeamFilter: ref(''),
      serviceStatusFilter: ref('all'),
      serviceCheckSourceFilter: ref(''),
      serviceAlertReasonFilter: ref(''),
      attackTeamFilter: ref(''),
      attackResultFilter: ref('all'),
      attackSourceFilter: ref('all'),
      getChallengeTitle: (challengeId: string) => `Challenge ${challengeId}`,
      getCheckStatusLabel: (value: unknown) => String(value || ''),
    })

    expect(serviceAlerts.value).toHaveLength(1)
    expect(serviceAlerts.value[0]?.samples[0]?.service_id).toBe('7009')
  })

  it('服务告警样本缺少运行态 service_id 时不应回退到行 id', () => {
    const services = ref([
      {
        id: 'service-row-legacy',
        round_id: 'round-2',
        team_id: 'team-1',
        team_name: 'Blue Team',
        awd_challenge_id: 'challenge-1',
        service_status: 'down' as const,
        checker_type: 'http_standard' as const,
        check_result: {
          error_code: 'connect_timeout',
          check_source: 'manual_current_round',
        },
        attack_received: 0,
        sla_score: 0,
        defense_score: 0,
        attack_score: 0,
        updated_at: '2026-04-12T10:02:00Z',
      },
    ])

    const { serviceAlerts } = useAwdInspectorDerivedData({
      services,
      attacks: ref([]),
      trafficSummary: ref(null),
      trafficEvents: ref([]),
      serviceTeamFilter: ref(''),
      serviceStatusFilter: ref('all'),
      serviceCheckSourceFilter: ref(''),
      serviceAlertReasonFilter: ref(''),
      attackTeamFilter: ref(''),
      attackResultFilter: ref('all'),
      attackSourceFilter: ref('all'),
      getChallengeTitle: (challengeId: string) => `Challenge ${challengeId}`,
      getCheckStatusLabel: (value: unknown) => String(value || ''),
    })

    expect(serviceAlerts.value).toHaveLength(1)
    expect(serviceAlerts.value[0]?.samples[0]?.service_id).toBe('')
  })
})

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ContestProjectorServiceMatrix from '@/components/platform/contest/projector/ContestProjectorServiceMatrix.vue'
import type { AWDTeamServiceData } from '@/api/contracts'

function buildService(overrides: Partial<AWDTeamServiceData> = {}): AWDTeamServiceData {
  return {
    id: 'service-state-1',
    round_id: 'round-1',
    team_id: 'team-1',
    team_name: 'Red Team',
    service_id: 'svc-1',
    service_name: 'Bank Portal',
    awd_challenge_id: 'challenge-1',
    awd_challenge_title: 'Web Bank',
    service_status: 'up',
    check_result: {},
    attack_received: 0,
    sla_score: 1,
    defense_score: 2,
    attack_score: 0,
    updated_at: '2026-04-12T09:00:00.000Z',
    ...overrides,
  }
}

describe('ContestProjectorServiceMatrix', () => {
  it('显示服务名和状态，避免只显示状态码', () => {
    const wrapper = mount(ContestProjectorServiceMatrix, {
      props: {
        rows: [
          {
            team_id: 'team-1',
            team_name: 'Red Team',
            services: [buildService()],
          },
        ],
        upCount: 1,
        downCount: 0,
        compromisedCount: 0,
      },
    })

    expect(wrapper.text()).toContain('Bank Portal')
    expect(wrapper.text()).toContain('UP')
    expect(wrapper.get('.service-cell').attributes('title')).toContain('Bank Portal')
  })

  it('没有服务名时回退显示服务标识', () => {
    const wrapper = mount(ContestProjectorServiceMatrix, {
      props: {
        rows: [
          {
            team_id: 'team-1',
            team_name: 'Red Team',
            services: [buildService({ service_name: undefined, awd_challenge_title: undefined, service_id: 'svc-2' })],
          },
        ],
        upCount: 1,
        downCount: 0,
        compromisedCount: 0,
      },
    })

    expect(wrapper.text()).toContain('服务 svc-2')
  })
})

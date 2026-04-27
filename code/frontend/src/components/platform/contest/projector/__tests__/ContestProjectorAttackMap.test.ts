import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ContestProjectorAttackMap from '@/components/platform/contest/projector/ContestProjectorAttackMap.vue'
import type { ContestProjectorAttackEdge } from '@/components/platform/contest/projector/contestProjectorTypes'

function buildEdge(overrides: Partial<ContestProjectorAttackEdge> = {}): ContestProjectorAttackEdge {
  return {
    id: 'blue->red',
    attacker_team_id: 'blue',
    attacker_team: 'Blue Team',
    victim_team_id: 'red',
    victim_team: 'Red Team',
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

describe('ContestProjectorAttackMap', () => {
  it('显示攻击方、目标方和互攻状态', () => {
    const wrapper = mount(ContestProjectorAttackMap, {
      props: {
        edges: [buildEdge()],
      },
    })

    expect(wrapper.text()).toContain('Blue Team')
    expect(wrapper.text()).toContain('Red Team')
    expect(wrapper.text()).toContain('1 HIT')
    expect(wrapper.text()).toContain('1 MISS')
    expect(wrapper.text()).toContain('Supply Ticket')
    expect(wrapper.text()).toContain('互攻 1')
  })

  it('没有攻击关系时显示空状态', () => {
    const wrapper = mount(ContestProjectorAttackMap, {
      props: {
        edges: [],
      },
    })

    expect(wrapper.text()).toContain('暂无队伍攻击关系')
  })
})

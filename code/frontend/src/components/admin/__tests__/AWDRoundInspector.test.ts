import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AWDRoundInspector from '../contest/AWDRoundInspector.vue'

describe('AWDRoundInspector', () => {
  it('攻击流量明细应显示显式 service 引用', () => {
    const wrapper = mount(AWDRoundInspector, {
      props: {
        contest: {
          id: 'contest-1',
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-04-18T09:00:00.000Z',
          ends_at: '2026-04-18T18:00:00.000Z',
        },
        rounds: [
          {
            id: 'round-1',
            contest_id: 'contest-1',
            round_number: 1,
            status: 'running',
            attack_score: 60,
            defense_score: 40,
            started_at: '2026-04-18T09:05:00.000Z',
            ended_at: undefined,
            created_at: '2026-04-18T09:05:00.000Z',
            updated_at: '2026-04-18T09:10:00.000Z',
          },
        ],
        selectedRoundId: 'round-1',
        services: [],
        attacks: [],
        challengeLinks: [
          {
            id: 'link-1',
            contest_id: 'contest-1',
            challenge_id: '101',
            awd_service_id: '7009',
            title: 'Bank Portal',
            category: 'web',
            difficulty: 'medium',
            points: 160,
            order: 1,
            is_visible: true,
            created_at: '2026-04-18T09:00:00.000Z',
          },
        ],
        summary: null,
        trafficSummary: null,
        trafficEvents: [
          {
            id: 'event-1',
            contest_id: 'contest-1',
            round_id: 'round-1',
            attacker_team_id: 'team-2',
            attacker_team_name: 'Red Team',
            victim_team_id: 'team-1',
            victim_team_name: 'Blue Team',
            service_id: '7009',
            challenge_id: '101',
            challenge_title: 'Bank Portal',
            method: 'GET',
            path: '/login',
            status_code: 500,
            status_group: 'server_error',
            is_error: true,
            source: 'runtime_proxy',
            occurred_at: '2026-04-18T09:09:00.000Z',
          },
        ],
        trafficEventsTotal: 1,
        trafficFilters: {
          attacker_team_id: '',
          victim_team_id: '',
          service_id: '',
          challenge_id: '',
          status_group: 'all',
          path_keyword: '',
          page: 1,
          page_size: 20,
        },
        scoreboardRows: [],
        scoreboardFrozen: false,
        loadingRounds: false,
        loadingRoundDetail: false,
        loadingTrafficSummary: false,
        loadingTrafficEvents: false,
        checking: false,
        shouldAutoRefresh: false,
        canRecordServiceChecks: false,
        canRecordAttackLogs: false,
      },
      global: {
        stubs: {
          AppCard: {
            template: '<div><slot name="header" /><slot /></div>',
          },
          SectionCard: {
            template: '<section><slot /></section>',
          },
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AppEmpty: {
            props: ['title', 'description'],
            template: '<div>{{ title }}{{ description }}</div>',
          },
          AdminPaginationControls: {
            template: '<div />',
          },
        },
      },
    })

    expect(wrapper.text()).toContain('Service #7009')
    expect(wrapper.text()).toContain('Bank Portal')
  })

  it('流量筛选应发出 service_id patch', async () => {
    const wrapper = mount(AWDRoundInspector, {
      props: {
        contest: {
          id: 'contest-1',
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-04-18T09:00:00.000Z',
          ends_at: '2026-04-18T18:00:00.000Z',
        },
        rounds: [
          {
            id: 'round-1',
            contest_id: 'contest-1',
            round_number: 1,
            status: 'running',
            attack_score: 60,
            defense_score: 40,
            started_at: '2026-04-18T09:05:00.000Z',
            ended_at: undefined,
            created_at: '2026-04-18T09:05:00.000Z',
            updated_at: '2026-04-18T09:10:00.000Z',
          },
        ],
        selectedRoundId: 'round-1',
        services: [],
        attacks: [],
        challengeLinks: [
          {
            id: 'link-1',
            contest_id: 'contest-1',
            challenge_id: '101',
            awd_service_id: '7009',
            title: 'Bank Portal',
            category: 'web',
            difficulty: 'medium',
            points: 160,
            order: 1,
            is_visible: true,
            created_at: '2026-04-18T09:00:00.000Z',
          },
        ],
        summary: null,
        trafficSummary: null,
        trafficEvents: [],
        trafficEventsTotal: 0,
        trafficFilters: {
          attacker_team_id: '',
          victim_team_id: '',
          service_id: '',
          challenge_id: '',
          status_group: 'all',
          path_keyword: '',
          page: 1,
          page_size: 20,
        },
        scoreboardRows: [],
        scoreboardFrozen: false,
        loadingRounds: false,
        loadingRoundDetail: false,
        loadingTrafficSummary: false,
        loadingTrafficEvents: false,
        checking: false,
        shouldAutoRefresh: false,
        canRecordServiceChecks: false,
        canRecordAttackLogs: false,
      },
      global: {
        stubs: {
          AppCard: {
            template: '<div><slot name="header" /><slot /></div>',
          },
          SectionCard: {
            template: '<section><slot /></section>',
          },
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AppEmpty: {
            props: ['title', 'description'],
            template: '<div>{{ title }}{{ description }}</div>',
          },
          AdminPaginationControls: {
            template: '<div />',
          },
        },
      },
    })

    await wrapper.get('#awd-traffic-filter-service').setValue('7009')

    expect(wrapper.emitted('applyTrafficFilters')).toEqual([[{ service_id: '7009' }]])
  })
})

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'

import AWDRoundInspector from '../contest/AWDRoundInspector.vue'

const baseContest = {
  id: 'contest-1',
  title: '2026 AWD 联赛',
  description: '攻防赛',
  mode: 'awd' as const,
  status: 'running' as const,
  starts_at: '2026-04-18T09:00:00.000Z',
  ends_at: '2026-04-18T18:00:00.000Z',
}

const baseRound = {
  id: 'round-1',
  contest_id: 'contest-1',
  round_number: 1,
  status: 'running' as const,
  attack_score: 60,
  defense_score: 40,
  started_at: '2026-04-18T09:05:00.000Z',
  ended_at: undefined,
  created_at: '2026-04-18T09:05:00.000Z',
  updated_at: '2026-04-18T09:10:00.000Z',
}

function buildRequiredProps(overrides: Record<string, unknown> = {}) {
  return {
    contest: baseContest,
    rounds: [baseRound],
    selectedRoundId: 'round-1',
    services: [],
    attacks: [],
    challengeLinks: [],
    summary: null,
    trafficSummary: null,
    trafficEvents: [],
    trafficEventsTotal: 0,
    trafficFilters: {
      attacker_team_id: '',
      victim_team_id: '',
      service_id: '',
      challenge_id: '',
      status_group: 'all' as const,
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
    ...overrides,
  }
}

const globalStubs = {
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
  PlatformPaginationControls: {
    template: '<div />',
  },
}

describe('AWDRoundInspector', () => {
  it('服务告警只通过 scoped slot 暴露，由上层页面决定是否展示', () => {
    const compromisedService = {
      id: 'service-row-1',
      round_id: 'round-1',
      team_id: 'team-1',
      team_name: 'Blue Team',
      service_id: '7009',
      challenge_id: '101',
      challenge_title: 'Bank Portal',
      service_status: 'compromised' as const,
      checker_type: 'http_standard' as const,
      check_result: {},
      attack_received: 1,
      sla_score: 0,
      defense_score: 0,
      attack_score: 60,
      updated_at: '2026-04-18T09:10:00.000Z',
    }

    const wrapperWithoutSlot = mount(AWDRoundInspector, {
      props: buildRequiredProps({
        services: [compromisedService],
      }),
      global: {
        stubs: globalStubs,
      },
    })

    expect(wrapperWithoutSlot.text()).not.toContain('重点异常告警')
    expect(wrapperWithoutSlot.text()).not.toContain('service_compromised')

    const wrapperWithSlot = mount(AWDRoundInspector, {
      props: buildRequiredProps({
        services: [compromisedService],
      }),
      slots: {
        'service-alerts': ({ serviceAlerts }) =>
          h(
            'div',
            { 'data-testid': 'service-alert-slot' },
            serviceAlerts.map((alert) => `${alert.label} (${alert.count})`).join(' ')
          ),
      },
      global: {
        stubs: globalStubs,
      },
    })

    expect(wrapperWithSlot.get('[data-testid="service-alert-slot"]').text()).toContain(
      '服务已失陷 (1)'
    )
  })

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
          PlatformPaginationControls: {
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
          PlatformPaginationControls: {
            template: '<div />',
          },
        },
      },
    })

    await wrapper.get('#awd-traffic-filter-service').setValue('7009')

    expect(wrapper.emitted('applyTrafficFilters')).toEqual([[{ service_id: '7009' }]])
  })

  it('轮次汇总应展示手动巡检合计，而不是渲染 undefined', () => {
    const wrapper = mount(AWDRoundInspector, {
      props: buildRequiredProps({
        services: [
          {
            id: 'service-row-1',
            round_id: 'round-1',
            team_id: 'team-1',
            team_name: 'Blue Team',
            service_id: '7009',
            challenge_id: '101',
            challenge_title: 'Bank Portal',
            service_status: 'up',
            checker_type: 'http_standard',
            check_result: {
              check_source: 'manual_service_check',
            },
            attack_received: 0,
            sla_score: 0,
            defense_score: 0,
            attack_score: 0,
            updated_at: '2026-04-18T09:10:00.000Z',
          },
        ],
        summary: {
          round: baseRound,
          metrics: {
            total_service_count: 1,
            service_up_count: 1,
            service_down_count: 0,
            service_compromised_count: 0,
            attacked_service_count: 0,
            defense_success_count: 0,
            total_attack_count: 3,
            successful_attack_count: 2,
            failed_attack_count: 1,
            scheduler_check_count: 0,
            manual_current_round_check_count: 0,
            manual_selected_round_check_count: 0,
            manual_service_check_count: 1,
            submission_attack_count: 3,
            manual_attack_log_count: 0,
            legacy_attack_log_count: 0,
          },
          items: [],
        },
      }),
      global: {
        stubs: globalStubs,
      },
    })

    expect(wrapper.text()).toContain('巡检 调度 0 / 手动 1')
    expect(wrapper.text()).toContain('日志 提交 3 / 人工 0 / 历史 0')
    expect(wrapper.text()).not.toContain('undefined')
  })
})

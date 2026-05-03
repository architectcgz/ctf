import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { AWDTeamServiceData } from '@/api/contracts'
import AWDServiceStatusPanel from '../contest/AWDServiceStatusPanel.vue'

function buildProps() {
  const services: AWDTeamServiceData[] = [
    {
      id: 'svc-1',
      round_id: 'round-1',
      team_id: 'team-1',
      team_name: 'Blue Team',
      service_id: '7009',
      service_name: 'Bank Portal',
      awd_challenge_id: 'challenge-1',
      awd_challenge_title: 'Bank Portal',
      service_status: 'up' as const,
      checker_type: 'http_standard' as const,
      check_result: {
        checker_type: 'http_standard',
        check_source: 'scheduler',
        status_reason: 'healthy',
        checked_at: '2026-04-18T09:10:00.000Z',
      },
      attack_received: 0,
      sla_score: 10,
      defense_score: 20,
      attack_score: 30,
      updated_at: '2026-04-18T09:10:00.000Z',
    },
    {
      id: 'svc-2',
      round_id: 'round-1',
      team_id: 'team-1',
      team_name: 'Blue Team',
      service_id: '7010',
      service_name: 'Mail Relay',
      awd_challenge_id: 'challenge-2',
      awd_challenge_title: 'Mail Relay',
      service_status: 'down' as const,
      checker_type: 'http_standard' as const,
      check_result: {
        checker_type: 'http_standard',
        check_source: 'manual_current_round',
        status_reason: 'http_request_failed',
        checked_at: '2026-04-18T09:10:00.000Z',
      },
      attack_received: 2,
      sla_score: 0,
      defense_score: 5,
      attack_score: 0,
      updated_at: '2026-04-18T09:10:00.000Z',
    },
    {
      id: 'svc-3',
      round_id: 'round-1',
      team_id: 'team-2',
      team_name: 'Red Team',
      service_id: '7009',
      service_name: 'Bank Portal',
      awd_challenge_id: 'challenge-1',
      awd_challenge_title: 'Bank Portal',
      service_status: 'compromised' as const,
      checker_type: 'http_standard' as const,
      check_result: {
        checker_type: 'http_standard',
        check_source: 'manual_selected_round',
        status_reason: 'service_compromised',
        checked_at: '2026-04-18T09:10:00.000Z',
      },
      attack_received: 4,
      sla_score: 0,
      defense_score: 0,
      attack_score: 0,
      updated_at: '2026-04-18T09:10:00.000Z',
    },
  ]

  return {
    services,
    filteredServices: services,
    summary: null,
    serviceAlerts: [],
    serviceTeamOptions: [
      { team_id: 'team-1', team_name: 'Blue Team' },
      { team_id: 'team-2', team_name: 'Red Team' },
    ],
    serviceCheckSourceOptions: ['scheduler'],
    serviceTeamFilter: '',
    serviceStatusFilter: 'all' as const,
    serviceCheckSourceFilter: '',
    serviceAlertReasonFilter: '',
    getChallengeTitle: (challengeId: string) =>
      challengeId === 'challenge-1' ? 'Bank Portal' : 'Mail Relay',
    getServiceStatusLabel: (status: 'up' | 'down' | 'compromised') =>
      ({
        up: '在线',
        down: '离线',
        compromised: '失陷',
      })[status],
    getServiceStatusClass: (status: 'up' | 'down' | 'compromised') => `status--${status}`,
    getCheckSourceLabel: (source: unknown) => String(source),
    getCheckerTypeLabel: (value: unknown) => String(value),
    getCheckStatusLabel: (value: unknown) => String(value),
    summarizeCheckResult: (checkResult: Record<string, unknown>) => String(checkResult.summary ?? ''),
    getCheckActions: () => [],
    getCheckTargets: () => [],
    getTargetActions: () => [],
    getTargetProbeSummary: () => '',
    getProbeStatusText: () => '',
    formatDateTime: (value?: string) => (value ? `fmt:${value}` : '未记录'),
    formatLatency: (value?: number) => (value == null ? '-' : `${value}ms`),
    getServiceCheckPresentationResult: (service: AWDTeamServiceData) => service.check_result,
  }
}

describe('AWDServiceStatusPanel', () => {
  it('应按队伍和服务题目渲染运行矩阵单元格', () => {
    const wrapper = mount(AWDServiceStatusPanel, {
      props: buildProps(),
    })

    const headers = wrapper.findAll('thead th').map((cell) => cell.text())
    expect(headers).toContain('队伍节点')
    expect(headers).toContain('Bank Portal')
    expect(headers).toContain('Mail Relay')

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('Blue Team')
    expect(rows[0].text()).toContain('在线')
    expect(rows[0].text()).toContain('离线')
    expect(rows[0].text()).toContain('Checker')
    expect(rows[0].text()).toContain('scheduler')
    expect(rows[0].text()).toContain('healthy')
    expect(rows[0].text()).toContain('fmt:2026-04-18T09:10:00.000Z')
    expect(rows[1].text()).toContain('Red Team')
    expect(rows[1].text()).toContain('失陷')
    expect(rows[1].text()).toContain('N/A')
  })
})

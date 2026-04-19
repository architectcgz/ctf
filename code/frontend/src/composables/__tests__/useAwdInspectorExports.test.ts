import { ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAwdInspectorExports } from '@/composables/useAwdInspectorExports'

const csvMocks = vi.hoisted(() => ({
  downloadJSONFileMock: vi.fn(),
  downloadCSVFileMock: vi.fn(),
}))

vi.mock('@/utils/csv', () => ({
  downloadJSONFile: csvMocks.downloadJSONFileMock,
  downloadCSVFile: csvMocks.downloadCSVFileMock,
}))

describe('useAwdInspectorExports', () => {
  beforeEach(() => {
    csvMocks.downloadJSONFileMock.mockReset()
    csvMocks.downloadCSVFileMock.mockReset()
  })

  it('导出复盘包时应保留服务与攻击日志的 service_id', () => {
    const { exportReviewPackage } = useAwdInspectorExports({
      contest: ref({
        id: 'contest-1',
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
        starts_at: '2026-04-18T09:00:00.000Z',
        ends_at: '2026-04-18T18:00:00.000Z',
      }),
      selectedRound: ref({
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
      }),
      summary: ref(null),
      scoreboardRows: ref([]),
      scoreboardFrozen: ref(false),
      serviceTeamFilter: ref(''),
      serviceStatusFilter: ref('all'),
      serviceCheckSourceFilter: ref(''),
      serviceAlertReasonFilter: ref(''),
      attackTeamFilter: ref(''),
      attackResultFilter: ref('all'),
      attackSourceFilter: ref('all'),
      serviceTeamOptions: ref([]),
      attackTeamOptions: ref([]),
      trafficTeamOptions: ref([]),
      serviceAlerts: ref([]),
      filteredServices: ref([
        {
          id: 'row-1',
          round_id: 'round-1',
          team_id: 'team-1',
          team_name: 'Blue Team',
          service_id: '7009',
          challenge_id: '101',
          service_status: 'up',
          checker_type: 'http_standard',
          check_result: { check_source: 'scheduler' },
          attack_received: 1,
          sla_score: 20,
          defense_score: 40,
          attack_score: 10,
          updated_at: '2026-04-18T09:08:00.000Z',
        },
      ]),
      filteredAttacks: ref([
        {
          id: 'attack-1',
          round_id: 'round-1',
          attacker_team_id: 'team-2',
          attacker_team: 'Red Team',
          victim_team_id: 'team-1',
          victim_team: 'Blue Team',
          service_id: '7009',
          challenge_id: '101',
          attack_type: 'flag_capture',
          source: 'submission',
          submitted_flag: 'flag{demo}',
          is_success: true,
          score_gained: 60,
          created_at: '2026-04-18T09:09:00.000Z',
        },
      ]),
      formatDateTime: (value?: string) => value || '',
      getChallengeTitle: (challengeId: string) => `Challenge ${challengeId}`,
      getSelectedRoundLabel: () => '第 1 轮',
      buildExportFilename: (suffix: string) => `awd-${suffix}.csv`,
      getServiceStatusLabel: (status) => status,
      getAttackTypeLabel: (type) => type,
      getAttackSourceLabel: (source) => source,
      getCheckSourceLabel: (source) => String(source || ''),
      getCheckerTypeLabel: (value) => String(value || ''),
      getServiceAlertLabel: (reason) => reason,
      summarizeCheckResult: () => 'ok',
      getServiceCheckSourceValue: (result) =>
        typeof result.check_source === 'string' ? result.check_source : '',
    })

    exportReviewPackage()

    expect(csvMocks.downloadJSONFileMock).toHaveBeenCalledTimes(1)
    const [, payload] = csvMocks.downloadJSONFileMock.mock.calls[0]
    expect(payload.services[0]?.service_id).toBe('7009')
    expect(payload.attacks[0]?.service_id).toBe('7009')
  })
})

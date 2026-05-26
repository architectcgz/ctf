import { computed, ref } from 'vue'
import { describe, expect, it } from 'vitest'

import type {
  ContestAWDWorkspaceData,
  ContestAWDWorkspaceServiceData,
  ScoreboardRow,
} from '@/api/contracts'
import { formatTime } from '@/utils/format'
import type { AWDRuntimeChallenge } from './awdChallengeIdentity'
import { useAwdWorkspaceSummary } from './useAwdWorkspaceSummary'

function runtimeChallenge(
  serviceId: string,
  challengeId: string,
  title: string
): AWDRuntimeChallenge {
  return {
    id: challengeId,
    challenge_id: `legacy-${challengeId}`,
    awd_challenge_id: challengeId,
    awd_service_id: serviceId,
    title,
    category: 'web',
    difficulty: 'medium',
    points: 100,
    solved_count: 0,
    is_solved: false,
  }
}

function service(
  serviceId: string,
  awdChallengeId: string,
  overrides: Partial<ContestAWDWorkspaceServiceData> = {}
): ContestAWDWorkspaceServiceData {
  return {
    service_id: serviceId,
    awd_challenge_id: awdChallengeId,
    attack_received: 0,
    sla_score: 100,
    defense_score: 100,
    attack_score: 0,
    ...overrides,
  }
}

describe('useAwdWorkspaceSummary', () => {
  it('应生成 HUD 摘要展示文案', () => {
    const workspace = ref<ContestAWDWorkspaceData | null>({
      contest_id: 'contest-1',
      current_round: {
        id: 'round-3',
        contest_id: 'contest-1',
        round_number: 3,
        status: 'finished',
        started_at: '2026-05-26T05:00:00.000Z',
        ended_at: '2026-05-26T05:15:00.000Z',
        attack_score: 120,
        defense_score: 90,
        created_at: '2026-05-26T04:59:00.000Z',
        updated_at: '2026-05-26T05:15:00.000Z',
      },
      my_team: {
        team_id: 'team-2',
        team_name: 'Blue Team',
      },
      services: [
        service('svc-1', 'awd-1'),
        service('svc-2', 'awd-2'),
      ],
      targets: [],
      recent_events: [],
    })
    const scoreboardRows = ref<ScoreboardRow[]>([
      { rank: 1, team_id: 'team-1', team_name: 'Alpha', score: 480, solved_count: 5 },
      { rank: 2, team_id: 'team-2', team_name: 'Blue Team', score: 420, solved_count: 4 },
    ])
    const runtimeChallenges = ref<AWDRuntimeChallenge[]>([])
    const servicesByServiceId = ref(
      new Map<string, ContestAWDWorkspaceServiceData>([
        ['svc-1', service('svc-1', 'awd-1')],
        ['svc-2', service('svc-2', 'awd-2')],
      ])
    )
    const lastSyncedAt = ref('2026-05-26T05:12:00.000Z')

    const state = useAwdWorkspaceSummary({
      workspace: computed(() => workspace.value),
      scoreboardRows: computed(() => scoreboardRows.value),
      runtimeChallenges: computed(() => runtimeChallenges.value),
      servicesByServiceId: computed(() => servicesByServiceId.value),
      lastSyncedAt: computed(() => lastSyncedAt.value),
    })

    expect(state.myTeam.value?.team_name).toBe('Blue Team')
    expect(state.currentRoundLabel.value).toBe('#03')
    expect(state.currentRoundStatusLabel.value).toBe('已结束')
    expect(state.myTeamRank.value).toBe(2)
    expect(state.serviceCount.value).toBe(2)
    expect(state.topScore.value).toBe(480)
    expect(state.lastSyncedLabel.value).toBe(formatTime(lastSyncedAt.value))
  })

  it('应按当前规则生成防守告警', () => {
    const workspace = ref<ContestAWDWorkspaceData | null>({
      contest_id: 'contest-1',
      services: [],
      targets: [],
      recent_events: [],
    })
    const scoreboardRows = ref<ScoreboardRow[]>([])
    const runtimeChallenges = ref<AWDRuntimeChallenge[]>([
      runtimeChallenge('svc-compromised', 'awd-1', 'Bank Portal'),
      runtimeChallenge('svc-down', 'awd-2', 'Patch Relay'),
      runtimeChallenge('svc-stable', 'awd-3', 'Stable API'),
    ])
    const servicesByServiceId = ref(
      new Map<string, ContestAWDWorkspaceServiceData>([
        [
          'svc-compromised',
          service('svc-compromised', 'awd-1', {
            service_status: 'compromised',
            attack_received: 2,
          }),
        ],
        [
          'svc-down',
          service('svc-down', 'awd-2', {
            service_status: 'down',
            instance_status: 'failed',
          }),
        ],
        [
          'svc-stable',
          service('svc-stable', 'awd-3', {
            service_status: 'up',
            attack_received: 0,
          }),
        ],
      ])
    )
    const lastSyncedAt = ref<string | null>(null)

    const state = useAwdWorkspaceSummary({
      workspace: computed(() => workspace.value),
      scoreboardRows: computed(() => scoreboardRows.value),
      runtimeChallenges: computed(() => runtimeChallenges.value),
      servicesByServiceId: computed(() => servicesByServiceId.value),
      lastSyncedAt: computed(() => lastSyncedAt.value),
    })

    expect(state.defenseAlerts.value).toEqual([
      {
        challengeId: 'awd-1',
        challengeTitle: 'Bank Portal',
        statusLabel: '严重',
        tone: 'danger',
        issues: ['已失陷', '检测到 2 次攻击'],
      },
      {
        challengeId: 'awd-2',
        challengeTitle: 'Patch Relay',
        statusLabel: '告警',
        tone: 'warning',
        issues: ['已离线'],
      },
    ])
    expect(state.lastSyncedLabel.value).toBe('未同步')
  })
})

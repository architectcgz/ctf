import { computed, toValue, type MaybeRefOrGetter } from 'vue'

import type {
  ContestAWDWorkspaceData,
  ContestAWDWorkspaceServiceData,
  ScoreboardRow,
} from '@/api/contracts'
import { formatTime } from '@/utils/format'
import type { AWDRuntimeChallenge } from './awdChallengeIdentity'

export interface AWDDefenseAlert {
  challengeId: string
  challengeTitle: string
  statusLabel: string
  tone: 'danger' | 'warning'
  issues: string[]
}

interface UseAwdWorkspaceSummaryOptions {
  workspace: MaybeRefOrGetter<ContestAWDWorkspaceData | null>
  scoreboardRows: MaybeRefOrGetter<ScoreboardRow[]>
  runtimeChallenges: MaybeRefOrGetter<AWDRuntimeChallenge[]>
  servicesByServiceId: MaybeRefOrGetter<Map<string, ContestAWDWorkspaceServiceData>>
  lastSyncedAt: MaybeRefOrGetter<string | null>
}

export function useAwdWorkspaceSummary(options: UseAwdWorkspaceSummaryOptions) {
  const currentRound = computed(() => toValue(options.workspace)?.current_round)
  const myTeam = computed(() => toValue(options.workspace)?.my_team ?? null)

  const currentRoundLabel = computed(() =>
    currentRound.value ? `#${String(currentRound.value.round_number).padStart(2, '0')}` : '--'
  )
  const currentRoundStatusLabel = computed(() => formatRoundStatusLabel(currentRound.value?.status))
  const myTeamRank = computed(
    () =>
      toValue(options.scoreboardRows).find((row) => row.team_id === myTeam.value?.team_id)?.rank || '--'
  )
  const serviceCount = computed(() => toValue(options.workspace)?.services.length || 0)
  const topScore = computed(() => toValue(options.scoreboardRows)[0]?.score ?? 0)
  const lastSyncedLabel = computed(() =>
    toValue(options.lastSyncedAt) ? formatTime(toValue(options.lastSyncedAt) as string) : '未同步'
  )

  const defenseAlerts = computed<AWDDefenseAlert[]>(() => {
    const items: AWDDefenseAlert[] = []

    for (const challenge of toValue(options.runtimeChallenges)) {
      const service = getWorkspaceService(challenge, toValue(options.servicesByServiceId))
      if (!service) continue

      const issues: string[] = []
      let statusLabel = '正常'
      let tone: 'danger' | 'warning' = 'warning'

      if (service.service_status === 'compromised') {
        issues.push('已失陷')
        statusLabel = '严重'
        tone = 'danger'
      } else if (service.service_status === 'down' && service.instance_status !== 'running') {
        issues.push('已离线')
        statusLabel = '告警'
      }

      if ((service.attack_received ?? 0) > 0) {
        issues.push(`检测到 ${service.attack_received} 次攻击`)
      }

      if (issues.length === 0) continue

      items.push({
        challengeId: challenge.awd_challenge_id,
        challengeTitle: challenge.title,
        statusLabel,
        tone,
        issues,
      })
    }

    return items
  })

  return {
    myTeam,
    currentRoundLabel,
    currentRoundStatusLabel,
    myTeamRank,
    serviceCount,
    topScore,
    lastSyncedLabel,
    defenseAlerts,
  }
}

function formatRoundStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'frozen':
      return '已冻结'
    case 'finished':
    case 'completed':
    case 'ended':
      return '已结束'
    default:
      return '等待中'
  }
}

function getWorkspaceService(
  challenge: AWDRuntimeChallenge,
  servicesByServiceId: Map<string, ContestAWDWorkspaceServiceData>
): ContestAWDWorkspaceServiceData | undefined {
  return servicesByServiceId.get(challenge.awd_service_id)
}

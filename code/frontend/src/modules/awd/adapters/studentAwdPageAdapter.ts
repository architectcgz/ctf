import type {
  AWDAttackLogData,
  ContestAWDWorkspaceData,
  ContestChallengeItem,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import type { AwdHeroMetric, AwdTimelineItem, StudentAwdPageKey } from '@/modules/awd/types'
import { formatTime } from '@/utils/format'

interface StudentAwdAdapterInput {
  contest: ContestDetailData | null
  challenges: ContestChallengeItem[]
  workspace: ContestAWDWorkspaceData | null
  scoreboardRows: ScoreboardRow[]
  selectedPage: StudentAwdPageKey
  submitResult?: AWDAttackLogData | null
}

interface StudentDefenseAlert {
  challengeId: string
  challengeTitle: string
  statusLabel: string
  tone: 'danger' | 'warning'
  issues: string[]
}

interface StudentServiceRow {
  challengeId: string
  challengeTitle: string
  category: string
  points: number
  serviceId?: string
  accessUrl?: string
  status: 'up' | 'down' | 'compromised' | 'pending'
  statusLabel: string
  attackReceived: number
  slaScore: number
  defenseScore: number
  attackScore: number
}

interface StudentTargetRow {
  teamId: string
  teamName: string
  challengeId: string
  challengeTitle: string
  serviceId?: string
  accessUrl?: string
}

interface StudentAttackRow {
  id: string
  time: string
  directionLabel: string
  resultLabel: string
  challengeTitle: string
  peerTeamName: string
  scoreGained: number
  serviceRef: string
  title: string
  description: string
}

interface StudentCollabModel {
  priorities: string[]
  notes: string[]
}

export interface StudentAwdPageModel {
  hero: {
    contestTitle: string
    pageTitle: string
    pageDescription: string
    metrics: AwdHeroMetric[]
  }
  overview: {
    scoreboardRows: ScoreboardRow[]
    recentEvents: AwdTimelineItem[]
    defenseAlerts: StudentDefenseAlert[]
  }
  services: {
    items: StudentServiceRow[]
  }
  targets: {
    activeChallengeId: string | null
    challengeOptions: Array<{ key: string; label: string }>
    rows: StudentTargetRow[]
  }
  attacks: {
    recent: StudentAttackRow[]
    submitResultMessage: string
  }
  collab: StudentCollabModel
}

const STUDENT_PAGE_META: Record<StudentAwdPageKey, { title: string; description: string }> = {
  overview: {
    title: '战场总览',
    description: '查看当前轮次、分数变化与最近战场事件。',
  },
  services: {
    title: '我的服务',
    description: '查看本队服务状态、可用地址与防守压力。',
  },
  targets: {
    title: '目标目录',
    description: '按当前攻击题目查看目标队伍与可用攻击面。',
  },
  attacks: {
    title: '攻击记录',
    description: '查看近期攻击提交结果与得分变化。',
  },
  collab: {
    title: '队伍协作',
    description: '整理协作优先级、队伍提醒与战术摘要。',
  },
}

export function buildStudentAwdPageModel(input: StudentAwdAdapterInput): StudentAwdPageModel {
  const pageMeta = STUDENT_PAGE_META[input.selectedPage]
  const runtimeChallenges = input.challenges.filter(
    (item): item is ContestChallengeItem & { awd_service_id: string } => Boolean(item.awd_service_id)
  )
  const challengeById = new Map(input.challenges.map((challenge) => [challenge.challenge_id, challenge]))
  const challengeByServiceId = new Map(
    runtimeChallenges.map((challenge) => [challenge.awd_service_id, challenge] as const)
  )
  const workspaceServices = new Map(
    (input.workspace?.services || [])
      .filter((item): item is NonNullable<typeof item> & { service_id: string } => Boolean(item.service_id))
      .map((item) => [item.service_id, item] as const)
  )

  const activeChallenge = runtimeChallenges[0] || null
  const defenseAlerts = runtimeChallenges.flatMap((challenge) => {
    const service = workspaceServices.get(challenge.awd_service_id)
    if (!service) {
      return []
    }

    const issues: string[] = []
    let statusLabel = '留意'
    let tone: 'danger' | 'warning' = 'warning'

    if (service.service_status === 'compromised') {
      issues.push('本轮服务状态为失陷')
      statusLabel = '失陷'
      tone = 'danger'
    } else if (service.service_status === 'down') {
      issues.push('本轮服务状态为离线')
      statusLabel = '离线'
    }

    if (service.attack_received > 0) {
      issues.push(`本轮收到 ${service.attack_received} 次攻击`)
    }

    if (issues.length === 0) {
      return []
    }

    return [
      {
        challengeId: challenge.challenge_id,
        challengeTitle: challenge.title,
        statusLabel,
        tone,
        issues,
      },
    ]
  })

  const attackRows = (input.workspace?.recent_events || []).map((event) => {
    const challenge =
      (event.service_id ? challengeByServiceId.get(event.service_id) : undefined) ||
      challengeById.get(event.challenge_id)
    const challengeTitle = challenge?.title || event.challenge_id
    const time = formatTime(event.created_at)
    const directionLabel = event.direction === 'attack_out' ? '我方攻击' : '我方被打'
    const resultLabel = event.is_success ? '命中' : '未命中'
    const serviceRef = event.service_id ? `Service #${event.service_id}` : 'Service #--'

    return {
      id: event.id,
      time,
      directionLabel,
      resultLabel,
      challengeTitle,
      peerTeamName: event.peer_team_name,
      scoreGained: event.score_gained,
      serviceRef,
      title: `${directionLabel} ${event.peer_team_name} / ${challengeTitle}`,
      description: `${resultLabel}，得分 ${event.score_gained}，${serviceRef}`,
    }
  })

  const serviceRows = runtimeChallenges.map((challenge) => {
    const service = workspaceServices.get(challenge.awd_service_id)
    const status: StudentServiceRow['status'] = service?.service_status || 'pending'
    return {
      challengeId: challenge.challenge_id,
      challengeTitle: challenge.title,
      category: challenge.category,
      points: challenge.points,
      serviceId: service?.service_id,
      accessUrl: service?.access_url,
      status,
      statusLabel: serviceStatusLabel(status),
      attackReceived: service?.attack_received || 0,
      slaScore: service?.sla_score || 0,
      defenseScore: service?.defense_score || 0,
      attackScore: service?.attack_score || 0,
    }
  })

  const targetRows = !activeChallenge
    ? []
    : (input.workspace?.targets || []).map((target) => {
        const service = target.services.find((item) => item.service_id === activeChallenge.awd_service_id)
        return {
          teamId: target.team_id,
          teamName: target.team_name,
          challengeId: activeChallenge.challenge_id,
          challengeTitle: activeChallenge.title,
          serviceId: service?.service_id,
          accessUrl: service?.access_url,
        }
      })

  const heroMetrics = buildHeroMetrics(input.contest, input.workspace, input.scoreboardRows)
  const submitResultMessage = input.submitResult
    ? input.submitResult.is_success
      ? `${resolveChallengeTitle(input.submitResult, challengeById, challengeByServiceId)} 攻击成功，+${input.submitResult.score_gained} 分`
      : `${resolveChallengeTitle(input.submitResult, challengeById, challengeByServiceId)} 攻击未命中有效 flag`
    : ''

  return {
    hero: {
      contestTitle: input.contest?.title || 'AWD 竞赛',
      pageTitle: pageMeta.title,
      pageDescription: pageMeta.description,
      metrics: heroMetrics,
    },
    overview: {
      scoreboardRows: input.scoreboardRows,
      recentEvents: attackRows.map((item) => ({
        id: item.id,
        time: item.time,
        title: item.title,
        description: item.description,
      })),
      defenseAlerts,
    },
    services: {
      items: serviceRows,
    },
    targets: {
      activeChallengeId: activeChallenge?.challenge_id || null,
      challengeOptions: runtimeChallenges.map((challenge) => ({
        key: challenge.challenge_id,
        label: challenge.title,
      })),
      rows: targetRows,
    },
    attacks: {
      recent: attackRows,
      submitResultMessage,
    },
    collab: buildCollabModel(input.workspace, targetRows, defenseAlerts),
  }
}

function buildHeroMetrics(
  contest: ContestDetailData | null,
  workspace: ContestAWDWorkspaceData | null,
  scoreboardRows: ScoreboardRow[]
): AwdHeroMetric[] {
  const round = workspace?.current_round
  const myTeamName = workspace?.my_team?.team_name || '未加入队伍'
  const scoreRow = workspace?.my_team
    ? scoreboardRows.find((item) => item.team_id === workspace.my_team?.team_id)
    : null

  return [
    {
      label: '当前轮',
      value: round ? `R${round.round_number}` : '待开始',
      helper: round ? `${round.attack_score}/${round.defense_score} 分值` : contest?.status || '',
    },
    {
      label: '我的队伍',
      value: myTeamName,
      helper: scoreRow ? `当前排名 #${scoreRow.rank}` : '尚未进入排行榜',
    },
    {
      label: '当前分数',
      value: String(scoreRow?.score || 0),
      helper: scoreboardRows[0] ? `榜首 ${scoreboardRows[0].score}` : '暂无榜单',
    },
  ]
}

function buildCollabModel(
  workspace: ContestAWDWorkspaceData | null,
  targetRows: StudentTargetRow[],
  defenseAlerts: StudentDefenseAlert[]
): StudentCollabModel {
  const firstTarget = targetRows[0]
  const firstAlert = defenseAlerts[0]
  const priorities: string[] = []
  const notes: string[] = []

  if (firstTarget) {
    priorities.push(`继续跟进 ${firstTarget.teamName} / ${firstTarget.challengeTitle} 的攻击链路`)
    notes.push(firstTarget.accessUrl ? `目标地址：${firstTarget.accessUrl}` : '目标当前没有可用地址')
  }

  if (firstAlert) {
    priorities.push(`优先修复 ${firstAlert.challengeTitle}，当前状态 ${firstAlert.statusLabel}`)
    notes.push(firstAlert.issues.join('；'))
  }

  if (!workspace?.my_team) {
    priorities.push('先加入队伍后再分配攻防职责')
  }

  if (priorities.length === 0) {
    priorities.push('当前没有紧急协作事项，保持轮次同步。')
  }

  if (notes.length === 0) {
    notes.push('建议在每轮结束后同步修补结果、有效 payload 与目标切换时机。')
  }

  return { priorities, notes }
}

function resolveChallengeTitle(
  event: { service_id?: string; challenge_id: string },
  challengeById: Map<string, ContestChallengeItem>,
  challengeByServiceId: Map<string, ContestChallengeItem>
): string {
  if (event.service_id) {
    const byService = challengeByServiceId.get(event.service_id)
    if (byService) {
      return byService.title
    }
  }
  return challengeById.get(event.challenge_id)?.title || event.challenge_id
}

function serviceStatusLabel(status: StudentServiceRow['status']): string {
  switch (status) {
    case 'up':
      return '在线'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return '待启动'
  }
}

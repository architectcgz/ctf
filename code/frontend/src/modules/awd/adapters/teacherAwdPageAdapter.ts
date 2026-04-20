import type { TeacherAWDReviewArchiveData } from '@/api/contracts'
import type { AwdHeroMetric, AwdTimelineItem, TeacherAwdPageKey } from '@/modules/awd/types'
import { formatTime } from '@/utils/format'

interface TeacherAwdAdapterInput {
  review: TeacherAWDReviewArchiveData | null
  selectedPage: TeacherAwdPageKey
}

interface TeacherOverviewCard {
  label: string
  value: string
  helper: string
}

interface TeacherTeamRow {
  teamId: string
  teamName: string
  totalScore: number
  memberCount: number
  lastSolveAt: string
  serviceIssueCount: number
  attackCount: number
}

interface TeacherServiceCard {
  challengeId: string
  challengeTitle: string
  teamCount: number
  healthyCount: number
  degradedCount: number
  attackReceived: number
  attackScore: number
  updatedAt: string
}

interface TeacherRoundRow {
  id: string
  roundNumber: number
  statusLabel: string
  attackScore: number
  defenseScore: number
  serviceCount: number
  attackCount: number
  trafficCount: number
}

export interface TeacherAwdPageModel {
  hero: {
    contestTitle: string
    pageTitle: string
    pageDescription: string
    metrics: AwdHeroMetric[]
  }
  overview: {
    cards: TeacherOverviewCard[]
    roundRows: TeacherRoundRow[]
    insights: string[]
  }
  teams: {
    rows: TeacherTeamRow[]
  }
  services: {
    cards: TeacherServiceCard[]
  }
  replay: {
    timeline: AwdTimelineItem[]
    roundRows: TeacherRoundRow[]
  }
  export: {
    canExportReport: boolean
    snapshotLabel: string
    cards: Array<{ title: string; description: string; stateLabel: string }>
  }
}

const TEACHER_PAGE_META: Record<TeacherAwdPageKey, { title: string; description: string }> = {
  overview: {
    title: '教学总览',
    description: '查看整场复盘摘要、轮次走势与教学提示。',
  },
  teams: {
    title: '队伍复盘',
    description: '按队伍查看得分、问题分布与最近活动。',
  },
  services: {
    title: 'Service 复盘',
    description: '按 Service 聚合稳定性、被打次数与得分表现。',
  },
  replay: {
    title: '轮次回放',
    description: '通过轮次与事件时间线回放关键攻防瞬间。',
  },
  export: {
    title: '报告导出',
    description: '统一管理复盘归档、教师报告与当前快照的导出状态。',
  },
}

export function buildTeacherAwdPageModel(input: TeacherAwdAdapterInput): TeacherAwdPageModel {
  const pageMeta = TEACHER_PAGE_META[input.selectedPage]
  const selectedRound = input.review?.selected_round
  const serviceCards = buildServiceCards(input.review)
  const teamRows = buildTeamRows(input.review)
  const roundRows = (input.review?.rounds || []).map((round) => ({
    id: round.id,
    roundNumber: round.round_number,
    statusLabel: teacherRoundStatusLabel(round.status),
    attackScore: round.attack_score,
    defenseScore: round.defense_score,
    serviceCount: round.service_count,
    attackCount: round.attack_count,
    trafficCount: round.traffic_count,
  }))
  const timeline = buildTimeline(input.review)

  return {
    hero: {
      contestTitle: input.review?.contest.title || '教师 AWD 复盘',
      pageTitle: pageMeta.title,
      pageDescription: pageMeta.description,
      metrics: buildHeroMetrics(input.review),
    },
    overview: {
      cards: [
        {
          label: '轮次规模',
          value: String(input.review?.overview?.round_count || input.review?.contest.round_count || 0),
          helper: `${input.review?.overview?.team_count || input.review?.contest.team_count || 0} 支队伍`,
        },
        {
          label: '服务 / 攻击 / 流量',
          value: `${input.review?.overview?.service_count || 0} / ${input.review?.overview?.attack_count || 0} / ${input.review?.overview?.traffic_count || 0}`,
          helper: selectedRound ? `当前聚焦第 ${selectedRound.round.round_number} 轮` : '当前为整场视角',
        },
        {
          label: '导出状态',
          value: input.review?.contest.export_ready ? '可导出' : '仅实时快照',
          helper: input.review?.scope.snapshot_type === 'final' ? '赛后快照' : '实时快照',
        },
      ],
      roundRows,
      insights: buildInsights(input.review, teamRows, serviceCards),
    },
    teams: {
      rows: teamRows,
    },
    services: {
      cards: serviceCards,
    },
    replay: {
      timeline,
      roundRows,
    },
    export: {
      canExportReport: Boolean(input.review?.contest.export_ready),
      snapshotLabel: input.review?.scope.snapshot_type === 'final' ? '赛后快照' : '实时快照',
      cards: [
        {
          title: '复盘归档',
          description: '导出整场 AWD 复盘证据归档，便于教学留存和复查。',
          stateLabel: input.review ? '随时可发起' : '等待数据',
        },
        {
          title: '教师报告',
          description: '导出适合课程总结与讲评的教师复盘报告。',
          stateLabel: input.review?.contest.export_ready ? '已开放' : '比赛结束后开放',
        },
        {
          title: '当前视图快照',
          description: selectedRound
            ? `当前聚焦第 ${selectedRound.round.round_number} 轮，可保留单轮讲解素材。`
            : '当前为整场总览，可作为教学总览快照。',
          stateLabel: input.review?.scope.snapshot_type === 'final' ? '已固化' : '实时更新中',
        },
      ],
    },
  }
}

function buildHeroMetrics(review: TeacherAWDReviewArchiveData | null): AwdHeroMetric[] {
  return [
    {
      label: '当前视图',
      value: review?.selected_round ? `第 ${review.selected_round.round.round_number} 轮` : '整场总览',
      helper: review?.scope.snapshot_type === 'final' ? '赛后快照' : '实时快照',
    },
    {
      label: '队伍数',
      value: String(review?.contest.team_count || 0),
      helper: `${review?.contest.round_count || 0} 轮复盘可用`,
    },
    {
      label: '报告状态',
      value: review?.contest.export_ready ? '可导出' : '待开放',
      helper: review?.contest.latest_evidence_at ? `最新信号 ${formatTime(review.contest.latest_evidence_at)}` : '暂无最新信号',
    },
  ]
}

function buildTeamRows(review: TeacherAWDReviewArchiveData | null): TeacherTeamRow[] {
  const selectedRound = review?.selected_round
  if (!selectedRound) {
    return []
  }

  return selectedRound.teams.map((team) => {
    const teamServices = selectedRound.services.filter((item) => item.team_id === team.team_id)
    const teamAttacks = selectedRound.attacks.filter(
      (item) => item.attacker_team_id === team.team_id || item.victim_team_id === team.team_id
    )

    return {
      teamId: team.team_id,
      teamName: team.team_name,
      totalScore: team.total_score,
      memberCount: team.member_count,
      lastSolveAt: team.last_solve_at ? formatTime(team.last_solve_at) : '暂无',
      serviceIssueCount: teamServices.filter((item) => item.service_status !== 'up').length,
      attackCount: teamAttacks.length,
    }
  })
}

function buildServiceCards(review: TeacherAWDReviewArchiveData | null): TeacherServiceCard[] {
  const selectedRound = review?.selected_round
  if (!selectedRound) {
    return []
  }

  const groups = new Map<
    string,
    {
      challengeTitle: string
      teamIds: Set<string>
      healthyCount: number
      degradedCount: number
      attackReceived: number
      attackScore: number
      updatedAt: string
    }
  >()

  selectedRound.services.forEach((item) => {
    const current = groups.get(item.challenge_id) || {
      challengeTitle: item.challenge_title,
      teamIds: new Set<string>(),
      healthyCount: 0,
      degradedCount: 0,
      attackReceived: 0,
      attackScore: 0,
      updatedAt: item.updated_at,
    }
    current.teamIds.add(item.team_id)
    if (item.service_status === 'up') {
      current.healthyCount += 1
    } else {
      current.degradedCount += 1
    }
    current.attackReceived += item.attack_received
    current.attackScore += item.attack_score
    if (item.updated_at > current.updatedAt) {
      current.updatedAt = item.updated_at
    }
    groups.set(item.challenge_id, current)
  })

  return [...groups.entries()].map(([challengeId, value]) => ({
    challengeId,
    challengeTitle: value.challengeTitle,
    teamCount: value.teamIds.size,
    healthyCount: value.healthyCount,
    degradedCount: value.degradedCount,
    attackReceived: value.attackReceived,
    attackScore: value.attackScore,
    updatedAt: formatTime(value.updatedAt),
  }))
}

function buildInsights(
  review: TeacherAWDReviewArchiveData | null,
  teamRows: TeacherTeamRow[],
  serviceCards: TeacherServiceCard[]
): string[] {
  const insights: string[] = []
  const leadingTeam = [...teamRows].sort((left, right) => right.totalScore - left.totalScore)[0]
  const unstableService = [...serviceCards].sort((left, right) => right.degradedCount - left.degradedCount)[0]

  if (leadingTeam) {
    insights.push(`${leadingTeam.teamName} 当前以 ${leadingTeam.totalScore} 分领先，可作为课堂案例。`)
  }
  if (unstableService && unstableService.degradedCount > 0) {
    insights.push(
      `${unstableService.challengeTitle} 出现 ${unstableService.degradedCount} 个异常实例，适合讲解恢复流程。`
    )
  }
  if (review?.scope.snapshot_type === 'live') {
    insights.push('当前仍是实时快照，课堂讲评时建议结合轮次切换观察趋势。')
  }
  if (insights.length === 0) {
    insights.push('当前复盘数据较少，可先从轮次回放页查看更多线索。')
  }

  return insights
}

function buildTimeline(review: TeacherAWDReviewArchiveData | null): AwdTimelineItem[] {
  if (!review) {
    return []
  }

  const roundItems = review.rounds.map((round) => ({
    id: `round-${round.id}`,
    time: round.ended_at ? formatTime(round.ended_at) : formatTime(round.started_at || review.generated_at),
    title: `第 ${round.round_number} 轮 ${teacherRoundStatusLabel(round.status)}`,
    description: `服务 ${round.service_count} · 攻击 ${round.attack_count} · 流量 ${round.traffic_count}`,
  }))

  const attackItems = (review.selected_round?.attacks || []).slice(0, 6).map((attack) => ({
    id: `attack-${attack.id}`,
    time: formatTime(attack.created_at),
    title: `${attack.attacker_team_name} -> ${attack.victim_team_name}`,
    description: `${attack.challenge_title} ${attack.is_success ? '命中' : '未命中'}，${attack.score_gained} 分`,
  }))

  const trafficItems = (review.selected_round?.traffic || []).slice(0, 6).map((traffic) => ({
    id: `traffic-${traffic.id}`,
    time: formatTime(traffic.created_at),
    title: `${traffic.attacker_team_name} 请求 ${traffic.path}`,
    description: `${traffic.challenge_title} · 状态码 ${traffic.status_code}`,
  }))

  return [...roundItems, ...attackItems, ...trafficItems]
    .sort((left, right) => String(right.time).localeCompare(String(left.time)))
    .slice(0, 12)
}

function teacherRoundStatusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'finished':
      return '已完成'
    default:
      return status || '未知'
  }
}

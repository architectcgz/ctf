import type {
  AWDAttackLogData,
  AWDReadinessData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import type { AdminAwdPageKey, AwdHeroMetric, AwdTimelineItem } from '@/modules/awd/types'
import { formatTime } from '@/utils/format'

interface AdminAwdAdapterInput {
  contest: ContestDetailData | null
  rounds: AWDRoundData[]
  summary: AWDRoundSummaryData | null
  readiness: AWDReadinessData | null
  services: AWDTeamServiceData[]
  attacks: AWDAttackLogData[]
  trafficSummary: AWDTrafficSummaryData | null
  trafficEvents: AWDTrafficEventData[]
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeData[]
  scoreboardRows: ScoreboardRow[]
  selectedPage: AdminAwdPageKey
}

interface AdminSummaryRow {
  teamId: string
  teamName: string
  serviceHealthLabel: string
  totalScore: number
  attackScore: number
  defenseScore: number
}

interface AdminServiceMatrixRow {
  id: string
  teamName: string
  challengeTitle: string
  serviceLabel: string
  status: AWDTeamServiceData['service_status']
  statusLabel: string
  attackReceived: number
  slaScore: number
  defenseScore: number
  attackScore: number
  updatedAt: string
}

interface AdminAttackRow {
  id: string
  time: string
  attackerTeam: string
  victimTeam: string
  challengeTitle: string
  resultLabel: string
  scoreLabel: string
  sourceLabel: string
}

interface AdminTrafficEventRow {
  id: string
  time: string
  challengeTitle: string
  routeLabel: string
  attackerTeam: string
  victimTeam: string
  statusLabel: string
}

interface AdminAlertItem {
  id: string
  severity: 'critical' | 'warning' | 'info'
  title: string
  description: string
  time?: string
}

interface AdminInstanceRow {
  teamId: string
  teamName: string
  totalServices: number
  upCount: number
  downCount: number
  compromisedCount: number
  latestUpdatedAt: string
}

interface AdminRoundRow {
  id: string
  roundNumber: number
  statusLabel: string
  scoreLabel: string
  serviceCount: number
  attackCount: number
  updatedAt: string
}

interface AdminOverviewCard {
  label: string
  value: string
  helper: string
}

export interface AdminAwdPageModel {
  hero: {
    contestTitle: string
    pageTitle: string
    pageDescription: string
    metrics: AwdHeroMetric[]
  }
  overview: {
    cards: AdminOverviewCard[]
    scoreboardRows: ScoreboardRow[]
    roundRows: AdminRoundRow[]
    summaryRows: AdminSummaryRow[]
  }
  readiness: {
    ready: boolean
    blockingCount: number
    actions: string[]
    items: Array<{
      id: string
      title: string
      statusLabel: string
      reasonLabel: string
      accessUrl: string
      updatedAt: string
    }>
  }
  rounds: {
    rows: AdminRoundRow[]
    summaryRows: AdminSummaryRow[]
  }
  services: {
    items: AdminServiceMatrixRow[]
  }
  attacks: {
    recent: AdminAttackRow[]
  }
  traffic: {
    headline: AdminOverviewCard[]
    topChallenges: Array<{ challengeTitle: string; requestCount: number; errorCount: number }>
    topPaths: Array<{ path: string; requestCount: number; errorCount: number; lastStatusCode: number }>
    events: AdminTrafficEventRow[]
  }
  alerts: {
    items: AdminAlertItem[]
  }
  instances: {
    rows: AdminInstanceRow[]
  }
  replay: {
    timeline: AwdTimelineItem[]
    rounds: AdminRoundRow[]
  }
}

const ADMIN_PAGE_META: Record<AdminAwdPageKey, { title: string; description: string }> = {
  overview: {
    title: 'AWD 总览',
    description: '汇总赛事健康度、轮次状态、真实榜单与关键信号。',
  },
  readiness: {
    title: 'Readiness',
    description: '查看赛前检查结果、阻塞项与 checker 验证状态。',
  },
  rounds: {
    title: '轮次控制台',
    description: '浏览轮次节奏、当前轮摘要与分数结构。',
  },
  services: {
    title: '服务矩阵',
    description: '按队伍与服务查看运行状态、被打次数和得分表现。',
  },
  attacks: {
    title: '攻击日志',
    description: '查看近期攻击流水、命中情况与来源。',
  },
  traffic: {
    title: '流量态势',
    description: '聚合热点路径、错误流量与最近请求事件。',
  },
  alerts: {
    title: '告警中心',
    description: '集中展示 readiness、服务异常与高风险流量告警。',
  },
  instances: {
    title: '实例运维',
    description: '按队伍聚合运行实例健康状态，便于快速排障。',
  },
  replay: {
    title: '赛后复盘与导出',
    description: '通过轮次与事件时间线回看整场 AWD 过程。',
  },
}

export function buildAdminAwdPageModel(input: AdminAwdAdapterInput): AdminAwdPageModel {
  const pageMeta = ADMIN_PAGE_META[input.selectedPage]
  const challengeMeta = new Map(
    input.challengeLinks.map((item) => [
      item.challenge_id,
      {
        title: item.title || item.awd_service_display_name || item.challenge_id,
        serviceLabel: item.awd_service_display_name || item.title || item.challenge_id,
      },
    ])
  )

  const roundRows = input.rounds.map((round) => ({
    id: round.id,
    roundNumber: round.round_number,
    statusLabel: roundStatusLabel(round.status),
    scoreLabel: `${round.attack_score}/${round.defense_score}`,
    serviceCount: input.summary?.round.id === round.id ? input.summary.metrics?.total_service_count || 0 : 0,
    attackCount: input.summary?.round.id === round.id ? input.summary.metrics?.total_attack_count || 0 : 0,
    updatedAt: formatTime(round.updated_at),
  }))

  const summaryRows: AdminSummaryRow[] = (input.summary?.items || []).map((item) => ({
    teamId: item.team_id,
    teamName: item.team_name,
    serviceHealthLabel: `${item.service_up_count} 正常 / ${item.service_down_count} 离线 / ${item.service_compromised_count} 失陷`,
    totalScore: item.total_score,
    attackScore: item.attack_score,
    defenseScore: item.defense_score,
  }))

  const serviceRows: AdminServiceMatrixRow[] = input.services.map((item) => {
    const meta = challengeMeta.get(item.challenge_id)
    const statusLabel = serviceStatusLabel(item.service_status)
    return {
      id: item.id,
      teamName: item.team_name,
      challengeTitle: meta?.title || item.challenge_id,
      serviceLabel: meta?.serviceLabel || item.service_id || item.challenge_id,
      status: item.service_status,
      statusLabel,
      attackReceived: item.attack_received,
      slaScore: item.sla_score,
      defenseScore: item.defense_score,
      attackScore: item.attack_score,
      updatedAt: formatTime(item.updated_at),
    }
  })

  const attackRows: AdminAttackRow[] = input.attacks.map((item) => ({
    id: item.id,
    time: formatTime(item.created_at),
    attackerTeam: item.attacker_team,
    victimTeam: item.victim_team,
    challengeTitle: challengeMeta.get(item.challenge_id)?.title || item.challenge_id,
    resultLabel: item.is_success ? '命中' : '未命中',
    scoreLabel: `${item.score_gained} 分`,
    sourceLabel: item.source,
  }))

  const trafficRows: AdminTrafficEventRow[] = input.trafficEvents.map((item) => ({
    id: item.id,
    time: formatTime(item.occurred_at),
    challengeTitle: item.challenge_title || challengeMeta.get(item.challenge_id)?.title || item.challenge_id,
    routeLabel: `${item.method} ${item.path}`,
    attackerTeam: item.attacker_team_name || item.attacker_team_id,
    victimTeam: item.victim_team_name || item.victim_team_id,
    statusLabel: `${item.status_code} / ${item.status_group}`,
  }))

  const instanceRows: AdminInstanceRow[] = input.teams.map((team) => {
    const teamServices = input.services.filter((item) => item.team_id === team.id)
    const latestUpdatedAt =
      teamServices
        .map((item) => item.updated_at)
        .sort()
        .at(-1) || team.created_at

    return {
      teamId: team.id,
      teamName: team.name,
      totalServices: teamServices.length,
      upCount: teamServices.filter((item) => item.service_status === 'up').length,
      downCount: teamServices.filter((item) => item.service_status === 'down').length,
      compromisedCount: teamServices.filter((item) => item.service_status === 'compromised').length,
      latestUpdatedAt: formatTime(latestUpdatedAt),
    }
  })

  const readinessItems = (input.readiness?.items || []).map((item) => ({
    id: item.challenge_id,
    title: item.title,
    statusLabel: readinessStatusLabel(item.validation_state),
    reasonLabel: readinessReasonLabel(item.blocking_reason),
    accessUrl: item.last_access_url || '--',
    updatedAt: item.last_preview_at ? formatTime(item.last_preview_at) : '暂无',
  }))

  const overviewCards: AdminOverviewCard[] = [
    {
      label: '轮次状态',
      value: input.summary?.round ? `第 ${input.summary.round.round_number} 轮` : '未开始',
      helper: input.summary?.round ? roundStatusLabel(input.summary.round.status) : '暂无轮次',
    },
    {
      label: '服务健康',
      value: `${input.summary?.metrics?.service_up_count || 0}/${input.summary?.metrics?.total_service_count || 0}`,
      helper: `离线 ${input.summary?.metrics?.service_down_count || 0} · 失陷 ${input.summary?.metrics?.service_compromised_count || 0}`,
    },
    {
      label: '攻击成功',
      value: String(input.summary?.metrics?.successful_attack_count || 0),
      helper: `总攻击 ${input.summary?.metrics?.total_attack_count || 0}`,
    },
    {
      label: '流量错误',
      value: String(input.trafficSummary?.error_request_count || 0),
      helper: input.trafficSummary?.latest_event_at
        ? `最新 ${formatTime(input.trafficSummary.latest_event_at)}`
        : '暂无流量数据',
    },
  ]

  const alertItems = buildAlertItems(input, challengeMeta)
  const replayTimeline = buildReplayTimeline(input, challengeMeta)

  return {
    hero: {
      contestTitle: input.contest?.title || 'AWD 管理工作台',
      pageTitle: pageMeta.title,
      pageDescription: pageMeta.description,
      metrics: buildHeroMetrics(input),
    },
    overview: {
      cards: overviewCards,
      scoreboardRows: input.scoreboardRows,
      roundRows,
      summaryRows,
    },
    readiness: {
      ready: input.readiness?.ready ?? false,
      blockingCount: input.readiness?.blocking_count || 0,
      actions: (input.readiness?.blocking_actions || []).map(readinessActionLabel),
      items: readinessItems,
    },
    rounds: {
      rows: roundRows,
      summaryRows,
    },
    services: {
      items: serviceRows,
    },
    attacks: {
      recent: attackRows,
    },
    traffic: {
      headline: [
        {
          label: '总请求量',
          value: String(input.trafficSummary?.total_request_count || 0),
          helper: `错误 ${input.trafficSummary?.error_request_count || 0}`,
        },
        {
          label: '活跃攻击队伍',
          value: String(input.trafficSummary?.active_attacker_team_count || 0),
          helper: `受害队伍 ${input.trafficSummary?.victim_team_count || 0}`,
        },
        {
          label: '唯一路径',
          value: String(input.trafficSummary?.unique_path_count || 0),
          helper: input.trafficSummary?.latest_event_at
            ? `最新 ${formatTime(input.trafficSummary.latest_event_at)}`
            : '暂无流量摘要',
        },
      ],
      topChallenges: (input.trafficSummary?.top_challenges || []).map((item) => ({
        challengeTitle: item.challenge_title || challengeMeta.get(item.challenge_id)?.title || item.challenge_id,
        requestCount: item.request_count,
        errorCount: item.error_count,
      })),
      topPaths: (input.trafficSummary?.top_paths || []).map((item) => ({
        path: item.path,
        requestCount: item.request_count,
        errorCount: item.error_count,
        lastStatusCode: item.last_status_code,
      })),
      events: trafficRows,
    },
    alerts: {
      items: alertItems,
    },
    instances: {
      rows: instanceRows,
    },
    replay: {
      timeline: replayTimeline,
      rounds: roundRows,
    },
  }
}

function buildHeroMetrics(input: AdminAwdAdapterInput): AwdHeroMetric[] {
  const currentRound = input.summary?.round || input.rounds.at(-1) || null
  return [
    {
      label: '当前轮',
      value: currentRound ? `R${currentRound.round_number}` : '待创建',
      helper: currentRound ? roundStatusLabel(currentRound.status) : '暂无轮次',
    },
    {
      label: '阻塞项',
      value: String(input.readiness?.blocking_count || 0),
      helper: `Readiness ${input.readiness?.ready ? '通过' : '待处理'}`,
    },
    {
      label: '队伍规模',
      value: String(input.teams.length),
      helper: `${input.scoreboardRows.length} 支队伍进入真实榜单`,
    },
  ]
}

function buildAlertItems(
  input: AdminAwdAdapterInput,
  challengeMeta: Map<string, { title: string; serviceLabel: string }>
): AdminAlertItem[] {
  const readinessAlerts = (input.readiness?.items || [])
    .filter((item) => item.validation_state === 'failed' || item.validation_state === 'stale')
    .map((item) => ({
      id: `readiness-${item.challenge_id}`,
      severity: item.validation_state === 'failed' ? ('critical' as const) : ('warning' as const),
      title: `${item.title} Readiness 异常`,
      description: readinessReasonLabel(item.blocking_reason),
      time: item.last_preview_at ? formatTime(item.last_preview_at) : undefined,
    }))

  const serviceAlerts = input.services
    .filter((item) => item.service_status === 'down' || item.service_status === 'compromised')
    .map((item) => ({
      id: `service-${item.id}`,
      severity: item.service_status === 'compromised' ? ('critical' as const) : ('warning' as const),
      title: `${item.team_name} / ${challengeMeta.get(item.challenge_id)?.title || item.challenge_id}`,
      description: `${serviceStatusLabel(item.service_status)}，本轮被打 ${item.attack_received} 次`,
      time: formatTime(item.updated_at),
    }))

  const trafficAlerts = (input.trafficSummary?.top_error_paths || []).slice(0, 3).map((item, index) => ({
    id: `traffic-${index}`,
    severity: 'info' as const,
    title: `高错误路径 ${item.path}`,
    description: `错误 ${item.error_count} / 请求 ${item.request_count}，最近状态码 ${item.last_status_code}`,
  }))

  return [...readinessAlerts, ...serviceAlerts, ...trafficAlerts].slice(0, 10)
}

function buildReplayTimeline(
  input: AdminAwdAdapterInput,
  challengeMeta: Map<string, { title: string; serviceLabel: string }>
): AwdTimelineItem[] {
  const roundItems = input.rounds.map((round) => ({
    id: `round-${round.id}`,
    time: formatTime(round.updated_at),
    title: `第 ${round.round_number} 轮 ${roundStatusLabel(round.status)}`,
    description: `攻击/防守分值 ${round.attack_score}/${round.defense_score}`,
  }))
  const attackItems = input.attacks.slice(0, 6).map((attack) => ({
    id: `attack-${attack.id}`,
    time: formatTime(attack.created_at),
    title: `${attack.attacker_team} -> ${attack.victim_team}`,
    description: `${challengeMeta.get(attack.challenge_id)?.title || attack.challenge_id} ${attack.is_success ? '命中' : '未命中'}，${attack.score_gained} 分`,
  }))
  const trafficItems = input.trafficEvents.slice(0, 6).map((event) => ({
    id: `traffic-${event.id}`,
    time: formatTime(event.occurred_at),
    title: `${event.attacker_team_name || event.attacker_team_id} 请求 ${event.path}`,
    description: `${event.challenge_title || challengeMeta.get(event.challenge_id)?.title || event.challenge_id} · ${event.status_code}`,
  }))

  return [...roundItems, ...attackItems, ...trafficItems]
    .sort((left, right) => String(right.time).localeCompare(String(left.time)))
    .slice(0, 12)
}

function roundStatusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'finished':
      return '已完成'
    case 'pending':
      return '待开始'
    default:
      return status || '未知'
  }
}

function serviceStatusLabel(status: AWDTeamServiceData['service_status']): string {
  switch (status) {
    case 'up':
      return '正常'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return status || '未知'
  }
}

function readinessStatusLabel(status: string): string {
  switch (status) {
    case 'passed':
      return '已通过'
    case 'failed':
      return '失败'
    case 'stale':
      return '过期'
    case 'pending':
      return '待验证'
    default:
      return status || '未知'
  }
}

function readinessReasonLabel(reason: string): string {
  switch (reason) {
    case 'missing_checker':
      return '缺少 checker 配置'
    case 'last_preview_failed':
      return '最近一次预检失败'
    case 'validation_stale':
      return '预检结果已过期'
    case 'preview_pending':
      return '尚未执行预检'
    default:
      return reason || '待确认'
  }
}

function readinessActionLabel(action: string): string {
  switch (action) {
    case 'create_round':
      return '允许创建轮次'
    case 'run_current_round_check':
      return '允许立即巡检'
    default:
      return action
  }
}

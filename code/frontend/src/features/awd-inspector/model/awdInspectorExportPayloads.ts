import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTeamServiceData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'

interface BuildServiceExportRowsOptions {
  contestTitle: string
  selectedRoundLabel: string
  serviceTeamFilter: string
  serviceStatusFilter: 'all' | AWDTeamServiceData['service_status']
  serviceCheckSourceFilter: string
  serviceAlertReasonFilter: string
  serviceTeamOptions: Array<{ team_id: string; team_name: string }>
  rows: AWDTeamServiceData[]
  getChallengeTitle: (challengeId: string) => string
  getServiceStatusLabel: (status: AWDTeamServiceData['service_status']) => string
  getCheckSourceLabel: (source: unknown) => string
  getCheckerTypeLabel: (value: unknown) => string
  summarizeCheckResult: (checkResult: Record<string, unknown>) => string
  getServiceAlertLabel: (reason: string) => string
  formatDateTime: (value?: string) => string
}

export function buildServiceExportRows(options: BuildServiceExportRowsOptions) {
  const {
    contestTitle,
    selectedRoundLabel,
    serviceTeamFilter,
    serviceStatusFilter,
    serviceCheckSourceFilter,
    serviceAlertReasonFilter,
    serviceTeamOptions,
    rows,
    getChallengeTitle,
    getServiceStatusLabel,
    getCheckSourceLabel,
    getCheckerTypeLabel,
    summarizeCheckResult,
    getServiceAlertLabel,
    formatDateTime,
  } = options

  return rows.map((item) => {
    const checkResult = {
      checker_type: item.checker_type,
      ...item.check_result,
    }

    return {
      赛事: contestTitle,
      轮次: selectedRoundLabel,
      筛选队伍: serviceTeamFilter
        ? serviceTeamOptions.find((team) => team.team_id === serviceTeamFilter)?.team_name ||
          serviceTeamFilter
        : '全部队伍',
      筛选状态:
        serviceStatusFilter === 'all' ? '全部状态' : getServiceStatusLabel(serviceStatusFilter),
      筛选来源: serviceCheckSourceFilter
        ? getCheckSourceLabel(serviceCheckSourceFilter) || serviceCheckSourceFilter
        : '全部来源',
      筛选告警: serviceAlertReasonFilter
        ? getServiceAlertLabel(serviceAlertReasonFilter)
        : '全部告警',
      队伍: item.team_name,
      靶题: getChallengeTitle(item.awd_challenge_id),
      服务编号: item.service_id || '',
      服务状态: getServiceStatusLabel(item.service_status),
      巡检来源: getCheckSourceLabel(item.check_result.check_source) || '',
      Checker类型: getCheckerTypeLabel(checkResult.checker_type) || '',
      检查摘要: summarizeCheckResult(checkResult),
      SLA得分: item.sla_score ?? 0,
      防守得分: item.defense_score,
      攻击得分: item.attack_score,
      受攻击次数: item.attack_received,
      更新时间: formatDateTime(item.updated_at),
    }
  })
}

interface BuildAttackExportRowsOptions {
  contestTitle: string
  selectedRoundLabel: string
  attackTeamFilter: string
  attackResultFilter: 'all' | 'success' | 'failed'
  attackSourceFilter: 'all' | AWDAttackLogData['source']
  attackTeamOptions: Array<{ id: string; name: string }>
  rows: AWDAttackLogData[]
  getChallengeTitle: (challengeId: string) => string
  getAttackTypeLabel: (type: AWDAttackLogData['attack_type']) => string
  getAttackSourceLabel: (source: AWDAttackLogData['source']) => string
  formatDateTime: (value?: string) => string
}

export function buildAttackExportRows(options: BuildAttackExportRowsOptions) {
  const {
    contestTitle,
    selectedRoundLabel,
    attackTeamFilter,
    attackResultFilter,
    attackSourceFilter,
    attackTeamOptions,
    rows,
    getChallengeTitle,
    getAttackTypeLabel,
    getAttackSourceLabel,
    formatDateTime,
  } = options

  return rows.map((item) => ({
    赛事: contestTitle,
    轮次: selectedRoundLabel,
    筛选队伍: attackTeamFilter
      ? attackTeamOptions.find((team) => team.id === attackTeamFilter)?.name || attackTeamFilter
      : '全部队伍',
    筛选结果:
      attackResultFilter === 'all'
        ? '全部结果'
        : attackResultFilter === 'success'
          ? '仅成功'
          : '仅失败',
    筛选来源:
      attackSourceFilter === 'all' ? '全部来源' : getAttackSourceLabel(attackSourceFilter),
    时间: formatDateTime(item.created_at),
    攻击方: item.attacker_team,
    受害方: item.victim_team,
    靶题: getChallengeTitle(item.awd_challenge_id),
    服务编号: item.service_id || '',
    攻击类型: getAttackTypeLabel(item.attack_type),
    记录来源: getAttackSourceLabel(item.source),
    攻击结果: item.is_success ? '成功' : '失败',
    得分: item.score_gained,
    提交Flag: item.submitted_flag || '',
  }))
}

interface BuildReviewPackagePayloadOptions {
  contest: ContestDetailData
  selectedRound: AWDRoundData
  summary: AWDRoundSummaryData | null
  scoreboardRows: ScoreboardRow[]
  scoreboardFrozen: boolean
  serviceTeamFilter: string
  serviceStatusFilter: 'all' | AWDTeamServiceData['service_status']
  serviceCheckSourceFilter: string
  serviceAlertReasonFilter: string
  attackTeamFilter: string
  attackResultFilter: 'all' | 'success' | 'failed'
  attackSourceFilter: 'all' | AWDAttackLogData['source']
  serviceTeamOptions: Array<{ team_id: string; team_name: string }>
  attackTeamOptions: Array<{ id: string; name: string }>
  serviceAlerts: Array<{
    key: string
    label: string
    count: number
    affected_teams: string[]
    samples: Array<{
      service_id: string
      team_name: string
      awd_challenge_title: string
    }>
  }>
  filteredServices: AWDTeamServiceData[]
  filteredAttacks: AWDAttackLogData[]
  getChallengeTitle: (challengeId: string) => string
  getServiceStatusLabel: (status: AWDTeamServiceData['service_status']) => string
  getCheckSourceLabel: (source: unknown) => string
  getAttackTypeLabel: (type: AWDAttackLogData['attack_type']) => string
  getAttackSourceLabel: (source: AWDAttackLogData['source']) => string
  getServiceAlertLabel: (reason: string) => string
  getServiceCheckSourceValue: (result: Record<string, unknown>) => string
}

export function buildReviewPackagePayload(options: BuildReviewPackagePayloadOptions) {
  const {
    contest,
    selectedRound,
    summary,
    scoreboardRows,
    scoreboardFrozen,
    serviceTeamFilter,
    serviceStatusFilter,
    serviceCheckSourceFilter,
    serviceAlertReasonFilter,
    attackTeamFilter,
    attackResultFilter,
    attackSourceFilter,
    serviceTeamOptions,
    attackTeamOptions,
    serviceAlerts,
    filteredServices,
    filteredAttacks,
    getChallengeTitle,
    getServiceStatusLabel,
    getCheckSourceLabel,
    getAttackTypeLabel,
    getAttackSourceLabel,
    getServiceAlertLabel,
    getServiceCheckSourceValue,
  } = options

  return {
    exported_at: new Date().toISOString(),
    contest: {
      id: contest.id,
      title: contest.title,
      status: contest.status,
      mode: contest.mode,
    },
    round: {
      id: selectedRound.id,
      round_number: selectedRound.round_number,
      status: selectedRound.status,
      attack_score: selectedRound.attack_score,
      defense_score: selectedRound.defense_score,
      started_at: selectedRound.started_at || null,
      ended_at: selectedRound.ended_at || null,
      updated_at: selectedRound.updated_at,
    },
    filters: {
      service: {
        team_id: serviceTeamFilter || null,
        team_name:
          serviceTeamOptions.find((team) => team.team_id === serviceTeamFilter)?.team_name || null,
        status: serviceStatusFilter,
        check_source: serviceCheckSourceFilter || null,
        check_source_label: serviceCheckSourceFilter
          ? getCheckSourceLabel(serviceCheckSourceFilter) || serviceCheckSourceFilter
          : '全部来源',
        alert_reason: serviceAlertReasonFilter || null,
        alert_reason_label: serviceAlertReasonFilter
          ? getServiceAlertLabel(serviceAlertReasonFilter)
          : '全部告警',
      },
      attack: {
        team_id: attackTeamFilter || null,
        team_name: attackTeamOptions.find((team) => team.id === attackTeamFilter)?.name || null,
        result: attackResultFilter,
        source: attackSourceFilter,
        source_label:
          attackSourceFilter === 'all' ? '全部来源' : getAttackSourceLabel(attackSourceFilter),
      },
    },
    summary: {
      round: summary?.round || null,
      metrics: summary?.metrics || null,
      items: summary?.items || [],
      service_alerts: serviceAlerts.map((alert) => ({
        key: alert.key,
        label: alert.label,
        count: alert.count,
        affected_teams: alert.affected_teams,
        samples: alert.samples,
      })),
    },
    scoreboard: {
      frozen: scoreboardFrozen,
      rows: scoreboardRows,
    },
    services: filteredServices.map((item) => ({
      id: item.id,
      team_id: item.team_id,
      team_name: item.team_name,
      service_id: item.service_id || null,
      awd_challenge_id: item.awd_challenge_id,
      awd_challenge_title: getChallengeTitle(item.awd_challenge_id),
      service_status: item.service_status,
      service_status_label: getServiceStatusLabel(item.service_status),
      checker_type: item.checker_type,
      check_source: getServiceCheckSourceValue(item.check_result) || null,
      check_source_label: getCheckSourceLabel(item.check_result.check_source) || '',
      check_result: item.check_result,
      sla_score: item.sla_score ?? 0,
      defense_score: item.defense_score,
      attack_received: item.attack_received,
      attack_score: item.attack_score,
      updated_at: item.updated_at,
    })),
    attacks: filteredAttacks.map((item) => ({
      id: item.id,
      attacker_team_id: item.attacker_team_id,
      attacker_team: item.attacker_team,
      victim_team_id: item.victim_team_id,
      victim_team: item.victim_team,
      service_id: item.service_id || null,
      awd_challenge_id: item.awd_challenge_id,
      awd_challenge_title: getChallengeTitle(item.awd_challenge_id),
      attack_type: item.attack_type,
      attack_type_label: getAttackTypeLabel(item.attack_type),
      source: item.source,
      source_label: getAttackSourceLabel(item.source),
      is_success: item.is_success,
      score_gained: item.score_gained,
      submitted_flag: item.submitted_flag || null,
      created_at: item.created_at,
    })),
  }
}

export function getAwdTrafficSourceLabel(source: string): string {
  if (source === 'proxy_audit' || source === 'runtime_proxy') {
    return '平台代理'
  }
  return source || '未标记'
}

import type { Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTeamServiceData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { downloadCSVFile, downloadJSONFile } from '@/utils/csv'

interface AWDServiceAlertView {
  key: string
  label: string
  count: number
  affected_teams: string[]
  samples: Array<{
    service_id: string
    team_name: string
    challenge_title: string
  }>
}

interface UseAwdInspectorExportsOptions {
  contest: Ref<ContestDetailData>
  selectedRound: Ref<AWDRoundData | null>
  summary: Ref<AWDRoundSummaryData | null>
  scoreboardRows: Ref<ScoreboardRow[]>
  scoreboardFrozen: Ref<boolean>
  serviceTeamFilter: Ref<string>
  serviceStatusFilter: Ref<'all' | AWDTeamServiceData['service_status']>
  serviceCheckSourceFilter: Ref<string>
  serviceAlertReasonFilter: Ref<string>
  attackTeamFilter: Ref<string>
  attackResultFilter: Ref<'all' | 'success' | 'failed'>
  attackSourceFilter: Ref<'all' | AWDAttackLogData['source']>
  serviceTeamOptions: Ref<Array<{ team_id: string; team_name: string }>>
  attackTeamOptions: Ref<Array<{ id: string; name: string }>>
  trafficTeamOptions: Ref<Array<{ id: string; name: string }>>
  serviceAlerts: Ref<AWDServiceAlertView[]>
  filteredServices: Ref<AWDTeamServiceData[]>
  filteredAttacks: Ref<AWDAttackLogData[]>
  formatDateTime: (value?: string) => string
  getChallengeTitle: (challengeId: string) => string
  getSelectedRoundLabel: () => string
  buildExportFilename: (suffix: string) => string
  getServiceStatusLabel: (status: AWDTeamServiceData['service_status']) => string
  getAttackTypeLabel: (type: AWDAttackLogData['attack_type']) => string
  getAttackSourceLabel: (source: AWDAttackLogData['source']) => string
  getCheckSourceLabel: (source: unknown) => string
  getCheckerTypeLabel: (value: unknown) => string
  getServiceAlertLabel: (reason: string) => string
  summarizeCheckResult: (checkResult: Record<string, unknown>) => string
  getServiceCheckSourceValue: (result: Record<string, unknown>) => string
}

export function useAwdInspectorExports({
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
  trafficTeamOptions,
  serviceAlerts,
  filteredServices,
  filteredAttacks,
  formatDateTime,
  getChallengeTitle,
  getSelectedRoundLabel,
  buildExportFilename,
  getServiceStatusLabel,
  getAttackTypeLabel,
  getAttackSourceLabel,
  getCheckSourceLabel,
  getCheckerTypeLabel,
  getServiceAlertLabel,
  summarizeCheckResult,
  getServiceCheckSourceValue,
}: UseAwdInspectorExportsOptions) {
  function getTrafficTeamName(teamId: string, teamName?: string): string {
    if (teamName && teamName.trim() !== '') {
      return teamName
    }
    return trafficTeamOptions.value.find((item) => item.id === teamId)?.name || `Team #${teamId}`
  }

  function getTrafficChallengeTitle(challengeId: string, fallbackTitle?: string): string {
    if (fallbackTitle && fallbackTitle.trim() !== '') {
      return fallbackTitle
    }
    return getChallengeTitle(challengeId)
  }

  function getTrafficSourceLabel(source: string): string {
    if (source === 'proxy_audit' || source === 'runtime_proxy') {
      return '平台代理'
    }
    return source || '未标记'
  }

  function exportFilteredServices() {
    if (filteredServices.value.length === 0) {
      return
    }
    downloadCSVFile(
      buildExportFilename('services'),
      filteredServices.value.map((item) => {
        const checkResult = {
          checker_type: item.checker_type,
          ...item.check_result,
        }

        return {
          赛事: contest.value.title,
          轮次: getSelectedRoundLabel(),
          筛选队伍: serviceTeamFilter.value
            ? serviceTeamOptions.value.find((team) => team.team_id === serviceTeamFilter.value)
                ?.team_name || serviceTeamFilter.value
            : '全部队伍',
          筛选状态:
            serviceStatusFilter.value === 'all'
              ? '全部状态'
              : getServiceStatusLabel(serviceStatusFilter.value),
          筛选来源: serviceCheckSourceFilter.value
            ? getCheckSourceLabel(serviceCheckSourceFilter.value) || serviceCheckSourceFilter.value
            : '全部来源',
          筛选告警: serviceAlertReasonFilter.value
            ? getServiceAlertLabel(serviceAlertReasonFilter.value)
            : '全部告警',
          队伍: item.team_name,
          靶题: getChallengeTitle(item.challenge_id),
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
    )
  }

  function exportFilteredAttacks() {
    if (filteredAttacks.value.length === 0) {
      return
    }
    downloadCSVFile(
      buildExportFilename('attacks'),
      filteredAttacks.value.map((item) => ({
        赛事: contest.value.title,
        轮次: getSelectedRoundLabel(),
        筛选队伍: attackTeamFilter.value
          ? attackTeamOptions.value.find((team) => team.id === attackTeamFilter.value)?.name ||
            attackTeamFilter.value
          : '全部队伍',
        筛选结果:
          attackResultFilter.value === 'all'
            ? '全部结果'
            : attackResultFilter.value === 'success'
              ? '仅成功'
              : '仅失败',
        筛选来源:
          attackSourceFilter.value === 'all'
            ? '全部来源'
            : getAttackSourceLabel(attackSourceFilter.value),
        时间: formatDateTime(item.created_at),
        攻击方: item.attacker_team,
        受害方: item.victim_team,
        靶题: getChallengeTitle(item.challenge_id),
        攻击类型: getAttackTypeLabel(item.attack_type),
        记录来源: getAttackSourceLabel(item.source),
        攻击结果: item.is_success ? '成功' : '失败',
        得分: item.score_gained,
        提交Flag: item.submitted_flag || '',
      }))
    )
  }

  function exportReviewPackage() {
    if (!selectedRound.value) {
      return
    }

    downloadJSONFile(buildExportFilename('review-package').replace(/\.csv$/, '.json'), {
      exported_at: new Date().toISOString(),
      contest: {
        id: contest.value.id,
        title: contest.value.title,
        status: contest.value.status,
        mode: contest.value.mode,
      },
      round: {
        id: selectedRound.value.id,
        round_number: selectedRound.value.round_number,
        status: selectedRound.value.status,
        attack_score: selectedRound.value.attack_score,
        defense_score: selectedRound.value.defense_score,
        started_at: selectedRound.value.started_at || null,
        ended_at: selectedRound.value.ended_at || null,
        updated_at: selectedRound.value.updated_at,
      },
      filters: {
        service: {
          team_id: serviceTeamFilter.value || null,
          team_name:
            serviceTeamOptions.value.find((team) => team.team_id === serviceTeamFilter.value)
              ?.team_name || null,
          status: serviceStatusFilter.value,
          check_source: serviceCheckSourceFilter.value || null,
          check_source_label: serviceCheckSourceFilter.value
            ? getCheckSourceLabel(serviceCheckSourceFilter.value) || serviceCheckSourceFilter.value
            : '全部来源',
          alert_reason: serviceAlertReasonFilter.value || null,
          alert_reason_label: serviceAlertReasonFilter.value
            ? getServiceAlertLabel(serviceAlertReasonFilter.value)
            : '全部告警',
        },
        attack: {
          team_id: attackTeamFilter.value || null,
          team_name:
            attackTeamOptions.value.find((team) => team.id === attackTeamFilter.value)?.name ||
            null,
          result: attackResultFilter.value,
          source: attackSourceFilter.value,
          source_label:
            attackSourceFilter.value === 'all'
              ? '全部来源'
              : getAttackSourceLabel(attackSourceFilter.value),
        },
      },
      summary: {
        round: summary.value?.round || null,
        metrics: summary.value?.metrics || null,
        items: summary.value?.items || [],
        service_alerts: serviceAlerts.value.map((alert) => ({
          key: alert.key,
          label: alert.label,
          count: alert.count,
          affected_teams: alert.affected_teams,
          samples: alert.samples,
        })),
      },
      scoreboard: {
        frozen: scoreboardFrozen.value,
        rows: scoreboardRows.value,
      },
      services: filteredServices.value.map((item) => ({
        id: item.id,
        team_id: item.team_id,
        team_name: item.team_name,
        challenge_id: item.challenge_id,
        challenge_title: getChallengeTitle(item.challenge_id),
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
      attacks: filteredAttacks.value.map((item) => ({
        id: item.id,
        attacker_team_id: item.attacker_team_id,
        attacker_team: item.attacker_team,
        victim_team_id: item.victim_team_id,
        victim_team: item.victim_team,
        challenge_id: item.challenge_id,
        challenge_title: getChallengeTitle(item.challenge_id),
        attack_type: item.attack_type,
        attack_type_label: getAttackTypeLabel(item.attack_type),
        source: item.source,
        source_label: getAttackSourceLabel(item.source),
        is_success: item.is_success,
        score_gained: item.score_gained,
        submitted_flag: item.submitted_flag || null,
        created_at: item.created_at,
      })),
    })
  }

  return {
    getTrafficTeamName,
    getTrafficChallengeTitle,
    getTrafficSourceLabel,
    exportFilteredServices,
    exportFilteredAttacks,
    exportReviewPackage,
  }
}

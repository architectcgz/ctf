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
import {
  buildAttackExportRows,
  buildReviewPackagePayload,
  buildServiceExportRows,
  getAwdTrafficSourceLabel,
} from './awdInspectorExportPayloads'

interface AWDServiceAlertView {
  key: string
  label: string
  count: number
  affected_teams: string[]
  samples: Array<{
    service_id: string
    team_name: string
    awd_challenge_title: string
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
    return getAwdTrafficSourceLabel(source)
  }

  function exportFilteredServices() {
    if (filteredServices.value.length === 0) {
      return
    }
    downloadCSVFile(
      buildExportFilename('services'),
      buildServiceExportRows({
        contestTitle: contest.value.title,
        selectedRoundLabel: getSelectedRoundLabel(),
        serviceTeamFilter: serviceTeamFilter.value,
        serviceStatusFilter: serviceStatusFilter.value,
        serviceCheckSourceFilter: serviceCheckSourceFilter.value,
        serviceAlertReasonFilter: serviceAlertReasonFilter.value,
        serviceTeamOptions: serviceTeamOptions.value,
        rows: filteredServices.value,
        getChallengeTitle,
        getServiceStatusLabel,
        getCheckSourceLabel,
        getCheckerTypeLabel,
        summarizeCheckResult,
        getServiceAlertLabel,
        formatDateTime,
      })
    )
  }

  function exportFilteredAttacks() {
    if (filteredAttacks.value.length === 0) {
      return
    }
    downloadCSVFile(
      buildExportFilename('attacks'),
      buildAttackExportRows({
        contestTitle: contest.value.title,
        selectedRoundLabel: getSelectedRoundLabel(),
        attackTeamFilter: attackTeamFilter.value,
        attackResultFilter: attackResultFilter.value,
        attackSourceFilter: attackSourceFilter.value,
        attackTeamOptions: attackTeamOptions.value,
        rows: filteredAttacks.value,
        getChallengeTitle,
        getAttackTypeLabel,
        getAttackSourceLabel,
        formatDateTime,
      })
    )
  }

  function exportReviewPackage() {
    if (!selectedRound.value) {
      return
    }

    downloadJSONFile(
      buildExportFilename('review-package').replace(/\.csv$/, '.json'),
      buildReviewPackagePayload({
        contest: contest.value,
        selectedRound: selectedRound.value,
        summary: summary.value,
        scoreboardRows: scoreboardRows.value,
        scoreboardFrozen: scoreboardFrozen.value,
        serviceTeamFilter: serviceTeamFilter.value,
        serviceStatusFilter: serviceStatusFilter.value,
        serviceCheckSourceFilter: serviceCheckSourceFilter.value,
        serviceAlertReasonFilter: serviceAlertReasonFilter.value,
        attackTeamFilter: attackTeamFilter.value,
        attackResultFilter: attackResultFilter.value,
        attackSourceFilter: attackSourceFilter.value,
        serviceTeamOptions: serviceTeamOptions.value,
        attackTeamOptions: attackTeamOptions.value,
        serviceAlerts: serviceAlerts.value,
        filteredServices: filteredServices.value,
        filteredAttacks: filteredAttacks.value,
        getChallengeTitle,
        getServiceStatusLabel,
        getCheckSourceLabel,
        getAttackTypeLabel,
        getAttackSourceLabel,
        getServiceAlertLabel,
        getServiceCheckSourceValue,
      })
    )
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

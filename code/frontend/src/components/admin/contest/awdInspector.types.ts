import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficStatusGroup,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import type { AWDTrafficFilters } from '@/composables/useAwdTrafficPanel'

export interface AWDRoundInspectorProps {
  contest: ContestDetailData
  rounds: AWDRoundData[]
  selectedRoundId: string | null
  services: AWDTeamServiceData[]
  attacks: AWDAttackLogData[]
  challengeLinks: AdminContestChallengeData[]
  summary: AWDRoundSummaryData | null
  trafficSummary: AWDTrafficSummaryData | null
  trafficEvents: AWDTrafficEventData[]
  trafficEventsTotal: number
  trafficFilters: AWDTrafficFilters
  scoreboardRows: ScoreboardRow[]
  scoreboardFrozen: boolean
  loadingRounds: boolean
  loadingRoundDetail: boolean
  loadingTrafficSummary: boolean
  loadingTrafficEvents: boolean
  checking: boolean
  shouldAutoRefresh: boolean
  canRecordServiceChecks: boolean
  canRecordAttackLogs: boolean
  serviceCheckHint?: string
  attackLogHint?: string
}

export type AWDRoundInspectorEmits = {
  refresh: []
  openCreateRoundDialog: []
  openServiceCheckDialog: []
  openAttackLogDialog: []
  runSelectedRoundCheck: []
  applyTrafficFilters: [filters: Partial<AWDTrafficFilters>]
  changeTrafficPage: [page: number]
  resetTrafficFilters: []
  'update:selectedRoundId': [roundId: string]
}

export interface AWDServiceAlertView {
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

export interface AWDCheckActionView {
  key: 'put_flag' | 'get_flag' | 'havoc'
  label: string
  healthy: boolean
  error_code?: string
  error?: string
  method?: string
  path?: string
}

export interface AWDCheckAttemptView {
  probe: string
  healthy: boolean
  error_code?: string
  error?: string
  latency_ms?: number
}

export interface AWDCheckTargetView {
  access_url?: string
  healthy: boolean
  error_code?: string
  error?: string
  probe?: string
  latency_ms?: number
  attempts: AWDCheckAttemptView[]
  actions: AWDCheckActionView[]
}

export interface AWDServiceStatusPanelProps {
  services: AWDTeamServiceData[]
  filteredServices: AWDTeamServiceData[]
  serviceAlerts: AWDServiceAlertView[]
  serviceTeamOptions: Array<{ team_id: string; team_name: string }>
  serviceCheckSourceOptions: string[]
  serviceTeamFilter: string
  serviceStatusFilter: 'all' | AWDTeamServiceData['service_status']
  serviceCheckSourceFilter: string
  serviceAlertReasonFilter: string
  getChallengeTitle: (challengeId: string) => string
  getServiceStatusLabel: (status: AWDTeamServiceData['service_status']) => string
  getServiceStatusClass: (status: AWDTeamServiceData['service_status']) => string
  getCheckSourceLabel: (source: unknown) => string
  summarizeCheckResult: (checkResult: Record<string, unknown>) => string
  getCheckActions: (checkResult: Record<string, unknown>) => AWDCheckActionView[]
  getCheckTargets: (checkResult: Record<string, unknown>) => AWDCheckTargetView[]
  getTargetActions: (target: Record<string, unknown> | AWDCheckTargetView) => AWDCheckActionView[]
  getTargetProbeSummary: (checkResult: Record<string, unknown>) => string
  getProbeStatusText: (healthy: boolean, errorCode?: string, error?: string) => string
  formatLatency: (value?: number) => string
  getServiceCheckPresentationResult: (service: AWDTeamServiceData) => Record<string, unknown>
}

export type AWDServiceStatusPanelEmits = {
  updateServiceTeamFilter: [value: string]
  updateServiceStatusFilter: [value: 'all' | AWDTeamServiceData['service_status']]
  updateServiceCheckSourceFilter: [value: string]
  updateServiceAlertReasonFilter: [value: string]
  exportServices: []
}

export interface AWDAttackLogPanelProps {
  attacks: AWDAttackLogData[]
  filteredAttacks: AWDAttackLogData[]
  attackTeamOptions: Array<{ id: string; name: string }>
  attackSourceOptions: AWDAttackLogData['source'][]
  attackTeamFilter: string
  attackResultFilter: 'all' | 'success' | 'failed'
  attackSourceFilter: 'all' | AWDAttackLogData['source']
  formatDateTime: (value?: string) => string
  getChallengeTitle: (challengeId: string) => string
  getAttackTypeLabel: (type: AWDAttackLogData['attack_type']) => string
  getAttackSourceLabel: (source: AWDAttackLogData['source']) => string
}

export type AWDAttackLogPanelEmits = {
  updateAttackTeamFilter: [value: string]
  updateAttackResultFilter: [value: 'all' | 'success' | 'failed']
  updateAttackSourceFilter: [value: 'all' | AWDAttackLogData['source']]
  exportAttacks: []
}

export interface AWDTrafficPanelProps {
  updatedAt?: string
  challengeLinks: AdminContestChallengeData[]
  trafficSummary: AWDTrafficSummaryData | null
  trafficEvents: AWDTrafficEventData[]
  trafficEventsTotal: number
  trafficFilters: AWDTrafficFilters
  trafficTeamOptions: Array<{ id: string; name: string }>
  loadingTrafficSummary: boolean
  loadingTrafficEvents: boolean
  formatDateTime: (value?: string) => string
  formatPercent: (value: number) => string
  getTrafficStatusGroupLabel: (statusGroup: AWDTrafficStatusGroup) => string
  getTrafficStatusGroupClass: (statusGroup: AWDTrafficStatusGroup) => string
  getTrafficTeamName: (teamId: string, teamName?: string) => string
  getTrafficChallengeTitle: (challengeId: string, fallbackTitle?: string) => string
  getTrafficSourceLabel: (source: string) => string
}

export type AWDTrafficPanelEmits = {
  applyTrafficFilters: [filters: Partial<AWDTrafficFilters>]
  changeTrafficPage: [page: number]
  resetTrafficFilters: []
}

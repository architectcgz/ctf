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

import type { AdminContestChallengeViewData } from '@/api/contracts'

export interface AWDChallengeConfigSummaryItem {
  key: string
  label: string
  value: string
  hint: string
}

export interface AWDChallengeConfigDirectoryItemView {
  source: AdminContestChallengeViewData
  challengeId: string
  title: string
  category?: string | null
  order: number
  checkerTypeLabel: string
  slaScore?: number
  defenseScore?: number
  configSummary: string
  validationState?: string
  validationStateText: string
  validationPrimaryHint: string
  previewRoute: {
    name: string
    params: {
      id: string
    }
  }
}

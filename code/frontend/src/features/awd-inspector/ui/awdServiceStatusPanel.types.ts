import type { AWDTeamServiceData } from '@/api/contracts'

export interface AWDServiceStatusTeamOptionView {
  teamId: string
  teamName: string
}

export interface AWDServiceStatusFilterOptionView {
  value: string
  label: string
}

export interface AWDServiceStatusChallengeColumnView {
  challengeId: string
  label: string
}

export interface AWDServiceStatusCellView {
  key: string
  status: AWDTeamServiceData['service_status'] | null
  statusClass: string
  statusLabel: string
  checkerLabel: string
  sourceLabel: string
  reasonLabel: string
  checkedAtLabel: string
}

export interface AWDServiceStatusRowView {
  teamId: string
  teamName: string
  cells: AWDServiceStatusCellView[]
}

export interface AWDInstanceOrchestrationHeaderView {
  runningCount: number
  totalTargetCount: number
  loading: boolean
  hasPendingStart: boolean
  canStartAll: boolean
}

export interface AWDInstanceOrchestrationServiceView {
  serviceId: string
  displayName: string
}

export interface AWDInstanceOrchestrationCellView {
  teamId: string
  serviceId: string
  status?: string
  statusLabel: string
  accessUrl?: string
  isStarting: boolean
}

export interface AWDInstanceOrchestrationRowView {
  teamId: string
  teamName: string
  captainId: string
  hasMissingService: boolean
  cells: AWDInstanceOrchestrationCellView[]
}

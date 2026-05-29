import type {
  AWDAttackLogData,
  AWDReadinessData,
  AWDRoundData,
  AWDTeamServiceData,
} from '@/api/contracts'

export interface AwdCreateRoundPayload {
  round_number: number
  status: AWDRoundData['status']
  attack_score: number
  defense_score: number
}

export interface AwdCreateServiceCheckPayload {
  team_id: number
  service_id: number
  service_status: AWDTeamServiceData['service_status']
  check_result?: Record<string, unknown>
}

export interface AwdCreateAttackLogPayload {
  attacker_team_id: number
  victim_team_id: number
  service_id: number
  attack_type: AWDAttackLogData['attack_type']
  submitted_flag?: string
  is_success: boolean
}

export interface AwdOperationsOverrideDialogState {
  open: boolean
  title: string
  readiness: AWDReadinessData | null
  confirmLoading: boolean
}

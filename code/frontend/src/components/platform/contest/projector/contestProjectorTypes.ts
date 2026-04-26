import type { AWDTeamServiceData } from '@/api/contracts'

export interface ContestProjectorServiceMatrixRow {
  team_id: string
  team_name: string
  services: AWDTeamServiceData[]
}

export interface ContestProjectorAttackLeader {
  team_id: string
  team_name: string
  success: number
  score: number
}

export interface ContestProjectorTrafficTrendBar {
  bucket_start_at: string
  request_count: number
  error_count: number
  height: string
  errorHeight: string
}

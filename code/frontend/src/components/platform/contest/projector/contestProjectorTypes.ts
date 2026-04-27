import type { AWDTeamServiceData } from '@/api/contracts'

export type ContestProjectorFocusPanel = 'leaderboard' | 'attack-map' | 'services' | 'traffic' | 'events'

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

export interface ContestProjectorAttackEdge {
  id: string
  attacker_team_id: string
  attacker_team: string
  victim_team_id: string
  victim_team: string
  success: number
  failed: number
  total: number
  score: number
  latest_at: string
  latest_service_label: string
  successRate: number
  reciprocalSuccess: number
}

export interface ContestProjectorTrafficTrendBar {
  bucket_start_at: string
  request_count: number
  error_count: number
  height: string
  errorHeight: string
}

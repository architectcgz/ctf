import type { ContestStatus } from '@/api/contracts'

export type ContestManageStatusFilter =
  | 'all'
  | Extract<ContestStatus, 'draft' | 'registering' | 'running' | 'frozen' | 'ended'>

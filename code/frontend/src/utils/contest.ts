import type { ContestStatus, ContestMode } from '@/api/contracts'

export function isStudentVisibleContestStatus(status: ContestStatus): boolean {
  return status !== 'draft'
}

export function getStatusLabel(status: ContestStatus): string {
  const labels: Record<ContestStatus, string> = {
    draft: '草稿',
    published: '已发布',
    registering: '报名中',
    running: '进行中',
    frozen: '已冻结',
    ended: '已结束',
    cancelled: '已取消',
    archived: '已归档',
  }
  return labels[status] || status
}

export function getModeLabel(mode: ContestMode): string {
  const labels: Record<ContestMode, string> = {
    jeopardy: 'Jeopardy',
    awd: 'AWD',
    awd_plus: 'AWD+',
    king_of_hill: 'King of the Hill',
  }
  return labels[mode] || mode
}

export function getStatusBadgeClass(status: ContestStatus): string {
  if (status === 'running') return 'contest-status-badge--running'
  if (status === 'registering') return 'contest-status-badge--registering'
  return 'contest-status-badge--neutral'
}

export function getContestAccentColor(status: ContestStatus): string {
  if (status === 'running') return 'var(--color-primary)'
  if (status === 'registering') return 'var(--color-warning)'
  if (status === 'draft' || status === 'published') {
    return 'color-mix(in srgb, var(--color-text-secondary) 45%, var(--color-primary))'
  }
  return 'color-mix(in srgb, var(--color-text-muted) 76%, var(--color-border-default))'
}

export function getContestActionLabel(status: ContestStatus): string {
  if (status === 'running') return '进入竞赛'
  if (status === 'registering') return '立即报名'
  return '查看详情'
}

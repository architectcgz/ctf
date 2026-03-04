import type { ContestStatus, ContestMode } from '@/api/contracts'

export function getStatusLabel(status: ContestStatus): string {
  const labels: Record<ContestStatus, string> = {
    draft: '草稿',
    published: '已发布',
    registering: '报名中',
    running: '进行中',
    frozen: '已冻结',
    ended: '已结束',
    cancelled: '已取消',
    archived: '已归档'
  }
  return labels[status] || status
}

export function getModeLabel(mode: ContestMode): string {
  const labels: Record<ContestMode, string> = {
    jeopardy: 'Jeopardy',
    awd: 'AWD',
    awd_plus: 'AWD Plus',
    king_of_hill: 'King of Hill'
  }
  return labels[mode] || mode
}

export function getStatusBadgeClass(status: ContestStatus): string {
  if (status === 'running') return 'bg-[var(--color-primary)]/10 text-[#06b6d4]'
  if (status === 'registering') return 'bg-[#f59e0b]/10 text-[#f59e0b]'
  return 'bg-[#30363d] text-[var(--color-text-secondary)]'
}

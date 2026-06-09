import type { AWDAttackLogData, AWDTeamServiceData } from '@/api/contracts'

export function getContestStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'frozen':
      return '已冻结'
    case 'ended':
      return '已结束'
    default:
      return '待同步'
  }
}

export function formatProjectorScore(value: number): string {
  return new Intl.NumberFormat('zh-CN', {
    maximumFractionDigits: 2,
  }).format(value)
}

export function formatProjectorTime(value?: string): string {
  if (!value) return '--'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function getRoundStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '运行中'
    case 'finished':
      return '已结束'
    case 'pending':
      return '待开始'
    default:
      return '未同步'
  }
}

export function getServiceStatusLabel(status: AWDTeamServiceData['service_status']): string {
  switch (status) {
    case 'up':
      return 'UP'
    case 'down':
      return 'DOWN'
    case 'compromised':
      return 'PWN'
    default:
      return status
  }
}

export function getAttackTypeLabel(type: AWDAttackLogData['attack_type']): string {
  switch (type) {
    case 'flag_capture':
      return 'Flag'
    case 'service_exploit':
      return 'Exploit'
    default:
      return type
  }
}

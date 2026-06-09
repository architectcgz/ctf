import type { AdminContestAWDServiceData, AWDCheckerType } from '@/api/contracts'

export function formatAwdCheckDateTime(value?: string): string {
  if (!value) return '未记录'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function getAwdCheckerTypeLabel(value?: AWDCheckerType): string {
  switch (value) {
    case 'http_standard':
      return 'HTTP 标准 Checker'
    case 'tcp_standard':
      return 'TCP 标准 Checker'
    case 'script_checker':
      return '脚本 Checker'
    case 'legacy_probe':
      return '基础探活'
    default:
      return '未声明 Checker'
  }
}

export function getAwdProtocolLabel(value?: AWDCheckerType): string {
  switch (value) {
    case 'http_standard':
      return 'Web HTTP'
    case 'tcp_standard':
      return 'Binary TCP'
    case 'script_checker':
      return '题目包脚本'
    case 'legacy_probe':
      return '基础探活'
    default:
      return '题目包未声明'
  }
}

export function getAwdValidationLabel(
  value?: AdminContestAWDServiceData['validation_state']
): string {
  switch (value) {
    case 'passed':
      return '已通过'
    case 'failed':
      return '未通过'
    case 'stale':
      return '待重验'
    case 'pending':
    default:
      return '待验证'
  }
}

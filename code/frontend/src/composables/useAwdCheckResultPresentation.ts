import type { AWDCheckerType } from '@/api/contracts'

interface AWDProbeAttemptView {
  probe: string
  healthy: boolean
  latency_ms?: number
  error_code?: string
  error?: string
}

interface AWDCheckerActionView {
  key: 'put_flag' | 'get_flag' | 'havoc'
  label: string
  healthy: boolean
  method?: string
  path?: string
  error_code?: string
  error?: string
}

interface AWDProbeTargetView {
  access_url?: string
  healthy: boolean
  probe?: string
  latency_ms?: number
  error_code?: string
  error?: string
  attempts: AWDProbeAttemptView[]
  actions: AWDCheckerActionView[]
}

interface UseAwdCheckResultPresentationOptions {
  formatDateTime: (value?: string) => string
}

export function useAwdCheckResultPresentation({
  formatDateTime,
}: UseAwdCheckResultPresentationOptions) {
  const checkerActionOptions: Array<{
    key: AWDCheckerActionView['key']
    label: AWDCheckerActionView['label']
  }> = [
    { key: 'put_flag', label: 'PUT Flag' },
    { key: 'get_flag', label: 'GET Flag' },
    { key: 'havoc', label: 'Havoc' },
  ]

  function getCheckSourceLabel(value: unknown): string {
    switch (value) {
      case 'checker_preview':
        return '配置试跑'
      case 'manual_current_round':
        return '当前轮手动巡检'
      case 'manual_selected_round':
        return '指定轮次重跑'
      case 'manual_service_check':
        return '人工补录'
      case 'scheduler':
        return '调度巡检'
      default:
        return ''
    }
  }

  function getCheckerTypeLabel(value: unknown): string {
    switch (value as AWDCheckerType | undefined) {
      case 'legacy_probe':
        return '基础探活'
      case 'http_standard':
        return 'HTTP 标准 Checker'
      default:
        return ''
    }
  }

  function getValidationStateLabel(value: unknown): string {
    switch (value) {
      case 'pending':
        return '未验证'
      case 'passed':
        return '最近通过'
      case 'failed':
        return '最近失败'
      case 'stale':
        return '待重新验证'
      default:
        return ''
    }
  }

  function getCheckStatusLabel(value: unknown): string {
    if (typeof value !== 'string' || value.trim() === '') {
      return ''
    }
    const labels: Record<string, string> = {
      healthy: '全部正常',
      partial_available: '部分可用',
      no_running_instances: '无运行实例',
      unexpected_http_status: 'HTTP 状态异常',
      http_request_failed: 'HTTP 请求失败',
      http_response_read_failed: 'HTTP 响应读取失败',
      invalid_access_url: '访问地址无效',
      all_probes_failed: '巡检失败',
      flag_mismatch: 'Flag 校验失败',
      invalid_checker_config: 'Checker 配置无效',
      flag_unavailable: '轮次 Flag 不可用',
      checker_action_not_configured: 'Checker 动作未配置',
      service_compromised: '服务已失陷',
      service_down: '服务下线',
    }
    return labels[value] || value
  }

  function readPreviewPassSummary(result: Record<string, unknown>): string {
    const passCount =
      typeof result.preview_pass_count === 'number' ? result.preview_pass_count : undefined
    const totalCount =
      typeof result.preview_total_count === 'number' ? result.preview_total_count : undefined
    if (
      typeof passCount === 'number' &&
      typeof totalCount === 'number' &&
      Number.isFinite(passCount) &&
      Number.isFinite(totalCount) &&
      totalCount > 0
    ) {
      return `${passCount}/${totalCount} 通过`
    }
    return ''
  }

  function readCheckerAction(
    key: AWDCheckerActionView['key'],
    label: AWDCheckerActionView['label'],
    value: unknown
  ): AWDCheckerActionView | null {
    if (!value || typeof value !== 'object') {
      return null
    }
    const item = value as Record<string, unknown>
    return {
      key,
      label,
      healthy: item.healthy === true,
      method: typeof item.method === 'string' ? item.method : undefined,
      path: typeof item.path === 'string' ? item.path : undefined,
      error_code: typeof item.error_code === 'string' ? item.error_code : undefined,
      error: typeof item.error === 'string' ? item.error : undefined,
    }
  }

  function getCheckActions(result: Record<string, unknown>): AWDCheckerActionView[] {
    return checkerActionOptions.flatMap(({ key, label }) => {
      const action = readCheckerAction(key, label, result[key])
      return action ? [action] : []
    })
  }

  function summarizeCheckResult(result: Record<string, unknown>): string {
    const checkerLabel = getCheckerTypeLabel(result.checker_type)
    const sourceLabel = getCheckSourceLabel(result.check_source)
    const statusLabel = readPreviewPassSummary(result) || getCheckStatusLabel(result.status_reason)
    const checkedAt =
      typeof result.checked_at === 'string' && result.checked_at.trim() !== ''
        ? formatDateTime(result.checked_at)
        : ''

    const entries = [
      checkerLabel ? `Checker: ${checkerLabel}` : '',
      sourceLabel ? `来源: ${sourceLabel}` : '',
      statusLabel ? `状态: ${statusLabel}` : '',
      checkedAt ? `时间: ${checkedAt}` : '',
    ].filter(Boolean)

    if (entries.length > 0) {
      return entries.join(' | ')
    }

    const fallbackEntries = Object.entries(result)
      .filter(([, value]) => value !== null && value !== undefined && value !== '')
      .slice(0, 3)
      .map(([key, value]) => `${key}: ${String(value)}`)

    return fallbackEntries.length > 0 ? fallbackEntries.join(' | ') : '未返回检查明细'
  }

  function readProbeAttempts(value: unknown): AWDProbeAttemptView[] {
    if (!Array.isArray(value)) {
      return []
    }
    return value
      .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
      .map((item) => ({
        probe: typeof item.probe === 'string' ? item.probe : '',
        healthy: item.healthy === true,
        latency_ms: typeof item.latency_ms === 'number' ? item.latency_ms : undefined,
        error_code: typeof item.error_code === 'string' ? item.error_code : undefined,
        error: typeof item.error === 'string' ? item.error : undefined,
      }))
  }

  function getTargetActions(
    target: Record<string, unknown> | AWDProbeTargetView
  ): AWDCheckerActionView[] {
    if (Array.isArray((target as AWDProbeTargetView).actions)) {
      return (target as AWDProbeTargetView).actions
    }
    return getCheckActions(target as Record<string, unknown>)
  }

  function getCheckTargets(result: Record<string, unknown>): AWDProbeTargetView[] {
    if (!Array.isArray(result.targets)) {
      return []
    }
    return result.targets
      .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
      .map((item) => ({
        access_url: typeof item.access_url === 'string' ? item.access_url : undefined,
        healthy: item.healthy === true,
        probe: typeof item.probe === 'string' ? item.probe : undefined,
        latency_ms: typeof item.latency_ms === 'number' ? item.latency_ms : undefined,
        error_code: typeof item.error_code === 'string' ? item.error_code : undefined,
        error: typeof item.error === 'string' ? item.error : undefined,
        attempts: readProbeAttempts(item.attempts),
        actions: getTargetActions(item),
      }))
  }

  function getPrimaryAccessURL(result: Record<string, unknown>): string {
    const previewContext =
      result.preview_context && typeof result.preview_context === 'object'
        ? (result.preview_context as Record<string, unknown>)
        : null
    const previewAccessURL =
      previewContext && typeof previewContext.access_url === 'string'
        ? previewContext.access_url.trim()
        : ''
    if (previewAccessURL) {
      return previewAccessURL
    }
    return getCheckTargets(result)[0]?.access_url || ''
  }

  function getTargetProbeSummary(result: Record<string, unknown>): string {
    const previewSummary = readPreviewPassSummary(result)
    if (previewSummary) {
      return `试跑 ${previewSummary}`
    }
    const targets = getCheckTargets(result)
    if (targets.length === 0) {
      return ''
    }
    const healthyTargets = targets.filter((item) => item.healthy).length
    return `探测目标 ${healthyTargets}/${targets.length} 正常`
  }

  function getProbeStatusText(healthy: boolean, errorCode?: string, error?: string): string {
    if (healthy) {
      return '探测成功'
    }
    return getCheckStatusLabel(errorCode) || getCheckStatusLabel(error) || error || '探测失败'
  }

  function formatLatency(value?: number): string {
    if (typeof value !== 'number' || Number.isNaN(value) || value <= 0) {
      return ''
    }
    return `${Math.round(value)} ms`
  }

  return {
    getCheckSourceLabel,
    getCheckerTypeLabel,
    getValidationStateLabel,
    getCheckStatusLabel,
    summarizeCheckResult,
    getCheckActions,
    getCheckTargets,
    getPrimaryAccessURL,
    getTargetActions,
    getTargetProbeSummary,
    getProbeStatusText,
    formatLatency,
  }
}

interface AWDProbeAttemptView {
  probe: string
  healthy: boolean
  latency_ms?: number
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
}

interface UseAwdCheckResultPresentationOptions {
  formatDateTime: (value?: string) => string
}

export function useAwdCheckResultPresentation({ formatDateTime }: UseAwdCheckResultPresentationOptions) {
  function getCheckSourceLabel(value: unknown): string {
    switch (value) {
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
      invalid_access_url: '访问地址无效',
      all_probes_failed: '巡检失败',
    }
    return labels[value] || value
  }

  function summarizeCheckResult(result: Record<string, unknown>): string {
    const sourceLabel = getCheckSourceLabel(result.check_source)
    const statusLabel = getCheckStatusLabel(result.status_reason)
    const checkedAt =
      typeof result.checked_at === 'string' && result.checked_at.trim() !== ''
        ? formatDateTime(result.checked_at)
        : ''

    const entries = [
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
      }))
  }

  function getTargetProbeSummary(result: Record<string, unknown>): string {
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
    return getCheckStatusLabel(errorCode) || error || '探测失败'
  }

  function formatLatency(value?: number): string {
    if (typeof value !== 'number' || Number.isNaN(value) || value <= 0) {
      return ''
    }
    return `${Math.round(value)} ms`
  }

  return {
    getCheckSourceLabel,
    getCheckStatusLabel,
    summarizeCheckResult,
    getCheckTargets,
    getTargetProbeSummary,
    getProbeStatusText,
    formatLatency,
  }
}

import { describe, expect, it } from 'vitest'

import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'

describe('useAwdCheckResultPresentation', () => {
  it('应该展示 http_standard checker 的类型和动作摘要', () => {
    const presentation = useAwdCheckResultPresentation({
      formatDateTime: (value?: string) => (value ? `fmt:${value}` : '未记录'),
    })

    expect(presentation.getCheckerTypeLabel('http_standard')).toBe('HTTP 标准 Checker')

    const summary = presentation.summarizeCheckResult({
      check_source: 'scheduler',
      checker_type: 'http_standard',
      status_reason: 'healthy',
      checked_at: '2026-03-12T10:05:00.000Z',
    })
    expect(summary).toContain('Checker: HTTP 标准 Checker')
    expect(summary).toContain('状态: 全部正常')

    expect(
      presentation.getCheckActions({
        put_flag: { healthy: true, method: 'PUT', path: '/api/flag' },
        get_flag: { healthy: true, method: 'GET', path: '/api/flag' },
        havoc: { healthy: false, error_code: 'unexpected_http_status' },
      })
    ).toEqual([
      expect.objectContaining({ key: 'put_flag', label: 'PUT Flag', healthy: true }),
      expect.objectContaining({ key: 'get_flag', label: 'GET Flag', healthy: true }),
      expect.objectContaining({ key: 'havoc', label: 'Havoc', healthy: false }),
    ])
  })

  it('应该解析 target 级 checker 动作结果', () => {
    const presentation = useAwdCheckResultPresentation({
      formatDateTime: (value?: string) => value || '未记录',
    })

    expect(
      presentation.getTargetActions({
        access_url: 'http://svc.local',
        healthy: false,
        put_flag: { healthy: true, method: 'PUT', path: '/api/flag' },
        get_flag: { healthy: false, error_code: 'flag_mismatch', error: 'flag_mismatch' },
        havoc: { healthy: false, error_code: 'unexpected_http_status' },
        attempts: [],
      })
    ).toEqual([
      expect.objectContaining({ key: 'put_flag', healthy: true }),
      expect.objectContaining({ key: 'get_flag', healthy: false, error_code: 'flag_mismatch' }),
      expect.objectContaining({
        key: 'havoc',
        healthy: false,
        error_code: 'unexpected_http_status',
      }),
    ])
  })
})

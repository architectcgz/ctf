import { createApp, type App } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useReportStatusPolling } from '../useReportStatusPolling'

function withSetup<T>(composable: () => T): [T, App] {
  let result!: T

  const app = createApp({
    setup() {
      result = composable()
      return () => null
    },
  })

  app.mount(document.createElement('div'))
  return [result, app]
}

let app: App | null = null

describe('useReportStatusPolling', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    app?.unmount()
    app = null
    vi.clearAllMocks()
    vi.useRealTimers()
  })

  it('启动后应该立即轮询并在完成后自动停止', async () => {
    const getReportStatus = vi.fn().mockResolvedValue({
      report_id: 'r-1',
      status: 'ready',
      download_url: '/reports/r-1',
      expires_at: undefined,
      error_message: '',
    })

    const [composable, testApp] = withSetup(() => useReportStatusPolling(getReportStatus))
    app = testApp
    const onUpdate = vi.fn()

    composable.start('r-1', onUpdate)
    await Promise.resolve()

    expect(getReportStatus).toHaveBeenCalledWith('r-1')
    expect(onUpdate).toHaveBeenCalledWith(
      expect.objectContaining({
        report_id: 'r-1',
        status: 'ready',
      })
    )
    expect(composable.polling.value).toBe(false)
  })

  it('processing 状态应该继续定时轮询，直到结果结束', async () => {
    const getReportStatus = vi
      .fn()
      .mockResolvedValueOnce({
        report_id: 'r-2',
        status: 'processing',
        download_url: '',
        expires_at: undefined,
        error_message: '',
      })
      .mockResolvedValueOnce({
        report_id: 'r-2',
        status: 'ready',
        download_url: '/reports/r-2',
        expires_at: undefined,
        error_message: '',
      })

    const [composable, testApp] = withSetup(() => useReportStatusPolling(getReportStatus))
    app = testApp
    const onUpdate = vi.fn()

    composable.start('r-2', onUpdate)
    await Promise.resolve()
    expect(composable.polling.value).toBe(true)

    await vi.advanceTimersByTimeAsync(3000)

    expect(getReportStatus).toHaveBeenCalledTimes(2)
    expect(onUpdate).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ report_id: 'r-2', status: 'processing' })
    )
    expect(onUpdate).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ report_id: 'r-2', status: 'ready' })
    )
    expect(composable.polling.value).toBe(false)
  })

  it('请求失败时应该停止轮询并回调错误', async () => {
    const expectedError = new Error('network failed')
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const getReportStatus = vi.fn().mockRejectedValue(expectedError)

    const [composable, testApp] = withSetup(() => useReportStatusPolling(getReportStatus))
    app = testApp
    const onError = vi.fn()

    composable.start('r-3', vi.fn(), onError)
    await Promise.resolve()

    expect(onError).toHaveBeenCalledWith(expectedError)
    expect(composable.polling.value).toBe(false)
    expect(consoleErrorSpy).toHaveBeenCalled()

    consoleErrorSpy.mockRestore()
  })
})

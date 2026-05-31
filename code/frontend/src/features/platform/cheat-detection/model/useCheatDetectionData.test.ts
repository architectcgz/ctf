import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useCheatDetectionData } from './useCheatDetectionData'

const adminApiMocks = vi.hoisted(() => ({
  getCheatDetection: vi.fn(),
}))

vi.mock('@/api/admin/platform', () => adminApiMocks)

describe('useCheatDetectionData', () => {
  beforeEach(() => {
    adminApiMocks.getCheatDetection.mockReset()
  })

  it('应在初始化时加载作弊检测结果', async () => {
    adminApiMocks.getCheatDetection.mockResolvedValue({
      generated_at: '2026-03-07T06:00:00.000Z',
      summary: {
        submit_burst_users: 1,
        shared_ip_groups: 1,
        affected_users: 2,
      },
      suspects: [],
      shared_ips: [],
    })

    let composable!: ReturnType<typeof useCheatDetectionData>
    const Harness = defineComponent({
      setup() {
        composable = useCheatDetectionData()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getCheatDetection).toHaveBeenCalledTimes(1)
    expect(composable.riskData.value?.summary.affected_users).toBe(2)
    expect(composable.error.value).toBe('')

    wrapper.unmount()
  })

  it('初始化失败时应暴露作弊检测错误状态', async () => {
    adminApiMocks.getCheatDetection.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useCheatDetectionData>
    const Harness = defineComponent({
      setup() {
        composable = useCheatDetectionData()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.error.value).toBe('加载作弊检测结果失败，请稍后重试。')
    expect(composable.loading.value).toBe(false)

    wrapper.unmount()
  })
})

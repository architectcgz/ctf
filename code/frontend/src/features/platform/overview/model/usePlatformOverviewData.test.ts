import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { usePlatformOverviewData } from './usePlatformOverviewData'

const adminApiMocks = vi.hoisted(() => ({
  getDashboard: vi.fn(),
}))

vi.mock('@/api/admin/platform', () => adminApiMocks)

describe('usePlatformOverviewData', () => {
  beforeEach(() => {
    adminApiMocks.getDashboard.mockReset()
  })

  it('应在初始化时加载平台概览数据', async () => {
    adminApiMocks.getDashboard.mockResolvedValue({
      online_users: 18,
      active_containers: 6,
      cpu_usage: 62,
      memory_usage: 47,
      container_stats: [],
      alerts: [],
    })

    let composable!: ReturnType<typeof usePlatformOverviewData>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformOverviewData()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getDashboard).toHaveBeenCalledTimes(1)
    expect(composable.dashboard.value?.online_users).toBe(18)
    expect(composable.loading.value).toBe(false)

    wrapper.unmount()
  })

  it('初始化失败时应暴露平台概览错误状态', async () => {
    adminApiMocks.getDashboard.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof usePlatformOverviewData>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformOverviewData()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.error.value).toBe('加载系统概览失败，请稍后重试')
    expect(composable.loading.value).toBe(false)

    wrapper.unmount()
  })
})

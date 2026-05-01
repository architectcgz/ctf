import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useContestAwdChallengePicker } from './useContestAwdChallengePicker'

const adminApiMocks = vi.hoisted(() => ({
  listAdminAwdChallenges: vi.fn(),
}))

vi.mock('@/api/admin/awd-authoring', () => ({
  listAdminAwdChallenges: adminApiMocks.listAdminAwdChallenges,
}))

function buildAwdChallenge(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: '11',
    name: 'Bank Portal AWD',
    slug: 'bank-portal-awd',
    category: 'web',
    difficulty: 'medium',
    description: 'bank target',
    service_type: 'web_http',
    deployment_mode: 'single_container',
    version: '1.0.0',
    status: 'published',
    readiness_status: 'passed',
    created_at: '2026-03-01T00:00:00.000Z',
    updated_at: '2026-03-01T00:00:00.000Z',
    ...overrides,
  }
}

describe('useContestAwdChallengePicker', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    adminApiMocks.listAdminAwdChallenges.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('初始刷新应使用第一页和固定分页大小加载已发布 AWD 题目', async () => {
    adminApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [buildAwdChallenge()],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const existingChallengeIds = ref<string[]>([])
    let composable!: ReturnType<typeof useContestAwdChallengePicker>

    mount(
      defineComponent({
        setup() {
          composable = useContestAwdChallengePicker({ existingChallengeIds })
          return () => null
        },
      })
    )

    await composable.refresh()

    expect(adminApiMocks.listAdminAwdChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: 'published',
    })
    expect(composable.list.value).toHaveLength(1)
    expect(composable.selectableList.value).toHaveLength(1)
  })

  it('关键词变化后应防抖回到第一页并带上筛选条件', async () => {
    adminApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [buildAwdChallenge()],
      total: 21,
      page: 1,
      page_size: 20,
    })

    const existingChallengeIds = ref<string[]>([])
    let composable!: ReturnType<typeof useContestAwdChallengePicker>

    mount(
      defineComponent({
        setup() {
          composable = useContestAwdChallengePicker({ existingChallengeIds })
          return () => null
        },
      })
    )

    await composable.changePage(3)
    composable.filters.keyword = 'bank'
    composable.filters.keyword = 'bank portal'

    expect(adminApiMocks.listAdminAwdChallenges).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(250)
    await flushPromises()

    expect(composable.page.value).toBe(1)
    expect(adminApiMocks.listAdminAwdChallenges).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      keyword: 'bank portal',
      status: 'published',
    })
  })

  it('服务类型、部署方式和就绪状态变化后应立即回到第一页', async () => {
    adminApiMocks.listAdminAwdChallenges
      .mockResolvedValueOnce({
        list: [buildAwdChallenge()],
        total: 5,
        page: 2,
        page_size: 20,
      })
      .mockResolvedValue({
        list: [buildAwdChallenge()],
        total: 5,
        page: 1,
        page_size: 20,
      })

    const existingChallengeIds = ref<string[]>([])
    let composable!: ReturnType<typeof useContestAwdChallengePicker>

    mount(
      defineComponent({
        setup() {
          composable = useContestAwdChallengePicker({ existingChallengeIds })
          return () => null
        },
      })
    )

    await composable.changePage(2)
    composable.filters.serviceType = 'binary_tcp'
    await flushPromises()
    composable.filters.deploymentMode = 'topology'
    await flushPromises()
    composable.filters.readinessStatus = 'failed'
    await flushPromises()

    expect(composable.page.value).toBe(1)
    expect(adminApiMocks.listAdminAwdChallenges).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      service_type: 'binary_tcp',
      deployment_mode: 'topology',
      readiness_status: 'failed',
      status: 'published',
    })
  })

  it('应从可选列表中过滤掉已关联题目', async () => {
    adminApiMocks.listAdminAwdChallenges.mockResolvedValue({
      list: [
        buildAwdChallenge({ id: '11', name: 'Bank Portal AWD' }),
        buildAwdChallenge({ id: '12', name: 'IoT Hub AWD', slug: 'iot-hub-awd' }),
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const existingChallengeIds = ref<string[]>(['11'])
    let composable!: ReturnType<typeof useContestAwdChallengePicker>

    mount(
      defineComponent({
        setup() {
          composable = useContestAwdChallengePicker({ existingChallengeIds })
          return () => null
        },
      })
    )

    await composable.refresh()

    expect(composable.list.value).toHaveLength(2)
    expect(composable.selectableList.value).toEqual([
      expect.objectContaining({ id: '12', name: 'IoT Hub AWD' }),
    ])
  })

  it('加载失败时应保留上次成功结果并暴露错误文本', async () => {
    adminApiMocks.listAdminAwdChallenges
      .mockResolvedValueOnce({
        list: [buildAwdChallenge({ id: '11', name: 'Bank Portal AWD' })],
        total: 1,
        page: 1,
        page_size: 20,
      })
      .mockRejectedValueOnce(new Error('catalog failed'))

    const existingChallengeIds = ref<string[]>([])
    let composable!: ReturnType<typeof useContestAwdChallengePicker>

    mount(
      defineComponent({
        setup() {
          composable = useContestAwdChallengePicker({ existingChallengeIds })
          return () => null
        },
      })
    )

    await composable.refresh()
    await composable.changePage(2)

    expect(composable.list.value).toEqual([
      expect.objectContaining({ id: '11', name: 'Bank Portal AWD' }),
    ])
    expect(composable.loadError.value).toBe('catalog failed')
  })
})

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useContestOperationsData } from './useContestOperationsData'

const adminApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
}))

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    getContest: adminApiMocks.getContest,
  }
})

describe('useContestOperationsData', () => {
  beforeEach(() => {
    adminApiMocks.getContest.mockReset()
  })

  it('应按 contestId 加载单场竞赛信息', async () => {
    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-ops-1',
      title: '2026 AWD 运维联赛',
      status: 'running',
    })

    let composable!: ReturnType<typeof useContestOperationsData>
    const Harness = defineComponent({
      setup() {
        composable = useContestOperationsData(computed(() => 'contest-ops-1'))
        return () => null
      },
    })

    mount(Harness)
    await composable.loadContest()
    await flushPromises()

    expect(adminApiMocks.getContest).toHaveBeenCalledWith('contest-ops-1')
    expect(composable.contest.value?.title).toBe('2026 AWD 运维联赛')
    expect(composable.loading.value).toBe(false)
  })

  it('缺少 contestId 时不应发起查询', async () => {
    let composable!: ReturnType<typeof useContestOperationsData>
    const Harness = defineComponent({
      setup() {
        composable = useContestOperationsData(ref(''))
        return () => null
      },
    })

    mount(Harness)
    await composable.loadContest()
    await flushPromises()

    expect(adminApiMocks.getContest).not.toHaveBeenCalled()
    expect(composable.loading.value).toBe(false)
    expect(composable.loadError.value).toBe('')
  })

  it('加载失败时应暴露错误状态', async () => {
    adminApiMocks.getContest.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useContestOperationsData>
    const Harness = defineComponent({
      setup() {
        composable = useContestOperationsData(ref('contest-ops-1'))
        return () => null
      },
    })

    mount(Harness)
    await composable.loadContest()
    await flushPromises()

    expect(composable.contest.value).toBeNull()
    expect(composable.loadError.value).toBe('boom')
    expect(composable.loading.value).toBe(false)
  })
})

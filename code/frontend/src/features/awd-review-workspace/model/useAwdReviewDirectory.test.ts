import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useAuthStore } from '@/stores/auth'

import { useAwdReviewDirectory } from './useAwdReviewDirectory'

const awdReviewApiMocks = vi.hoisted(() => ({
  listAwdReviewsByRole: vi.fn(),
}))

vi.mock('@/api/awd-reviews', () => ({
  listAwdReviewsByRole: awdReviewApiMocks.listAwdReviewsByRole,
}))

describe('useAwdReviewDirectory', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    useAuthStore().user = { id: 'teacher-1', role: 'teacher' } as never
    awdReviewApiMocks.listAwdReviewsByRole.mockReset()
  })

  it('应按当前 role 和筛选参数加载复盘目录', async () => {
    awdReviewApiMocks.listAwdReviewsByRole.mockResolvedValue({
      list: [{ id: 'contest-1', title: '春季 AWD 联训', status: 'running' }],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        running_count: 1,
        export_ready_count: 0,
      },
    })

    let composable!: ReturnType<typeof useAwdReviewDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useAwdReviewDirectory()
        composable.filters.value.keyword = '春季'
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.loadContests()
    await flushPromises()

    expect(awdReviewApiMocks.listAwdReviewsByRole).toHaveBeenCalledWith(
      'teacher',
      {
        status: undefined,
        keyword: '春季',
        page: 1,
        page_size: 20,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )
    expect(composable.contests.value).toHaveLength(1)
    expect(composable.summary.value.running_count).toBe(1)

    wrapper.unmount()
  })

  it('筛选变更后应 debounce 刷新并回到第一页', async () => {
    awdReviewApiMocks.listAwdReviewsByRole.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
      summary: {
        running_count: 0,
        export_ready_count: 0,
      },
    })

    let composable!: ReturnType<typeof useAwdReviewDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useAwdReviewDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)

    composable.page.value = 3
    composable.filters.value.keyword = '期末'
    composable.triggerFilterRefresh()
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(composable.page.value).toBe(1)
    expect(awdReviewApiMocks.listAwdReviewsByRole).toHaveBeenCalledTimes(1)
    expect(awdReviewApiMocks.listAwdReviewsByRole).toHaveBeenLastCalledWith(
      'teacher',
      {
        status: undefined,
        keyword: '期末',
        page: 1,
        page_size: 20,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )

    wrapper.unmount()
  })

  it('较晚返回的旧请求不应覆盖最新结果', async () => {
    let resolveFirst!: (value: unknown) => void
    let resolveSecond!: (value: unknown) => void
    awdReviewApiMocks.listAwdReviewsByRole
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve
          })
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveSecond = resolve
          })
      )

    let composable!: ReturnType<typeof useAwdReviewDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useAwdReviewDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)

    const first = composable.loadContests()
    const second = composable.loadContests()

    resolveSecond({
      list: [{ id: 'contest-new', title: '新复盘', status: 'running' }],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        running_count: 1,
        export_ready_count: 0,
      },
    })
    await second
    await flushPromises()

    resolveFirst({
      list: [{ id: 'contest-old', title: '旧复盘', status: 'ended' }],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        running_count: 0,
        export_ready_count: 1,
      },
    })
    await first
    await flushPromises()

    expect(composable.contests.value[0]?.id).toBe('contest-new')

    wrapper.unmount()
  })

  it('加载失败时应回填统一错误状态并清空目录', async () => {
    awdReviewApiMocks.listAwdReviewsByRole.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useAwdReviewDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useAwdReviewDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.loadContests()
    await flushPromises()

    expect(composable.error.value).toBe('加载 AWD 复盘目录失败，请稍后重试')
    expect(composable.contests.value).toEqual([])
    expect(composable.summary.value.export_ready_count).toBe(0)

    wrapper.unmount()
  })
})

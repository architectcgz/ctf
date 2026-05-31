import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useManagedInstanceDirectory } from './useManagedInstanceDirectory'

const instanceAccessApiMocks = vi.hoisted(() => ({
  getInstanceDirectoryByRole: vi.fn(),
}))

vi.mock('@/api/instances', () => ({
  getInstanceDirectoryByRole: instanceAccessApiMocks.getInstanceDirectoryByRole,
}))

describe('useManagedInstanceDirectory', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockReset()
  })

  it('应按 role 和 query 回填目录分页结果', async () => {
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockResolvedValue({
      list: [{ id: 'inst-1' }],
      total: 1,
      page: 1,
      page_size: 15,
      summary: {
        total_count: 1,
        running_count: 1,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    })

    let composable!: ReturnType<typeof useManagedInstanceDirectory>
    const loadedSpy = vi.fn()
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDirectory({
          role: 'admin',
          initialPageSize: 15,
          buildQuery: ({ page, pageSize }) => ({
            keyword: 'alice',
            status: 'running',
            page,
            page_size: pageSize,
          }),
          errorMessage: '加载实例列表失败，请稍后重试',
          onLoaded: loadedSpy,
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.loadInstances()
    await flushPromises()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenCalledWith(
      'admin',
      {
        keyword: 'alice',
        status: 'running',
        page: 1,
        page_size: 15,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(composable.list.value).toEqual([{ id: 'inst-1' }])
    expect(composable.total.value).toBe(1)
    expect(loadedSpy).toHaveBeenCalled()

    wrapper.unmount()
  })

  it('应在 debounce 搜索时只保留最后一次请求', async () => {
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 0,
        running_count: 0,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    })

    let keyword = ''
    let composable!: ReturnType<typeof useManagedInstanceDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDirectory({
          role: 'teacher',
          buildQuery: ({ page, pageSize }) => ({
            keyword,
            page,
            page_size: pageSize,
          }),
          errorMessage: '加载实例列表失败，请稍后重试',
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    keyword = 'a'
    composable.scheduleSearch()
    keyword = 'alice'
    composable.scheduleSearch()
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenCalledTimes(1)
    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenLastCalledWith(
      'teacher',
      {
        keyword: 'alice',
        page: 1,
        page_size: 20,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )

    wrapper.unmount()
  })

  it('旧请求返回较晚时不应覆盖最新结果', async () => {
    let resolveFirst!: (value: unknown) => void
    let resolveSecond!: (value: unknown) => void
    instanceAccessApiMocks.getInstanceDirectoryByRole
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

    let composable!: ReturnType<typeof useManagedInstanceDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDirectory({
          role: 'admin',
          buildQuery: ({ page, pageSize }) => ({
            page,
            page_size: pageSize,
          }),
          errorMessage: '加载实例列表失败，请稍后重试',
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    const first = composable.loadInstances()
    const second = composable.loadInstances()
    resolveSecond({
      list: [{ id: 'inst-new' }],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 1,
        running_count: 1,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    })
    await second
    await flushPromises()

    resolveFirst({
      list: [{ id: 'inst-old' }],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 1,
        running_count: 1,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    })
    await first
    await flushPromises()

    expect(composable.list.value).toEqual([{ id: 'inst-new' }])

    wrapper.unmount()
  })

  it('加载失败时应清空列表并回填统一错误状态', async () => {
    const errorSpy = vi.fn()
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useManagedInstanceDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useManagedInstanceDirectory({
          role: 'teacher',
          buildQuery: ({ page, pageSize }) => ({
            page,
            page_size: pageSize,
          }),
          errorMessage: '加载实例列表失败，请稍后重试',
          onLoadError: errorSpy,
        })
        return () => null
      },
    })

    const wrapper = mount(Harness)

    await composable.loadInstances()
    await flushPromises()

    expect(composable.list.value).toEqual([])
    expect(composable.total.value).toBe(0)
    expect(composable.error.value).toBe('加载实例列表失败，请稍后重试')
    expect(errorSpy).toHaveBeenCalled()

    wrapper.unmount()
  })
})

import { describe, it, expect, vi } from 'vitest'
import { usePagination } from '../usePagination'

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((nextResolve, nextReject) => {
    resolve = nextResolve
    reject = nextReject
  })
  return { promise, resolve, reject }
}

describe('usePagination', () => {
  it('应该初始化默认值', () => {
    const mockFetch = vi.fn().mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const { page, pageSize, total, list } = usePagination(mockFetch)

    expect(page.value).toBe(1)
    expect(pageSize.value).toBe(20)
    expect(total.value).toBe(0)
    expect(list.value).toEqual([])
  })

  it('应该正确加载数据', async () => {
    const mockData = {
      list: [{ id: 1 }, { id: 2 }],
      total: 10,
      page: 1,
      page_size: 20,
    }
    const mockFetch = vi.fn().mockResolvedValue(mockData)

    const { refresh, list, total } = usePagination(mockFetch)
    await refresh()

    expect(list.value).toEqual(mockData.list)
    expect(total.value).toBe(10)
    expect(mockFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        signal: expect.any(AbortSignal),
      })
    )
  })

  it('应该正确切换页码', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      list: [],
      total: 100,
      page: 2,
      page_size: 20,
    })

    const { changePage, page } = usePagination(mockFetch)
    await changePage(2)

    expect(page.value).toBe(2)
    expect(mockFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 2,
        page_size: 20,
        signal: expect.any(AbortSignal),
      })
    )
  })

  it('应该正确切换每页大小', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      list: [],
      total: 100,
      page: 1,
      page_size: 50,
    })

    const { changePageSize, pageSize, page } = usePagination(mockFetch)
    await changePageSize(50)

    expect(pageSize.value).toBe(50)
    expect(page.value).toBe(1)
    expect(mockFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 50,
        signal: expect.any(AbortSignal),
      })
    )
  })

  it('应该暴露加载错误', async () => {
    const expectedError = new Error('加载失败')
    const mockFetch = vi.fn().mockRejectedValue(expectedError)

    const { refresh, error, loading } = usePagination(mockFetch)

    await refresh()
    expect(error.value).toBe(expectedError)
    expect(loading.value).toBe(false)
  })

  it('应该在响应缺少合法 page_size 时报错', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      list: [],
      total: 3,
      page: 1,
      page_size: 0,
    })

    const { refresh, error } = usePagination(mockFetch)
    await refresh()

    expect(error.value).toBeInstanceOf(Error)
    expect((error.value as Error).message).toContain('page_size')
  })

  it('应该为新请求中止旧请求并向 fetchFn 传入 signal', async () => {
    const firstRequest = deferred<{
      list: Array<{ id: number }>
      total: number
      page: number
      page_size: number
    }>()
    const secondRequest = deferred<{
      list: Array<{ id: number }>
      total: number
      page: number
      page_size: number
    }>()
    const mockFetch = vi
      .fn()
      .mockImplementationOnce(({ signal }) => {
        expect(signal).toBeInstanceOf(AbortSignal)
        return firstRequest.promise
      })
      .mockImplementationOnce(({ signal }) => {
        expect(signal).toBeInstanceOf(AbortSignal)
        return secondRequest.promise
      })

    const { refresh, list, error } = usePagination(mockFetch)

    const firstRefresh = refresh()
    const firstSignal = mockFetch.mock.calls[0]?.[0]?.signal as AbortSignal
    expect(firstSignal.aborted).toBe(false)

    const secondRefresh = refresh()
    expect(firstSignal.aborted).toBe(true)

    secondRequest.resolve({
      list: [{ id: 2 }],
      total: 1,
      page: 1,
      page_size: 20,
    })
    await secondRefresh

    firstRequest.reject(Object.assign(new Error('canceled'), { code: 'ERR_CANCELED' }))
    await firstRefresh

    expect(list.value).toEqual([{ id: 2 }])
    expect(error.value).toBeNull()
  })
})

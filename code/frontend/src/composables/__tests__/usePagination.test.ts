import { describe, it, expect, vi } from 'vitest'
import { usePagination } from '../usePagination'

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
    expect(mockFetch).toHaveBeenCalledWith({ page: 1, page_size: 20 })
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
    expect(mockFetch).toHaveBeenCalledWith({ page: 2, page_size: 20 })
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
    expect(mockFetch).toHaveBeenCalledWith({ page: 1, page_size: 50 })
  })

  it('应该暴露加载错误', async () => {
    const expectedError = new Error('加载失败')
    const mockFetch = vi.fn().mockRejectedValue(expectedError)

    const { refresh, error, loading } = usePagination(mockFetch)

    await refresh()
    expect(error.value).toBe(expectedError)
    expect(loading.value).toBe(false)
  })
})

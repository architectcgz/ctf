import { ref, type Ref } from 'vue'

import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

export interface PaginationState<T> {
  list: Ref<T[]>
  total: Ref<number>
  page: Ref<number>
  pageSize: Ref<number>
  loading: Ref<boolean>
  error: Ref<unknown | null>
  changePage: (next: number) => Promise<void>
  changePageSize: (next: number) => Promise<void>
  refresh: () => Promise<void>
}

export function usePagination<T>(
  fetchFn: (params: { page: number; page_size: number }) => Promise<{ list: T[]; total: number; page: number; page_size: number }>
): PaginationState<T> {
  const list = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const loading = ref(false)
  const error = ref<unknown | null>(null)

  async function refresh(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const data = await fetchFn({ page: page.value, page_size: pageSize.value })
      if (!Number.isInteger(data.page_size) || data.page_size < 1) {
        throw new Error('分页响应缺少合法的 page_size 字段')
      }
      list.value = data.list
      total.value = data.total
      page.value = data.page
      pageSize.value = data.page_size
    } catch (err) {
      error.value = err
    } finally {
      loading.value = false
    }
  }

  async function changePage(next: number): Promise<void> {
    page.value = Math.max(1, Math.floor(next))
    await refresh()
  }

  async function changePageSize(next: number): Promise<void> {
    pageSize.value = Math.max(1, Math.floor(next))
    page.value = 1
    await refresh()
  }

  return { list, total, page, pageSize, loading, error, changePage, changePageSize, refresh }
}

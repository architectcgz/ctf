import { ref, shallowRef, type Ref } from 'vue'

import { useAbortController } from '@/composables/useAbortController'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import type { PageResult } from '@/api/contracts'

export interface PaginationState<T, R extends PageResult<T> = PageResult<T>> {
  list: Ref<T[]>
  total: Ref<number>
  page: Ref<number>
  pageSize: Ref<number>
  loading: Ref<boolean>
  error: Ref<unknown | null>
  response: Ref<R | null>
  changePage: (next: number) => Promise<void>
  changePageSize: (next: number) => Promise<void>
  refresh: () => Promise<void>
}

export function usePagination<T>(
  fetchFn: (params: {
    page: number
    page_size: number
    signal: AbortSignal
  }) => Promise<PageResult<T>>
): PaginationState<T>
export function usePagination<T, R extends PageResult<T>>(
  fetchFn: (params: {
    page: number
    page_size: number
    signal: AbortSignal
  }) => Promise<R>
): PaginationState<T, R>
export function usePagination<T, R extends PageResult<T>>(
  fetchFn: (params: {
    page: number
    page_size: number
    signal: AbortSignal
  }) => Promise<R>
): PaginationState<T, R> {

  const list = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const loading = ref(false)
  const error = ref<unknown | null>(null)
  const response = shallowRef<R | null>(null) as Ref<R | null>
  let latestRequestId = 0
  const { createController } = useAbortController()

  function isCanceledError(err: unknown): boolean {
    return (
      !!err &&
      typeof err === 'object' &&
      ('code' in err ? (err as { code?: unknown }).code === 'ERR_CANCELED' : false)
    )
  }

  async function load(nextPage = page.value, nextPageSize = pageSize.value): Promise<void> {
    const requestId = ++latestRequestId
    const controller = createController()
    loading.value = true
    error.value = null
    try {
      const data = await fetchFn({
        page: nextPage,
        page_size: nextPageSize,
        signal: controller.signal,
      })
      if (requestId !== latestRequestId) return
      if (!Number.isInteger(data.page_size) || data.page_size < 1) {
        throw new Error('分页响应缺少合法的 page_size 字段')
      }
      response.value = data
      list.value = data.list
      total.value = data.total
      page.value = data.page
      pageSize.value = data.page_size
    } catch (err) {
      if (requestId !== latestRequestId) return
      if (isCanceledError(err)) {
        error.value = null
        return
      }
      error.value = err
    } finally {
      if (requestId !== latestRequestId) return
      loading.value = false
    }
  }

  async function refresh(): Promise<void> {
    await load(page.value, pageSize.value)
  }

  async function changePage(next: number): Promise<void> {
    await load(Math.max(1, Math.floor(next)), pageSize.value)
  }

  async function changePageSize(next: number): Promise<void> {
    await load(1, Math.max(1, Math.floor(next)))
  }

  return { list, total, page, pageSize, loading, error, response, changePage, changePageSize, refresh }
}

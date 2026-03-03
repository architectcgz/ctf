import { ref, type Ref } from 'vue'

export interface PaginationState<T> {
  list: Ref<T[]>
  total: Ref<number>
  page: Ref<number>
  pageSize: Ref<number>
  loading: Ref<boolean>
  changePage: (next: number) => Promise<void>
  changePageSize: (next: number) => Promise<void>
  refresh: () => Promise<void>
}

export function usePagination<T>(
  fetchFn: (params: { page: number; page_size: number }) => Promise<{ items: T[]; total: number }>
): PaginationState<T> {
  const list = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)

  async function refresh(): Promise<void> {
    loading.value = true
    try {
      const data = await fetchFn({ page: page.value, page_size: pageSize.value })
      list.value = data.items
      total.value = data.total
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

  return { list, total, page, pageSize, loading, changePage, changePageSize, refresh }
}

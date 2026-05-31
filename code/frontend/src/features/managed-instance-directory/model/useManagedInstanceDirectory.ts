import { computed, onUnmounted, ref, toValue, type MaybeRefOrGetter } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import type { InstanceDirectoryItem, InstanceDirectoryPageData } from '@/api/contracts'
import { getInstanceDirectoryByRole } from '@/api/instances'
import { useAbortController } from '@/shared/lib/request/useAbortController'
import type { UserRole } from '@/utils/constants'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

type ManagedInstanceDirectoryQuery = Parameters<typeof getInstanceDirectoryByRole>[1]

interface UseManagedInstanceDirectoryOptions {
  role: MaybeRefOrGetter<UserRole | null | undefined>
  buildQuery: (context: { page: number; pageSize: number }) => ManagedInstanceDirectoryQuery
  errorMessage: string
  initialPageSize?: number
  debounceMs?: number
  onLoaded?: (response: InstanceDirectoryPageData<InstanceDirectoryItem>) => void
  onLoadError?: (error: unknown) => void
}

function isCanceledError(err: unknown): boolean {
  return (
    !!err &&
    typeof err === 'object' &&
    ('code' in err ? (err as { code?: unknown }).code === 'ERR_CANCELED' : false)
  )
}

export function useManagedInstanceDirectory(options: UseManagedInstanceDirectoryOptions) {
  const list = ref<InstanceDirectoryItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(options.initialPageSize ?? DEFAULT_PAGE_SIZE)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))

  let latestRequestId = 0
  const { createController, abort } = useAbortController()

  async function loadInstances(): Promise<void> {
    const requestId = ++latestRequestId
    const controller = createController()
    loading.value = true
    error.value = null

    try {
      const response = await getInstanceDirectoryByRole(
        toValue(options.role),
        options.buildQuery({
          page: page.value,
          pageSize: pageSize.value,
        }),
        {
          signal: controller.signal,
        }
      )
      if (requestId !== latestRequestId) {
        return
      }

      list.value = response.list
      total.value = response.total
      page.value = response.page
      pageSize.value = response.page_size
      options.onLoaded?.(response)

      if (page.value > totalPages.value) {
        page.value = totalPages.value
      }
    } catch (err) {
      if (requestId !== latestRequestId) {
        return
      }
      if (isCanceledError(err)) {
        return
      }

      list.value = []
      total.value = 0
      error.value = options.errorMessage
      options.onLoadError?.(err)
    } finally {
      if (requestId === latestRequestId) {
        loading.value = false
      }
    }
  }

  type DebouncedDirectoryLoader = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const debouncedLoadInstances = useDebounceFn(() => {
    void loadInstances()
  }, options.debounceMs ?? 250) as DebouncedDirectoryLoader

  function scheduleSearch(): void {
    page.value = 1
    debouncedLoadInstances()
  }

  async function handlePageChange(nextPage: number): Promise<void> {
    const normalizedPage = Math.max(1, Math.floor(nextPage))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }

    page.value = normalizedPage
    await loadInstances()
  }

  onUnmounted(() => {
    debouncedLoadInstances.cancel?.()
    abort()
  })

  return {
    list,
    total,
    page,
    pageSize,
    loading,
    error,
    totalPages,
    loadInstances,
    scheduleSearch,
    handlePageChange,
  }
}

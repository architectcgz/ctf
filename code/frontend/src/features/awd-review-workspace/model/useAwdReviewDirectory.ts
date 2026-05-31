import { computed, onUnmounted, ref } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import { listAwdReviewsByRole } from '@/api/awd-reviews'
import type { AwdReviewContestItemData } from '@/api/contracts'
import { useAbortController } from '@/shared/lib/request/useAbortController'
import { useAuthStore } from '@/stores/auth'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { reportFrontendError } from '@/utils/reportFrontendError'

export interface AwdReviewDirectoryFilters {
  status: '' | AwdReviewContestItemData['status']
  keyword: string
}

export interface AwdReviewDirectorySummary {
  running_count: number
  export_ready_count: number
}

type AwdReviewContestPage = Awaited<ReturnType<typeof listAwdReviewsByRole>>

export function useAwdReviewDirectory() {
  const authStore = useAuthStore()
  const { createController, abort } = useAbortController()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const contests = ref<AwdReviewContestItemData[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const filters = ref<AwdReviewDirectoryFilters>({
    status: '',
    keyword: '',
  })
  const summary = ref<AwdReviewDirectorySummary>({
    running_count: 0,
    export_ready_count: 0,
  })

  let latestRequestId = 0

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))

  function applyPage(nextPage: AwdReviewContestPage): void {
    contests.value = nextPage.list
    total.value = nextPage.total
    page.value = nextPage.page
    pageSize.value = nextPage.page_size
    summary.value = {
      running_count: nextPage.summary.running_count,
      export_ready_count: nextPage.summary.export_ready_count,
    }
  }

  async function loadContests(): Promise<void> {
    const requestId = ++latestRequestId
    const controller = createController()
    loading.value = true
    error.value = null

    try {
      const nextPage = await listAwdReviewsByRole(
        authStore.user?.role,
        {
          status: filters.value.status || undefined,
          keyword: filters.value.keyword.trim() || undefined,
          page: page.value,
          page_size: pageSize.value,
        },
        {
          signal: controller.signal,
        }
      )
      if (requestId !== latestRequestId) {
        return
      }

      applyPage(nextPage)
    } catch (err) {
      if (requestId !== latestRequestId) {
        return
      }
      if (
        err &&
        typeof err === 'object' &&
        ('code' in err ? (err as { code?: unknown }).code === 'ERR_CANCELED' : false)
      ) {
        error.value = null
        return
      }

      reportFrontendError('加载 AWD 复盘目录失败:', err)
      contests.value = []
      total.value = 0
      summary.value = {
        running_count: 0,
        export_ready_count: 0,
      }
      error.value = '加载 AWD 复盘目录失败，请稍后重试'
    } finally {
      if (requestId === latestRequestId) {
        loading.value = false
      }
    }
  }

  type DebouncedContestLoader = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const scheduleContestSearch = useDebounceFn(() => {
    void loadContests()
  }, 250) as DebouncedContestLoader

  function triggerFilterRefresh(): void {
    page.value = 1
    scheduleContestSearch()
  }

  async function changePage(next: number): Promise<void> {
    const normalized = Math.max(1, Math.min(totalPages.value, Math.floor(next)))
    if (normalized === page.value && contests.value.length > 0) {
      return
    }
    page.value = normalized
    await loadContests()
  }

  onUnmounted(() => {
    scheduleContestSearch.cancel?.()
    abort()
  })

  return {
    loading,
    error,
    contests,
    total,
    page,
    totalPages,
    filters,
    summary,
    loadContests,
    changePage,
    triggerFilterRefresh,
  }
}

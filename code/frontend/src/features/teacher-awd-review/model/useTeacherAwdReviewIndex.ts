import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { useRouter } from 'vue-router'

import { listTeacherAWDReviews } from '@/api/teacher'
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import { useAbortController } from '@/composables/useAbortController'
import { useAuthStore } from '@/stores/auth'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { resolveAwdReviewDetailRouteName } from '@/utils/teachingWorkspaceRouting'

export interface PlatformAwdReviewRow extends TeacherAWDReviewContestItemData {
  contestCode: string
}

export function useTeacherAwdReviewIndex() {
  const router = useRouter()
  const authStore = useAuthStore()
  const { createController, abort } = useAbortController()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const contests = ref<TeacherAWDReviewContestItemData[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const runningCount = ref(0)
  const exportReadyCount = ref(0)
  const filters = ref({
    status: '' as '' | TeacherAWDReviewContestItemData['status'],
    keyword: '',
  })
  let latestRequestId = 0

  const hasContests = computed(() => contests.value.length > 0)
  const statusOptions = [
    { value: '', label: '全部状态' },
    { value: 'running', label: '进行中' },
    { value: 'ended', label: '已结束' },
    { value: 'frozen', label: '冻结中' },
  ] as const
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
  const contestSummary = computed(() => ({
    totalCount: total.value,
    runningCount: runningCount.value,
    exportReadyCount: exportReadyCount.value,
  }))
  const hasActiveFilters = computed(() =>
    Boolean(filters.value.status || filters.value.keyword.trim())
  )
  const reviewRows = computed<PlatformAwdReviewRow[]>(() =>
    contests.value.map((contest) => ({
      ...contest,
      contestCode: `AWD-${contest.id}`,
    }))
  )

  async function loadContests(): Promise<void> {
    const requestId = ++latestRequestId
    const controller = createController()
    loading.value = true
    error.value = null

    try {
      const nextPage = await listTeacherAWDReviews({
        status: filters.value.status || undefined,
        keyword: filters.value.keyword.trim() || undefined,
        page: page.value,
        page_size: pageSize.value,
      }, {
        signal: controller.signal,
      })
      if (requestId !== latestRequestId) {
        return
      }
      contests.value = nextPage.list
      total.value = nextPage.total
      page.value = nextPage.page
      pageSize.value = nextPage.page_size
      runningCount.value = nextPage.summary.running_count
      exportReadyCount.value = nextPage.summary.export_ready_count
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
      console.error('加载 AWD 复盘目录失败:', err)
      contests.value = []
      total.value = 0
      runningCount.value = 0
      exportReadyCount.value = 0
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

  function openContest(contestId: string): void {
    router.push({
      name: resolveAwdReviewDetailRouteName(authStore.user?.role),
      params: { contestId },
    })
  }

  function openDashboard(): void {
    router.push({ name: 'TeacherDashboard' })
  }

  function openPlatformOverview(): void {
    router.push({ name: 'PlatformOverview' })
  }

  function resetFilters(): void {
    filters.value.status = ''
    filters.value.keyword = ''
  }

  function contestStatusLabel(status: string): string {
    switch (status) {
      case 'running':
        return '进行中'
      case 'ended':
        return '已结束'
      case 'frozen':
        return '冻结中'
      case 'published':
        return '已发布'
      default:
        return status || '未开始'
    }
  }

  onMounted(() => {
    void loadContests()
  })

  watch(
    () => [filters.value.status, filters.value.keyword],
    () => {
      page.value = 1
      scheduleContestSearch()
    }
  )

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
    hasContests,
    statusOptions,
    contestSummary,
    hasActiveFilters,
    reviewRows,
    loadContests,
    changePage,
    resetFilters,
    openDashboard,
    openPlatformOverview,
    openContest,
    contestStatusLabel,
  }
}

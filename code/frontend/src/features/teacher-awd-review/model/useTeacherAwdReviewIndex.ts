import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { useRouter } from 'vue-router'

import { listTeacherAWDReviews } from '@/api/teacher'
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { resolveAwdReviewDetailRouteName } from '@/utils/teachingWorkspaceRouting'

export function useTeacherAwdReviewIndex() {
  const router = useRouter()
  const authStore = useAuthStore()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const contests = ref<TeacherAWDReviewContestItemData[]>([])
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
  const contestSummary = computed(() => ({
    totalCount: contests.value.length,
    runningCount: contests.value.filter((item) => item.status === 'running').length,
    exportReadyCount: contests.value.filter((item) => item.export_ready).length,
  }))

  async function loadContests(): Promise<void> {
    const requestId = ++latestRequestId
    loading.value = true
    error.value = null

    try {
      const nextContests = await listTeacherAWDReviews({
        status: filters.value.status || undefined,
        keyword: filters.value.keyword.trim() || undefined,
      })
      if (requestId !== latestRequestId) {
        return
      }
      contests.value = nextContests
    } catch (err) {
      if (requestId !== latestRequestId) {
        return
      }
      console.error('加载 AWD 复盘目录失败:', err)
      contests.value = []
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
      scheduleContestSearch()
    }
  )

  onUnmounted(() => {
    scheduleContestSearch.cancel?.()
  })

  return {
    router,
    loading,
    error,
    contests,
    filters,
    hasContests,
    statusOptions,
    contestSummary,
    loadContests,
    openContest,
    contestStatusLabel,
  }
}

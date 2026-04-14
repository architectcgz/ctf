import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { useRouter } from 'vue-router'

import { listTeacherAWDReviews } from '@/api/teacher'
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'

export function useTeacherAwdReviewIndex() {
  const router = useRouter()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const contests = ref<TeacherAWDReviewContestItemData[]>([])
  const filters = ref({
    status: '' as '' | TeacherAWDReviewContestItemData['status'],
    keyword: '',
  })
  let latestRequestId = 0

  const hasContests = computed(() => contests.value.length > 0)

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
      name: 'TeacherAWDReviewDetail',
      params: { contestId },
    })
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
    loadContests,
    openContest,
  }
}

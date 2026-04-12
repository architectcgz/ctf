import { computed, onMounted, ref } from 'vue'
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

  const hasContests = computed(() => contests.value.length > 0)

  async function loadContests(): Promise<void> {
    loading.value = true
    error.value = null

    try {
      contests.value = await listTeacherAWDReviews({
        status: filters.value.status || undefined,
        keyword: filters.value.keyword.trim() || undefined,
      })
    } catch (err) {
      console.error('加载 AWD 复盘目录失败:', err)
      contests.value = []
      error.value = '加载 AWD 复盘目录失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  function openContest(contestId: string): void {
    router.push({
      name: 'TeacherAWDReviewDetail',
      params: { contestId },
    })
  }

  onMounted(() => {
    void loadContests()
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

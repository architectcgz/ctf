import { ref, type ComputedRef, type Ref } from 'vue'

import { getContest } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'

export function useContestOperationsData(contestId: Ref<string> | ComputedRef<string>) {
  const loading = ref(true)
  const contest = ref<ContestDetailData | null>(null)
  const loadError = ref('')

  async function loadContest() {
    if (!contestId.value) {
      contest.value = null
      loadError.value = ''
      loading.value = false
      return null
    }

    loading.value = true
    loadError.value = ''
    try {
      const result = await getContest(contestId.value)
      contest.value = result
      return result
    } catch (error) {
      contest.value = null
      loadError.value = error instanceof Error ? error.message : '加载竞赛信息失败'
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    contest,
    loadError,
    loadContest,
  }
}

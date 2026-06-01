import { onMounted, ref, type ComputedRef, type Ref } from 'vue'

import { getContest } from '@/api/admin/contest-manage'
import type { ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

export function useContestAnnouncementsData(
  contestId: Ref<string> | ComputedRef<string>,
  loadAnnouncements: () => Promise<unknown>
) {
  const contest = ref<ContestDetailData | null>(null)
  const loading = ref(true)
  const loadError = ref('')

  async function loadPage(): Promise<void> {
    if (!contestId.value) {
      contest.value = null
      loadError.value = '缺少竞赛编号。'
      loading.value = false
      return
    }

    loading.value = true
    loadError.value = ''
    try {
      contest.value = await getContest(contestId.value)
      await loadAnnouncements()
    } catch (error) {
      loadError.value = humanizeRequestError(error, '竞赛公告加载失败')
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadPage()
  })

  return {
    contest,
    loading,
    loadError,
    loadPage,
  }
}

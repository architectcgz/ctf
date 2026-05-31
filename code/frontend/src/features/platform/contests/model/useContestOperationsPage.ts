import { computed, onMounted, onUnmounted, ref, type ComputedRef, type Ref } from 'vue'

import { getContest } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'
import { useBackofficeBreadcrumbDetail } from '@/shared/model/layout/useBackofficeBreadcrumbDetail'
import { useToast } from '@/shared/model/common/useToast'

export function useContestOperationsPage(contestId: Ref<string> | ComputedRef<string>) {
  const toast = useToast()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const loading = ref(true)
  const contest = ref<ContestDetailData | null>(null)
  const runtimeStageReady = computed(
    () =>
      contest.value?.status === 'running' ||
      contest.value?.status === 'frozen' ||
      contest.value?.status === 'ended'
  )
  const inspectorRuntimeContent = computed(() =>
    runtimeStageReady.value ? 'round-inspector' : 'readiness'
  )

  async function loadContest() {
    if (!contestId.value) {
      loading.value = false
      setBreadcrumbDetailTitle()
      return
    }
    loading.value = true
    try {
      contest.value = await getContest(contestId.value)
      setBreadcrumbDetailTitle(contest.value.title)
    } catch {
      setBreadcrumbDetailTitle()
      toast.error('加载竞赛信息失败')
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadContest()
  })

  onUnmounted(() => {
    setBreadcrumbDetailTitle()
  })

  return {
    loading,
    contest,
    inspectorRuntimeContent,
  }
}

import { computed, onMounted, onUnmounted, type ComputedRef, type Ref } from 'vue'

import { useBackofficeBreadcrumbDetail } from '@/shared/model/layout/useBackofficeBreadcrumbDetail'
import { useToast } from '@/shared/model/common/useToast'
import { useContestOperationsData } from './useContestOperationsData'

export function useContestOperationsPage(contestId: Ref<string> | ComputedRef<string>) {
  const toast = useToast()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()
  const contestData = useContestOperationsData(contestId)
  const loading = contestData.loading
  const contest = contestData.contest
  const runtimeStageReady = computed(
    () =>
      contest.value?.status === 'running' ||
      contest.value?.status === 'frozen' ||
      contest.value?.status === 'ended'
  )
  const inspectorRuntimeContent = computed(() =>
    runtimeStageReady.value ? 'round-inspector' : 'readiness'
  )

  async function initialize(): Promise<void> {
    const loadedContest = await contestData.loadContest()
    if (loadedContest) {
      setBreadcrumbDetailTitle(loadedContest.title)
      return
    }

    setBreadcrumbDetailTitle()
    if (contestId.value && contestData.loadError.value) {
      toast.error('加载竞赛信息失败')
    }
  }

  onMounted(() => {
    void initialize()
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

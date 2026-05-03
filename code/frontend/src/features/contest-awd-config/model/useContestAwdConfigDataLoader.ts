import { ref, type Ref } from 'vue'

import { getContest, listContestAWDServices } from '@/api/admin/contests'
import type { AdminContestAWDServiceData, ContestDetailData } from '@/api/contracts'

interface UseContestAwdConfigDataLoaderOptions {
  contestId: Readonly<Ref<string>>
  setBreadcrumbDetailTitle: (title?: string) => void
}

export function useContestAwdConfigDataLoader(options: UseContestAwdConfigDataLoaderOptions) {
  const { contestId, setBreadcrumbDetailTitle } = options

  const loading = ref(true)
  const refreshing = ref(false)
  const loadError = ref('')
  const contest = ref<ContestDetailData | null>(null)
  const services = ref<AdminContestAWDServiceData[]>([])

  let loadVersion = 0
  let afterLoadHandler: (() => void) | null = null

  function setAfterLoadHandler(handler: () => void) {
    afterLoadHandler = handler
  }

  async function loadPage(initial = false) {
    if (!contestId.value) return
    const version = ++loadVersion
    if (initial) loading.value = true
    refreshing.value = !initial
    try {
      const [contestDetail, serviceList] = await Promise.all([
        getContest(contestId.value),
        listContestAWDServices(contestId.value),
      ])
      if (version !== loadVersion) return
      contest.value = contestDetail
      services.value = serviceList
      setBreadcrumbDetailTitle(contestDetail.title)
      afterLoadHandler?.()
      loadError.value = ''
    } catch (error) {
      if (version !== loadVersion) return
      loadError.value =
        error instanceof Error && error.message.trim() ? error.message : 'AWD 配置加载失败'
    } finally {
      if (version === loadVersion) {
        loading.value = false
        refreshing.value = false
      }
    }
  }

  function clearBreadcrumbDetailTitle() {
    setBreadcrumbDetailTitle()
  }

  return {
    clearBreadcrumbDetailTitle,
    contest,
    loadError,
    loading,
    loadPage,
    refreshing,
    services,
    setAfterLoadHandler,
  }
}

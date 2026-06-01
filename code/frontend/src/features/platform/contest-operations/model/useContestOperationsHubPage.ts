import { computed, onMounted, ref } from 'vue'

import { getContests } from '@/api/admin/contest-manage'
import { usePagination } from '@/shared/model/common/usePagination'
import { useAbortController } from '@/shared/lib/request/useAbortController'
import type { ContestDetailData, ContestListSummaryData, ContestPageData } from '@/api/contracts'
import { buildContestManageListRoute, buildContestOperationsRoute } from './contestOperationsHubRoutes'

const OPERABLE_AWD_STATUSES = ['registering', 'running', 'frozen', 'ended'] as const
const PREFERRED_AWD_STATUSES = ['running', 'frozen'] as const

function buildFallbackSummary(contests: ContestDetailData[]): ContestListSummaryData {
  return {
    draft_count: 0,
    registering_count: contests.filter((contest) => contest.status === 'registering').length,
    running_count: contests.filter((contest) => contest.status === 'running').length,
    frozen_count: contests.filter((contest) => contest.status === 'frozen').length,
    ended_count: contests.filter((contest) => contest.status === 'ended').length,
  }
}

export function useContestOperationsHubPage() {
  const {
    list,
    total,
    page,
    pageSize,
    loading,
    error,
    response,
    changePage,
    refresh: refreshDirectory,
  } = usePagination<ContestDetailData, ContestPageData<ContestDetailData>>(({ page, page_size, signal }) =>
    getContests(
      {
        page,
        page_size,
        mode: 'awd',
        statuses: [...OPERABLE_AWD_STATUSES],
        sort_key: 'start_time',
        sort_order: 'desc',
      },
      { signal }
    )
  )
  const preferredContest = ref<ContestDetailData | null>(null)
  const { createController } = useAbortController()
  let preferredRequestToken = 0

  const contestSummary = computed(() => response.value?.summary ?? buildFallbackSummary(list.value))
  const loadError = computed(() => {
    if (!error.value) {
      return ''
    }
    return error.value instanceof Error ? error.value.message : '赛事运维目录加载失败'
  })
  const operableContests = computed(() => list.value)
  const runningContestCount = computed(() => contestSummary.value.running_count)
  const frozenContestCount = computed(() => contestSummary.value.frozen_count)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))

  async function refreshPreferredContest(): Promise<void> {
    const requestToken = ++preferredRequestToken
    const controller = createController()

    try {
      const preferredActive = await getContests(
        {
          page: 1,
          page_size: 1,
          mode: 'awd',
          statuses: [...PREFERRED_AWD_STATUSES],
          sort_key: 'start_time',
          sort_order: 'desc',
        },
        { signal: controller.signal }
      )
      if (requestToken !== preferredRequestToken) {
        return
      }
      if (preferredActive.list.length > 0) {
        preferredContest.value = preferredActive.list[0]
        return
      }

      const fallback = await getContests(
        {
          page: 1,
          page_size: 1,
          mode: 'awd',
          statuses: [...OPERABLE_AWD_STATUSES],
          sort_key: 'start_time',
          sort_order: 'desc',
        },
        { signal: controller.signal }
      )
      if (requestToken !== preferredRequestToken) {
        return
      }
      preferredContest.value = fallback.list[0] ?? null
    } catch {
      if (requestToken !== preferredRequestToken) {
        return
      }
      preferredContest.value = null
    }
  }

  async function loadContests(): Promise<void> {
    await Promise.all([refreshDirectory(), refreshPreferredContest()])
  }

  async function changeContestPage(nextPage: number): Promise<void> {
    await changePage(nextPage)
  }

  onMounted(() => {
    void loadContests()
  })

  return {
    changeContestPage,
    backToContestDirectoryRoute: buildContestManageListRoute(),
    buildContestOperationsRoute,
    loadContests,
    loadError,
    loading,
    operableContests,
    page,
    pageSize,
    preferredContest,
    runningContestCount,
    frozenContestCount,
    total,
    totalPages,
  }
}

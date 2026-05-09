import { computed, ref, watch } from 'vue'

import { getContests } from '@/api/admin/contests'
import type { ContestDetailData, ContestListSummaryData, ContestPageData } from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'

import type { PlatformContestStatus } from './contestFormSupport'

export type ContestStatusFilter = 'all' | PlatformContestStatus

function buildFallbackSummary(contests: ContestDetailData[]): ContestListSummaryData {
  return {
    draft_count: contests.filter((contest) => contest.status === 'draft').length,
    registering_count: contests.filter((contest) => contest.status === 'registering').length,
    running_count: contests.filter((contest) => contest.status === 'running').length,
    frozen_count: contests.filter((contest) => contest.status === 'frozen').length,
    ended_count: contests.filter((contest) => contest.status === 'ended').length,
  }
}

export function useContestListState() {
  const statusFilter = ref<ContestStatusFilter>('all')
  const pagination = usePagination<ContestDetailData, ContestPageData<ContestDetailData>>(
    ({ page, page_size, signal }) =>
      getContests(
        {
          page,
          page_size,
          status: statusFilter.value === 'all' ? undefined : statusFilter.value,
        },
        { signal }
      )
  )
  const summary = computed(() => pagination.response.value?.summary ?? buildFallbackSummary(pagination.list.value))

  watch(statusFilter, async () => {
    await pagination.changePage(1)
  })

  return {
    pagination,
    summary,
    statusFilter,
  }
}

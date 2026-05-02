import { ref, watch } from 'vue'

import { getContests } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'

import type { PlatformContestStatus } from './contestFormSupport'

export type ContestStatusFilter = 'all' | PlatformContestStatus

export function useContestListState() {
  const statusFilter = ref<ContestStatusFilter>('all')
  const pagination = usePagination<ContestDetailData>(({ page, page_size }) =>
    getContests({
      page,
      page_size,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
    })
  )

  watch(statusFilter, async () => {
    await pagination.changePage(1)
  })

  return {
    pagination,
    statusFilter,
  }
}

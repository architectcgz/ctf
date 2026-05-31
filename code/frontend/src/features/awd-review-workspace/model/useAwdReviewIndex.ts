import { computed, onMounted, watch } from 'vue'
import type { AwdReviewContestItemData } from '@/api/contracts'

import { useAwdReviewDirectory } from './useAwdReviewDirectory'

export interface PlatformAwdReviewRow extends AwdReviewContestItemData {
  contestCode: string
}

export function useAwdReviewIndex() {
  const {
    loading,
    error,
    contests,
    total,
    page,
    totalPages,
    filters,
    summary,
    loadContests,
    changePage,
    triggerFilterRefresh,
  } = useAwdReviewDirectory()
  const hasContests = computed(() => contests.value.length > 0)
  const statusOptions = [
    { value: '', label: '全部状态' },
    { value: 'running', label: '进行中' },
    { value: 'ended', label: '已结束' },
    { value: 'frozen', label: '冻结中' },
  ] as const
  const contestSummary = computed(() => ({
    totalCount: total.value,
    runningCount: summary.value.running_count,
    exportReadyCount: summary.value.export_ready_count,
  }))
  const hasActiveFilters = computed(() =>
    Boolean(filters.value.status || filters.value.keyword.trim())
  )
  const reviewRows = computed<PlatformAwdReviewRow[]>(() =>
    contests.value.map((contest) => ({
      ...contest,
      contestCode: `AWD-${contest.id}`,
    }))
  )

  function resetFilters(): void {
    filters.value.status = ''
    filters.value.keyword = ''
  }

  function contestStatusLabel(status: string): string {
    switch (status) {
      case 'running':
        return '进行中'
      case 'ended':
        return '已结束'
      case 'frozen':
        return '冻结中'
      case 'published':
        return '已发布'
      default:
        return status || '未开始'
    }
  }

  onMounted(() => {
    void loadContests()
  })

  watch(
    () => [filters.value.status, filters.value.keyword],
    () => {
      triggerFilterRefresh()
    }
  )

  return {
    loading,
    error,
    contests,
    total,
    page,
    totalPages,
    filters,
    hasContests,
    statusOptions,
    contestSummary,
    hasActiveFilters,
    reviewRows,
    loadContests,
    changePage,
    resetFilters,
    contestStatusLabel,
  }
}

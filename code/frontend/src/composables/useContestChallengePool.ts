import { computed, ref, type Ref } from 'vue'

import type { AdminContestChallengeData, ContestMode } from '@/api/contracts'

export type ContestChallengePoolFilter = 'all' | 'unconfigured' | 'validation-failed'

interface ContestChallengePoolSummaryItem {
  key: string
  label: string
  value: string
  hint: string
}

interface ContestChallengePoolFilterItem {
  key: ContestChallengePoolFilter
  label: string
  hint: string
  count: number
}

function hasAwdConfiguration(item: AdminContestChallengeData): boolean {
  return Boolean(item.awd_checker_type)
}

function isValidationFailure(item: AdminContestChallengeData): boolean {
  return item.awd_checker_validation_state === 'failed' || item.awd_checker_validation_state === 'stale'
}

export function useContestChallengePool(
  challengeLinks: Ref<AdminContestChallengeData[]>,
  contestMode: Ref<ContestMode | null>
) {
  const activeFilter = ref<ContestChallengePoolFilter>('all')

  const isAwdContest = computed(() => contestMode.value === 'awd')

  const sortedItems = computed(() =>
    [...challengeLinks.value].sort(
      (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
    )
  )

  const summaryItems = computed<ContestChallengePoolSummaryItem[]>(() => {
    const items: ContestChallengePoolSummaryItem[] = [
      {
        key: 'total',
        label: '已关联题目',
        value: String(challengeLinks.value.length),
        hint: '当前竞赛中已经挂载的题目数量',
      },
      {
        key: 'visible',
        label: '对选手可见',
        value: String(challengeLinks.value.filter((item) => item.is_visible).length),
        hint: '进入竞赛详情后默认可见的题目数量',
      },
      {
        key: 'hidden',
        label: '暂时隐藏',
        value: String(challengeLinks.value.filter((item) => !item.is_visible).length),
        hint: '已关联但不会直接展示给选手的题目数量',
      },
    ]

    if (!isAwdContest.value) {
      return items
    }

    return [
      ...items,
      {
        key: 'configured',
        label: '已配 Checker',
        value: String(sortedItems.value.filter(hasAwdConfiguration).length),
        hint: '已经挂上基础 AWD Checker 的题目数量',
      },
      {
        key: 'attention',
        label: '待处理验证',
        value: String(sortedItems.value.filter((item) => !hasAwdConfiguration(item) || isValidationFailure(item)).length),
        hint: '仍需补齐配置或重新处理验证状态的题目数量',
      },
    ]
  })

  const filterItems = computed<ContestChallengePoolFilterItem[]>(() => {
    if (!isAwdContest.value) {
      return []
    }

    return [
      {
        key: 'all',
        label: '全部',
        hint: '查看当前赛事下的所有已关联题目',
        count: sortedItems.value.length,
      },
      {
        key: 'unconfigured',
        label: '未配置 AWD',
        hint: '还没有挂上 Checker 的题目',
        count: sortedItems.value.filter((item) => !hasAwdConfiguration(item)).length,
      },
      {
        key: 'validation-failed',
        label: '预检失败',
        hint: '最近失败或待重新验证的题目',
        count: sortedItems.value.filter(isValidationFailure).length,
      },
    ]
  })

  const visibleItems = computed(() => {
    switch (activeFilter.value) {
      case 'unconfigured':
        return sortedItems.value.filter((item) => !hasAwdConfiguration(item))
      case 'validation-failed':
        return sortedItems.value.filter(isValidationFailure)
      case 'all':
      default:
        return sortedItems.value
    }
  })

  function setFilter(value: ContestChallengePoolFilter) {
    if (!isAwdContest.value) {
      activeFilter.value = 'all'
      return
    }
    activeFilter.value = value
  }

  return {
    sortedItems,
    visibleItems,
    summaryItems,
    filterItems,
    activeFilter,
    isAwdContest,
    setFilter,
  }
}

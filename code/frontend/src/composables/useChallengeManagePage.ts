import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowDownWideNarrow, ArrowUpNarrowWide, Calendar, SortAsc } from 'lucide-vue-next'

import type { WorkspaceDirectorySortOption } from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useChallengeManagePresentation } from '@/composables/useChallengeManagePresentation'
import { usePlatformChallenges, type PlatformChallengeListRow } from '@/composables/usePlatformChallenges'

type ChallengeSortOption = WorkspaceDirectorySortOption & {
  order: 'asc' | 'desc'
}

const sortOptions: ChallengeSortOption[] = [
  { key: 'updateTime', order: 'desc', label: '最近更新', icon: Calendar },
  { key: 'points', order: 'desc', label: '分值由高到低', icon: ArrowDownWideNarrow },
  { key: 'points', order: 'asc', label: '分值由低到高', icon: ArrowUpNarrowWide },
  { key: 'title', order: 'asc', label: '标题 A-Z', icon: SortAsc },
]

export function useChallengeManagePage() {
  const router = useRouter()
  const {
    list,
    total,
    page,
    pageSize,
    loading,
    error,
    keyword,
    categoryFilter,
    difficultyFilter,
    statusFilter,
    clearFilters,
    changePage,
    refresh,
    publish,
    remove,
  } = usePlatformChallenges()

  const sortConfig = ref<ChallengeSortOption>(sortOptions[0]!)
  const publishedCount = computed(() => list.value.filter((item) => item.status === 'published').length)
  const draftCount = computed(() => list.value.filter((item) => item.status === 'draft').length)
  const archivedCount = computed(() => list.value.filter((item) => item.status === 'archived').length)
  const hasActiveFilters = computed(() =>
    Boolean(
      keyword.value.trim() || categoryFilter.value || difficultyFilter.value || statusFilter.value
    )
  )
  const manageEmptyTitle = computed(() => (hasActiveFilters.value ? '没有匹配题目' : '暂无题目'))
  const manageEmptyMessage = computed(() =>
    hasActiveFilters.value
      ? '当前筛选条件下没有匹配题目。'
      : '当前还没有题目，请先前往导入页上传题目包。'
  )
  const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
  const loadErrorMessage = computed(() =>
    error.value instanceof Error && error.value.message.trim().length > 0
      ? error.value.message
      : '题目目录暂时无法加载，请稍后重试。'
  )
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const sortedChallenges = computed<PlatformChallengeListRow[]>(() => {
    const nextRows = [...list.value]

    nextRows.sort((left, right) => {
      switch (sortConfig.value.key) {
        case 'points': {
          const delta = left.points - right.points
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
        case 'title': {
          const delta = left.title.localeCompare(right.title, 'zh-CN')
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
        case 'updateTime':
        default: {
          const leftTime = new Date(left.updated_at || left.created_at).getTime()
          const rightTime = new Date(right.updated_at || right.created_at).getTime()
          return sortConfig.value.order === 'asc' ? leftTime - rightTime : rightTime - leftTime
        }
      }
    })

    return nextRows
  })

  const {
    openActionMenuId,
    getCategoryLabel,
    getDifficultyLabel,
    closeActionMenu,
    openChallengeDetail,
    openChallengeTopology,
    openChallengeWriteup,
    submitPublishCheck,
    removeChallenge,
  } = useChallengeManagePresentation({
    router,
    publish,
    remove,
  })

  function setSort(option: WorkspaceDirectorySortOption) {
    const matchedOption =
      sortOptions.find((item) => item.key === option.key && item.label === option.label) ??
      sortOptions[0]

    if (!matchedOption) {
      return
    }

    sortConfig.value = matchedOption
  }

  function setActionMenuOpen(challengeId: string, nextOpen: boolean): void {
    if (nextOpen) {
      openActionMenuId.value = challengeId
      return
    }

    if (openActionMenuId.value === challengeId) {
      closeActionMenu()
    }
  }

  async function openImportWorkspace(): Promise<void> {
    await router.push({ name: 'PlatformChallengeImportManage' })
  }

  return {
    archivedCount,
    categoryFilter,
    changePage,
    clearFilters,
    difficultyFilter,
    draftCount,
    getCategoryLabel,
    getDifficultyLabel,
    hasActiveFilters,
    hasLoadError,
    keyword,
    loadErrorMessage,
    loading,
    manageEmptyMessage,
    manageEmptyTitle,
    openActionMenuId,
    openChallengeDetail,
    openChallengeTopology,
    openChallengeWriteup,
    openImportWorkspace,
    page,
    publishedCount,
    refresh,
    removeChallenge,
    selectedSortLabel: computed(() => sortConfig.value.label),
    setActionMenuOpen,
    setSort,
    sortOptions,
    sortedChallenges,
    statusFilter,
    submitPublishCheck,
    total,
    totalPages,
  }
}

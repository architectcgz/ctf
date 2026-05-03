import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getChallenges } from '@/api/challenge'
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import { ApiError } from '@/api/request'
import { usePagination } from '@/composables/usePagination'

const validCategories = [
  'web',
  'pwn',
  'reverse',
  'crypto',
  'misc',
  'forensics',
] satisfies ChallengeCategory[]

const validDifficulties = [
  'beginner',
  'easy',
  'medium',
  'hard',
  'insane',
] satisfies ChallengeDifficulty[]

function getQueryValue(value: unknown): string {
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return ''
}

function parseCategoryFilter(value: unknown): ChallengeCategory | '' {
  const queryValue = getQueryValue(value)
  return validCategories.includes(queryValue as ChallengeCategory)
    ? (queryValue as ChallengeCategory)
    : ''
}

function parseDifficultyFilter(value: unknown): ChallengeDifficulty | '' {
  const queryValue = getQueryValue(value)
  return validDifficulties.includes(queryValue as ChallengeDifficulty)
    ? (queryValue as ChallengeDifficulty)
    : ''
}

export function useChallengeListPage() {
  const route = useRoute()
  const router = useRouter()
  const searchQuery = ref('')
  const categoryFilter = ref<ChallengeCategory | ''>('')
  const difficultyFilter = ref<ChallengeDifficulty | ''>('')

  async function syncFilterQuery(): Promise<boolean> {
    const nextQuery = { ...route.query }

    if (categoryFilter.value) nextQuery.category = categoryFilter.value
    else delete nextQuery.category

    if (difficultyFilter.value) nextQuery.difficulty = difficultyFilter.value
    else delete nextQuery.difficulty

    if (
      getQueryValue(route.query.category) === getQueryValue(nextQuery.category) &&
      getQueryValue(route.query.difficulty) === getQueryValue(nextQuery.difficulty)
    ) {
      return false
    }

    await router.replace({ query: nextQuery })
    return true
  }

  const { list, total, page, pageSize, loading, error, changePage, refresh } = usePagination(
    (params) => {
      const filters: Record<string, unknown> = { ...params }
      if (searchQuery.value) filters.keyword = searchQuery.value
      if (categoryFilter.value) filters.category = categoryFilter.value
      if (difficultyFilter.value) filters.difficulty = difficultyFilter.value
      return getChallenges(filters)
    }
  )

  const hasActiveFilters = computed(() =>
    Boolean(searchQuery.value || categoryFilter.value || difficultyFilter.value)
  )
  const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
  const errorMessage = computed(() => {
    if (error.value instanceof ApiError) return error.value.message
    if (error.value instanceof Error) return error.value.message
    return '题目列表暂时无法加载，请稍后重试。'
  })
  const emptyTitle = computed(() => (hasActiveFilters.value ? '没有匹配的题目' : '目前还没有题目'))
  const emptyDescription = computed(() =>
    hasActiveFilters.value
      ? '当前筛选条件下没有找到可训练的题目，建议放宽分类、难度或搜索词。'
      : '管理员还没有发布训练题目，稍后再来查看即可。'
  )
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
  const solvedCount = computed(() => list.value.filter((challenge) => challenge.is_solved).length)
  const unsolvedCount = computed(() => list.value.filter((challenge) => !challenge.is_solved).length)
  const summaryStats = computed(() => [
    { key: 'total', label: '题目总数', value: total.value, helper: '当前可训练题库规模' },
    {
      key: 'visible',
      label: '当前结果',
      value: list.value.length,
      helper: hasActiveFilters.value ? '已按筛选条件收束范围' : '当前页展示结果数',
    },
    { key: 'solved', label: '已解出', value: solvedCount.value, helper: '当前列表内已完成题目' },
    { key: 'unsolved', label: '待攻克', value: unsolvedCount.value, helper: '仍可直接进入训练' },
  ])

  function onSearch(): void {
    page.value = 1
    void refresh()
  }

  function onFilterChange(): void {
    page.value = 1
    void syncFilterQuery().then((updated) => {
      if (!updated) {
        void refresh()
      }
    })
  }

  function resetFilters(): void {
    searchQuery.value = ''
    categoryFilter.value = ''
    difficultyFilter.value = ''
    page.value = 1
    void syncFilterQuery().then((updated) => {
      if (!updated) {
        void refresh()
      }
    })
  }

  function goToDashboard(): void {
    void router.push({ name: 'Dashboard' })
  }

  function openSkillProfile(): void {
    void router.push({ name: 'SkillProfile' })
  }

  function goToDetail(id: string): void {
    void router.push(`/challenges/${id}`)
  }

  watch(
    () => [route.query.category, route.query.difficulty] as const,
    ([nextCategoryQuery, nextDifficultyQuery], previousQuery) => {
      const nextCategory = parseCategoryFilter(nextCategoryQuery)
      const nextDifficulty = parseDifficultyFilter(nextDifficultyQuery)
      const previousCategory = parseCategoryFilter(previousQuery?.[0])
      const previousDifficulty = parseDifficultyFilter(previousQuery?.[1])

      categoryFilter.value = nextCategory
      difficultyFilter.value = nextDifficulty

      if (
        previousQuery === undefined ||
        previousCategory !== nextCategory ||
        previousDifficulty !== nextDifficulty
      ) {
        page.value = 1
        void refresh()
      }
    },
    { immediate: true }
  )

  return {
    list,
    total,
    page,
    loading,
    searchQuery,
    categoryFilter,
    difficultyFilter,
    hasActiveFilters,
    hasLoadError,
    errorMessage,
    emptyTitle,
    emptyDescription,
    totalPages,
    summaryStats,
    changePage,
    refresh,
    onSearch,
    onFilterChange,
    resetFilters,
    goToDashboard,
    openSkillProfile,
    goToDetail,
  }
}

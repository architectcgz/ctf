import { computed, ref } from 'vue'

import { getContests, type GetContestsData } from '@/api/contest'
import { getPracticeRanking } from '@/api/scoreboard'
import { usePagination } from '@/composables/usePagination'
import type {
  ContestListItem,
  ContestMode,
  ContestListSummaryData,
  ContestStatus,
  PracticeRankingItemData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface ScoreboardSection {
  contest: ContestListItem
  frozen: boolean
}

const SCOREBOARD_CONTEST_STATUSES: ContestStatus[] = ['running', 'frozen', 'ended']
type ScoreboardStatusFilter = '' | 'running' | 'frozen' | 'ended'
type ScoreboardModeFilter = '' | Extract<ContestMode, 'jeopardy' | 'awd'>

function buildFallbackSummary(contests: ContestListItem[]): ContestListSummaryData {
  return {
    draft_count: 0,
    registering_count: 0,
    running_count: contests.filter((contest) => contest.status === 'running').length,
    frozen_count: contests.filter((contest) => contest.status === 'frozen').length,
    ended_count: contests.filter((contest) => contest.status === 'ended').length,
  }
}

export function useScoreboardView() {
  const toast = useToast()
  const contestStatusFilter = ref<ScoreboardStatusFilter>('')
  const contestModeFilter = ref<ScoreboardModeFilter>('')
  const activeContestStatuses = computed<ContestStatus[]>(() =>
    contestStatusFilter.value ? [contestStatusFilter.value] : SCOREBOARD_CONTEST_STATUSES
  )
  const {
    list,
    total,
    page,
    pageSize,
    loading,
    error,
    response,
    changePage,
    refresh: refreshContestDirectory,
  } = usePagination<ContestListItem, GetContestsData>(({ page, page_size, signal }) =>
    getContests(
      {
        page,
        page_size,
        statuses: activeContestStatuses.value,
        ...(contestModeFilter.value ? { mode: contestModeFilter.value } : {}),
        sort_key: 'start_time',
        sort_order: 'desc',
      },
      { signal }
    )
  )

  const rankingRows = ref<PracticeRankingItemData[]>([])
  const rankingLoading = ref(false)
  const rankingError = ref(false)
  const rankingHint = ref('展示全站练习积分排行榜，按积分高低排序。')
  const sections = computed<ScoreboardSection[]>(() =>
    list.value.map((contest) => ({
      contest,
      frozen: Boolean(contest.scoreboard_frozen) || contest.status === 'frozen',
    }))
  )
  const contestSummary = computed(() => response.value?.summary ?? buildFallbackSummary(list.value))
  const contestTotalPages = computed(() =>
    Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1)))
  )
  const contestPageStartIndex = computed(() => (page.value - 1) * pageSize.value)
  const selectionHint = computed(() => {
    if (error.value) {
      return '竞赛列表加载失败，请稍后重试。'
    }
    if (!sections.value.length) {
      return '当前没有进行中或已结束的竞赛排行榜。'
    }
    return '按竞赛开始时间倒序展示可查看排行榜，点击竞赛进入完整排行。'
  })

  const hasSections = computed(() => sections.value.length > 0)
  const hasRankingRows = computed(() => rankingRows.value.length > 0)

  async function refreshPracticeRanking(): Promise<void> {
    rankingLoading.value = true
    rankingError.value = false

    try {
      rankingRows.value = await getPracticeRanking({ limit: 100 })
      rankingHint.value = rankingRows.value.length
        ? '展示全站练习积分排行榜，按积分高低排序。'
        : '当前还没有可展示的积分排行榜数据。'
    } catch {
      rankingRows.value = []
      rankingError.value = true
      rankingHint.value = '积分排行榜加载失败，请稍后重试。'
      toast.warning('积分排行榜加载失败')
    } finally {
      rankingLoading.value = false
    }
  }

  async function refresh(): Promise<void> {
    await Promise.all([refreshContestDirectory(), refreshPracticeRanking()])
  }

  async function changeContestPage(nextPage: number): Promise<void> {
    await changePage(nextPage)
  }

  async function applyContestFilters(): Promise<void> {
    await changePage(1)
  }

  async function updateContestStatusFilter(value: ScoreboardStatusFilter): Promise<void> {
    contestStatusFilter.value = value
    await applyContestFilters()
  }

  async function updateContestModeFilter(value: ScoreboardModeFilter): Promise<void> {
    contestModeFilter.value = value
    await applyContestFilters()
  }

  async function resetContestFilters(): Promise<void> {
    contestStatusFilter.value = ''
    contestModeFilter.value = ''
    await applyContestFilters()
  }

  void refresh()

  return {
    contestPage: page,
    contestPageSize: pageSize,
    contestPageStartIndex,
    contestSummary,
    contestTotal: total,
    contestTotalPages,
    changeContestPage,
    contestStatusFilter,
    contestModeFilter,
    updateContestStatusFilter,
    updateContestModeFilter,
    resetContestFilters,
    hasSections,
    hasRankingRows,
    loading,
    rankingError,
    rankingHint,
    rankingLoading,
    rankingRows,
    refresh,
    refreshPracticeRanking,
    sections,
    selectionHint,
  }
}

import { computed, ref } from 'vue'

import { getContests, getScoreboard } from '@/api/contest'
import { getPracticeRanking } from '@/api/scoreboard'
import type { ContestListItem, PracticeRankingItemData, ScoreboardRow } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface ScoreboardSection {
  contest: ContestListItem
  frozen: boolean
  rows: ScoreboardRow[]
  error: boolean
}

const DISPLAYABLE_CONTEST_STATUSES = new Set(['running', 'frozen', 'ended'])

function toTimestamp(value: string): number {
  const parsed = Date.parse(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function sortContestsByLatest(contests: ContestListItem[]): ContestListItem[] {
  return [...contests].sort(
    (left, right) => toTimestamp(right.starts_at) - toTimestamp(left.starts_at)
  )
}

export function useScoreboardView() {
  const toast = useToast()

  const sections = ref<ScoreboardSection[]>([])
  const loading = ref(false)
  const selectionHint = ref('按竞赛开始时间倒序展示排行榜，最新的竞赛排在最前面。')
  const rankingRows = ref<PracticeRankingItemData[]>([])
  const rankingLoading = ref(false)
  const rankingError = ref(false)
  const rankingHint = ref('展示全站练习积分排行榜，按积分高低排序。')
  const refreshingContestIDs = new Set<string>()

  const hasSections = computed(() => sections.value.length > 0)
  const hasRankingRows = computed(() => rankingRows.value.length > 0)

  async function loadScoreboardSection(contest: ContestListItem): Promise<ScoreboardSection> {
    try {
      const payload = await getScoreboard(contest.id, { page: 1, page_size: 100 })
      return {
        contest,
        frozen: payload.frozen,
        rows: payload.scoreboard.list,
        error: false,
      }
    } catch (error) {
      return {
        contest,
        frozen: false,
        rows: [],
        error: true,
      }
    }
  }

  function replaceSection(nextSection: ScoreboardSection): void {
    sections.value = sections.value.map((section) =>
      section.contest.id === nextSection.contest.id ? nextSection : section
    )
  }

  async function refreshContestScoreboard(contestId: string): Promise<void> {
    const currentSection = sections.value.find((section) => section.contest.id === contestId)
    if (!currentSection || refreshingContestIDs.has(contestId)) {
      return
    }

    refreshingContestIDs.add(contestId)

    try {
      const payload = await getScoreboard(contestId, { page: 1, page_size: 100 })
      replaceSection({
        contest: currentSection.contest,
        frozen: payload.frozen,
        rows: payload.scoreboard.list,
        error: false,
      })
    } catch (error) {
      replaceSection({
        ...currentSection,
        error: true,
      })
    } finally {
      refreshingContestIDs.delete(contestId)
    }
  }

  async function refreshPracticeRanking(): Promise<void> {
    rankingLoading.value = true
    rankingError.value = false

    try {
      rankingRows.value = await getPracticeRanking({ limit: 100 })
      rankingHint.value = rankingRows.value.length
        ? '展示全站练习积分排行榜，按积分高低排序。'
        : '当前还没有可展示的积分排行榜数据。'
    } catch (error) {
      rankingRows.value = []
      rankingError.value = true
      rankingHint.value = '积分排行榜加载失败，请稍后重试。'
      toast.warning('积分排行榜加载失败')
    } finally {
      rankingLoading.value = false
    }
  }

  async function refresh(): Promise<void> {
    loading.value = true
    const contestTask = (async () => {
      try {
        const payload = await getContests({ page: 1, page_size: 100 })
        const contests = sortContestsByLatest(
          payload.list.filter((item) => DISPLAYABLE_CONTEST_STATUSES.has(item.status))
        )

        if (!contests.length) {
          sections.value = []
          selectionHint.value = '当前没有进行中或已结束的竞赛排行榜。'
          return
        }

        selectionHint.value = '按竞赛开始时间倒序展示排行榜，最新的竞赛排在最前面。'

        const scoreboardSections = await Promise.all(
          contests.map((contest) => loadScoreboardSection(contest))
        )

        sections.value = scoreboardSections

        if (scoreboardSections.some((item) => item.error)) {
          toast.warning('部分竞赛排行榜加载失败')
        }
      } catch (error) {
        sections.value = []
        selectionHint.value = '竞赛列表加载失败，请稍后重试。'
        toast.error('加载竞赛列表失败')
      } finally {
        loading.value = false
      }
    })()

    await Promise.all([contestTask, refreshPracticeRanking()])
  }

  void refresh()

  return {
    hasSections,
    hasRankingRows,
    loading,
    rankingError,
    rankingHint,
    rankingLoading,
    rankingRows,
    refresh,
    refreshContestScoreboard,
    refreshPracticeRanking,
    sections,
    selectionHint,
  }
}

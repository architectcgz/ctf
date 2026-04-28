import { computed, ref } from 'vue'

import { getContests } from '@/api/contest'
import { getPracticeRanking } from '@/api/scoreboard'
import type { ContestListItem, PracticeRankingItemData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface ScoreboardSection {
  contest: ContestListItem
  frozen: boolean
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
  const selectionHint = ref('按竞赛开始时间倒序展示可查看排行榜，点击竞赛进入完整排行。')
  const rankingRows = ref<PracticeRankingItemData[]>([])
  const rankingLoading = ref(false)
  const rankingError = ref(false)
  const rankingHint = ref('展示全站练习积分排行榜，按积分高低排序。')

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

        selectionHint.value = '按竞赛开始时间倒序展示可查看排行榜，点击竞赛进入完整排行。'
        sections.value = contests.map((contest) => ({
          contest,
          frozen: Boolean(contest.scoreboard_frozen) || contest.status === 'frozen',
        }))
      } catch {
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
    refreshPracticeRanking,
    sections,
    selectionHint,
  }
}

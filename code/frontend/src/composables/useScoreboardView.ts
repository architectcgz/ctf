import { computed, ref } from 'vue'

import { getContests, getScoreboard } from '@/api/contest'
import type { ContestListItem, ScoreboardRow } from '@/api/contracts'
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
  return [...contests].sort((left, right) => toTimestamp(right.starts_at) - toTimestamp(left.starts_at))
}

export function useScoreboardView() {
  const toast = useToast()

  const sections = ref<ScoreboardSection[]>([])
  const loading = ref(false)
  const selectionHint = ref('按竞赛开始时间倒序展示排行榜，最新的竞赛排在最前面。')

  const hasSections = computed(() => sections.value.length > 0)

  async function refresh(): Promise<void> {
    loading.value = true

    try {
      const payload = await getContests({ page: 1, size: 100 })
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
        contests.map(async (contest) => {
          try {
            const payload = await getScoreboard(contest.id, { page: 1, page_size: 100 })
            return {
              contest,
              frozen: payload.frozen,
              rows: payload.scoreboard.list,
              error: false,
            } satisfies ScoreboardSection
          } catch (error) {
            return {
              contest,
              frozen: false,
              rows: [],
              error: true,
            } satisfies ScoreboardSection
          }
        })
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
  }

  void refresh()

  return {
    hasSections,
    loading,
    refresh,
    sections,
    selectionHint,
  }
}

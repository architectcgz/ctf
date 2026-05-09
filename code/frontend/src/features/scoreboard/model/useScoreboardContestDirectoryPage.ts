import { computed, type Ref } from 'vue'

import type { ContestListItem, ContestListSummaryData, ContestStatus } from '@/api/contracts'
import { getContestAccentColor } from '@/utils/contest'

interface ScoreboardSectionItem {
  contest: ContestListItem
  frozen: boolean
}

interface UseScoreboardContestDirectoryPageOptions {
  sections: Ref<ScoreboardSectionItem[]>
  contestSummary: Ref<ContestListSummaryData>
  contestPage: Ref<number>
  contestPageSize: Ref<number>
  contestTotal: Ref<number>
  selectionHint: Ref<string>
  rankingError: Ref<boolean>
}

export function useScoreboardContestDirectoryPage(options: UseScoreboardContestDirectoryPageOptions) {
  const contestCount = computed(() => options.contestTotal.value)
  const runningCount = computed(() => options.contestSummary.value.running_count)
  const frozenCount = computed(() => options.contestSummary.value.frozen_count)
  const endedCount = computed(() => options.contestSummary.value.ended_count)
  const contestPageStartIndex = computed(
    () => (options.contestPage.value - 1) * options.contestPageSize.value
  )
  const paginatedSections = computed(() => options.sections.value)
  const emptyTitle = computed(() =>
    options.selectionHint.value.includes('失败') ? '排行榜加载失败' : '暂无可查看的竞赛排行榜'
  )
  const pointsEmptyTitle = computed(() =>
    options.rankingError.value ? '积分排行榜加载失败' : '暂无可查看的积分排行榜'
  )

  function formatDateTime(value?: string): string {
    if (!value) return '未记录'
    return new Date(value).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  function formatContestWindow(startsAt: string, endsAt: string): string {
    return `${formatDateTime(startsAt)} ~ ${formatDateTime(endsAt)}`
  }

  function sectionAccentStyle(status: ContestStatus): Record<string, string> {
    return { '--scoreboard-accent': getContestAccentColor(status) }
  }

  function getRowClass(rank: number): string {
    if (rank === 1) return 'sb-row sb-row--top1'
    if (rank === 2) return 'sb-row sb-row--top2'
    if (rank === 3) return 'sb-row sb-row--top3'
    return 'sb-row'
  }

  function getRankPillClass(rank: number): string[] {
    return [
      'sb-rank-pill',
      rank === 1 ? 'sb-rank-pill--top1' : '',
      rank === 2 ? 'sb-rank-pill--top2' : '',
      rank === 3 ? 'sb-rank-pill--top3' : '',
    ]
  }

  function getCardDescription(status: ContestStatus, frozen: boolean): string {
    if (frozen || status === 'frozen') {
      return '封榜阶段先展示竞赛入口，进入后查看冻结前排名。'
    }

    if (status === 'running') {
      return '进行中竞赛进入详情后支持实时刷新，提交后榜单会自动更新。'
    }

    return '历史竞赛进入详情后展示最终成绩，可用于复盘队伍解题表现。'
  }

  return {
    contestCount,
    runningCount,
    frozenCount,
    endedCount,
    contestPageStartIndex,
    paginatedSections,
    emptyTitle,
    pointsEmptyTitle,
    formatDateTime,
    formatContestWindow,
    sectionAccentStyle,
    getRowClass,
    getRankPillClass,
    getCardDescription,
  }
}

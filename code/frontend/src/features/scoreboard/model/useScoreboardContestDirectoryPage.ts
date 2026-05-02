import { computed, ref, watch, type Ref } from 'vue'

import type { ContestListItem, ContestStatus } from '@/api/contracts'
import { getContestAccentColor } from '@/utils/contest'

interface ScoreboardSectionItem {
  contest: ContestListItem
  frozen: boolean
}

interface UseScoreboardContestDirectoryPageOptions {
  sections: Ref<ScoreboardSectionItem[]>
  selectionHint: Ref<string>
  rankingError: Ref<boolean>
}

const CONTEST_PAGE_SIZE = 6

export function useScoreboardContestDirectoryPage(options: UseScoreboardContestDirectoryPageOptions) {
  const contestCount = computed(() => options.sections.value.length)
  const runningCount = computed(
    () => options.sections.value.filter((section) => section.contest.status === 'running').length
  )
  const frozenCount = computed(() => options.sections.value.filter((section) => section.frozen).length)
  const endedCount = computed(
    () => options.sections.value.filter((section) => section.contest.status === 'ended').length
  )
  const contestPage = ref(1)
  const contestTotalPages = computed(() =>
    Math.max(1, Math.ceil(options.sections.value.length / CONTEST_PAGE_SIZE))
  )
  const contestPageStartIndex = computed(() => (contestPage.value - 1) * CONTEST_PAGE_SIZE)
  const paginatedSections = computed(() =>
    options.sections.value.slice(
      contestPageStartIndex.value,
      contestPageStartIndex.value + CONTEST_PAGE_SIZE
    )
  )
  const emptyTitle = computed(() =>
    options.selectionHint.value.includes('失败') ? '排行榜加载失败' : '暂无可查看的竞赛排行榜'
  )
  const pointsEmptyTitle = computed(() =>
    options.rankingError.value ? '积分排行榜加载失败' : '暂无可查看的积分排行榜'
  )

  watch(
    () => options.sections.value,
    () => {
      contestPage.value = 1
    }
  )

  watch(contestTotalPages, (totalPages) => {
    if (contestPage.value > totalPages) {
      contestPage.value = totalPages
    }
  })

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

  function changeContestPage(page: number): void {
    contestPage.value = page
  }

  return {
    contestCount,
    runningCount,
    frozenCount,
    endedCount,
    contestPage,
    contestTotalPages,
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
    changeContestPage,
  }
}

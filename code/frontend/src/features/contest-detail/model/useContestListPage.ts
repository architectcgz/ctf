import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type {
  ContestListItem,
  ContestListSummaryData,
  ContestMode,
  ContestStatus,
} from '@/api/contracts'
import { getContests, type GetContestsData } from '@/api/contest'
import { usePagination } from '@/composables/usePagination'
import {
  getContestAccentColor,
  getContestActionLabel,
  getModeLabel,
  getStatusLabel,
  isStudentVisibleContestStatus,
} from '@/utils/contest'

interface ContestSummaryMetric {
  key: string
  label: string
  value: number
  hint: string
}

const VISIBLE_CONTEST_STATUSES: ContestStatus[] = ['registering', 'running', 'frozen', 'ended']
type ContestStatusFilter = '' | 'registering' | 'running' | 'frozen' | 'ended'
type ContestModeFilter = '' | Extract<ContestMode, 'jeopardy' | 'awd'>

function buildFallbackSummary(contests: ContestListItem[]): ContestListSummaryData {
  return {
    draft_count: 0,
    registering_count: contests.filter((contest) => contest.status === 'registering').length,
    running_count: contests.filter((contest) => contest.status === 'running').length,
    frozen_count: contests.filter((contest) => contest.status === 'frozen').length,
    ended_count: contests.filter((contest) => contest.status === 'ended').length,
  }
}

export function useContestListPage() {
  const router = useRouter()
  const statusFilter = ref<ContestStatusFilter>('')
  const modeFilter = ref<ContestModeFilter>('')
  const activeStatuses = computed<ContestStatus[]>(() =>
    statusFilter.value ? [statusFilter.value] : VISIBLE_CONTEST_STATUSES
  )
  const { list, total, page, pageSize, loading, error, response, changePage, refresh: refreshPage } =
    usePagination<ContestListItem, GetContestsData>(({ page, page_size, signal }) =>
      getContests(
        {
          page,
          page_size,
          statuses: activeStatuses.value,
          ...(modeFilter.value ? { mode: modeFilter.value } : {}),
        },
        { signal }
      )
    )
  const contestSummary = computed(() => response.value?.summary ?? buildFallbackSummary(list.value))
  const visibleTotal = computed(() => Math.max(0, total.value - contestSummary.value.draft_count))
  const runningCount = computed(() => contestSummary.value.running_count)
  const registeringCount = computed(() => contestSummary.value.registering_count)
  const endedCount = computed(() => contestSummary.value.ended_count + contestSummary.value.frozen_count)

  const visibleContests = computed(() =>
    list.value.filter((contest) => isStudentVisibleContestStatus(contest.status))
  )
  const totalPages = computed(() =>
    Math.max(1, Math.ceil(visibleTotal.value / Math.max(pageSize.value, 1)))
  )

  const summaryMetrics = computed<ContestSummaryMetric[]>(() => {
    return [
      {
        key: 'total',
        label: '竞赛总数',
        value: visibleTotal.value,
        hint: '当前可查看的竞赛总数',
      },
      { key: 'running', label: '进行中', value: runningCount.value, hint: '当前仍可直接进入竞赛工作区的赛事' },
      { key: 'registering', label: '报名中', value: registeringCount.value, hint: '当前仍可报名参与的赛事' },
      { key: 'ended', label: '已结束', value: endedCount.value, hint: '当前已结束或已封榜、可用于回看的赛事' },
    ]
  })

  const loadErrorMessage = computed(() => {
    if (!error.value) return ''
    if (error.value instanceof Error && error.value.message.trim().length > 0) {
      return error.value.message
    }
    return '竞赛列表加载失败，请稍后重试。'
  })

  const nowMs = computed(() => Date.now())

  function formatTime(value: string): string {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return '--'
    return new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date)
  }

  function formatDuration(ms: number): string {
    const minute = 60 * 1000
    const hour = 60 * minute
    const day = 24 * hour

    if (ms <= minute) return '少于 1 分钟'
    if (ms >= day) {
      const days = Math.floor(ms / day)
      const hours = Math.floor((ms % day) / hour)
      return `${days} 天 ${hours} 小时`
    }
    if (ms >= hour) {
      const hours = Math.floor(ms / hour)
      const minutes = Math.floor((ms % hour) / minute)
      return `${hours} 小时 ${minutes} 分钟`
    }
    return `${Math.floor(ms / minute)} 分钟`
  }

  function getTimelineHint(contest: ContestListItem): string {
    const startMs = new Date(contest.starts_at).getTime()
    const endMs = new Date(contest.ends_at).getTime()
    if (Number.isNaN(startMs) || Number.isNaN(endMs)) return '时间待定'

    if (nowMs.value < startMs) {
      return `距开始 ${formatDuration(startMs - nowMs.value)}`
    }
    if (nowMs.value <= endMs && ['running', 'registering', 'published'].includes(contest.status)) {
      return `距结束 ${formatDuration(endMs - nowMs.value)}`
    }
    return '竞赛已结束'
  }

  function goToDetail(id: string): void {
    void router.push(`/contests/${id}`)
  }

  function openContest(contest: ContestListItem): void {
    goToDetail(contest.id)
  }

  function contestAccentStyle(status: ContestStatus): Record<string, string> {
    return { '--contest-row-accent': getContestAccentColor(status) }
  }

  async function refresh(): Promise<void> {
    await refreshPage()
  }

  async function applyFilters(): Promise<void> {
    await changePage(1)
  }

  async function updateStatusFilter(value: ContestStatusFilter): Promise<void> {
    statusFilter.value = value
    await applyFilters()
  }

  async function updateModeFilter(value: ContestModeFilter): Promise<void> {
    modeFilter.value = value
    await applyFilters()
  }

  async function resetFilters(): Promise<void> {
    statusFilter.value = ''
    modeFilter.value = ''
    await applyFilters()
  }

  onMounted(() => {
    void refresh()
  })

  return {
    loading,
    total: visibleTotal,
    page,
    pageSize,
    totalPages,
    changePage,
    refresh,
    statusFilter,
    modeFilter,
    updateStatusFilter,
    updateModeFilter,
    resetFilters,
    visibleContests,
    summaryMetrics,
    loadErrorMessage,
    formatTime,
    getTimelineHint,
    openContest,
    contestAccentStyle,
    getStatusLabel,
    getModeLabel,
    getContestActionLabel,
  }
}

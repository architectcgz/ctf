import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import type { ContestListItem, ContestStatus } from '@/api/contracts'
import { getContests } from '@/api/contest'
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

export function useContestListPage() {
  const router = useRouter()
  const { list, loading, error, refresh } = usePagination(getContests)

  const visibleContests = computed(() =>
    list.value.filter((contest) => isStudentVisibleContestStatus(contest.status))
  )

  const summaryMetrics = computed<ContestSummaryMetric[]>(() => {
    const runningCount = visibleContests.value.filter((contest) => contest.status === 'running').length
    const registeringCount = visibleContests.value.filter((contest) => contest.status === 'registering').length
    const endedCount = visibleContests.value.filter((contest) =>
      ['ended', 'cancelled', 'archived', 'frozen'].includes(contest.status)
    ).length

    return [
      {
        key: 'total',
        label: '竞赛总数',
        value: visibleContests.value.length,
        hint: '当前可查看的竞赛数量',
      },
      { key: 'running', label: '进行中', value: runningCount, hint: '已经开赛且仍可参与' },
      { key: 'registering', label: '报名中', value: registeringCount, hint: '近期可以报名的竞赛' },
      { key: 'ended', label: '已结束', value: endedCount, hint: '可用于复盘或排行回看' },
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

  onMounted(() => {
    void refresh()
  })

  return {
    loading,
    refresh,
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

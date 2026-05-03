import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import type { ContestScoreboardData, ContestStatus } from '@/api/contracts'
import { getScoreboard } from '@/api/contest'
import { useToast } from '@/composables/useToast'
import { getContestAccentColor, getStatusLabel } from '@/utils/contest'

export function useScoreboardDetailPage() {
  const route = useRoute()
  const toast = useToast()

  const scoreboard = ref<ContestScoreboardData | null>(null)
  const loading = ref(false)
  const refreshing = ref(false)
  const error = ref(false)
  let requestToken = 0

  const contestId = computed(() => String(route.params.contestId ?? ''))
  const rows = computed(() => scoreboard.value?.scoreboard.list ?? [])
  const contest = computed(() => scoreboard.value?.contest)
  const supportsRealtime = computed(() => {
    const status = contest.value?.status
    return status === 'running' || status === 'frozen'
  })
  const accentStyle = computed<Record<string, string>>(() => ({
    '--scoreboard-accent': getContestAccentColor(contest.value?.status ?? 'ended'),
  }))
  const emptyTitle = computed(() => (error.value ? '排行榜加载失败' : '暂无排行榜数据'))
  const emptyDescription = computed(() =>
    error.value ? '该竞赛排行榜暂时不可用，请稍后重新加载。' : '当前还没有队伍进入榜单。'
  )
  const topScore = computed(() => rows.value[0]?.score ?? 0)
  const solvedCount = computed(() => rows.value.reduce((sum, row) => sum + row.solved_count, 0))

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

  function formatContestWindow(payload?: ContestScoreboardData['contest']): string {
    if (!payload) return '未记录'
    return `${formatDateTime(payload.started_at)} ~ ${formatDateTime(payload.ends_at)}`
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

  function getStatusCopy(status?: ContestStatus, frozen?: boolean): string {
    if (frozen || status === 'frozen') {
      return '封榜阶段展示冻结前排名，后续解封后同步最终成绩。'
    }
    if (status === 'running') {
      return '进行中竞赛支持实时更新，也可以手动刷新最新排名。'
    }
    return '历史竞赛展示最终成绩，用于复盘队伍表现。'
  }

  async function loadScoreboard(silent = false): Promise<void> {
    const currentContestId = contestId.value
    if (!currentContestId) {
      return
    }

    const token = ++requestToken
    const hadScoreboard = Boolean(scoreboard.value)
    if (silent) {
      refreshing.value = true
    } else {
      loading.value = true
    }
    error.value = false

    try {
      const payload = await getScoreboard(currentContestId, { page: 1, page_size: 100 })
      if (token !== requestToken) {
        return
      }
      scoreboard.value = payload
    } catch {
      if (token !== requestToken) {
        return
      }
      if (silent && hadScoreboard) {
        toast.error('排行榜刷新失败')
        return
      }
      scoreboard.value = null
      error.value = true
    } finally {
      if (token === requestToken) {
        loading.value = false
        refreshing.value = false
      }
    }
  }

  watch(
    contestId,
    () => {
      void loadScoreboard()
    },
    { immediate: true }
  )

  return {
    contest,
    rows,
    scoreboard,
    loading,
    refreshing,
    supportsRealtime,
    accentStyle,
    emptyTitle,
    emptyDescription,
    topScore,
    solvedCount,
    getStatusLabel,
    formatDateTime,
    formatContestWindow,
    getRowClass,
    getRankPillClass,
    getStatusCopy,
    loadScoreboard,
  }
}

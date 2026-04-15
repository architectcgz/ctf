import { computed, onBeforeUnmount, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  getContestAWDWorkspace,
  getScoreboard,
  startContestChallengeInstance,
  submitContestAWDAttack,
} from '@/api/contest'
import type {
  AWDAttackLogData,
  ContestAWDWorkspaceData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

const AWD_WORKSPACE_AUTO_REFRESH_INTERVAL_MS = 15_000

interface UseContestAWDWorkspaceOptions {
  contestId: MaybeRefOrGetter<string>
  contestStatus?: MaybeRefOrGetter<ContestDetailData['status'] | null | undefined>
}

export function useContestAWDWorkspace(options: UseContestAWDWorkspaceOptions) {
  const toast = useToast()

  const workspace = ref<ContestAWDWorkspaceData | null>(null)
  const scoreboardRows = ref<ScoreboardRow[]>([])
  const loading = ref(false)
  const error = ref('')
  const submitResult = ref<AWDAttackLogData | null>(null)
  const startingChallengeId = ref('')
  const submittingKey = ref('')
  const lastSyncedAt = ref<string | null>(null)

  let requestToken = 0
  let autoRefreshTimer: number | null = null

  const hasTeam = computed(() => Boolean(workspace.value?.my_team))
  const shouldAutoRefresh = computed(() => {
    const status = toValue(options.contestStatus)
    return status === 'running' || status === 'frozen'
  })

  async function loadWorkspace(contestId: string): Promise<void> {
    workspace.value = await getContestAWDWorkspace(contestId)
  }

  async function loadScoreboard(contestId: string): Promise<void> {
    const payload = await getScoreboard(contestId, { page: 1, page_size: 10 })
    scoreboardRows.value = payload.scoreboard.list
  }

  async function refreshAll(): Promise<void> {
    const contestId = toValue(options.contestId)
    if (!contestId) {
      workspace.value = null
      scoreboardRows.value = []
      error.value = ''
      loading.value = false
      lastSyncedAt.value = null
      return
    }

    const currentToken = ++requestToken
    loading.value = true

    try {
      const [nextWorkspace] = await Promise.all([
        getContestAWDWorkspace(contestId),
        loadScoreboard(contestId),
      ])

      if (currentToken !== requestToken) {
        return
      }

      workspace.value = nextWorkspace
      error.value = ''
      lastSyncedAt.value = new Date().toISOString()
    } catch (err) {
      if (currentToken !== requestToken) {
        return
      }

      workspace.value = null
      scoreboardRows.value = []
      error.value = '加载 AWD 战场失败，请稍后刷新重试'
    } finally {
      if (currentToken === requestToken) {
        loading.value = false
      }
    }
  }

  function stopAutoRefresh(): void {
    if (autoRefreshTimer !== null) {
      window.clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
  }

  async function startService(challengeId: string): Promise<void> {
    const contestId = toValue(options.contestId)
    if (!contestId || startingChallengeId.value) {
      return
    }

    startingChallengeId.value = challengeId
    try {
      const instance = await startContestChallengeInstance(contestId, challengeId)
      await refreshAll()
      if (instance.access_url) {
        toast.success('服务已就绪，可直接进入')
      } else {
        toast.success('服务启动请求已提交')
      }
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '启动服务失败')
    } finally {
      startingChallengeId.value = ''
    }
  }

  async function submitAttack(
    challengeId: string,
    victimTeamId: number,
    flag: string
  ): Promise<AWDAttackLogData | null> {
    const contestId = toValue(options.contestId)
    const normalizedFlag = flag.trim()
    if (!contestId || !victimTeamId || !normalizedFlag) {
      return null
    }
    if (submittingKey.value) {
      return null
    }

    submittingKey.value = `${challengeId}:${victimTeamId}`
    submitResult.value = null

    try {
      const result = await submitContestAWDAttack(contestId, challengeId, {
        victim_team_id: victimTeamId,
        flag: normalizedFlag,
      })
      submitResult.value = result
      await refreshAll()
      toast.success(
        result.is_success ? `攻击成功，+${result.score_gained} 分` : '攻击未命中有效 flag'
      )
      return result
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '提交 stolen flag 失败')
      return null
    } finally {
      submittingKey.value = ''
    }
  }

  watch(
    () => toValue(options.contestId),
    () => {
      void refreshAll()
    },
    { immediate: true }
  )

  watch(
    () => [toValue(options.contestId), shouldAutoRefresh.value] as const,
    ([contestId, enabled]) => {
      stopAutoRefresh()
      if (!contestId || !enabled || typeof window === 'undefined') {
        return
      }
      autoRefreshTimer = window.setInterval(() => {
        void refreshAll()
      }, AWD_WORKSPACE_AUTO_REFRESH_INTERVAL_MS)
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    stopAutoRefresh()
  })

  return {
    workspace,
    scoreboardRows,
    loading,
    error,
    hasTeam,
    submitResult,
    startingChallengeId,
    submittingKey,
    shouldAutoRefresh,
    lastSyncedAt,
    refreshAll,
    loadWorkspace,
    startService,
    submitAttack,
  }
}

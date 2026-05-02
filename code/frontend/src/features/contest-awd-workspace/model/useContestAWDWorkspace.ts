import { computed, onBeforeUnmount, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  getContestAWDWorkspace,
  getScoreboard,
} from '@/api/contest'
import type {
  AWDAttackLogData,
  ContestAWDWorkspaceData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { useAwdWorkspaceAccessActions } from './useAwdWorkspaceAccessActions'
import { useAwdWorkspaceAttackSubmission } from './useAwdWorkspaceAttackSubmission'
import { useAwdWorkspaceServiceActions } from './useAwdWorkspaceServiceActions'

const AWD_WORKSPACE_AUTO_REFRESH_INTERVAL_MS = 15_000

interface UseContestAWDWorkspaceOptions {
  contestId: MaybeRefOrGetter<string>
  contestStatus?: MaybeRefOrGetter<ContestDetailData['status'] | null | undefined>
  formatAttackResultToast?: (result: AWDAttackLogData) => string
}

export function useContestAWDWorkspace(options: UseContestAWDWorkspaceOptions) {
  const workspace = ref<ContestAWDWorkspaceData | null>(null)
  const scoreboardRows = ref<ScoreboardRow[]>([])
  const loading = ref(false)
  const error = ref('')
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
      clearSSHAccess()
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
      clearSettledServiceActions(nextWorkspace)
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

  const {
    clearSSHAccess,
    openDefenseSSH,
    openService,
    openTarget,
    openingServiceKey,
    openingSSHKey,
    openingTargetKey,
    sshAccessByServiceId,
  } = useAwdWorkspaceAccessActions({
    contestId: options.contestId,
  })
  const {
    submitAttack,
    submitResult,
    submittingKey,
  } = useAwdWorkspaceAttackSubmission({
    contestId: options.contestId,
    refreshAll,
    formatAttackResultToast: options.formatAttackResultToast,
  })
  const {
    clearSettledServiceActions,
    restartService,
    serviceActionPendingById,
    startService,
    startingServiceKey,
  } = useAwdWorkspaceServiceActions({
    contestId: options.contestId,
    refreshAll,
  })

  function stopAutoRefresh(): void {
    if (autoRefreshTimer !== null) {
      window.clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
  }

  watch(
    () => toValue(options.contestId),
    () => {
      clearSSHAccess()
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
    startingServiceKey,
    serviceActionPendingById,
    openingServiceKey,
    openingSSHKey,
    sshAccessByServiceId,
    openingTargetKey,
    submittingKey,
    shouldAutoRefresh,
    lastSyncedAt,
    refreshAll,
    loadWorkspace,
    startService,
    restartService,
    openService,
    openDefenseSSH,
    openTarget,
    submitAttack,
  }
}

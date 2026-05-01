import { computed, onBeforeUnmount, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  getContestAWDWorkspace,
  getScoreboard,
  requestContestAWDDefenseSSH,
  requestContestAWDTargetAccess,
  restartContestAWDServiceInstance,
  startContestAWDServiceInstance,
  submitContestAWDAttack,
} from '@/api/contest'
import { requestInstanceAccess } from '@/api/instance'
import type {
  AWDAttackLogData,
  AWDDefenseSSHAccessData,
  ContestAWDWorkspaceData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

const AWD_WORKSPACE_AUTO_REFRESH_INTERVAL_MS = 15_000

interface UseContestAWDWorkspaceOptions {
  contestId: MaybeRefOrGetter<string>
  contestStatus?: MaybeRefOrGetter<ContestDetailData['status'] | null | undefined>
  formatAttackResultToast?: (result: AWDAttackLogData) => string
}

export function useContestAWDWorkspace(options: UseContestAWDWorkspaceOptions) {
  const toast = useToast()

  const workspace = ref<ContestAWDWorkspaceData | null>(null)
  const scoreboardRows = ref<ScoreboardRow[]>([])
  const loading = ref(false)
  const error = ref('')
  const submitResult = ref<AWDAttackLogData | null>(null)
  const startingServiceKey = ref('')
  const openingServiceKey = ref('')
  const openingSSHKey = ref('')
  const sshAccessByServiceId = ref<Record<string, AWDDefenseSSHAccessData>>({})
  const openingTargetKey = ref('')
  const submittingKey = ref('')
  const lastSyncedAt = ref<string | null>(null)
  const serviceActionPendingById = ref<Record<string, boolean>>({})

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

  function isServiceRuntimeBusy(status?: string): boolean {
    return status === 'pending' || status === 'creating'
  }

  function isServiceOperationBusy(status?: string): boolean {
    return status === 'requested' || status === 'provisioning' || status === 'recovering'
  }

  function clearSettledServiceActions(nextWorkspace: ContestAWDWorkspaceData): void {
    const nextPending = { ...serviceActionPendingById.value }
    for (const item of nextWorkspace.services || []) {
      const serviceId = item.service_id
      if (!serviceId) {
        continue
      }
      if (!isServiceRuntimeBusy(item.instance_status) && !isServiceOperationBusy(item.operation_status)) {
        delete nextPending[serviceId]
      }
    }
    serviceActionPendingById.value = nextPending
  }

  async function refreshAll(): Promise<void> {
    const contestId = toValue(options.contestId)
    if (!contestId) {
      workspace.value = null
      scoreboardRows.value = []
      error.value = ''
      loading.value = false
      lastSyncedAt.value = null
      sshAccessByServiceId.value = {}
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

  function stopAutoRefresh(): void {
    if (autoRefreshTimer !== null) {
      window.clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
  }

  async function startService(serviceId: string): Promise<void> {
    const contestId = toValue(options.contestId)
    if (!contestId || !serviceId || startingServiceKey.value) {
      return
    }

    startingServiceKey.value = serviceId
    try {
      const instance = await startContestAWDServiceInstance(contestId, serviceId)
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
      startingServiceKey.value = ''
    }
  }

  async function restartService(serviceId: string): Promise<void> {
    const contestId = toValue(options.contestId)
    if (!contestId || !serviceId || startingServiceKey.value || serviceActionPendingById.value[serviceId]) {
      return
    }

    startingServiceKey.value = serviceId
    serviceActionPendingById.value = {
      ...serviceActionPendingById.value,
      [serviceId]: true,
    }
    try {
      await restartContestAWDServiceInstance(contestId, serviceId)
      await refreshAll()
      toast.success('服务重启请求已提交')
    } catch (err) {
      console.error(err)
      const nextPending = { ...serviceActionPendingById.value }
      delete nextPending[serviceId]
      serviceActionPendingById.value = nextPending
      toast.error(err instanceof Error ? err.message : '重启服务失败')
    } finally {
      startingServiceKey.value = ''
    }
  }

  async function openService(instanceId: string): Promise<string | null> {
    if (!instanceId || openingServiceKey.value) {
      return null
    }

    openingServiceKey.value = instanceId
    try {
      const result = await requestInstanceAccess(instanceId)
      if (typeof window !== 'undefined') {
        window.open(result.access_url, '_blank', 'noopener,noreferrer')
      }
      return result.access_url
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '打开本队服务失败')
      return null
    } finally {
      openingServiceKey.value = ''
    }
  }

  async function openDefenseSSH(serviceId: string): Promise<AWDDefenseSSHAccessData | null> {
    const contestId = toValue(options.contestId)
    if (!contestId || !serviceId || openingSSHKey.value) {
      return null
    }

    openingSSHKey.value = serviceId
    try {
      const result = await requestContestAWDDefenseSSH(contestId, serviceId)
      sshAccessByServiceId.value = {
        ...sshAccessByServiceId.value,
        [serviceId]: result,
      }
      toast.success('SSH 防守连接已生成')
      return result
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '生成 SSH 防守连接失败')
      return null
    } finally {
      openingSSHKey.value = ''
    }
  }

  async function openTarget(serviceId: string, victimTeamId: string): Promise<string | null> {
    const contestId = toValue(options.contestId)
    if (!contestId || !serviceId || !victimTeamId || openingTargetKey.value) {
      return null
    }

    const targetKey = `${serviceId}:${victimTeamId}`
    openingTargetKey.value = targetKey

    try {
      const result = await requestContestAWDTargetAccess(contestId, serviceId, victimTeamId)
      if (typeof window !== 'undefined') {
        window.open(result.access_url, '_blank', 'noopener,noreferrer')
      }
      return result.access_url
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '打开目标服务失败')
      return null
    } finally {
      openingTargetKey.value = ''
    }
  }

  async function submitAttack(
    serviceId: string,
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

    submittingKey.value = `${serviceId}:${victimTeamId}`
    submitResult.value = null

    try {
      const result = await submitContestAWDAttack(contestId, serviceId, {
        victim_team_id: victimTeamId,
        flag: normalizedFlag,
      })
      submitResult.value = result
      await refreshAll()
      const formattedMessage = options.formatAttackResultToast?.(result)
      toast.success(
        formattedMessage ||
          (result.is_success ? `攻击成功，+${result.score_gained} 分` : '攻击未命中有效 flag')
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
      sshAccessByServiceId.value = {}
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

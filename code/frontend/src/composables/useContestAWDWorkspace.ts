import { computed, onBeforeUnmount, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  getContestAWDWorkspace,
  getScoreboard,
  listContestAWDDefenseDirectory,
  readContestAWDDefenseFile,
  requestContestAWDDefenseSSH,
  requestContestAWDTargetAccess,
  runContestAWDDefenseCommand,
  saveContestAWDDefenseFile,
  startContestAWDServiceInstance,
  submitContestAWDAttack,
} from '@/api/contest'
import { requestInstanceAccess } from '@/api/instance'
import type {
  AWDAttackLogData,
  AWDDefenseCommandData,
  AWDDefenseDirectoryData,
  AWDDefenseFileData,
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
  const activeDefenseServiceId = ref('')
  const defenseDirectory = ref<AWDDefenseDirectoryData | null>(null)
  const defenseDirectoryPath = ref('.')
  const defenseFile = ref<AWDDefenseFileData | null>(null)
  const defenseDraft = ref('')
  const defenseFilePath = ref('app.py')
  const loadingDefenseDirectory = ref(false)
  const loadingDefenseFile = ref(false)
  const savingDefenseFile = ref(false)
  const runningDefenseCommand = ref(false)
  const defenseCommand = ref('ls')
  const defenseCommandResult = ref<AWDDefenseCommandData | null>(null)
  const openingTargetKey = ref('')
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
      sshAccessByServiceId.value = {}
      activeDefenseServiceId.value = ''
      defenseDirectory.value = null
      defenseDirectoryPath.value = '.'
      defenseFile.value = null
      defenseDraft.value = ''
      defenseCommandResult.value = null
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

  async function openDefenseDirectory(dirPath = defenseDirectoryPath.value): Promise<void> {
    const contestId = toValue(options.contestId)
    const serviceId = activeDefenseServiceId.value
    if (!contestId || !serviceId || loadingDefenseDirectory.value) {
      return
    }

    loadingDefenseDirectory.value = true
    try {
      const result = await listContestAWDDefenseDirectory(contestId, serviceId, dirPath || '.')
      defenseDirectory.value = result
      defenseDirectoryPath.value = result.path
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '读取文件列表失败')
    } finally {
      loadingDefenseDirectory.value = false
    }
  }

  async function openDefenseFile(filePath: string): Promise<void> {
    const contestId = toValue(options.contestId)
    const serviceId = activeDefenseServiceId.value
    if (!contestId || !serviceId || !filePath || loadingDefenseFile.value) {
      return
    }

    defenseFilePath.value = filePath
    loadingDefenseFile.value = true
    defenseFile.value = null
    defenseDraft.value = ''
    defenseCommandResult.value = null
    try {
      const result = await readContestAWDDefenseFile(contestId, serviceId, filePath)
      defenseFile.value = result
      defenseDraft.value = result.content
      toast.success('防守文件已载入')
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '读取防守文件失败')
    } finally {
      loadingDefenseFile.value = false
    }
  }

  async function openDefenseWorkbench(serviceId: string, filePath = 'app.py'): Promise<void> {
    const contestId = toValue(options.contestId)
    if (!contestId || !serviceId || loadingDefenseFile.value || loadingDefenseDirectory.value) {
      return
    }

    activeDefenseServiceId.value = serviceId
    defenseFilePath.value = filePath
    loadingDefenseDirectory.value = true
    loadingDefenseFile.value = true
    defenseDirectory.value = null
    defenseDirectoryPath.value = '.'
    defenseFile.value = null
    defenseDraft.value = ''
    defenseCommandResult.value = null
    try {
      const [directory, result] = await Promise.all([
        listContestAWDDefenseDirectory(contestId, serviceId, '.'),
        readContestAWDDefenseFile(contestId, serviceId, filePath),
      ])
      defenseDirectory.value = directory
      defenseDirectoryPath.value = directory.path
      defenseFile.value = result
      defenseDraft.value = result.content
      toast.success('防守文件已载入')
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '读取防守文件失败')
    } finally {
      loadingDefenseDirectory.value = false
      loadingDefenseFile.value = false
    }
  }

  async function saveDefenseFile(): Promise<void> {
    const contestId = toValue(options.contestId)
    const serviceId = activeDefenseServiceId.value
    if (!contestId || !serviceId || !defenseFilePath.value || savingDefenseFile.value) {
      return
    }

    savingDefenseFile.value = true
    try {
      const result = await saveContestAWDDefenseFile(contestId, serviceId, {
        path: defenseFilePath.value,
        content: defenseDraft.value,
        backup: true,
      })
      defenseFile.value = {
        path: result.path,
        content: defenseDraft.value,
        size: result.size,
      }
      toast.success(result.backup_path ? `已保存，备份 ${result.backup_path}` : '已保存防守文件')
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '保存防守文件失败')
    } finally {
      savingDefenseFile.value = false
    }
  }

  async function runDefenseCommand(
    command = defenseCommand.value
  ): Promise<AWDDefenseCommandData | null> {
    const contestId = toValue(options.contestId)
    const serviceId = activeDefenseServiceId.value
    const normalizedCommand = command.trim()
    if (!contestId || !serviceId || !normalizedCommand || runningDefenseCommand.value) {
      return null
    }

    runningDefenseCommand.value = true
    try {
      const result = await runContestAWDDefenseCommand(contestId, serviceId, normalizedCommand)
      defenseCommand.value = normalizedCommand
      defenseCommandResult.value = result
      return result
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '执行防守命令失败')
      return null
    } finally {
      runningDefenseCommand.value = false
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
      activeDefenseServiceId.value = ''
      defenseDirectory.value = null
      defenseDirectoryPath.value = '.'
      defenseFile.value = null
      defenseDraft.value = ''
      defenseCommandResult.value = null
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
    openingServiceKey,
    openingSSHKey,
    sshAccessByServiceId,
    activeDefenseServiceId,
    defenseDirectory,
    defenseDirectoryPath,
    defenseFile,
    defenseDraft,
    defenseFilePath,
    loadingDefenseDirectory,
    loadingDefenseFile,
    savingDefenseFile,
    runningDefenseCommand,
    defenseCommand,
    defenseCommandResult,
    openingTargetKey,
    submittingKey,
    shouldAutoRefresh,
    lastSyncedAt,
    refreshAll,
    loadWorkspace,
    startService,
    openService,
    openDefenseSSH,
    openDefenseDirectory,
    openDefenseFile,
    openDefenseWorkbench,
    saveDefenseFile,
    runDefenseCommand,
    openTarget,
    submitAttack,
  }
}

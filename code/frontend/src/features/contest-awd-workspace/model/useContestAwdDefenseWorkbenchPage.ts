import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import {
  getContestAWDWorkspace,
  getContestChallenges,
  requestContestAWDDefenseDirectory,
  requestContestAWDDefenseFile,
  requestContestAWDDefenseFileSave,
} from '@/api/contest'
import type { AWDDefenseDirectoryData, AWDDefenseFileData } from '@/api/contracts'
import { toDefenseServiceCards, type AWDDefenseServiceCard } from './awdDefensePresentation'

function toStatus(err: unknown): number | undefined {
  if (err && typeof err === 'object' && 'status' in err) {
    const status = (err as { status?: number }).status
    return typeof status === 'number' ? status : undefined
  }
  return undefined
}

function normalizeScopePath(path: string): string {
  return path.replace(/^\.?\//, '').replace(/\/+/g, '/').replace(/\/$/, '')
}

function toDefenseDirectoryPath(path: string): string {
  const normalized = normalizeScopePath(path.trim())
  if (!normalized) {
    return '.'
  }
  const parts = normalized.split('/').filter(Boolean)
  if (parts.length <= 1) {
    return '.'
  }
  return parts.slice(0, -1).join('/')
}

function getInitialDirectoryPath(serviceCard: AWDDefenseServiceCard | null): string {
  const editablePaths = serviceCard?.defenseScope?.editable_paths || []
  const firstEditablePath = editablePaths.find((path) => path.trim().length > 0)
  return firstEditablePath ? toDefenseDirectoryPath(firstEditablePath) : '.'
}

function getDefenseWorkbenchUnavailableMessage(serviceCard: AWDDefenseServiceCard | null): string {
  if (!serviceCard) {
    return '当前防守容器暂不可用，请稍后重试。'
  }
  if (!serviceCard.instanceId) {
    return '当前服务还没有分配实例，暂时不能查看防守内容。'
  }
  if (serviceCard.instanceStatusLabel === '重启队列中' || serviceCard.instanceStatusLabel === '正在启动') {
    return '当前服务实例正在启动，容器就绪后再试。'
  }
  if (serviceCard.instanceStatusLabel === '启动失败') {
    return '当前服务实例启动失败，暂时不能查看防守内容。'
  }
  if (!serviceCard.canOpenService) {
    return `当前服务暂未就绪（${serviceCard.instanceStatusLabel}），暂时不能查看防守内容。`
  }
  return '当前防守容器暂不可用，可能是容器尚未就绪，请稍后重试。'
}

export function useContestAwdDefenseWorkbenchPage() {
  const route = useRoute()
  const contestId = computed(() => String(route.params.id ?? ''))
  const serviceId = computed(() => String(route.params.serviceId ?? ''))
  const backLink = computed(() => ({
    name: 'ContestDetail',
    params: { id: contestId.value },
    query: { panel: 'challenges' },
  }))

  const serviceCard = ref<AWDDefenseServiceCard | null>(null)
  const directory = ref<AWDDefenseDirectoryData | null>(null)
  const file = ref<AWDDefenseFileData | null>(null)
  const error = ref('')
  const fileError = ref('')
  const pageLoading = ref(false)
  const directoryLoading = ref(false)
  const fileLoading = ref(false)
  const saveLoading = ref(false)
  const saveError = ref('')

  let pageRequestSeq = 0
  let directoryRequestSeq = 0
  let fileRequestSeq = 0
  let saveRequestSeq = 0

  const serviceTitle = computed(() => serviceCard.value?.title || `服务 #${serviceId.value || '--'}`)
  const currentDirectoryPath = computed(() => directory.value?.path || '.')
  const loading = computed(
    () => pageLoading.value || directoryLoading.value || fileLoading.value
  )

  async function loadDirectory(path: string, ownerSeq = pageRequestSeq): Promise<void> {
    const currentContestId = contestId.value
    const currentServiceId = serviceId.value
    if (!currentContestId || !currentServiceId) {
      return
    }
    if (!serviceCard.value?.canOpenService) {
      directory.value = null
      file.value = null
      error.value = getDefenseWorkbenchUnavailableMessage(serviceCard.value)
      fileError.value = ''
      saveRequestSeq += 1
      saveLoading.value = false
      saveError.value = ''
      directoryLoading.value = false
      return
    }

    const currentSeq = ++directoryRequestSeq
    fileRequestSeq += 1
    saveRequestSeq += 1
    directoryLoading.value = true
    error.value = ''
    fileError.value = ''
    saveLoading.value = false
    saveError.value = ''
    file.value = null

    try {
      const nextDirectory = await requestContestAWDDefenseDirectory(
        currentContestId,
        currentServiceId,
        path
      )
      if (ownerSeq != pageRequestSeq || currentSeq !== directoryRequestSeq) {
        return
      }
      directory.value = nextDirectory
      error.value = ''
    } catch (err) {
      if (ownerSeq != pageRequestSeq || currentSeq !== directoryRequestSeq) {
        return
      }
      directory.value = null
      error.value =
        toStatus(err) === 403
          ? getDefenseWorkbenchUnavailableMessage(serviceCard.value)
          : '加载防守文件列表失败，请稍后重试。'
    } finally {
      if (ownerSeq === pageRequestSeq && currentSeq === directoryRequestSeq) {
        directoryLoading.value = false
      }
    }
  }

  async function openFile(path: string): Promise<void> {
    const currentContestId = contestId.value
    const currentServiceId = serviceId.value
    if (!currentContestId || !currentServiceId) {
      return
    }
    if (!serviceCard.value?.canOpenService) {
      file.value = null
      fileError.value = getDefenseWorkbenchUnavailableMessage(serviceCard.value)
      saveRequestSeq += 1
      saveLoading.value = false
      saveError.value = ''
      fileLoading.value = false
      return
    }

    const currentSeq = ++fileRequestSeq
    saveRequestSeq += 1
    fileLoading.value = true
    fileError.value = ''
    saveLoading.value = false
    saveError.value = ''
    file.value = null

    try {
      const nextFile = await requestContestAWDDefenseFile(
        currentContestId,
        currentServiceId,
        path
      )
      if (currentSeq !== fileRequestSeq) {
        return
      }
      file.value = nextFile
      fileError.value = ''
    } catch (err) {
      if (currentSeq !== fileRequestSeq) {
        return
      }
      fileError.value =
        toStatus(err) === 403
          ? getDefenseWorkbenchUnavailableMessage(serviceCard.value)
          : '加载文件内容失败，请稍后重试。'
    } finally {
      if (currentSeq === fileRequestSeq) {
        fileLoading.value = false
      }
    }
  }

  async function refreshDirectory(): Promise<void> {
    await loadDirectory(currentDirectoryPath.value)
  }

  async function saveFile(path: string, content: string): Promise<void> {
    const currentContestId = contestId.value
    const currentServiceId = serviceId.value
    if (!currentContestId || !currentServiceId || saveLoading.value) {
      return
    }
    if (!serviceCard.value?.canOpenService) {
      saveError.value = getDefenseWorkbenchUnavailableMessage(serviceCard.value)
      return
    }

    const currentSeq = ++saveRequestSeq
    saveLoading.value = true
    saveError.value = ''

    try {
      const saved = await requestContestAWDDefenseFileSave(currentContestId, currentServiceId, {
        path,
        content,
        backup: true,
      })
      if (currentSeq !== saveRequestSeq) {
        return
      }
      if (file.value?.path === path) {
        file.value = {
          path: saved.path,
          content,
          size: saved.size,
        }
      }
      fileError.value = ''
    } catch (err) {
      if (currentSeq !== saveRequestSeq) {
        return
      }
      saveError.value =
        toStatus(err) === 403
          ? getDefenseWorkbenchUnavailableMessage(serviceCard.value)
          : '保存文件失败，请稍后重试。'
    } finally {
      if (currentSeq === saveRequestSeq) {
        saveLoading.value = false
      }
    }
  }

  async function loadPage(): Promise<void> {
    if (!contestId.value || !serviceId.value) {
      pageRequestSeq += 1
      directoryRequestSeq += 1
      fileRequestSeq += 1
      saveRequestSeq += 1
      serviceCard.value = null
      directory.value = null
      file.value = null
      error.value = '防守内容页参数不完整。'
      fileError.value = ''
      saveError.value = ''
      pageLoading.value = false
      directoryLoading.value = false
      fileLoading.value = false
      saveLoading.value = false
      return
    }

    const currentSeq = ++pageRequestSeq
    directoryRequestSeq += 1
    fileRequestSeq += 1
    saveRequestSeq += 1
    pageLoading.value = true
    directoryLoading.value = false
    fileLoading.value = false
    saveLoading.value = false
    directory.value = null
    file.value = null
    serviceCard.value = null
    error.value = ''
    fileError.value = ''
    saveError.value = ''

    try {
      const [workspace, challenges] = await Promise.all([
        getContestAWDWorkspace(contestId.value),
        getContestChallenges(contestId.value),
      ])
      if (currentSeq !== pageRequestSeq) {
        return
      }

      const matchedCard =
        toDefenseServiceCards({
          challenges,
          services: workspace.services,
        }).find((item) => item.serviceId === serviceId.value) || null

      if (!matchedCard) {
        error.value = '未找到该防守服务。'
        return
      }

      serviceCard.value = matchedCard
      error.value = ''
      if (!matchedCard.canOpenService) {
        error.value = getDefenseWorkbenchUnavailableMessage(matchedCard)
        return
      }
      await loadDirectory(getInitialDirectoryPath(matchedCard), currentSeq)
    } catch (err) {
      if (currentSeq !== pageRequestSeq) {
        return
      }
      error.value =
        toStatus(err) === 403
          ? getDefenseWorkbenchUnavailableMessage(serviceCard.value)
          : '加载防守内容页失败，请稍后重试。'
    } finally {
      if (currentSeq === pageRequestSeq) {
        pageLoading.value = false
      }
    }
  }

  watch(
    () => [contestId.value, serviceId.value],
    () => {
      void loadPage()
    },
    { immediate: true }
  )

  return {
    contestId,
    serviceId,
    backLink,
    serviceCard,
    serviceTitle,
    currentDirectoryPath,
    directory,
    file,
    error,
    fileError,
    saveError,
    loading,
    pageLoading,
    directoryLoading,
    fileLoading,
    saveLoading,
    loadPage,
    loadDirectory,
    refreshDirectory,
    openFile,
    saveFile,
  }
}

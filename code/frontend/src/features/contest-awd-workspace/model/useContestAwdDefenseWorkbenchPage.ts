import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import {
  getContestAWDWorkspace,
  getContestChallenges,
  requestContestAWDDefenseDirectory,
  requestContestAWDDefenseFile,
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

  let pageRequestSeq = 0
  let directoryRequestSeq = 0
  let fileRequestSeq = 0

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

    const currentSeq = ++directoryRequestSeq
    fileRequestSeq += 1
    directoryLoading.value = true
    error.value = ''
    fileError.value = ''
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
        toStatus(err) === 403 ? '当前环境未开启防守内容页。' : '加载防守文件列表失败，请稍后重试。'
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

    const currentSeq = ++fileRequestSeq
    fileLoading.value = true
    fileError.value = ''
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
        toStatus(err) === 403 ? '当前文件不可查看。' : '加载文件内容失败，请稍后重试。'
    } finally {
      if (currentSeq === fileRequestSeq) {
        fileLoading.value = false
      }
    }
  }

  async function refreshDirectory(): Promise<void> {
    await loadDirectory(currentDirectoryPath.value)
  }

  async function loadPage(): Promise<void> {
    if (!contestId.value || !serviceId.value) {
      pageRequestSeq += 1
      directoryRequestSeq += 1
      fileRequestSeq += 1
      serviceCard.value = null
      directory.value = null
      file.value = null
      error.value = '防守内容页参数不完整。'
      fileError.value = ''
      pageLoading.value = false
      directoryLoading.value = false
      fileLoading.value = false
      return
    }

    const currentSeq = ++pageRequestSeq
    directoryRequestSeq += 1
    fileRequestSeq += 1
    pageLoading.value = true
    directoryLoading.value = false
    fileLoading.value = false
    directory.value = null
    file.value = null
    serviceCard.value = null
    error.value = ''
    fileError.value = ''

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
      await loadDirectory('.', currentSeq)
    } catch (err) {
      if (currentSeq !== pageRequestSeq) {
        return
      }
      error.value =
        toStatus(err) === 403 ? '当前环境未开启防守内容页。' : '加载防守内容页失败，请稍后重试。'
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
    loading,
    pageLoading,
    directoryLoading,
    fileLoading,
    loadPage,
    loadDirectory,
    refreshDirectory,
    openFile,
  }
}

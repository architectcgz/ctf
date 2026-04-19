import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import {
  createChallengePublishRequest,
  deleteChallenge,
  getChallenges,
  getLatestChallengePublishRequest,
} from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminChallengePublishRequestData,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

const POLL_INTERVAL_MS = 3000

export interface PlatformChallengeListRow extends AdminChallengeListItem {
  latestPublishRequest: AdminChallengePublishRequestData | null
}

type ChallengeManageStatusFilter = Extract<ChallengeStatus, 'draft' | 'published' | 'archived'>

export function usePlatformChallenges() {
  const toast = useToast()
  const keyword = ref('')
  const categoryFilter = ref<ChallengeCategory | ''>('')
  const difficultyFilter = ref<ChallengeDifficulty | ''>('')
  const statusFilter = ref<ChallengeManageStatusFilter | ''>('')
  const autoFilterReady = ref(false)
  const pagination = usePagination(({ page, page_size, signal }) =>
    getChallenges({
      page,
      page_size,
      keyword: keyword.value.trim() || undefined,
      category: categoryFilter.value || undefined,
      difficulty: difficultyFilter.value || undefined,
      status: statusFilter.value || undefined,
    }, {
      signal,
    })
  )
  const latestPublishRequests = ref<Record<string, AdminChallengePublishRequestData | null>>({})
  let pollTimer: number | null = null
  let latestPublishRequestsToken = 0

  const list = computed<PlatformChallengeListRow[]>(() =>
    pagination.list.value.map((item) => ({
      ...item,
      latestPublishRequest: latestPublishRequests.value[item.id] ?? null,
    }))
  )

  function stopPolling() {
    if (pollTimer !== null) {
      window.clearInterval(pollTimer)
      pollTimer = null
    }
  }

  function syncPolling() {
    const hasActiveJob = Object.values(latestPublishRequests.value).some(
      (request) => request?.active
    )
    if (!hasActiveJob) {
      stopPolling()
      return
    }
    if (pollTimer !== null) {
      return
    }
    pollTimer = window.setInterval(() => {
      void refreshLatestPublishRequests()
    }, POLL_INTERVAL_MS)
  }

  function didAnyActiveRequestFinish(
    previousRequests: Record<string, AdminChallengePublishRequestData | null>,
    nextRequests: Record<string, AdminChallengePublishRequestData | null>
  ): boolean {
    return Object.entries(nextRequests).some(
      ([id, request]) => previousRequests[id]?.active && !request?.active
    )
  }

  async function loadLatestPublishRequests(): Promise<boolean> {
    if (pagination.list.value.length === 0) {
      latestPublishRequestsToken += 1
      latestPublishRequests.value = {}
      stopPolling()
      return false
    }

    const requestToken = ++latestPublishRequestsToken
    const previousRequests = latestPublishRequests.value
    const listSnapshot = [...pagination.list.value]
    const latestEntries = await Promise.all(
      listSnapshot.map(
        async (item) => [item.id, await getLatestChallengePublishRequest(item.id)] as const
      )
    )

    if (requestToken !== latestPublishRequestsToken) {
      return false
    }

    const nextRequests = Object.fromEntries(latestEntries)
    const finishedActiveRequest = didAnyActiveRequestFinish(previousRequests, nextRequests)

    latestPublishRequests.value = nextRequests
    syncPolling()
    return finishedActiveRequest
  }

  async function refreshLatestPublishRequests() {
    try {
      const finishedActiveRequest = await loadLatestPublishRequests()
      if (finishedActiveRequest) {
        await pagination.refresh()
      }
    } catch {
      stopPolling()
    }
  }

  async function refresh() {
    await pagination.refresh()
    await refreshLatestPublishRequests()
  }

  async function publish(row: AdminChallengeListItem) {
    try {
      await createChallengePublishRequest(row.id)
      toast.success('已提交发布检查')
      await refreshLatestPublishRequests()
    } catch {
      toast.error('提交发布检查失败，请稍后重试')
    }
  }

  async function remove(id: string) {
    const confirmed = await confirmDestructiveAction({
      message: '确定要删除这道题目吗？',
    })
    if (!confirmed) {
      return
    }

    try {
      await deleteChallenge(id)
      toast.success('删除成功')
      await refresh()
    } catch (error) {
      const message = error instanceof Error && error.message.trim() ? error.message : '删除失败'
      toast.error(message)
    }
  }

  async function changePage(next: number) {
    await pagination.changePage(next)
    await refreshLatestPublishRequests()
  }

  async function changePageSize(next: number) {
    await pagination.changePageSize(next)
    await refreshLatestPublishRequests()
  }

  type DebouncedRefresh = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const scheduleKeywordRefresh = useDebounceFn(() => {
    pagination.page.value = 1
    void refresh()
  }, 250) as DebouncedRefresh

  async function clearFilters() {
    autoFilterReady.value = false
    scheduleKeywordRefresh.cancel?.()
    keyword.value = ''
    categoryFilter.value = ''
    difficultyFilter.value = ''
    statusFilter.value = ''
    pagination.page.value = 1
    await refresh()
    autoFilterReady.value = true
  }

  watch(keyword, () => {
    if (!autoFilterReady.value) return
    scheduleKeywordRefresh()
  })

  watch([categoryFilter, difficultyFilter, statusFilter], async () => {
    if (!autoFilterReady.value) return
    scheduleKeywordRefresh.cancel?.()
    pagination.page.value = 1
    await refresh()
  })

  onMounted(async () => {
    await refresh()
    autoFilterReady.value = true
  })

  onUnmounted(() => {
    scheduleKeywordRefresh.cancel?.()
    stopPolling()
  })

  return {
    ...pagination,
    list,
    error: pagination.error,
    keyword,
    categoryFilter,
    difficultyFilter,
    statusFilter,
    changePage,
    changePageSize,
    refresh,
    clearFilters,
    publish,
    remove,
  }
}

import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessageBox } from 'element-plus'

import {
  createChallengePublishRequest,
  deleteChallenge,
  getChallenges,
  getLatestChallengePublishRequest,
} from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminChallengePublishRequestData,
} from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

const POLL_INTERVAL_MS = 3000

export interface AdminChallengeListRow extends AdminChallengeListItem {
  latestPublishRequest: AdminChallengePublishRequestData | null
}

export function useAdminChallenges() {
  const toast = useToast()
  const pagination = usePagination(getChallenges)
  const latestPublishRequests = ref<Record<string, AdminChallengePublishRequestData | null>>({})
  let pollTimer: number | null = null

  const list = computed<AdminChallengeListRow[]>(() =>
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
    const hasActiveJob = Object.values(latestPublishRequests.value).some((request) => request?.active)
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
      latestPublishRequests.value = {}
      stopPolling()
      return false
    }

    const previousRequests = latestPublishRequests.value
    const latestEntries = await Promise.all(
      pagination.list.value.map(async (item) => [
        item.id,
        await getLatestChallengePublishRequest(item.id),
      ] as const)
    )

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
    try {
      await ElMessageBox.confirm('确定要删除这道题目吗？', '确认删除', {
        type: 'warning',
        showClose: true,
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        cancelButtonClass: 'challenge-delete-confirm-cancel',
        confirmButtonClass: 'challenge-delete-confirm-primary',
        customClass: 'challenge-delete-confirm-box',
        modalClass: 'challenge-delete-confirm-modal',
        closeOnClickModal: false,
      })
      await deleteChallenge(id)
      toast.success('删除成功')
      await refresh()
    } catch (error) {
      if (error !== 'cancel') {
        toast.error('删除失败')
      }
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

  onMounted(() => {
    void refresh()
  })

  onUnmounted(() => {
    stopPolling()
  })

  return {
    ...pagination,
    list,
    changePage,
    changePageSize,
    refresh,
    publish,
    remove,
  }
}

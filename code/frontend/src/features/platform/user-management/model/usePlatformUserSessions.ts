import { onUnmounted, ref, watch } from 'vue'

import {
  getUserSessions,
  revokeAllUserSessions,
  revokeUserSession,
} from '@/api/admin/users'
import { ApiError } from '@/api/request'
import type { UserSessionData } from '@/api/contracts'
import { useToast } from '@/shared/model/common/useToast'

export function usePlatformUserSessions(userId: () => string | undefined) {
  const toast = useToast()
  const sessions = ref<UserSessionData[]>([])
  const loading = ref(false)
  const revokingSessionId = ref<string | null>(null)
  const revokingAll = ref(false)
  const copiedSessionId = ref<string | null>(null)

  let fetchController: AbortController | undefined

  async function fetch() {
    const uid = userId()
    if (!uid) return
    loading.value = true
    try {
      sessions.value = await getUserSessions(uid)
    } catch (e) {
      console.warn('[usePlatformUserSessions] Failed to fetch sessions:', e)
      sessions.value = []
    } finally {
      loading.value = false
    }
  }

  async function revokeOne(sessionId: string) {
    const uid = userId()
    if (!uid) return
    revokingSessionId.value = sessionId
    try {
      await revokeUserSession(uid, sessionId)
      toast.success('会话已撤销')
      sessions.value = sessions.value.filter((s) => s.id !== sessionId)
    } catch (e) {
      if (e instanceof ApiError) {
        toast.error(e.message || '撤销会话失败')
      } else {
        toast.error('撤销会话失败')
      }
    } finally {
      revokingSessionId.value = null
    }
  }

  async function revokeAll() {
    const uid = userId()
    if (!uid) return
    revokingAll.value = true
    try {
      await revokeAllUserSessions(uid)
      toast.success('已撤销所有会话')
      sessions.value = []
    } catch (e) {
      if (e instanceof ApiError) {
        toast.error(e.message || '撤销所有会话失败')
      } else {
        toast.error('撤销所有会话失败')
      }
    } finally {
      revokingAll.value = false
    }
  }

  async function copySessionIdToClipboard(id: string) {
    try {
      await navigator.clipboard.writeText(id)
      copiedSessionId.value = id
      setTimeout(() => {
        if (copiedSessionId.value === id) {
          copiedSessionId.value = null
        }
      }, 2000)
    } catch {
      // 降级：让用户手动选择复制
    }
  }

  function reset() {
    fetchController?.abort()
    sessions.value = []
    revokingSessionId.value = null
    revokingAll.value = false
  }

  watch(
    userId,
    (newId) => {
      fetchController?.abort()
      if (newId) {
        reset()
        void fetch()
      }
    }
  )

  onUnmounted(() => {
    fetchController?.abort()
  })

  return {
    sessions,
    loading,
    revokingSessionId,
    revokingAll,
    copiedSessionId,
    fetch,
    revokeOne,
    revokeAll,
    copySessionIdToClipboard,
    reset,
  }
}

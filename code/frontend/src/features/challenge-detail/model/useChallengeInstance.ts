import { onUnmounted, ref, watch, type MaybeRefOrGetter, toValue } from 'vue'

import { createInstance } from '@/api/challenge'
import { getMyInstances } from '@/api/instance'
import type { InstanceData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import { useInstanceWorkflowActions } from '@/features/instance-workflow'
import { useToast } from '@/shared/model/common/useToast'

const CHALLENGE_INSTANCE_POLL_INTERVAL_MS = 3000

export function useChallengeInstance(challengeId: MaybeRefOrGetter<string | undefined>) {
  const toast = useToast()

  const instance = ref<InstanceData | null>(null)
  const loading = ref(false)
  const creating = ref(false)
  let pollingTimer: number | null = null

  function isWaitingStatus(status: InstanceData['status'] | undefined) {
    return status === 'pending' || status === 'creating'
  }

  function clearPollingTimer() {
    if (pollingTimer !== null) {
      window.clearTimeout(pollingTimer)
      pollingTimer = null
    }
  }

  function schedulePolling() {
    if (pollingTimer !== null) return
    pollingTimer = window.setTimeout(() => {
      pollingTimer = null
      void refresh({ silent: true })
    }, CHALLENGE_INSTANCE_POLL_INTERVAL_MS)
  }

  function syncPollingState() {
    const currentChallengeId = toValue(challengeId)
    if (!currentChallengeId || !instance.value || !isWaitingStatus(instance.value.status)) {
      clearPollingTimer()
      return
    }
    schedulePolling()
  }

  async function refresh(options?: { silent?: boolean }) {
    const currentChallengeId = toValue(challengeId)
    if (!currentChallengeId) {
      instance.value = null
      clearPollingTimer()
      return
    }

    loading.value = true
    try {
      const instances = await getMyInstances()
      instance.value =
        instances.find((item) => String(item.challenge_id) === currentChallengeId) ?? null
    } catch (error) {
      if (!options?.silent) {
        toast.error('加载实例状态失败')
      }
    } finally {
      loading.value = false
      syncPollingState()
    }
  }

  async function start() {
    const currentChallengeId = toValue(challengeId)
    if (!currentChallengeId) return

    creating.value = true
    try {
      instance.value = await createInstance(currentChallengeId)
      if (instance.value.status === 'pending') {
        toast.info('实例已进入排队，正在等待创建')
      } else if (instance.value.status === 'creating') {
        toast.info('实例创建中，请稍候')
      } else if (instance.value.status === 'running') {
        toast.success('实例创建成功')
      } else {
        toast.info('实例状态已更新，请稍后查看')
      }
      syncPollingState()
    } catch (error) {
      if (error instanceof ApiError && error.message.includes('不需要靶机')) {
        toast.error('该题目不需要靶机，请直接提交 Flag')
        return
      }
      if (error instanceof ApiError && error.message) {
        toast.error(error.message)
        return
      }
      toast.error('创建实例失败')
    } finally {
      creating.value = false
    }
  }

  const {
    opening,
    extending,
    destroying,
    openInstance: open,
    extendInstance: extend,
    destroyInstance: destroy,
  } = useInstanceWorkflowActions({
    resolveTarget: () => instance.value,
    getExtendBlockedMessage: (target) =>
      target.share_scope === 'shared' ? '共享实例不支持手动延时' : null,
    getDestroyBlockedMessage: (target) =>
      target.share_scope === 'shared' ? '共享实例不支持手动销毁' : null,
    onExtended: async ({ target, result }) => {
      if (result) {
        instance.value = {
          ...target,
          expires_at: result.expires_at,
          remaining_extends: result.remaining_extends,
        }
        return
      }
      await refresh()
    },
    onDestroyed: () => {
      instance.value = null
      clearPollingTimer()
    },
    extendSuccessMessage: '延时成功',
    extendErrorMessage: '延时失败',
    destroySuccessMessage: '实例已销毁',
    destroyErrorMessage: '销毁实例失败',
  })

  watch(
    () => toValue(challengeId),
    () => {
      clearPollingTimer()
      void refresh()
    },
    { immediate: true }
  )

  onUnmounted(() => {
    clearPollingTimer()
  })

  return {
    instance,
    loading,
    creating,
    opening,
    extending,
    destroying,
    refresh,
    start,
    open,
    extend,
    destroy,
  }
}

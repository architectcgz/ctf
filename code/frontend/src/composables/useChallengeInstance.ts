import { ref, watch, type MaybeRefOrGetter, toValue } from 'vue'

import { createInstance } from '@/api/challenge'
import {
  destroyInstance as apiDestroyInstance,
  extendInstance as apiExtendInstance,
  getMyInstances,
  requestInstanceAccess,
} from '@/api/instance'
import type { InstanceData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import { useToast } from '@/composables/useToast'

export function useChallengeInstance(challengeId: MaybeRefOrGetter<string | undefined>) {
  const toast = useToast()

  const instance = ref<InstanceData | null>(null)
  const loading = ref(false)
  const creating = ref(false)
  const opening = ref(false)
  const extending = ref(false)
  const destroying = ref(false)

  async function refresh() {
    const currentChallengeId = toValue(challengeId)
    if (!currentChallengeId) {
      instance.value = null
      return
    }

    loading.value = true
    try {
      const instances = await getMyInstances()
      instance.value = instances.find((item) => String(item.challenge_id) === currentChallengeId) ?? null
    } catch (error) {
      toast.error('加载实例状态失败')
    } finally {
      loading.value = false
    }
  }

  async function start() {
    const currentChallengeId = toValue(challengeId)
    if (!currentChallengeId) return

    creating.value = true
    try {
      instance.value = await createInstance(currentChallengeId)
      toast.success('实例创建成功')
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

  async function open() {
    if (!instance.value) return

    opening.value = true
    try {
      const result = await requestInstanceAccess(instance.value.id)
      window.open(result.access_url, '_blank', 'noopener,noreferrer')
    } catch (error) {
      toast.error('打开目标失败')
    } finally {
      opening.value = false
    }
  }

  async function extend() {
    if (!instance.value) return

    extending.value = true
    try {
      const result = await apiExtendInstance(instance.value.id)
      if (result) {
        instance.value = {
          ...instance.value,
          expires_at: result.expires_at,
          remaining_extends: result.remaining_extends,
        }
      } else {
        await refresh()
      }
      toast.success('延时成功')
    } catch (error) {
      toast.error('延时失败')
    } finally {
      extending.value = false
    }
  }

  async function destroy() {
    if (!instance.value) return

    destroying.value = true
    try {
      await apiDestroyInstance(instance.value.id)
      instance.value = null
      toast.success('实例已销毁')
    } catch (error) {
      toast.error('销毁实例失败')
    } finally {
      destroying.value = false
    }
  }

  watch(
    () => toValue(challengeId),
    () => {
      void refresh()
    },
    { immediate: true }
  )

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

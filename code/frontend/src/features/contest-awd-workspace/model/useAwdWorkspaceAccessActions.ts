import { ref, toValue, type MaybeRefOrGetter } from 'vue'

import {
  requestContestAWDDefenseSSH,
  requestContestAWDTargetAccess,
} from '@/api/contest'
import { requestInstanceAccess } from '@/api/instance'
import type { AWDDefenseSSHAccessData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseAwdWorkspaceAccessActionsOptions {
  contestId: MaybeRefOrGetter<string>
}

export function useAwdWorkspaceAccessActions(options: UseAwdWorkspaceAccessActionsOptions) {
  const { contestId } = options
  const toast = useToast()

  const openingServiceKey = ref('')
  const openingSSHKey = ref('')
  const sshAccessByServiceId = ref<Record<string, AWDDefenseSSHAccessData>>({})
  const openingTargetKey = ref('')

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
    const resolvedContestId = toValue(contestId)
    if (!resolvedContestId || !serviceId || openingSSHKey.value) {
      return null
    }

    openingSSHKey.value = serviceId
    try {
      const result = await requestContestAWDDefenseSSH(resolvedContestId, serviceId)
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
    const resolvedContestId = toValue(contestId)
    if (!resolvedContestId || !serviceId || !victimTeamId || openingTargetKey.value) {
      return null
    }

    const targetKey = `${serviceId}:${victimTeamId}`
    openingTargetKey.value = targetKey

    try {
      const result = await requestContestAWDTargetAccess(resolvedContestId, serviceId, victimTeamId)
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

  function clearSSHAccess() {
    sshAccessByServiceId.value = {}
  }

  return {
    clearSSHAccess,
    openDefenseSSH,
    openService,
    openTarget,
    openingServiceKey,
    openingSSHKey,
    openingTargetKey,
    sshAccessByServiceId,
  }
}

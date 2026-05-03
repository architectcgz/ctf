import { ref, toValue, type MaybeRefOrGetter } from 'vue'

import {
  restartContestAWDServiceInstance,
  startContestAWDServiceInstance,
} from '@/api/contest'
import type { ContestAWDWorkspaceData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseAwdWorkspaceServiceActionsOptions {
  contestId: MaybeRefOrGetter<string>
  refreshAll: () => Promise<void>
}

export function useAwdWorkspaceServiceActions(options: UseAwdWorkspaceServiceActionsOptions) {
  const { contestId, refreshAll } = options
  const toast = useToast()

  const startingServiceKey = ref('')
  const serviceActionPendingById = ref<Record<string, boolean>>({})

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

  async function startService(serviceId: string): Promise<void> {
    const resolvedContestId = toValue(contestId)
    if (!resolvedContestId || !serviceId || startingServiceKey.value) {
      return
    }

    startingServiceKey.value = serviceId
    try {
      const instance = await startContestAWDServiceInstance(resolvedContestId, serviceId)
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
    const resolvedContestId = toValue(contestId)
    if (
      !resolvedContestId ||
      !serviceId ||
      startingServiceKey.value ||
      serviceActionPendingById.value[serviceId]
    ) {
      return
    }

    startingServiceKey.value = serviceId
    serviceActionPendingById.value = {
      ...serviceActionPendingById.value,
      [serviceId]: true,
    }
    try {
      await restartContestAWDServiceInstance(resolvedContestId, serviceId)
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

  return {
    clearSettledServiceActions,
    restartService,
    serviceActionPendingById,
    startService,
    startingServiceKey,
  }
}

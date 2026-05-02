import { ref, type Ref } from 'vue'

import {
  getContestAWDInstanceOrchestration,
  startContestAWDTeamServiceInstance,
} from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

import {
  createEmptyInstanceOrchestration,
  humanizeRequestError,
} from './awdAdminSupport'

interface UseAwdServiceOperationsOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
}

export function useAwdServiceOperations(options: UseAwdServiceOperationsOptions) {
  const { selectedContest } = options
  const toast = useToast()

  const instanceOrchestration = ref(createEmptyInstanceOrchestration())
  const loadingInstanceOrchestration = ref(false)
  const startingInstanceKey = ref<string | null>(null)

  function findInstanceItem(teamId: string, serviceId: string) {
    return instanceOrchestration.value.instances.find(
      (item) => item.team_id === teamId && item.service_id === serviceId && item.instance
    )
  }

  async function refreshInstanceOrchestration() {
    if (!selectedContest.value || selectedContest.value.mode !== 'awd') {
      instanceOrchestration.value = createEmptyInstanceOrchestration()
      return
    }

    loadingInstanceOrchestration.value = true
    try {
      instanceOrchestration.value = await getContestAWDInstanceOrchestration(selectedContest.value.id)
    } finally {
      loadingInstanceOrchestration.value = false
    }
  }

  async function startTeamServiceInstance(teamId: string, serviceId: string) {
    if (!selectedContest.value || startingInstanceKey.value) {
      return
    }

    const instanceKey = `${teamId}:${serviceId}`
    startingInstanceKey.value = instanceKey
    try {
      await startContestAWDTeamServiceInstance(selectedContest.value.id, {
        team_id: teamId,
        service_id: serviceId,
      })
      toast.success('队伍服务实例已启动')
      await refreshInstanceOrchestration()
    } catch (error) {
      toast.error(humanizeRequestError(error, '启动队伍服务实例失败'))
    } finally {
      startingInstanceKey.value = null
    }
  }

  async function startTeamAllServices(teamId: string) {
    if (!selectedContest.value || startingInstanceKey.value) {
      return
    }

    const serviceIds = instanceOrchestration.value.services
      .filter((service) => service.is_visible)
      .map((service) => service.service_id)
      .filter((serviceId) => !findInstanceItem(teamId, serviceId))
    if (serviceIds.length === 0) {
      toast.success('该队伍服务实例已全部启动')
      return
    }

    startingInstanceKey.value = `team:${teamId}`
    try {
      for (const serviceId of serviceIds) {
        await startContestAWDTeamServiceInstance(selectedContest.value.id, {
          team_id: teamId,
          service_id: serviceId,
        })
      }
      toast.success('队伍服务实例已批量启动')
      await refreshInstanceOrchestration()
    } catch (error) {
      toast.error(humanizeRequestError(error, '批量启动队伍服务实例失败'))
      await refreshInstanceOrchestration()
    } finally {
      startingInstanceKey.value = null
    }
  }

  async function startAllTeamServices() {
    if (!selectedContest.value || startingInstanceKey.value) {
      return
    }

    const targets = instanceOrchestration.value.teams.flatMap((team) =>
      instanceOrchestration.value.services
        .filter((service) => service.is_visible)
        .filter((service) => !findInstanceItem(team.team_id, service.service_id))
        .map((service) => ({
          teamId: team.team_id,
          serviceId: service.service_id,
        }))
    )
    if (targets.length === 0) {
      toast.success('所有队伍服务实例已启动')
      return
    }

    startingInstanceKey.value = 'all'
    try {
      for (const target of targets) {
        await startContestAWDTeamServiceInstance(selectedContest.value.id, {
          team_id: target.teamId,
          service_id: target.serviceId,
        })
      }
      toast.success('全部队伍服务实例已批量启动')
      await refreshInstanceOrchestration()
    } catch (error) {
      toast.error(humanizeRequestError(error, '批量启动全部实例失败'))
      await refreshInstanceOrchestration()
    } finally {
      startingInstanceKey.value = null
    }
  }

  return {
    instanceOrchestration,
    loadingInstanceOrchestration,
    startingInstanceKey,
    refreshInstanceOrchestration,
    startTeamServiceInstance,
    startTeamAllServices,
    startAllTeamServices,
  }
}

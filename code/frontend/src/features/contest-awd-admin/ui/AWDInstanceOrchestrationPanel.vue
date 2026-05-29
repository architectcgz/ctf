<script setup lang="ts">
import { computed } from 'vue'

import type {
  AdminContestAWDInstanceItemData,
  AdminContestAWDInstanceOrchestrationData,
  AdminContestAWDInstanceServiceData,
  AdminContestAWDInstanceTeamData,
} from '@/api/contracts'
import AWDInstanceOrchestrationHeader from './AWDInstanceOrchestrationHeader.vue'
import AWDInstanceOrchestrationMatrix from './AWDInstanceOrchestrationMatrix.vue'
import './awdInstanceOrchestrationPanel.css'
import type {
  AWDInstanceOrchestrationHeaderView,
  AWDInstanceOrchestrationRowView,
  AWDInstanceOrchestrationServiceView,
} from './awdInstanceOrchestration.types'

const props = defineProps<{
  orchestration: AdminContestAWDInstanceOrchestrationData
  loading: boolean
  startingKey: string | null
}>()

const emit = defineEmits<{
  refresh: []
  'start-cell': [teamId: string, serviceId: string]
  'start-team': [teamId: string]
  'start-all': []
}>()

const instanceMap = computed(() => {
  const map = new Map<string, AdminContestAWDInstanceItemData>()
  for (const item of props.orchestration.instances) {
    map.set(`${item.team_id}:${item.service_id}`, item)
  }
  return map
})

const visibleServices = computed(() =>
  props.orchestration.services.filter((service) => service.is_visible)
)
const serviceViews = computed<AWDInstanceOrchestrationServiceView[]>(() =>
  visibleServices.value.map((service) => ({
    serviceId: service.service_id,
    displayName: service.display_name,
  }))
)

const totalTargetCount = computed(
  () => props.orchestration.teams.length * visibleServices.value.length
)

const runningCount = computed(
  () =>
    props.orchestration.instances.filter(
      (item) =>
        item.instance &&
        visibleServices.value.some((service) => service.service_id === item.service_id)
    ).length
)

function getInstance(teamId: string, serviceId: string) {
  return instanceMap.value.get(`${teamId}:${serviceId}`)?.instance
}

function getCellKey(team: AdminContestAWDInstanceTeamData, service: AdminContestAWDInstanceServiceData) {
  return `${team.team_id}:${service.service_id}`
}

function isCellStarting(team: AdminContestAWDInstanceTeamData, service: AdminContestAWDInstanceServiceData) {
  const key = getCellKey(team, service)
  return props.startingKey === key || props.startingKey === `team:${team.team_id}` || props.startingKey === 'all'
}

function hasMissingService(teamId: string) {
  return visibleServices.value.some((service) => !getInstance(teamId, service.service_id))
}

function getStatusLabel(status?: string) {
  switch (status) {
    case 'pending':
      return '排队中'
    case 'creating':
      return '创建中'
    case 'running':
      return '运行中'
    case 'failed':
      return '失败'
    default:
      return '未启动'
  }
}

const hasPendingStart = computed(() => Boolean(props.startingKey))
const headerSummary = computed<AWDInstanceOrchestrationHeaderView>(() => ({
  runningCount: runningCount.value,
  totalTargetCount: totalTargetCount.value,
  loading: props.loading,
  hasPendingStart: hasPendingStart.value,
  canStartAll:
    !props.loading &&
    !hasPendingStart.value &&
    runningCount.value < totalTargetCount.value &&
    totalTargetCount.value > 0,
}))

const rowViews = computed<AWDInstanceOrchestrationRowView[]>(() =>
  props.orchestration.teams.map((team) => ({
    teamId: team.team_id,
    teamName: team.team_name,
    captainId: team.captain_id,
    hasMissingService: hasMissingService(team.team_id),
    cells: visibleServices.value.map((service) => {
      const instance = getInstance(team.team_id, service.service_id)
      return {
        teamId: team.team_id,
        serviceId: service.service_id,
        status: instance?.status,
        statusLabel: getStatusLabel(instance?.status),
        accessUrl: instance?.access_url,
        isStarting: isCellStarting(team, service),
      }
    }),
  }))
)
</script>

<template>
  <section class="awd-instance-orchestration workspace-directory-section">
    <AWDInstanceOrchestrationHeader
      :summary="headerSummary"
      @refresh="emit('refresh')"
      @start-all="emit('start-all')"
    />
    <AWDInstanceOrchestrationMatrix
      :services="serviceViews"
      :rows="rowViews"
      :loading="loading"
      :has-pending-start="hasPendingStart"
      @start-cell="(teamId, serviceId) => emit('start-cell', teamId, serviceId)"
      @start-team="(teamId) => emit('start-team', teamId)"
    />
  </section>
</template>

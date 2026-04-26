<script setup lang="ts">
import { computed } from 'vue'
import { Play, RefreshCw, Server, ShieldCheck } from 'lucide-vue-next'

import type {
  AdminContestAWDInstanceItemData,
  AdminContestAWDInstanceOrchestrationData,
  AdminContestAWDInstanceServiceData,
  AdminContestAWDInstanceTeamData,
} from '@/api/contracts'

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
</script>

<template>
  <section class="awd-instance-orchestration workspace-directory-section">
    <header class="orchestration-header">
      <div class="orchestration-heading">
        <div class="orchestration-overline">
          Runtime / Team Instances
        </div>
        <h3 class="orchestration-title">
          队伍实例编排
        </h3>
      </div>
      <div class="orchestration-actions">
        <div class="orchestration-summary">
          <ShieldCheck class="summary-icon" />
          <span>{{ runningCount }} / {{ totalTargetCount }}</span>
        </div>
        <button
          type="button"
          class="ops-btn ops-btn--neutral"
          :disabled="loading || Boolean(startingKey)"
          title="刷新实例编排"
          @click="emit('refresh')"
        >
          <RefreshCw
            class="btn-icon"
            :class="{ 'animate-spin': loading }"
          />
        </button>
        <button
          type="button"
          class="ops-btn ops-btn--primary"
          :disabled="loading || Boolean(startingKey) || runningCount >= totalTargetCount || totalTargetCount === 0"
          @click="emit('start-all')"
        >
          <Play class="btn-icon" />
          <span>启动全部</span>
        </button>
      </div>
    </header>

    <div
      v-if="orchestration.teams.length === 0 || visibleServices.length === 0"
      class="orchestration-empty"
    >
      <Server class="empty-icon" />
      <span>暂无可编排的队伍或可见服务</span>
    </div>

    <div
      v-else
      class="orchestration-table-wrap"
    >
      <table class="orchestration-table">
        <colgroup>
          <col class="team-col-track">
          <col
            v-for="service in visibleServices"
            :key="`service-col-${service.service_id}`"
            class="service-col-track"
          >
          <col class="action-col-track">
        </colgroup>
        <thead>
          <tr>
            <th class="team-col">
              队伍
            </th>
            <th
              v-for="service in visibleServices"
              :key="service.service_id"
              class="service-col"
            >
              {{ service.display_name }}
            </th>
            <th class="action-col">
              操作
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="team in orchestration.teams"
            :key="team.team_id"
          >
            <th class="team-cell">
              <span class="team-name">{{ team.team_name }}</span>
              <span class="team-meta">Captain {{ team.captain_id }}</span>
            </th>
            <td
              v-for="service in visibleServices"
              :key="getCellKey(team, service)"
              class="service-cell"
            >
              <div class="instance-cell">
                <template v-if="getInstance(team.team_id, service.service_id)">
                  <span
                    class="instance-status"
                    :class="`instance-status--${getInstance(team.team_id, service.service_id)?.status}`"
                  >
                    {{ getStatusLabel(getInstance(team.team_id, service.service_id)?.status) }}
                  </span>
                  <a
                    v-if="getInstance(team.team_id, service.service_id)?.access_url"
                    class="instance-link"
                    :href="getInstance(team.team_id, service.service_id)?.access_url"
                    target="_blank"
                    rel="noreferrer"
                  >
                    访问
                  </a>
                </template>
                <button
                  v-else
                  type="button"
                  class="cell-start-btn"
                  :disabled="loading || Boolean(startingKey)"
                  @click="emit('start-cell', team.team_id, service.service_id)"
                >
                  <Play
                    class="btn-icon"
                    :class="{ 'animate-spin': isCellStarting(team, service) }"
                  />
                  <span>{{ isCellStarting(team, service) ? '启动中' : '启动' }}</span>
                </button>
              </div>
            </td>
            <td class="row-action-cell">
              <button
                type="button"
                class="row-start-btn"
                :disabled="loading || Boolean(startingKey) || !hasMissingService(team.team_id)"
                @click="emit('start-team', team.team_id)"
              >
                启动本队
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.awd-instance-orchestration {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.orchestration-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
}

.orchestration-overline {
  margin-bottom: var(--space-1);
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.orchestration-title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-1-125);
  font-weight: 900;
}

.orchestration-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.orchestration-summary {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-3);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.summary-icon,
.btn-icon,
.empty-icon {
  width: var(--space-4);
  height: var(--space-4);
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-4);
  border-radius: 0.5rem;
  font-size: var(--font-size-13);
  font-weight: 800;
  transition: all 0.2s ease;
}

.ops-btn--neutral {
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
}

.ops-btn--primary {
  border: 1px solid color-mix(in srgb, var(--color-primary) 70%, var(--color-border-default));
  background: var(--color-primary);
  color: var(--color-bg-base);
}

.ops-btn:hover:not(:disabled),
.cell-start-btn:hover:not(:disabled),
.row-start-btn:hover:not(:disabled) {
  filter: brightness(1.04);
}

.ops-btn:disabled,
.cell-start-btn:disabled,
.row-start-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.orchestration-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: calc(var(--ui-control-height-lg) * 2);
  color: var(--color-text-muted);
  font-size: var(--font-size-13);
  font-weight: 700;
}

.orchestration-table-wrap {
  overflow-x: auto;
}

.orchestration-table {
  width: 100%;
  min-width: 48rem;
  border-collapse: collapse;
  table-layout: fixed;
}

.orchestration-table th,
.orchestration-table td {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: var(--space-3);
  text-align: left;
  vertical-align: middle;
}

.orchestration-table thead th {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
  text-transform: uppercase;
}

.team-col-track {
  width: 12rem;
}

.action-col-track {
  width: 7rem;
}

.team-col,
.team-cell {
  text-align: left;
}

.service-col,
.service-cell {
  text-align: center;
}

.action-col,
.row-action-cell {
  text-align: center;
}

.team-cell {
  color: var(--color-text-primary);
}

.team-name,
.team-meta {
  display: block;
}

.team-name {
  font-size: var(--font-size-13);
  font-weight: 900;
}

.team-meta {
  margin-top: var(--space-0-5);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.instance-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-md);
}

.instance-status {
  display: inline-flex;
  align-items: center;
  min-height: var(--ui-control-height-sm);
  padding: 0 var(--space-2-5);
  border-radius: 0.375rem;
  background: color-mix(in srgb, var(--color-text-muted) 12%, transparent);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.instance-status--running {
  background: color-mix(in srgb, var(--color-success) 16%, transparent);
  color: var(--color-success);
}

.instance-status--creating,
.instance-status--pending {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: var(--color-warning);
}

.instance-link {
  color: var(--color-primary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.cell-start-btn,
.row-start-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1-5);
  min-height: var(--ui-control-height-sm);
  padding: 0 var(--space-3);
  border: 1px solid var(--color-border-default);
  border-radius: 0.375rem;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.row-action-cell {
  white-space: nowrap;
}

@media (max-width: 720px) {
  .orchestration-header {
    align-items: stretch;
    flex-direction: column;
  }

  .orchestration-actions {
    flex-wrap: wrap;
  }

  .orchestration-table {
    min-width: 42rem;
  }
}
</style>

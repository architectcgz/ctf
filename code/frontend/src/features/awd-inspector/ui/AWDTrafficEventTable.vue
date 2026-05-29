<script setup lang="ts">
import { Search } from 'lucide-vue-next'

import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import type { AWDTrafficStatusGroup, AWDTrafficEventData } from '@/api/contracts'
import type { AWDTrafficFilters } from '@/features/awd-inspector/model'

import type { AWDTrafficServiceOptionView } from './awdTrafficPanel.types'

defineProps<{
  updatedAt?: string
  serviceOptions: AWDTrafficServiceOptionView[]
  trafficEvents: AWDTrafficEventData[]
  trafficEventsTotal: number
  trafficFilters: AWDTrafficFilters
  trafficTeamOptions: Array<{ id: string; name: string }>
  trafficStatusGroupOptions: Array<{ value: 'all' | AWDTrafficStatusGroup; label: string }>
  trafficPathKeywordInput: string
  trafficTotalPages: number
  loadingTrafficEvents: boolean
  formatDateTime: (value?: string) => string
  getTrafficTeamName: (teamId: string, teamName?: string) => string
  getTrafficChallengeTitle: (challengeId: string, fallbackTitle?: string) => string
  getTrafficStatusGroupLabel: (statusGroup: AWDTrafficStatusGroup) => string
  getTrafficStatusGroupClass: (statusGroup: AWDTrafficStatusGroup) => string
}>()

const emit = defineEmits<{
  updateTrafficPathKeywordInput: [value: string]
  applyTrafficKeywordFilter: []
  handleTrafficStatusGroupChange: [value: string]
  applyTrafficFilterPatch: [patch: Partial<AWDTrafficFilters>]
  resetTrafficFilters: []
  handleTrafficPageChange: [page: number]
}>()
</script>

<template>
  <div class="drill-down-area">
    <header class="drill-down-toolbar">
      <div class="toolbar-left">
        <h3 class="toolbar-title">
          流量审计明细
        </h3>
        <p class="toolbar-hint">
          最后同步：{{ formatDateTime(updatedAt) }}
        </p>
      </div>

      <div class="toolbar-right">
        <div class="filter-row">
          <label class="ui-field awd-round-filter-field">
            <span class="ui-field__label">攻击队</span>
            <span class="ui-control-wrap awd-round-filter-control">
              <select
                id="awd-traffic-filter-attacker"
                :value="trafficFilters.attacker_team_id"
                class="ui-control"
                @change="emit('applyTrafficFilterPatch', { attacker_team_id: ($event.target as HTMLSelectElement).value })"
              >
                <option value="">
                  全部攻击队
                </option>
                <option
                  v-for="team in trafficTeamOptions"
                  :key="`attacker-${team.id}`"
                  :value="team.id"
                >
                  {{ team.name }}
                </option>
              </select>
            </span>
          </label>
          <label class="ui-field awd-round-filter-field">
            <span class="ui-field__label">受害队</span>
            <span class="ui-control-wrap awd-round-filter-control">
              <select
                id="awd-traffic-filter-victim"
                :value="trafficFilters.victim_team_id"
                class="ui-control"
                @change="emit('applyTrafficFilterPatch', { victim_team_id: ($event.target as HTMLSelectElement).value })"
              >
                <option value="">
                  全部受害队
                </option>
                <option
                  v-for="team in trafficTeamOptions"
                  :key="`victim-${team.id}`"
                  :value="team.id"
                >
                  {{ team.name }}
                </option>
              </select>
            </span>
          </label>
          <label class="ui-field awd-round-filter-field">
            <span class="ui-field__label">服务引用</span>
            <span class="ui-control-wrap awd-round-filter-control">
              <select
                id="awd-traffic-filter-service"
                :value="trafficFilters.service_id"
                class="ui-control"
                @change="emit('applyTrafficFilterPatch', { service_id: ($event.target as HTMLSelectElement).value })"
              >
                <option value="">
                  所有服务
                </option>
                <option
                  v-for="item in serviceOptions"
                  :key="item.serviceId"
                  :value="item.serviceId"
                >
                  {{ item.title }} · Service #{{ item.serviceId }}
                </option>
              </select>
            </span>
          </label>
          <label class="ui-field awd-round-filter-field">
            <span class="ui-field__label">状态分桶</span>
            <span class="ui-control-wrap awd-round-filter-control">
              <select
                id="awd-traffic-filter-status-group"
                :value="trafficFilters.status_group"
                class="ui-control"
                @change="emit('handleTrafficStatusGroupChange', ($event.target as HTMLSelectElement).value)"
              >
                <option
                  v-for="item in trafficStatusGroupOptions"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ item.label }}
                </option>
              </select>
            </span>
          </label>
          <label class="ui-field awd-round-filter-field">
            <span class="ui-field__label">路径搜索</span>
            <span class="ui-control-wrap awd-round-filter-control">
              <Search class="h-3 w-3 search-icon" />
              <input
                :value="trafficPathKeywordInput"
                class="ui-control"
                placeholder="过滤路径..."
                @input="emit('updateTrafficPathKeywordInput', ($event.target as HTMLInputElement).value)"
                @keydown.enter.prevent="emit('applyTrafficKeywordFilter')"
              >
            </span>
          </label>
          <button
            type="button"
            class="ui-btn ui-btn--ghost awd-round-filter-search"
            @click="emit('applyTrafficKeywordFilter')"
          >
            搜索
          </button>
          <button
            id="awd-traffic-reset-filters"
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="emit('resetTrafficFilters')"
          >
            重置
          </button>
        </div>
      </div>
    </header>

    <div class="log-table-wrap">
      <table class="studio-table">
        <thead>
          <tr>
            <th class="w-32">
              捕获时间
            </th>
            <th>交互矢量</th>
            <th>关联靶题</th>
            <th>请求方法 & 路径</th>
            <th class="text-right">
              响应状态
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="event in trafficEvents"
            :key="event.occurred_at"
            class="studio-row"
          >
            <td class="traffic-time-cell font-mono">
              {{ formatDateTime(event.occurred_at).split(' ')[1] }}
            </td>
            <td>
              <div class="vector-cell">
                <span class="team-label">{{ getTrafficTeamName(event.attacker_team_id, event.attacker_team_name) }}</span>
                <span class="traffic-vector-separator">→</span>
                <span class="team-label">{{ getTrafficTeamName(event.victim_team_id, event.victim_team_name) }}</span>
              </div>
            </td>
            <td>
              <div class="challenge-cell">
                <span class="challenge-name">{{ getTrafficChallengeTitle(event.awd_challenge_id, event.awd_challenge_title) }}</span>
                <span
                  v-if="event.service_id"
                  class="source-tag font-mono"
                >Service #{{ event.service_id }}</span>
              </div>
            </td>
            <td>
              <div class="request-cell font-mono">
                <span class="method-tag">{{ event.method }}</span>
                <span class="path-text truncate">{{ event.path }}</span>
              </div>
            </td>
            <td class="text-right">
              <span
                class="status-badge"
                :class="getTrafficStatusGroupClass(event.status_group)"
              >
                {{ event.status_code }} · {{ getTrafficStatusGroupLabel(event.status_group) }}
              </span>
            </td>
          </tr>
          <tr v-if="!loadingTrafficEvents && trafficEvents.length === 0">
            <td
              colspan="5"
              class="traffic-empty-cell py-20 text-center font-medium"
            >
              满足当前过滤条件的流量记录为空
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination-footer">
      <PlatformPaginationControls
        :page="trafficFilters.page"
        :total-pages="trafficTotalPages"
        :total="trafficEventsTotal"
        :disabled="loadingTrafficEvents"
        @change-page="emit('handleTrafficPageChange', $event)"
      />
    </div>
  </div>
</template>

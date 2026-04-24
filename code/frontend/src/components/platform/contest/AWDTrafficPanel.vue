<script setup lang="ts">
import { computed, toRef } from 'vue'
import { Search } from 'lucide-vue-next'

import type { AWDTrafficStatusGroup } from '@/api/contracts'
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import type {
  AWDTrafficPanelEmits,
  AWDTrafficPanelProps,
} from '@/components/platform/contest/awdInspector.types'
import { useAwdTrafficPanel } from '@/composables/useAwdTrafficPanel'

const props = defineProps<AWDTrafficPanelProps>()
const emit = defineEmits<AWDTrafficPanelEmits>()

const serviceOptions = computed(() => {
  const seen = new Set<string>()
  return props.challengeLinks
    .filter((item) => {
      const serviceId = item.awd_service_id?.trim()
      if (!serviceId || seen.has(serviceId)) {
        return false
      }
      seen.add(serviceId)
      return true
    })
    .map((item) => ({
      serviceId: item.awd_service_id!,
      title: item.title,
    }))
})

const {
  trafficPathKeywordInput,
  trafficTotalPages,
  trafficTrendRows,
  trafficSummaryStats,
  trafficTrendNarrative,
  trafficStatusGroupOptions,
  applyTrafficKeywordFilter,
  onTrafficStatusGroupChange,
  applyTrafficFilterPatch,
  handleTrafficPageChange,
} = useAwdTrafficPanel({
  trafficSummary: toRef(props, 'trafficSummary'),
  trafficEventsTotal: toRef(props, 'trafficEventsTotal'),
  trafficFilters: toRef(props, 'trafficFilters'),
  loadingTrafficEvents: toRef(props, 'loadingTrafficEvents'),
  trafficPathKeyword: computed(() => props.trafficFilters.path_keyword),
  formatDateTime: props.formatDateTime,
  formatPercent: props.formatPercent,
  applyTrafficFilters: (patch) => emit('applyTrafficFilters', patch),
  changeTrafficPage: (page) => emit('changeTrafficPage', page),
})

function handleTrafficStatusGroupChange(value: string): void {
  onTrafficStatusGroupChange(value)
}

function getTrafficStatusGroupLabel(statusGroup: AWDTrafficStatusGroup): string {
  return props.getTrafficStatusGroupLabel(statusGroup)
}
</script>

<template>
  <div class="studio-traffic-analysis">
    <!-- 1. Integrated Stats Band -->
    <div
      v-if="trafficSummary"
      class="studio-metric-band"
    >
      <div
        v-for="item in trafficSummaryStats"
        :key="item.key"
        class="metric-pill awd-traffic-summary-card"
      >
        <span class="metric-pill__label">{{ item.label }}</span>
        <span class="metric-pill__value">{{ item.value }}</span>
      </div>
    </div>

    <!-- 2. Intelligence Grid -->
    <div
      v-if="trafficSummary"
      class="intelligence-grid"
    >
      <div class="intel-column">
        <header class="intel-header">
          热点实体分析
        </header>
        <div class="intel-cards">
          <div class="intel-sub-card">
            <div class="sub-card-label">
              主力攻击队
            </div>
            <div class="sub-card-list">
              <div
                v-for="item in trafficSummary.top_attackers.slice(0, 3)"
                :key="item.team_id"
                class="list-row"
              >
                <span class="row-name">{{ item.team_name }}</span>
                <span class="row-count font-mono">{{ item.request_count }}</span>
              </div>
            </div>
          </div>
          <div class="intel-sub-card">
            <div class="sub-card-label">
              高频受害队
            </div>
            <div class="sub-card-list">
              <div
                v-for="item in trafficSummary.top_victims.slice(0, 3)"
                :key="item.team_id"
                class="list-row"
              >
                <span class="row-name">{{ item.team_name }}</span>
                <span class="row-count font-mono">{{ item.request_count }}</span>
              </div>
            </div>
          </div>
          <div class="intel-sub-card">
            <div class="sub-card-label">
              异常交互路径
            </div>
            <div class="sub-card-list">
              <div
                v-for="item in trafficSummary.top_error_paths.slice(0, 3)"
                :key="item.path"
                class="list-row"
              >
                <span class="row-name row-name--path truncate">{{ item.path }}</span>
                <span class="row-count row-count--danger">{{ item.error_count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="intel-column">
        <header class="intel-header">
          流量趋势 (12-Bucket Trend)
        </header>
        <div class="trend-canvas">
          <div
            v-for="bucket in trafficTrendRows"
            :key="bucket.bucket_start_at"
            class="trend-unit"
          >
            <div class="trend-meta">
              <span class="trend-label">{{ bucket.label }}</span>
              <span class="trend-data">{{ bucket.request_count }} REQS</span>
            </div>
            <div class="trend-bar-track">
              <div
                class="trend-bar-fill"
                :style="{ width: `${bucket.ratio}%` }"
              />
            </div>
          </div>
          <p
            v-if="trafficTrendRows.length === 0"
            class="trend-empty py-4"
          >
            等待数据注入趋势桶...
          </p>
        </div>
      </div>
    </div>

    <!-- 3. Event Drill-down Table -->
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
                  @change="applyTrafficFilterPatch({ attacker_team_id: ($event.target as HTMLSelectElement).value })"
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
                  @change="applyTrafficFilterPatch({ victim_team_id: ($event.target as HTMLSelectElement).value })"
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
                  @change="applyTrafficFilterPatch({ service_id: ($event.target as HTMLSelectElement).value })"
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
                  @change="handleTrafficStatusGroupChange(($event.target as HTMLSelectElement).value)"
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
                  v-model="trafficPathKeywordInput"
                  class="ui-control"
                  placeholder="过滤路径..."
                  @keydown.enter.prevent="applyTrafficKeywordFilter"
                >
              </span>
            </label>
            <button
              type="button"
              class="ui-btn ui-btn--ghost awd-round-filter-search"
              @click="applyTrafficKeywordFilter"
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
              <td class="traffic-time-cell">
                {{ formatDateTime(event.occurred_at).split(' ')[1] }}
              </td>
              <td>
                <div class="vector-cell">
                  <span class="team-label">{{ getTrafficTeamName(event.attacker_team_id, event.attacker_team_name) }}</span>
                  <span class="vector-divider">→</span>
                  <span class="team-label">{{ getTrafficTeamName(event.victim_team_id, event.victim_team_name) }}</span>
                </div>
              </td>
              <td>
                <div class="challenge-cell">
                  <span class="challenge-name">{{ getTrafficChallengeTitle(event.challenge_id, event.challenge_title) }}</span>
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
                class="table-empty-state py-20 text-center"
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
          @change-page="handleTrafficPageChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.studio-traffic-analysis { display: flex; flex-direction: column; gap: var(--space-8); }

/* Metric Band */
.studio-metric-band { display: flex; gap: var(--space-2); background: var(--color-bg-elevated); padding: 1rem; border-radius: 1rem; border: 1px solid var(--color-border-default); }
.metric-pill { background: var(--color-bg-surface); border: 1px solid var(--color-border-default); padding: 0.45rem var(--space-4); border-radius: 0.75rem; display: flex; align-items: baseline; gap: var(--space-3); }
.metric-pill__label { font-size: var(--font-size-10); font-weight: 800; text-transform: uppercase; color: var(--color-text-secondary); letter-spacing: 0.05em; }
.metric-pill__value { font-size: var(--font-size-14); font-weight: 900; color: var(--color-text-primary); font-family: var(--font-family-mono); }

/* Intel Grid - Flattened */
.intelligence-grid { display: grid; grid-template-columns: 1fr 20rem; gap: var(--space-8); }
.intel-header { font-size: var(--font-size-11); font-weight: 900; text-transform: uppercase; color: var(--color-text-muted); letter-spacing: 0.1em; margin-bottom: 1.25rem; }
.intel-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-6); }
.intel-sub-card { background: transparent; border: none; padding: 0; }
.sub-card-label { font-size: var(--font-size-10); font-weight: 800; color: var(--color-text-secondary); text-transform: uppercase; margin-bottom: var(--space-3); }
.list-row { display: flex; justify-content: space-between; align-items: center; padding: 0.45rem 0; border-bottom: 1px solid var(--color-border-subtle); }
.row-name { font-size: var(--font-size-12); font-weight: 700; color: var(--color-text-primary); }
.row-count { font-size: var(--font-size-12); font-weight: 800; color: var(--color-primary); }
.row-name--path {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
}
.row-count--danger { color: var(--color-danger); }

.trend-canvas { background: var(--color-bg-elevated); border: 1px solid var(--color-border-default); border-radius: 1rem; padding: 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.trend-unit { display: flex; flex-direction: column; gap: 0.25rem; }
.trend-meta { display: flex; justify-content: space-between; font-size: var(--font-size-10); font-weight: 800; color: var(--color-text-muted); }
.trend-bar-track { height: 4px; background: var(--color-border-default); border-radius: 2px; overflow: hidden; }
.trend-bar-fill { height: 100%; background: var(--color-primary); border-radius: 2px; }
.trend-empty {
  font-size: var(--font-size-11);
  color: color-mix(in srgb, var(--color-text-muted) 90%, var(--color-text-secondary));
}

/* Drill-down area */
.drill-down-toolbar { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1.5rem; }
.toolbar-title { font-size: var(--font-size-15); font-weight: 900; color: var(--color-text-primary); margin: 0; }
.toolbar-hint { font-size: var(--font-size-12); color: var(--color-text-muted); margin-top: 0.25rem; }
.filter-row { display: flex; gap: 0.5rem; align-items: center; }

.search-input-wrap { position: relative; width: 12rem; }
.search-icon { position: absolute; left: 0.75rem; top: 50%; transform: translateY(-50%); color: var(--color-text-muted); }
.log-input { width: 100%; height: 2rem; padding: 0 0.75rem 0 2.25rem; font-size: var(--font-size-11); font-weight: 700; border-radius: 0.5rem; border: 1px solid var(--color-border-default); background: var(--color-bg-surface); color: var(--color-text-primary); outline: none; }
.log-select { height: 2rem; padding: 0 0.5rem; font-size: var(--font-size-11); font-weight: 700; border-radius: 0.5rem; border: 1px solid var(--color-border-default); background: var(--color-bg-surface); color: var(--color-text-secondary); }

.log-table-wrap { border: none; border-radius: 0; background: transparent; overflow: hidden; }
.studio-table { width: 100%; border-collapse: collapse; }
.studio-table th { background: var(--color-bg-elevated); padding: 0.75rem 1rem; text-align: left; font-size: var(--font-size-10); font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); border-bottom: 1px solid var(--color-border-default); border-top: 1px solid var(--color-border-default); }
.studio-table td { padding: 0.85rem 1rem; border-bottom: 1px solid var(--color-border-subtle); }
.studio-row:hover { background: var(--color-bg-elevated); }

.team-label { font-size: var(--font-size-12); font-weight: 800; color: var(--color-text-primary); }
.vector-divider { color: color-mix(in srgb, var(--color-text-muted) 88%, var(--color-border-default)); }
.challenge-cell { display: flex; flex-direction: column; }
.challenge-name { font-size: var(--font-size-12); font-weight: 700; color: var(--color-text-primary); }
.source-tag { font-size: var(--font-size-10); font-weight: 800; color: var(--color-text-muted); }
.traffic-time-cell {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  color: color-mix(in srgb, var(--color-text-muted) 90%, var(--color-text-secondary));
}

.request-cell { display: flex; align-items: center; gap: 0.75rem; font-size: var(--font-size-11); max-width: 20rem; }
.method-tag { color: var(--color-primary); font-weight: 900; }
.path-text { color: var(--color-text-secondary); }

.status-badge { font-size: var(--font-size-10); font-weight: 900; padding: 0.15rem 0.6rem; border-radius: 99px; }
.status-badge.status-group-success { background: var(--color-success-soft); color: var(--color-success); }
.status-badge.status-group-client-error { background: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface)); color: var(--color-warning); }
.status-badge.status-group-server-error { background: var(--color-danger-soft); color: var(--color-danger); }

.pagination-footer { padding: 1rem 0; border-top: 1px solid var(--color-border-default); }
.w-32 { width: 8rem; }
.table-empty-state {
  color: color-mix(in srgb, var(--color-text-muted) 90%, var(--color-text-secondary));
  font-weight: 500;
}
</style>

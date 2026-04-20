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
    <div v-if="trafficSummary" class="studio-metric-band">
      <div v-for="item in trafficSummaryStats" :key="item.key" class="metric-pill">
        <span class="metric-pill__label">{{ item.label }}</span>
        <span class="metric-pill__value">{{ item.value }}</span>
      </div>
    </div>

    <!-- 2. Intelligence Grid -->
    <div v-if="trafficSummary" class="intelligence-grid">
      <div class="intel-column">
        <header class="intel-header">热点实体分析</header>
        <div class="intel-cards">
          <div class="intel-sub-card">
            <div class="sub-card-label">主力攻击队</div>
            <div class="sub-card-list">
              <div v-for="item in trafficSummary.top_attackers.slice(0, 3)" :key="item.team_id" class="list-row">
                <span class="row-name">{{ item.team_name }}</span>
                <span class="row-count font-mono">{{ item.request_count }}</span>
              </div>
            </div>
          </div>
          <div class="intel-sub-card">
            <div class="sub-card-label">高频受害队</div>
            <div class="sub-card-list">
              <div v-for="item in trafficSummary.top_victims.slice(0, 3)" :key="item.team_id" class="list-row">
                <span class="row-name">{{ item.team_name }}</span>
                <span class="row-count font-mono">{{ item.request_count }}</span>
              </div>
            </div>
          </div>
          <div class="intel-sub-card">
            <div class="sub-card-label">异常交互路径</div>
            <div class="sub-card-list">
              <div v-for="item in trafficSummary.top_error_paths.slice(0, 3)" :key="item.path" class="list-row">
                <span class="row-name truncate font-mono text-[10px]">{{ item.path }}</span>
                <span class="row-count text-red-500 font-mono">{{ item.error_count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="intel-column">
        <header class="intel-header">流量趋势 (12-Bucket Trend)</header>
        <div class="trend-canvas">
          <div v-for="bucket in trafficTrendRows" :key="bucket.bucket_start_at" class="trend-unit">
            <div class="trend-meta">
              <span class="trend-label">{{ bucket.label }}</span>
              <span class="trend-data">{{ bucket.request_count }} REQS</span>
            </div>
            <div class="trend-bar-track">
              <div class="trend-bar-fill" :style="{ width: `${bucket.ratio}%` }"></div>
            </div>
          </div>
          <p v-if="trafficTrendRows.length === 0" class="text-[11px] text-slate-400 py-4">等待数据注入趋势桶...</p>
        </div>
      </div>
    </div>

    <!-- 3. Event Drill-down Table -->
    <div class="drill-down-area">
      <header class="drill-down-toolbar">
        <div class="toolbar-left">
          <h3 class="toolbar-title">流量审计明细</h3>
          <p class="toolbar-hint">最后同步：{{ formatDateTime(updatedAt) }}</p>
        </div>
        
        <div class="toolbar-right">
          <div class="filter-row">
            <select :value="trafficFilters.status_group" class="log-select" @change="handleTrafficStatusGroupChange(($event.target as HTMLSelectElement).value)">
              <option v-for="item in trafficStatusGroupOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
            <div class="search-input-wrap">
              <Search class="h-3 w-3 search-icon" />
              <input v-model="trafficPathKeywordInput" class="log-input" placeholder="过滤路径..." @keydown.enter.prevent="applyTrafficKeywordFilter" />
            </div>
            <button class="ops-btn ops-btn--neutral" @click="emit('resetTrafficFilters')">重置</button>
          </div>
        </div>
      </header>

      <div class="log-table-wrap">
        <table class="studio-table">
          <thead>
            <tr>
              <th class="w-32">捕获时间</th>
              <th>交互矢量</th>
              <th>关联靶题</th>
              <th>请求方法 & 路径</th>
              <th class="text-right">响应状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="event in trafficEvents" :key="event.occurred_at" class="studio-row">
              <td class="font-mono text-[11px] text-slate-400">{{ formatDateTime(event.occurred_at).split(' ')[1] }}</td>
              <td>
                <div class="vector-cell">
                  <span class="team-label">{{ getTrafficTeamName(event.attacker_team_id, event.attacker_team_name) }}</span>
                  <span class="text-slate-300">→</span>
                  <span class="team-label">{{ getTrafficTeamName(event.victim_team_id, event.victim_team_name) }}</span>
                </div>
              </td>
              <td>
                <div class="challenge-cell">
                  <span class="challenge-name">{{ getTrafficChallengeTitle(event.challenge_id, event.challenge_title) }}</span>
                  <span v-if="event.service_id" class="source-tag font-mono">#{{ event.service_id }}</span>
                </div>
              </td>
              <td>
                <div class="request-cell font-mono">
                  <span class="method-tag">{{ event.method }}</span>
                  <span class="path-text truncate">{{ event.path }}</span>
                </div>
              </td>
              <td class="text-right">
                <span class="status-badge" :class="getTrafficStatusGroupClass(event.status_group)">
                  {{ event.status_code }} · {{ getTrafficStatusGroupLabel(event.status_group) }}
                </span>
              </td>
            </tr>
            <tr v-if="!loadingTrafficEvents && trafficEvents.length === 0">
              <td colspan="5" class="py-20 text-center text-slate-400 font-medium">满足当前过滤条件的流量记录为空</td>
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
.studio-traffic-analysis { display: flex; flex-direction: column; gap: 2rem; }

/* Metric Band */
.studio-metric-band { display: flex; gap: 0.5rem; background: #f1f5f9; padding: 1rem; border-radius: 1rem; border: 1px solid #e2e8f0; }
.metric-pill { background: white; border: 1px solid #e2e8f0; padding: 0.45rem 1rem; border-radius: 0.75rem; display: flex; align-items: baseline; gap: 0.75rem; }
.metric-pill__label { font-size: 8px; font-weight: 800; text-transform: uppercase; color: #64748b; letter-spacing: 0.05em; }
.metric-pill__value { font-size: 13px; font-weight: 900; color: #0f172a; font-family: var(--font-family-mono); }

/* Intel Grid */
.intelligence-grid { display: grid; grid-template-columns: 1fr 20rem; gap: 1.5rem; }
.intel-header { font-size: 11px; font-weight: 900; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.1em; margin-bottom: 1rem; }
.intel-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
.intel-sub-card { background: #f8fafc; border: none; border-radius: 0.75rem; padding: 1.25rem; }
.sub-card-label { font-size: 9px; font-weight: 800; color: #64748b; text-transform: uppercase; margin-bottom: 0.75rem; }
.list-row { display: flex; justify-content: space-between; align-items: center; padding: 0.35rem 0; border-bottom: 1px solid #edf2f7; }
.row-name { font-size: 11px; font-weight: 700; color: #1e293b; }
.row-count { font-size: 11px; font-weight: 800; color: #3b82f6; }

.trend-canvas { background: #f8fafc; border: none; border-radius: 0.75rem; padding: 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.trend-unit { display: flex; flex-direction: column; gap: 0.25rem; }
.trend-meta { display: flex; justify-content: space-between; font-size: 9px; font-weight: 800; color: #94a3b8; }
.trend-bar-track { height: 4px; background: #e2e8f0; border-radius: 2px; overflow: hidden; }
.trend-bar-fill { height: 100%; background: #3b82f6; border-radius: 2px; }

/* Drill-down area */
.drill-down-toolbar { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1.5rem; }
.toolbar-title { font-size: 14px; font-weight: 900; color: #0f172a; margin: 0; }
.toolbar-hint { font-size: 11px; color: #94a3b8; margin-top: 0.25rem; }
.filter-row { display: flex; gap: 0.5rem; align-items: center; }

.search-input-wrap { position: relative; width: 12rem; }
.search-icon { position: absolute; left: 0.75rem; top: 50%; transform: translateY(-50%); color: #94a3b8; }
.log-input { width: 100%; height: 2rem; padding: 0 0.75rem 0 2.25rem; font-size: 11px; font-weight: 700; border-radius: 0.5rem; border: 1px solid #e2e8f0; outline: none; }
.log-select { height: 2rem; padding: 0 0.5rem; font-size: 11px; font-weight: 700; border-radius: 0.5rem; border: 1px solid #e2e8f0; background: white; color: #475569; }

.log-table-wrap { border: none; border-radius: 0; background: white; overflow: hidden; }
.studio-table { width: 100%; border-collapse: collapse; }
.studio-table th { background: #f8fafc; padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; border-bottom: 1px solid #e2e8f0; border-top: 1px solid #e2e8f0; }
.studio-table td { padding: 0.85rem 1rem; border-bottom: 1px solid #f1f5f9; }
.studio-row:hover { background: #f8fafc; }

.team-label { font-size: 11px; font-weight: 800; color: #1e293b; }
.challenge-cell { display: flex; flex-direction: column; }
.challenge-name { font-size: 11px; font-weight: 700; color: #1e293b; }
.source-tag { font-size: 9px; font-weight: 800; color: #94a3b8; }

.request-cell { display: flex; align-items: center; gap: 0.75rem; font-size: 11px; max-width: 20rem; }
.method-tag { color: #2563eb; font-weight: 900; }
.path-text { color: #64748b; }

.status-badge { font-size: 9px; font-weight: 900; padding: 0.15rem 0.6rem; border-radius: 99px; }
.status-badge.status-group-success { background: #dcfce7; color: #166534; }
.status-badge.status-group-client-error { background: #fff7ed; color: #9a3412; }
.status-badge.status-group-server-error { background: #fee2e2; color: #991b1b; }

.pagination-footer { padding: 1rem; border-top: 1px solid #e2e8f0; }
.ops-btn { display: inline-flex; align-items: center; justify-content: center; height: 2rem; padding: 0 0.85rem; border-radius: 0.6rem; font-size: 11px; font-weight: 700; background: white; border: 1px solid #e2e8f0; color: #475569; cursor: pointer; }
.w-32 { width: 8rem; }
</style>

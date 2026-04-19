<script setup lang="ts">
import { computed, toRef } from 'vue'
import { Search } from 'lucide-vue-next'

import type { AWDTrafficStatusGroup } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import type {
  AWDTrafficPanelEmits,
  AWDTrafficPanelProps,
} from '@/components/admin/contest/awdInspector.types'
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
  clearTrafficKeywordFilter,
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

function getTrafficStatusGroupClass(statusGroup: AWDTrafficStatusGroup): string {
  return props.getTrafficStatusGroupClass(statusGroup)
}

function getTrafficStatusGroupLabel(statusGroup: AWDTrafficStatusGroup): string {
  return props.getTrafficStatusGroupLabel(statusGroup)
}
</script>

<template>
  <section class="space-y-4 border-t border-border pt-6">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h3 class="awd-traffic-title text-base font-semibold">攻击流量态势</h3>
        <p class="awd-traffic-muted mt-1 text-xs">
          代理请求摘要，不等同于已确认攻破结果。
        </p>
      </div>
      <span class="awd-traffic-muted text-xs">最近更新时间：{{ formatDateTime(updatedAt) }}</span>
    </div>

    <div
      v-if="loadingTrafficSummary"
      class="awd-traffic-empty rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm"
    >
      正在加载攻击流量摘要...
    </div>
    <div
      v-else-if="!trafficSummary"
      class="awd-traffic-empty rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm"
    >
      当前轮次暂未返回攻击流量摘要。
    </div>
    <template v-else>
      <div class="awd-traffic-summary-strip border-y border-border/80">
        <div class="grid gap-0 md:grid-cols-2 xl:grid-cols-5">
          <div
            v-for="item in trafficSummaryStats"
            :key="item.key"
            class="awd-traffic-summary-card border-b border-border/70 px-4 py-4 last:border-b-0 xl:border-b-0 xl:border-r xl:last:border-r-0"
          >
            <p class="awd-traffic-stat-label">
              {{ item.label }}
            </p>
            <p class="awd-traffic-stat-value mt-3 text-2xl font-semibold tracking-tight">
              {{ item.value }}
            </p>
            <p class="awd-traffic-stat-hint mt-2 text-xs leading-6">
              {{ item.hint }}
            </p>
          </div>
        </div>
      </div>

      <div class="awd-traffic-overview-grid grid gap-4">
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-xl border border-border/80">
            <div class="awd-traffic-block-head border-b border-border bg-surface-alt/40 px-3 py-2">
              热点攻击队
            </div>
            <ol class="divide-y divide-border/70">
              <li
                v-for="item in trafficSummary.top_attackers.slice(0, 5)"
                :key="`traffic-attacker-${item.team_id}`"
                class="flex items-center justify-between px-3 py-2 text-sm"
              >
                <span class="awd-traffic-primary">{{ item.team_name }}</span>
                <span class="awd-traffic-secondary font-medium">{{
                  item.request_count
                }}</span>
              </li>
              <li v-if="trafficSummary.top_attackers.length === 0" class="awd-traffic-empty-copy px-3 py-3 text-xs">
                暂无攻击队热点数据
              </li>
            </ol>
          </div>
          <div class="rounded-xl border border-border/80">
            <div class="awd-traffic-block-head border-b border-border bg-surface-alt/40 px-3 py-2">
              热点受害队
            </div>
            <ol class="divide-y divide-border/70">
              <li
                v-for="item in trafficSummary.top_victims.slice(0, 5)"
                :key="`traffic-victim-${item.team_id}`"
                class="flex items-center justify-between px-3 py-2 text-sm"
              >
                <span class="awd-traffic-primary">{{ item.team_name }}</span>
                <span class="awd-traffic-secondary font-medium">{{
                  item.request_count
                }}</span>
              </li>
              <li v-if="trafficSummary.top_victims.length === 0" class="awd-traffic-empty-copy px-3 py-3 text-xs">
                暂无目标热点数据
              </li>
            </ol>
          </div>
          <div class="rounded-xl border border-border/80">
            <div class="awd-traffic-block-head border-b border-border bg-surface-alt/40 px-3 py-2">
              热点题目
            </div>
            <ol class="divide-y divide-border/70">
              <li
                v-for="item in trafficSummary.top_challenges.slice(0, 5)"
                :key="`traffic-challenge-${item.challenge_id}`"
                class="flex items-center justify-between gap-3 px-3 py-2 text-sm"
              >
                <span class="awd-traffic-primary truncate">{{
                  getTrafficChallengeTitle(item.challenge_id, item.challenge_title)
                }}</span>
                <span class="awd-traffic-secondary shrink-0 font-medium">{{
                  item.request_count
                }}</span>
              </li>
              <li v-if="trafficSummary.top_challenges.length === 0" class="awd-traffic-empty-copy px-3 py-3 text-xs">
                暂无题目热点数据
              </li>
            </ol>
          </div>
          <div class="rounded-xl border border-border/80">
            <div class="awd-traffic-block-head border-b border-border bg-surface-alt/40 px-3 py-2">
              异常路径
            </div>
            <ol class="divide-y divide-border/70">
              <li
                v-for="item in trafficSummary.top_error_paths.slice(0, 5)"
                :key="`traffic-path-${item.path}`"
                class="px-3 py-2 text-sm"
              >
                <p class="awd-traffic-primary truncate font-mono">
                  {{ item.path }}
                </p>
                <p class="awd-traffic-muted mt-1 text-xs">
                  请求 {{ item.request_count }} / 错误 {{ item.error_count }}
                </p>
              </li>
              <li v-if="trafficSummary.top_error_paths.length === 0" class="awd-traffic-empty-copy px-3 py-3 text-xs">
                暂无异常路径数据
              </li>
            </ol>
          </div>
        </div>

        <div class="rounded-xl border border-border/80">
          <div class="border-b border-border bg-surface-alt/40 px-3 py-2">
            <p class="awd-traffic-block-head">
              趋势摘要（最近 12 桶）
            </p>
            <p class="awd-traffic-muted mt-2 text-xs leading-6">
              {{ trafficTrendNarrative }}
            </p>
          </div>
          <div class="space-y-2 px-3 py-3">
            <div
              v-for="bucket in trafficTrendRows"
              :key="bucket.bucket_start_at"
              class="space-y-1"
            >
              <div class="awd-traffic-muted flex items-center justify-between text-xs">
                <span>{{ bucket.label }}</span>
                <span>请求 {{ bucket.request_count }} / 错误 {{ bucket.error_count }}</span>
              </div>
              <div class="awd-traffic-trend-track h-1.5 overflow-hidden rounded-full">
                <div
                  class="awd-traffic-trend-bar h-full rounded-full"
                  :style="{ width: `${bucket.ratio}%` }"
                />
              </div>
            </div>
            <p v-if="trafficTrendRows.length === 0" class="awd-traffic-muted text-xs">
              当前没有趋势桶数据。
            </p>
          </div>
        </div>
      </div>
    </template>

    <div class="overflow-hidden rounded-xl border border-border">
      <div class="awd-traffic-filter-grid grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3">
        <label class="ui-field awd-round-filter-field">
          <span class="ui-field__label">攻击方</span>
          <span class="ui-control-wrap awd-round-filter-control">
            <select
              id="awd-traffic-filter-attacker"
              :value="trafficFilters.attacker_team_id"
              class="ui-control"
              @change="
                applyTrafficFilterPatch({
                  attacker_team_id: ($event.target as HTMLSelectElement).value,
                })
              "
            >
              <option value="">全部攻击方</option>
              <option
                v-for="team in trafficTeamOptions"
                :key="`traffic-attacker-option-${team.id}`"
                :value="team.id"
              >
                {{ team.name }}
              </option>
            </select>
          </span>
        </label>
        <label class="ui-field awd-round-filter-field">
          <span class="ui-field__label">受害方</span>
          <span class="ui-control-wrap awd-round-filter-control">
            <select
              id="awd-traffic-filter-victim"
              :value="trafficFilters.victim_team_id"
              class="ui-control"
              @change="
                applyTrafficFilterPatch({
                  victim_team_id: ($event.target as HTMLSelectElement).value,
                })
              "
            >
              <option value="">全部受害方</option>
              <option
                v-for="team in trafficTeamOptions"
                :key="`traffic-victim-option-${team.id}`"
                :value="team.id"
              >
                {{ team.name }}
              </option>
            </select>
          </span>
        </label>
        <label class="ui-field awd-round-filter-field">
          <span class="ui-field__label">题目</span>
          <span class="ui-control-wrap awd-round-filter-control">
            <select
              id="awd-traffic-filter-challenge"
              :value="trafficFilters.challenge_id"
              class="ui-control"
              @change="
                applyTrafficFilterPatch({
                  challenge_id: ($event.target as HTMLSelectElement).value,
                })
              "
            >
              <option value="">全部题目</option>
              <option
                v-for="challenge in challengeLinks"
                :key="challenge.id"
                :value="challenge.challenge_id"
              >
                {{ challenge.title || `Challenge #${challenge.challenge_id}` }}
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
              <option v-for="item in trafficStatusGroupOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </option>
            </select>
          </span>
        </label>
        <label class="ui-field awd-round-filter-field">
          <span class="ui-field__label">路径关键字</span>
          <div class="flex items-center gap-2">
            <span class="ui-control-wrap awd-round-filter-control">
              <input
                id="awd-traffic-filter-path"
                v-model="trafficPathKeywordInput"
                type="text"
                class="ui-control"
                placeholder="/api/..."
                @keydown.enter.prevent="applyTrafficKeywordFilter"
              />
            </span>
            <button
              id="awd-traffic-filter-search"
              type="button"
              class="ui-btn ui-btn--ghost awd-round-filter-search"
              @click="applyTrafficKeywordFilter"
            >
              <Search class="h-4 w-4" />
            </button>
          </div>
          <button
            v-if="trafficFilters.path_keyword"
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="clearTrafficKeywordFilter"
          >
            清除路径关键字
          </button>
        </label>
        <div class="flex items-end md:justify-end">
          <button
            id="awd-traffic-reset-filters"
            type="button"
            class="ui-btn ui-btn--secondary awd-traffic-reset-button"
            @click="emit('resetTrafficFilters')"
          >
            重置筛选
          </button>
        </div>
      </div>

      <table class="min-w-full divide-y divide-border">
        <thead class="awd-traffic-table-head">
          <tr>
            <th class="px-4 py-3">时间</th>
            <th class="px-4 py-3">攻击方 / 受害方</th>
            <th class="px-4 py-3">靶题</th>
            <th class="px-4 py-3">请求</th>
            <th class="px-4 py-3">状态</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border bg-surface/70">
          <tr v-if="loadingTrafficEvents">
            <td colspan="5" class="awd-traffic-empty-row">
              正在加载流量明细...
            </td>
          </tr>
          <tr
            v-for="event in trafficEvents"
            :key="`${event.occurred_at}-${event.attacker_team_id}-${event.victim_team_id}-${event.challenge_id}-${event.method}-${event.path}`"
          >
            <td class="awd-traffic-table-cell awd-traffic-table-cell--secondary">
              {{ formatDateTime(event.occurred_at) }}
            </td>
            <td class="awd-traffic-table-cell awd-traffic-table-cell--primary">
              <p>
                {{ getTrafficTeamName(event.attacker_team_id, event.attacker_team_name) }}
              </p>
              <p class="awd-traffic-muted mt-1 text-xs">
                → {{ getTrafficTeamName(event.victim_team_id, event.victim_team_name) }}
              </p>
            </td>
            <td class="awd-traffic-table-cell awd-traffic-table-cell--secondary">
              {{ getTrafficChallengeTitle(event.challenge_id, event.challenge_title) }}
            </td>
            <td class="awd-traffic-table-cell">
              <p class="awd-traffic-primary font-mono">
                {{ event.method.toUpperCase() }} {{ event.path }}
              </p>
              <p class="awd-traffic-muted mt-1 text-xs">HTTP {{ event.status_code }}</p>
            </td>
            <td class="awd-traffic-table-cell">
              <span
                class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                :class="getTrafficStatusGroupClass(event.status_group)"
              >
                {{ getTrafficStatusGroupLabel(event.status_group) }}
              </span>
              <p class="awd-traffic-muted mt-1 text-xs">
                {{ getTrafficSourceLabel(event.source) }}
              </p>
            </td>
          </tr>
          <tr v-if="!loadingTrafficEvents && trafficEvents.length === 0">
            <td colspan="5" class="awd-traffic-empty-row">
              当前筛选条件下没有流量事件。
            </td>
          </tr>
        </tbody>
      </table>

      <div class="awd-traffic-pagination border-t border-border bg-surface-alt/20 px-4 py-3 text-xs">
        <AdminPaginationControls
          :page="trafficFilters.page"
          :total-pages="trafficTotalPages"
          :total="trafficEventsTotal"
          :disabled="loadingTrafficEvents"
          :total-label="`共 ${trafficEventsTotal} 条流量事件`"
          prev-button-id="awd-traffic-page-prev"
          next-button-id="awd-traffic-page-next"
          @change-page="handleTrafficPageChange"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.awd-traffic-summary-strip {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-primary) 6%, transparent),
    color-mix(in srgb, var(--color-primary) 0%, transparent)
  );
}

.awd-traffic-title,
.awd-traffic-primary {
  color: var(--color-text-primary);
}

.awd-traffic-muted,
.awd-traffic-stat-hint,
.awd-traffic-empty,
.awd-traffic-empty-copy {
  color: var(--color-text-muted);
}

.awd-traffic-secondary {
  color: var(--color-text-secondary);
}

.awd-traffic-table-head {
  background: color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-bg-base));
  text-align: left;
  font-size: var(--font-size-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  color: var(--color-text-muted);
}

.awd-traffic-table-cell {
  padding: var(--space-4);
  font-size: var(--font-size-sm);
}

.awd-traffic-table-cell--primary {
  color: var(--color-text-primary);
}

.awd-traffic-table-cell--secondary,
.awd-traffic-pagination {
  color: var(--color-text-secondary);
}

.awd-traffic-empty-row {
  padding: var(--space-8) var(--space-4);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.awd-traffic-stat-label,
.awd-traffic-block-head {
  font-size: var(--font-size-11);
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.awd-traffic-stat-value {
  color: var(--color-text-primary);
}

.awd-traffic-trend-track {
  background: color-mix(in srgb, var(--color-text-muted) 20%, transparent);
}

.awd-traffic-trend-bar {
  background: color-mix(in srgb, var(--color-primary) 70%, transparent);
}

.awd-round-filter-field {
  --ui-field-gap: var(--space-2);
  --ui-field-label-size: var(--font-size-11);
  --ui-field-label-weight: 700;
  --ui-field-label-color: var(--color-text-muted);
  min-width: 0;
}

.awd-round-filter-field .ui-field__label {
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.awd-round-filter-control {
  width: 100%;
}

.awd-round-filter-search {
  flex-shrink: 0;
  min-width: var(--ui-control-height-md);
  padding: 0;
}

@media (min-width: 1280px) {
  .awd-traffic-overview-grid {
    grid-template-columns: 1.3fr 0.7fr;
  }

  .awd-traffic-filter-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr)) minmax(0, 1.35fr) auto;
  }

  .awd-traffic-summary-card {
    border-bottom-width: 0;
  }
}

@media (min-width: 768px) and (max-width: 1279px) {
  .awd-traffic-summary-card:nth-last-child(-n + 2) {
    border-bottom-width: 0;
  }
}
</style>

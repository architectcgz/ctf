<script setup lang="ts">
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  FileDown,
  Info,
  Layers,
  SearchCheck,
  ShieldCheck,
  ShieldX,
  Zap,
} from 'lucide-vue-next'
import type {
  AWDServiceStatusPanelEmits,
  AWDServiceStatusPanelProps,
} from '@/components/platform/contest/awdInspector.types'

const props = defineProps<AWDServiceStatusPanelProps>()
const emit = defineEmits<AWDServiceStatusPanelEmits>()

function updateServiceStatusFilter(value: string): void {
  if (value !== 'all' && value !== 'up' && value !== 'down' && value !== 'compromised') {
    return
  }
  emit('updateServiceStatusFilter', value)
}

function getServiceCheckActions(checkResult: Record<string, unknown>) {
  return props.getCheckActions(checkResult)
}

function getServiceCheckTargets(checkResult: Record<string, unknown>) {
  return props.getCheckTargets(checkResult)
}
</script>

<template>
  <div class="awd-status-panel">
    <header class="awd-panel-header">
      <div class="awd-panel-identity">
        <Layers class="h-4.5 w-4.5 text-primary" />
        <h3 class="awd-panel-title">服务运行矩阵</h3>
      </div>
      <button
        id="awd-export-services"
        type="button"
        class="ui-btn ui-btn--sm ui-btn--ghost awd-service-export-button"
        :disabled="filteredServices.length === 0"
        @click="emit('exportServices')"
      >
        <FileDown class="h-3.5 w-3.5" />
        导出报告
      </button>
    </header>

    <div class="awd-filter-bar">
      <label class="awd-filter-field">
        <span class="awd-filter-label">队伍范围</span>
        <select
          id="awd-service-filter-team"
          :value="serviceTeamFilter"
          class="awd-filter-control"
          @change="emit('updateServiceTeamFilter', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">全部队伍</option>
          <option v-for="team in serviceTeamOptions" :key="team.team_id" :value="team.team_id">
            {{ team.team_name }}
          </option>
        </select>
      </label>

      <label class="awd-filter-field">
        <span class="awd-filter-label">实时状态</span>
        <select
          id="awd-service-filter-status"
          :value="serviceStatusFilter"
          class="awd-filter-control"
          @change="updateServiceStatusFilter(($event.target as HTMLSelectElement).value)"
        >
          <option value="all">全部状态</option>
          <option value="up">在线 (UP)</option>
          <option value="down">下线 (DOWN)</option>
          <option value="compromised">失陷 (EXP)</option>
        </select>
      </label>

      <label class="awd-filter-field">
        <span class="awd-filter-label">巡检源</span>
        <select
          id="awd-service-filter-source"
          :value="serviceCheckSourceFilter"
          class="awd-filter-control"
          @change="emit('updateServiceCheckSourceFilter', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">全部来源</option>
          <option v-for="source in serviceCheckSourceOptions" :key="source" :value="source">
            {{ getCheckSourceLabel(source) || source }}
          </option>
        </select>
      </label>

      <label class="awd-filter-field">
        <span class="awd-filter-label">告警特征</span>
        <select
          id="awd-service-filter-alert"
          :value="serviceAlertReasonFilter"
          class="awd-filter-control"
          @change="emit('updateServiceAlertReasonFilter', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">不限告警</option>
          <option v-for="alert in serviceAlerts" :key="alert.key" :value="alert.key">
            {{ alert.label }}
          </option>
        </select>
      </label>
    </div>

    <div class="awd-matrix-scroll overflow-x-auto">
      <table class="awd-matrix-table">
        <thead>
          <tr>
            <th>队伍节点</th>
            <th>服务靶题</th>
            <th class="text-center">官方判定</th>
            <th>计分权重 (SLA / Def / Atk)</th>
            <th>检查流水</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="service in filteredServices" :key="service.id">
            <td class="awd-cell-team">
              <div class="flex items-center gap-2.5">
                <div class="h-2 w-2 rounded-full bg-primary/40 shadow-[0_0_8px_var(--color-primary)]" />
                <span class="font-bold text-text-primary">{{ service.team_name }}</span>
              </div>
            </td>
            <td class="awd-cell-challenge">
              <span class="font-medium text-text-secondary">
                {{ getChallengeTitle(service.challenge_id) }}
              </span>
            </td>
            <td class="text-center">
              <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg font-black text-[11px] uppercase tracking-wider" :class="getServiceStatusClass(service.service_status)">
                <component 
                  :is="service.service_status === 'up' ? CheckCircle2 : service.service_status === 'compromised' ? ShieldX : AlertCircle" 
                  class="h-3.5 w-3.5" 
                />
                {{ getServiceStatusLabel(service.service_status) }}
              </div>
            </td>
            <td class="awd-cell-scores">
              <div class="awd-score-strip">
                <span class="awd-score-item" title="SLA 分数">
                  <Activity class="h-3 w-3" /> {{ service.sla_score ?? 0 }}
                </span>
                <span class="awd-score-item" title="防守分数">
                  <ShieldCheck class="h-3 w-3" /> {{ service.defense_score }}
                </span>
                <span class="awd-score-item" title="攻击得分">
                  <Zap class="h-3 w-3" /> {{ service.attack_score }}
                </span>
              </div>
              <div class="mt-1.5 flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-widest text-text-muted">
                <span>Total Hit:</span>
                <span class="text-danger">{{ service.attack_received }}</span>
              </div>
            </td>
            <td class="awd-cell-check">
              <div class="awd-check-summary font-medium text-xs">
                {{ summarizeCheckResult(getServiceCheckPresentationResult(service)) }}
              </div>
              
              <div v-if="getServiceCheckActions(service.check_result).length > 0" class="awd-action-chips mt-2">
                <span
                  v-for="action in getServiceCheckActions(service.check_result)"
                  :key="`${service.id}-action-${action.key}`"
                  class="awd-action-chip"
                  :class="{ 'awd-action-chip--error': !action.healthy }"
                >
                  {{ action.label }}
                </span>
              </div>

              <details v-if="getServiceCheckTargets(service.check_result).length > 0" class="awd-mini-report mt-2">
                <summary class="awd-report-trigger">
                  <Info class="h-3.5 w-3.5" />
                  <span>Inspect Probe Details</span>
                </summary>
                <div class="awd-report-body mt-2">
                  <div
                    v-for="(target, targetIndex) in getServiceCheckTargets(service.check_result)"
                    :key="`${service.id}-target-${targetIndex}`"
                    class="awd-report-item"
                  >
                    <div class="flex items-center justify-between gap-4">
                      <code class="text-[10px] text-text-primary">{{ target.access_url || 'Default Entry' }}</code>
                      <span class="text-[10px] font-bold" :class="target.healthy ? 'text-success' : 'text-danger'">
                        {{ target.latency_ms ? `${target.latency_ms}ms` : 'FAIL' }}
                      </span>
                    </div>
                  </div>
                </div>
              </details>
            </td>
          </tr>
          <tr v-if="filteredServices.length === 0">
            <td colspan="5">
              <div class="py-12 flex flex-col items-center gap-3 text-text-muted">
                <SearchCheck class="h-8 w-8 opacity-20" />
                <span class="text-sm font-medium">
                  {{ services.length === 0 ? '当前轮次还没有服务巡检记录' : '没有找到匹配的服务项' }}
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.awd-status-panel {
  background: color-mix(in srgb, var(--journal-surface) 40%, transparent);
  border: 1px solid var(--workspace-line-soft);
  border-radius: 1.5rem;
  overflow: hidden;
  backdrop-filter: blur(12px);
}

.awd-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--workspace-line-soft);
  background: color-mix(in srgb, var(--journal-surface) 90%, var(--color-bg-base));
}

.awd-panel-identity {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.awd-panel-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  color: var(--journal-ink);
}

.awd-filter-bar {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
  padding: var(--space-3) var(--space-5);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  border-bottom: 1px solid var(--workspace-line-soft);
}

.awd-filter-field {
  display: grid;
  gap: var(--space-1.5);
}

.awd-filter-label {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.awd-filter-control {
  width: 100%;
  height: 2.25rem;
  padding: 0 var(--space-3);
  font-size: 13px;
  font-weight: 600;
  border-radius: 0.65rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: var(--journal-surface);
  color: var(--journal-ink);
  outline: none;
  transition: all 0.2s ease;
}

.awd-filter-control:focus {
  border-color: var(--color-primary);
  background: var(--color-bg-surface);
}

.awd-matrix-table {
  width: 100%;
  border-collapse: collapse;
}

.awd-matrix-table th {
  padding: var(--space-3) var(--space-5);
  text-align: left;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  border-bottom: 1px solid var(--workspace-line-soft);
}

.awd-matrix-table td {
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid color-mix(in srgb, var(--workspace-line-soft) 60%, transparent);
  vertical-align: middle;
}

.awd-matrix-table tr:last-child td {
  border-bottom: none;
}

.awd-cell-team {
  min-width: 10rem;
}

.awd-score-strip {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.awd-score-item {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-family-mono);
  font-size: 13px;
  font-weight: 700;
  color: var(--journal-ink);
}

.awd-score-item svg {
  color: var(--journal-muted);
}

.awd-status-pill--up {
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: var(--color-success);
  border: 1px solid color-mix(in srgb, var(--color-success) 24%, transparent);
}

.awd-status-pill--down {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
  border: 1px solid color-mix(in srgb, var(--color-danger) 24%, transparent);
}

.awd-status-pill--compromised {
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
  border: 1px solid color-mix(in srgb, var(--color-warning) 24%, transparent);
}

.awd-action-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.awd-action-chip {
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  border: 1px solid var(--workspace-line-soft);
  font-size: 10px;
  font-weight: 700;
  color: var(--journal-muted);
}

.awd-action-chip--error {
  border-color: color-mix(in srgb, var(--color-danger) 24%, transparent);
  color: var(--color-danger);
}

.awd-report-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-primary);
  cursor: pointer;
  opacity: 0.8;
  transition: opacity 0.2s ease;
}

.awd-report-trigger:hover {
  opacity: 1;
}

.awd-report-body {
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  border-radius: 0.75rem;
  padding: var(--space-2);
  border: 1px solid var(--workspace-line-soft);
}

.awd-report-item {
  padding: var(--space-1.5) var(--space-2);
}

.awd-report-item + .awd-report-item {
  border-top: 1px solid var(--workspace-line-soft);
}

@media (max-width: 1024px) {
  .awd-filter-bar {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .awd-filter-bar {
    grid-template-columns: 1fr;
  }
}
</style>

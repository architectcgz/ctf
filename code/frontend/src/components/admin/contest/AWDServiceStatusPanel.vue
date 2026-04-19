<script setup lang="ts">
import type {
  AWDServiceStatusPanelEmits,
  AWDServiceStatusPanelProps,
} from '@/components/admin/contest/awdInspector.types'

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
  <div class="overflow-hidden rounded-2xl border border-border">
    <div class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3">
      <div class="awd-panel-title text-sm font-semibold">服务状态表</div>
      <button
        id="awd-export-services"
        type="button"
        class="ui-btn ui-btn--secondary awd-service-export-button"
        :disabled="filteredServices.length === 0"
        @click="emit('exportServices')"
      >
        导出当前筛选
      </button>
    </div>
    <div class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-4">
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">队伍</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-service-filter-team"
            :value="serviceTeamFilter"
            class="ui-control"
            @change="emit('updateServiceTeamFilter', ($event.target as HTMLSelectElement).value)"
          >
            <option value="">全部队伍</option>
            <option v-for="team in serviceTeamOptions" :key="team.team_id" :value="team.team_id">
              {{ team.team_name }}
            </option>
          </select>
        </span>
      </label>
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">状态</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-service-filter-status"
            :value="serviceStatusFilter"
            class="ui-control"
            @change="updateServiceStatusFilter(($event.target as HTMLSelectElement).value)"
          >
            <option value="all">全部状态</option>
            <option value="up">正常</option>
            <option value="down">下线</option>
            <option value="compromised">已失陷</option>
          </select>
        </span>
      </label>
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">巡检来源</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-service-filter-source"
            :value="serviceCheckSourceFilter"
            class="ui-control"
            @change="
              emit('updateServiceCheckSourceFilter', ($event.target as HTMLSelectElement).value)
            "
          >
            <option value="">全部来源</option>
            <option v-for="source in serviceCheckSourceOptions" :key="source" :value="source">
              {{ getCheckSourceLabel(source) || source }}
            </option>
          </select>
        </span>
      </label>
      <label class="ui-field awd-round-filter-field">
        <span class="ui-field__label">告警类型</span>
        <span class="ui-control-wrap awd-round-filter-control">
          <select
            id="awd-service-filter-alert"
            :value="serviceAlertReasonFilter"
            class="ui-control"
            @change="
              emit('updateServiceAlertReasonFilter', ($event.target as HTMLSelectElement).value)
            "
          >
            <option value="">全部告警</option>
            <option v-for="alert in serviceAlerts" :key="alert.key" :value="alert.key">
              {{ alert.label }}
            </option>
          </select>
        </span>
      </label>
    </div>
    <table class="min-w-full divide-y divide-border">
      <thead class="awd-table-head">
        <tr>
          <th class="px-4 py-3">队伍</th>
          <th class="px-4 py-3">靶题</th>
          <th class="px-4 py-3">服务状态</th>
          <th class="px-4 py-3">得分</th>
          <th class="px-4 py-3">检查结果</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-border bg-surface/70">
        <tr v-for="service in filteredServices" :key="service.id">
          <td class="awd-table-cell awd-table-cell--primary awd-table-cell--strong">
            {{ service.team_name }}
          </td>
          <td class="awd-table-cell awd-table-cell--secondary">
            {{ getChallengeTitle(service.challenge_id) }}
          </td>
          <td class="awd-table-cell">
            <span
              class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
              :class="getServiceStatusClass(service.service_status)"
            >
              {{ getServiceStatusLabel(service.service_status) }}
            </span>
          </td>
          <td class="awd-table-cell awd-table-cell--secondary">
            <div>
              SLA {{ service.sla_score ?? 0 }} / 防守 {{ service.defense_score }} / 攻击
              {{ service.attack_score }}
            </div>
            <div class="awd-service-muted mt-1 text-xs">
              受攻击 {{ service.attack_received }}
            </div>
          </td>
          <td class="awd-table-cell awd-table-cell--muted">
            <div>
              {{ summarizeCheckResult(getServiceCheckPresentationResult(service)) }}
            </div>
            <div
              v-if="getServiceCheckActions(service.check_result).length > 0"
              class="awd-service-action-list mt-2 flex flex-wrap gap-2 text-xs"
            >
              <span
                v-for="action in getServiceCheckActions(service.check_result)"
                :key="`${service.id}-action-${action.key}`"
                class="rounded-full border border-border/70 bg-surface-alt/40 px-2 py-1"
              >
                {{ action.label }} ·
                {{ getProbeStatusText(action.healthy, action.error_code, action.error) }}
                <span v-if="action.method || action.path">
                  ·
                  {{ [action.method?.toUpperCase(), action.path].filter(Boolean).join(' ') }}
                </span>
              </span>
            </div>
            <div
              v-if="getTargetProbeSummary(service.check_result)"
              class="awd-service-muted mt-2 text-xs"
            >
              {{ getTargetProbeSummary(service.check_result) }}
            </div>
            <details
              v-if="getServiceCheckTargets(service.check_result).length > 0"
              class="awd-service-details mt-2 rounded-xl border border-border/80 bg-surface-alt/40 p-3 text-xs"
            >
              <summary class="awd-service-details-summary cursor-pointer select-none">
                查看检查明细
              </summary>
              <div class="mt-3 space-y-3">
                <div
                  v-for="(target, targetIndex) in getServiceCheckTargets(service.check_result)"
                  :key="`${service.id}-target-${targetIndex}`"
                  class="rounded-xl border border-border/70 bg-surface/70 p-3"
                >
                  <div class="awd-service-details-title font-medium">
                    {{ target.access_url || `Target #${targetIndex + 1}` }}
                  </div>
                  <div class="awd-service-muted mt-1">
                    {{ getProbeStatusText(target.healthy, target.error_code, target.error) }}
                    <span v-if="target.probe"> · {{ target.probe.toUpperCase() }}</span>
                    <span v-if="formatLatency(target.latency_ms)">
                      · {{ formatLatency(target.latency_ms) }}</span
                    >
                  </div>
                  <div
                    v-if="getTargetActions(target).length > 0"
                    class="mt-2 space-y-1 border-t border-border/60 pt-2"
                  >
                    <div
                      v-for="action in getTargetActions(target)"
                      :key="`${service.id}-target-${targetIndex}-action-${action.key}`"
                    >
                      {{ action.label }} ·
                      {{ getProbeStatusText(action.healthy, action.error_code, action.error) }}
                      <span v-if="action.method || action.path">
                        ·
                        {{ [action.method?.toUpperCase(), action.path].filter(Boolean).join(' ') }}
                      </span>
                    </div>
                  </div>
                  <div
                    v-if="target.attempts.length > 0"
                    class="mt-2 space-y-1 border-t border-border/60 pt-2"
                  >
                    <div
                      v-for="(attempt, attemptIndex) in target.attempts"
                      :key="`${service.id}-target-${targetIndex}-attempt-${attemptIndex}`"
                    >
                      Attempt {{ attemptIndex + 1 }}:
                      {{ attempt.probe.toUpperCase() || 'UNKNOWN' }}
                      ·
                      {{ getProbeStatusText(attempt.healthy, attempt.error_code, attempt.error) }}
                      <span v-if="formatLatency(attempt.latency_ms)">
                        · {{ formatLatency(attempt.latency_ms) }}</span
                      >
                    </div>
                  </div>
                </div>
              </div>
            </details>
          </td>
        </tr>
        <tr v-if="filteredServices.length === 0">
          <td colspan="5" class="awd-empty-row">
            {{
              services.length === 0 ? '当前轮次还没有服务巡检记录。' : '当前筛选条件下没有服务记录。'
            }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.awd-panel-title,
.awd-service-details-summary,
.awd-service-details-title {
  color: var(--color-text-primary);
}

.awd-table-head {
  background: color-mix(in srgb, var(--color-surface-alt, var(--color-bg-surface)) 40%, transparent);
  text-align: left;
  font-size: var(--font-size-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  color: var(--color-text-muted);
}

.awd-table-cell {
  padding: var(--space-4);
  font-size: var(--font-size-sm);
}

.awd-table-cell--primary {
  color: var(--color-text-primary);
}

.awd-table-cell--secondary,
.awd-service-action-list,
.awd-service-details {
  color: var(--color-text-secondary);
}

.awd-table-cell--muted,
.awd-service-muted,
.awd-empty-row {
  color: var(--color-text-muted);
}

.awd-table-cell--strong {
  font-weight: 500;
}

.awd-empty-row {
  padding: var(--space-8) var(--space-4);
  text-align: center;
  font-size: var(--font-size-sm);
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
</style>

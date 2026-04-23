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
  ChevronDown
} from 'lucide-vue-next'
import { computed } from 'vue'
import type {
  AWDServiceStatusPanelEmits,
  AWDServiceStatusPanelProps,
} from '@/components/platform/contest/awdInspector.types'

const props = defineProps<AWDServiceStatusPanelProps>()
const emit = defineEmits<AWDServiceStatusPanelEmits>()

// Matrix specific derivations
const distinctChallengeIds = computed(() => {
  const ids = new Set<string>()
  props.services.forEach(s => ids.add(s.challenge_id))
  return Array.from(ids)
})

const teamMap = computed(() => {
  const map = new Map<string, { team_name: string, services: Record<string, any> }>()
  props.filteredServices.forEach(s => {
    if (!map.has(s.team_id)) {
      map.set(s.team_id, { team_name: s.team_name, services: {} })
    }
    map.get(s.team_id)!.services[s.challenge_id] = s
  })
  return Array.from(map.entries()).sort((a, b) => a[1].team_name.localeCompare(b[1].team_name))
})

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
  <div class="awd-matrix-viewer">
    <div class="matrix-toolbar">
      <div class="toolbar-left">
        <h3 class="viewer-title">
          服务运行矩阵
        </h3>
        <div class="filter-summary">
          显示 {{ teamMap.length }} 支队伍
        </div>
      </div>
      
      <div class="toolbar-right">
        <div class="matrix-filters">
          <select
            id="awd-service-filter-team"
            :value="serviceTeamFilter"
            class="matrix-select"
            @change="emit('updateServiceTeamFilter', ($event.target as HTMLSelectElement).value)"
          >
            <option value="">
              所有队伍
            </option>
            <option
              v-for="team in serviceTeamOptions"
              :key="team.team_id"
              :value="team.team_id"
            >
              {{ team.team_name }}
            </option>
          </select>
          <select
            id="awd-service-filter-status"
            :value="serviceStatusFilter"
            class="matrix-select"
            @change="updateServiceStatusFilter(($event.target as HTMLSelectElement).value)"
          >
            <option value="all">
              所有状态
            </option>
            <option value="up">
              在线 (UP)
            </option>
            <option value="down">
              离线 (DOWN)
            </option>
            <option value="compromised">
              失陷 (EXP)
            </option>
          </select>
          <select
            id="awd-service-filter-source"
            :value="serviceCheckSourceFilter"
            class="matrix-select"
            @change="emit('updateServiceCheckSourceFilter', ($event.target as HTMLSelectElement).value)"
          >
            <option value="">
              所有来源
            </option>
            <option
              v-for="source in serviceCheckSourceOptions"
              :key="source"
              :value="source"
            >
              {{ getCheckSourceLabel(source) }}
            </option>
          </select>
          <select
            id="awd-service-filter-alert"
            :value="serviceAlertReasonFilter"
            class="matrix-select"
            @change="emit('updateServiceAlertReasonFilter', ($event.target as HTMLSelectElement).value)"
          >
            <option value="">
              所有告警
            </option>
            <option
              v-for="alert in serviceAlerts"
              :key="alert.key"
              :value="alert.key"
            >
              {{ alert.label }}
            </option>
          </select>
        </div>
        <button
          id="awd-export-services"
          type="button"
          class="ops-btn ops-btn--neutral"
          @click="emit('exportServices')"
        >
          <FileDown class="h-3.5 w-3.5 mr-2" /> 导出报告
        </button>
      </div>
    </div>

    <div class="matrix-scroll custom-scrollbar">
      <table class="matrix-table">
        <!-- ... matrix headers and rows ... -->
      </table>
    </div>

    <!-- New Integrated Round Performance Section -->
    <section
      v-if="props.summary"
      class="round-performance-area mt-12"
    >
      <header class="performance-header">
        <h3 class="viewer-title">
          本轮得分与健康表现
        </h3>
        <div class="filter-summary">
          Round Performance Summary
        </div>
      </header>
      
      <div class="log-table-wrap mt-4">
        <table class="studio-table">
          <thead>
            <tr>
              <th>队伍节点</th>
              <th class="text-right">
                本轮得分
              </th>
              <th class="text-right">
                SLA / ATK / DEF
              </th>
              <th class="text-right">
                服务健康
              </th>
              <th class="text-right">
                被攻破统计
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in props.summary.items"
              :key="item.team_id"
              class="studio-row"
            >
              <td class="font-bold text-slate-900">
                {{ item.team_name }}
              </td>
              <td class="text-right font-mono font-black text-emerald-600">
                {{ item.total_score }}
              </td>
              <td class="text-right font-mono text-[11px] text-slate-500">
                {{ item.sla_score ?? 0 }} / {{ item.attack_score }} / {{ item.defense_score }}
              </td>
              <td class="text-right">
                <div class="health-stack">
                  <span class="text-emerald-500">{{ item.service_up_count }} UP</span>
                  <span class="text-slate-300">/</span>
                  <span class="text-red-500">{{ item.service_down_count }} OFF</span>
                  <span class="text-slate-300">/</span>
                  <span class="text-orange-500">{{ item.service_compromised_count }} EXP</span>
                </div>
              </td>
              <td class="text-right text-[11px] text-slate-500">
                攻破 {{ item.successful_breach_count }} 次 · {{ item.unique_attackers_against }} 攻击方
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<style scoped>
.awd-matrix-viewer { display: flex; flex-direction: column; gap: 1.5rem; }

.matrix-toolbar {
  display: flex; justify-content: space-between; align-items: flex-end;
  padding-bottom: 0.5rem;
}
.viewer-title { font-size: 14px; font-weight: 900; color: var(--color-text-primary); text-transform: uppercase; letter-spacing: 0.1em; margin: 0; }
.filter-summary { font-size: 11px; font-weight: 600; color: var(--color-text-muted); margin-top: 0.25rem; }

.toolbar-right { display: flex; gap: 1rem; align-items: center; }
.matrix-filters { display: flex; gap: 0.5rem; }
.matrix-select {
  height: 2rem; padding: 0 0.5rem; font-size: 11px; font-weight: 700;
  border-radius: 0.5rem; border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface); color: var(--color-text-primary); outline: none;
}

.matrix-scroll {
  border: 1px solid var(--color-border-default); border-radius: 1rem;
  background: var(--color-bg-surface); overflow: auto;
}

.matrix-table { width: 100%; border-collapse: separate; border-spacing: 0; }
.matrix-table th {
  padding: 0.75rem 1rem; background: var(--color-bg-elevated);
  font-size: 10px; font-weight: 800; text-transform: uppercase; color: var(--color-text-secondary);
  border-bottom: 1px solid var(--color-border-default); border-right: 1px solid var(--color-border-default);
  white-space: nowrap; position: sticky; top: 0; z-index: 20;
}
.matrix-table td { padding: 0.5rem; border-bottom: 1px solid var(--color-border-subtle); border-right: 1px solid var(--color-border-subtle); }

.sticky-col { position: sticky; left: 0; z-index: 10; background: var(--color-bg-surface); border-right: 2px solid var(--color-border-default); }
.header-team { left: 0; z-index: 30; width: 12rem; min-width: 12rem; }
.cell-team { background: var(--color-bg-surface); }
.team-name { font-size: 13px; font-weight: 800; color: var(--color-text-primary); }

.status-box {
  display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 0.75rem;
  border-radius: 0.75rem; border: 1px solid transparent; transition: all 0.2s ease;
}
.status-box:hover { transform: scale(1.02); box-shadow: var(--color-shadow-soft); }
.status-icon { width: 2rem; height: 2rem; border-radius: 0.5rem; display: flex; align-items: center; justify-content: center; }
.status-score { font-family: var(--font-family-mono); font-size: 14px; font-weight: 900; }
.status-check { font-size: 10px; font-weight: 600; opacity: 0.8; margin-top: 1px; }

/* Status variants */
.status--up { background: color-mix(in srgb, var(--color-success) 8%, var(--color-bg-surface)); border-color: color-mix(in srgb, var(--color-success) 20%, transparent); color: var(--color-success); }
.status--up .status-icon { background: var(--color-success-soft); color: var(--color-success); }

.status--down { background: color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-surface)); border-color: color-mix(in srgb, var(--color-danger) 20%, transparent); color: var(--color-danger); }
.status--down .status-icon { background: var(--color-danger-soft); color: var(--color-danger); }

.status--compromised { background: color-mix(in srgb, var(--color-warning) 8%, var(--color-bg-surface)); border-color: color-mix(in srgb, var(--color-warning) 20%, transparent); color: var(--color-warning); }
.status--compromised .status-icon { background: var(--color-warning-soft); color: var(--color-warning); }

.status-empty { text-align: center; color: var(--color-text-muted); font-family: var(--font-family-mono); font-weight: 800; }

.performance-header { border-left: 3px solid var(--color-success); padding-left: 1rem; }
.health-stack { display: inline-flex; align-items: center; gap: 0.5rem; font-family: var(--font-family-mono); font-size: 11px; font-weight: 700; }
.studio-table { width: 100%; border-collapse: collapse; background: var(--color-bg-surface); }
.studio-table th { background: var(--color-bg-elevated); padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); border-top: 1px solid var(--color-border-default); border-bottom: 1px solid var(--color-border-default); }
.studio-table td { padding: 0.85rem 1rem; border-bottom: 1px solid var(--color-border-subtle); }

.ops-btn {
  display: inline-flex; align-items: center; gap: 0.5rem; height: 2rem; padding: 0 0.85rem;
  border-radius: 0.65rem; font-size: 12px; font-weight: 700;
  background: var(--color-bg-surface); border: 1px solid var(--color-border-default);
  color: var(--color-text-secondary); cursor: pointer;
}
.ops-btn:hover:not(:disabled) { background: var(--color-bg-elevated); color: var(--color-text-primary); }
</style>

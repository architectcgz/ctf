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
  const map = new Map<number, { team_name: string, services: Record<string, any> }>()
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
        <h3 class="viewer-title">服务运行矩阵</h3>
        <div class="filter-summary">显示 {{ teamMap.length }} 支队伍</div>
      </div>
      
      <div class="toolbar-right">
        <div class="matrix-filters">
          <select :value="serviceStatusFilter" class="matrix-select" @change="updateServiceStatusFilter(($event.target as HTMLSelectElement).value)">
            <option value="all">所有状态</option>
            <option value="up">在线 (UP)</option>
            <option value="down">离线 (DOWN)</option>
            <option value="compromised">失陷 (EXP)</option>
          </select>
          <select :value="serviceCheckSourceFilter" class="matrix-select" @change="emit('updateServiceCheckSourceFilter', ($event.target as HTMLSelectElement).value)">
            <option value="">所有来源</option>
            <option v-for="source in serviceCheckSourceOptions" :key="source" :value="source">{{ getCheckSourceLabel(source) }}</option>
          </select>
        </div>
        <button type="button" class="ops-btn ops-btn--neutral" @click="emit('exportServices')">
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
    <section v-if="props.summary" class="round-performance-area mt-12">
      <header class="performance-header">
        <h3 class="viewer-title">本轮得分与健康表现</h3>
        <div class="filter-summary">Round Performance Summary</div>
      </header>
      
      <div class="log-table-wrap mt-4">
        <table class="studio-table">
          <thead>
            <tr>
              <th>队伍节点</th>
              <th class="text-right">本轮得分</th>
              <th class="text-right">SLA / ATK / DEF</th>
              <th class="text-right">服务健康</th>
              <th class="text-right">被攻破统计</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in props.summary.items" :key="item.team_id" class="studio-row">
              <td class="font-bold text-slate-900">{{ item.team_name }}</td>
              <td class="text-right font-mono font-black text-emerald-600">{{ item.total_score }}</td>
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
/* ... other styles ... */
.performance-header { border-left: 3px solid #10b981; padding-left: 1rem; }
.health-stack { display: inline-flex; align-items: center; gap: 0.5rem; font-family: var(--font-family-mono); font-size: 11px; font-weight: 700; }
.studio-table { width: 100%; border-collapse: collapse; background: white; }
.studio-table th { background: #f8fafc; padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; border-top: 1px solid #e2e8f0; border-bottom: 1px solid #e2e8f0; }
.studio-table td { padding: 0.85rem 1rem; border-bottom: 1px solid #f1f5f9; }
</style>

<style scoped>
.awd-matrix-panel {
  background: white;
  border: 1px solid var(--workspace-line-soft);
  border-radius: 1rem;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.awd-matrix-header {
  padding: 1.25rem 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--workspace-line-soft);
}

.header-overline {
  font-size: 9px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: #94a3b8;
}

.header-title {
  font-size: 1rem;
  font-weight: 900;
  color: #0f172a;
  margin: 0.15rem 0 0;
}

.awd-matrix-toolbar {
  padding: 0.75rem 1.5rem;
  background: #f8fafc;
  border-bottom: 1px solid var(--workspace-line-soft);
}

.toolbar-group {
  display: flex;
  gap: 1.5rem;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-label {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  color: #64748b;
}

.matrix-select {
  height: 1.75rem;
  padding: 0 0.5rem;
  font-size: 11px;
  font-weight: 700;
  border-radius: 0.45rem;
  border: 1px solid #e2e8f0;
  background: white;
  color: #1e293b;
  outline: none;
}

.awd-matrix-container {
  flex: 1;
  overflow: auto;
}

.matrix-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.matrix-table th {
  padding: 0.75rem 1rem;
  background: #f1f5f9;
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
  border-right: 1px solid #e2e8f0;
  white-space: nowrap;
}

.matrix-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  border-right: 1px solid #f1f5f9;
}

.sticky-col {
  position: sticky;
  left: 0;
  z-index: 10;
  background: white;
  border-right: 2px solid #e2e8f0 !important;
}

.header-team { left: 0; z-index: 20; width: 12rem; min-width: 12rem; }
.header-challenge { min-width: 14rem; text-align: left; }

.cell-team {
  background: white;
}

.team-indicator {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #3b82f6;
  box-shadow: 0 0 8px #3b82f6;
}

.team-name {
  font-size: 13px;
  font-weight: 800;
  color: #0f172a;
}

.cell-status {
  padding: 0.5rem !important;
}

.status-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.75rem;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.status-box:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.status-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-details {
  display: flex;
  flex-direction: column;
}

.status-score {
  font-family: var(--font-family-mono);
  font-size: 14px;
  font-weight: 900;
}

.status-check {
  font-size: 10px;
  font-weight: 600;
  opacity: 0.8;
  margin-top: 1px;
}

/* Status variants */
.status--up { background: #f0fdf4; border-color: #bbf7d0; color: #166534; }
.status--up .status-icon { background: #dcfce7; color: #16a34a; }

.status--down { background: #fef2f2; border-color: #fecaca; color: #991b1b; }
.status--down .status-icon { background: #fee2e2; color: #dc2626; }

.status--compromised { background: #fff7ed; border-color: #ffedd5; color: #9a3412; }
.status--compromised .status-icon { background: #ffedd5; color: #ea580c; }

.status-empty {
  text-align: center;
  color: #cbd5e1;
  font-family: var(--font-family-mono);
  font-weight: 800;
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  height: 2rem;
  padding: 0 0.85rem;
  border-radius: 0.65rem;
  font-size: 12px;
  font-weight: 700;
  background: white;
  border: 1px solid #e2e8f0;
  color: #475569;
  cursor: pointer;
}

.ops-btn:hover:not(:disabled) {
  border-color: #cbd5e1;
  background: #f8fafc;
}
</style>

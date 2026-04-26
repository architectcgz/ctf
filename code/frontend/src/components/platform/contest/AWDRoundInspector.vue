<script setup lang="ts">
import { ref, toRef, computed } from 'vue'
import { 
  Activity, 
  BarChart3, 
  Download, 
  History, 
  LayoutGrid, 
  ShieldAlert, 
  Sword, 
  Zap,
  Radar
} from 'lucide-vue-next'
import type { AWDTeamServiceData } from '@/api/contracts'
import AWDAttackLogPanel from '@/components/platform/contest/AWDAttackLogPanel.vue'
import AWDRoundHeaderPanel from '@/components/platform/contest/AWDRoundHeaderPanel.vue'
import AWDScoreboardSummaryPanel from '@/components/platform/contest/AWDScoreboardSummaryPanel.vue'
import AWDServiceStatusPanel from '@/components/platform/contest/AWDServiceStatusPanel.vue'
import AWDTrafficPanel from '@/components/platform/contest/AWDTrafficPanel.vue'
import type {
  AWDRoundInspectorEmits,
  AWDRoundInspectorProps,
} from '@/components/platform/contest/awdInspector.types'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useAwdInspectorCoreState } from '@/composables/useAwdInspectorCoreState'
import { useAwdInspectorDerivedData } from '@/composables/useAwdInspectorDerivedData'
import { useAwdInspectorExports } from '@/composables/useAwdInspectorExports'
import { useAwdInspectorFormatting } from '@/composables/useAwdInspectorFormatting'

const props = defineProps<AWDRoundInspectorProps & { initialTab?: 'matrix' | 'attacks' | 'traffic' | 'scoreboard', hideStudioLink?: boolean }>()

const emit = defineEmits<AWDRoundInspectorEmits & { 'open:contestEdit': [] }>()

const activeSubTab = ref<'matrix' | 'attacks' | 'traffic' | 'scoreboard'>(props.initialTab || 'matrix')

const {
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  selectedRound,
  summaryMetrics,
  totalServiceCount,
  totalAttackCount,
  upCount,
  compromisedCount,
  downCount,
  successfulAttackCount,
  failedAttackCount,
  attackedServiceCount,
  checkButtonLabel,
} = useAwdInspectorCoreState({
  contest: toRef(props, 'contest'),
  selectedRoundId: toRef(props, 'selectedRoundId'),
  rounds: toRef(props, 'rounds'),
  services: toRef(props, 'services'),
  attacks: toRef(props, 'attacks'),
  summary: toRef(props, 'summary'),
  checking: toRef(props, 'checking'),
})

const {
  formatDateTime,
  getRoundStatusLabel,
  getRoundStatusClass,
  getServiceStatusLabel,
  getServiceStatusClass,
  getAttackTypeLabel,
  getAttackSourceLabel,
  formatPercent,
  getTrafficStatusGroupLabel,
  getTrafficStatusGroupClass,
  getChallengeTitle,
  buildExportFilename,
  getSelectedRoundLabel,
  formatScore,
  getSourceOverviewLabel,
  getSourceOverviewDescription,
} = useAwdInspectorFormatting({
  contest: toRef(props, 'contest'),
  challengeLinks: toRef(props, 'challengeLinks'),
  selectedRound,
  summaryMetrics,
  manualCheckCount: toRef(props, 'manualCheckCount' as any),
})

const {
  getCheckSourceLabel,
  getCheckerTypeLabel,
  summarizeCheckResult,
  getCheckActions,
  getCheckTargets,
  getTargetActions,
  getTargetProbeSummary,
  getProbeStatusText,
  formatLatency,
} = useAwdCheckResultPresentation({
  formatDateTime,
})

const {
  getServiceCheckSourceValue,
  serviceTeamOptions,
  serviceCheckSourceOptions,
  serviceAlerts,
  filteredServices,
  attackTeamOptions,
  attackSourceOptions,
  trafficTeamOptions,
  filteredAttacks,
  getServiceAlertClass,
  applyServiceAlertFilter,
} = useAwdInspectorDerivedData({
  services: toRef(props, 'services'),
  attacks: toRef(props, 'attacks'),
  trafficSummary: toRef(props, 'trafficSummary'),
  trafficEvents: toRef(props, 'trafficEvents'),
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  getChallengeTitle,
  getCheckStatusLabel: (s: any) => s,
})

const {
  getTrafficTeamName,
  getTrafficChallengeTitle,
  getTrafficSourceLabel,
  exportFilteredServices,
  exportFilteredAttacks,
  exportReviewPackage,
} = useAwdInspectorExports({
  contest: toRef(props, 'contest'),
  selectedRound,
  summary: toRef(props, 'summary'),
  scoreboardRows: toRef(props, 'scoreboardRows'),
  scoreboardFrozen: toRef(props, 'scoreboardFrozen'),
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  serviceTeamOptions,
  attackTeamOptions,
  trafficTeamOptions,
  serviceAlerts,
  filteredServices,
  filteredAttacks,
  formatDateTime,
  getChallengeTitle,
  getSelectedRoundLabel,
  buildExportFilename,
  getServiceStatusLabel,
  getAttackTypeLabel,
  getAttackSourceLabel,
  getCheckSourceLabel,
  getCheckerTypeLabel,
  getServiceAlertLabel: (s: any) => s,
  summarizeCheckResult,
  getServiceCheckSourceValue,
})

function getServiceCheckPresentationResult(service: AWDTeamServiceData): Record<string, unknown> {
  return {
    checker_type: service.checker_type,
    ...service.check_result,
  }
}
</script>

<template>
  <div class="awd-inspector-workbench">
    <!-- 1. Header (Stateless Header) -->
    <AWDRoundHeaderPanel
      :contest="contest"
      :rounds="rounds"
      :rounds-count="rounds.length"
      :selected-round="selectedRound"
      :selected-round-id="selectedRoundId"
      :loading-rounds="loadingRounds"
      :loading-round-detail="loadingRoundDetail"
      :checking="checking"
      :should-auto-refresh="shouldAutoRefresh"
      :can-record-service-checks="canRecordServiceChecks"
      :can-record-attack-logs="canRecordAttackLogs"
      :service-check-hint="serviceCheckHint"
      :attack-log-hint="attackLogHint"
      :compromised-count="compromisedCount"
      :total-attack-count="totalAttackCount"
      :successful-attack-count="successfulAttackCount"
      :failed-attack-count="failedAttackCount"
      :get-round-status-label="getRoundStatusLabel"
      :get-round-status-class="getRoundStatusClass"
      :check-button-label="checkButtonLabel"
      :hide-studio-link="hideStudioLink"
      @refresh="emit('refresh')"
      @open-create-round-dialog="emit('openCreateRoundDialog')"
      @open-service-check-dialog="emit('openServiceCheckDialog')"
      @open-attack-log-dialog="emit('openAttackLogDialog')"
      @run-selected-round-check="emit('runSelectedRoundCheck')"
      @update:selected-round-id="emit('update:selectedRoundId', $event)"
      @open:contest-edit="emit('open:contestEdit')"
    />

    <!-- 2. Integrated Metric HUD (Modern Dashboard Style) -->
    <div
      v-if="selectedRound"
      class="awd-stats-hud"
    >
      <div class="stat-card">
        <div class="stat-card__icon stat-card__icon--blue">
          <Activity class="h-4 w-4" />
        </div>
        <div class="stat-card__content">
          <div class="unit-label">Infrastructure</div>
          <div class="unit-value font-mono">
            {{ totalServiceCount }} <small>SRV</small>
          </div>
          <div class="unit-helper">
            ONLINE: {{ upCount }} · OFF: {{ downCount }}
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-card__icon stat-card__icon--green">
          <Radar class="h-4 w-4" />
        </div>
        <div class="stat-card__content">
          <div class="unit-label">Battle Traffic</div>
          <div class="unit-value font-mono">
            {{ totalAttackCount }} <small>HITS</small>
          </div>
          <div class="unit-helper unit-helper--success">
            SUCCESS: {{ successfulAttackCount }}
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-card__icon stat-card__icon--orange">
          <ShieldAlert class="h-4 w-4" />
        </div>
        <div class="stat-card__content">
          <div class="unit-label">Compromised</div>
          <div class="unit-value unit-value--warning font-mono">
            {{ compromisedCount }} <small>EXP</small>
          </div>
          <div class="unit-helper">
            AFFECTED: {{ attackedServiceCount }} TEAMS
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-card__icon stat-card__icon--purple">
          <LayoutGrid class="h-4 w-4" />
        </div>
        <div class="stat-card__content">
          <div class="unit-label">Composition</div>
          <div class="unit-value">
            {{ getSourceOverviewLabel() }}
          </div>
          <div class="unit-helper">
            {{ getSourceOverviewDescription() }}
          </div>
        </div>
      </div>
    </div>

    <!-- 3. Main Workspace (Flat Design) -->
    <div class="awd-detail-canvas">
      <header class="canvas-tabs-header">
        <nav class="sub-tabs">
          <button
            class="sub-tab"
            :class="{ active: activeSubTab === 'matrix' }"
            @click="activeSubTab = 'matrix'"
          >
            <LayoutGrid class="h-3.5 w-3.5" /> 运行矩阵
          </button>
          <button
            class="sub-tab"
            :class="{ active: activeSubTab === 'scoreboard' }"
            @click="activeSubTab = 'scoreboard'"
          >
            <BarChart3 class="h-3.5 w-3.5" /> 排行榜单
          </button>
          <button
            class="sub-tab"
            :class="{ active: activeSubTab === 'attacks' }"
            @click="activeSubTab = 'attacks'"
          >
            <Sword class="h-3.5 w-3.5" /> 攻击流水
          </button>
          <button
            class="sub-tab"
            :class="{ active: activeSubTab === 'traffic' }"
            @click="activeSubTab = 'traffic'"
          >
            <Zap class="h-3.5 w-3.5" /> 流量分析
          </button>
        </nav>

        <div class="canvas-actions">
          <button
            type="button"
            class="ops-btn ops-btn--neutral"
            @click="exportReviewPackage"
          >
            <Download class="h-3.5 w-3.5 mr-2" /> 导出复盘包
          </button>
        </div>
      </header>

      <div class="canvas-content custom-scrollbar">
        <div
          v-if="loadingRoundDetail"
          class="canvas-loading-overlay"
        >
          <AppLoading>同步态势中...</AppLoading>
        </div>

        <AppEmpty
          v-else-if="!selectedRound"
          title="等待激活"
          description="在上方选择轮次以展开战场监控。"
          icon="History"
          class="py-24"
        />

        <div
          v-else
          class="pane-wrap"
        >
          <div
            v-show="activeSubTab === 'matrix'"
            class="pane-matrix"
          >
            <div
              v-if="serviceAlerts.length > 0"
              class="alert-banner mb-8"
            >
              <span class="banner-tag">重点异常告警</span>
              <div class="alert-pills">
                <button
                  v-for="alert in serviceAlerts"
                  :key="alert.key"
                  class="alert-pill"
                  :class="[getServiceAlertClass(alert.key), { 'is-active': serviceAlertReasonFilter === alert.key }]"
                  @click="applyServiceAlertFilter(alert.key)"
                >
                  {{ alert.label }} ({{ alert.count }})
                </button>
              </div>
            </div>
            <AWDServiceStatusPanel
              :services="services"
              :filtered-services="filteredServices"
              :summary="summary"
              :service-alerts="serviceAlerts"
              :service-team-options="serviceTeamOptions"
              :service-check-source-options="serviceCheckSourceOptions"
              :service-team-filter="serviceTeamFilter"
              :service-status-filter="serviceStatusFilter"
              :service-check-source-filter="serviceCheckSourceFilter"
              :service-alert-reason-filter="serviceAlertReasonFilter"
              :get-challenge-title="getChallengeTitle"
              :get-service-status-label="getServiceStatusLabel"
              :get-service-status-class="getServiceStatusClass"
              :get-check-source-label="getCheckSourceLabel"
              :summarize-check-result="summarizeCheckResult"
              :get-check-actions="getCheckActions"
              :get-check-targets="getCheckTargets"
              :get-target-actions="getTargetActions"
              :get-target-probe-summary="getTargetProbeSummary"
              :get-probe-status-text="getProbeStatusText"
              :format-latency="formatLatency"
              :get-service-check-presentation-result="getServiceCheckPresentationResult"
              @update-service-team-filter="serviceTeamFilter = $event"
              @update-service-status-filter="serviceStatusFilter = $event"
              @update-service-check-source-filter="serviceCheckSourceFilter = $event"
              @update-service-alert-reason-filter="serviceAlertReasonFilter = $event"
              @export-services="exportFilteredServices"
            />
          </div>

          <div
            v-show="activeSubTab === 'scoreboard'"
            class="pane-scoreboard"
          >
            <AWDScoreboardSummaryPanel
              :scoreboard-rows="scoreboardRows"
              :scoreboard-frozen="scoreboardFrozen"
              :summary="summary"
              :format-score="formatScore"
              :format-date-time="formatDateTime"
            />
          </div>

          <div
            v-show="activeSubTab === 'attacks'"
            class="pane-attacks"
          >
            <AWDAttackLogPanel
              :attacks="attacks"
              :filtered-attacks="filteredAttacks"
              :attack-team-options="attackTeamOptions"
              :attack-source-options="attackSourceOptions"
              :attack-team-filter="attackTeamFilter"
              :attack-result-filter="attackResultFilter"
              :attack-source-filter="attackSourceFilter"
              :format-date-time="formatDateTime"
              :get-challenge-title="getChallengeTitle"
              :get-attack-type-label="getAttackTypeLabel"
              :get-attack-source-label="getAttackSourceLabel"
              @update-attack-team-filter="attackTeamFilter = $event"
              @update-attack-result-filter="attackResultFilter = $event"
              @update-attack-source-filter="attackSourceFilter = $event"
              @export-attacks="exportFilteredAttacks"
            />
          </div>

          <div
            v-show="activeSubTab === 'traffic'"
            class="pane-traffic"
          >
            <AWDTrafficPanel
              :updated-at="selectedRound.updated_at"
              :challenge-links="challengeLinks"
              :traffic-summary="trafficSummary"
              :traffic-events="trafficEvents"
              :traffic-events-total="trafficEventsTotal"
              :traffic-filters="trafficFilters"
              :traffic-team-options="trafficTeamOptions"
              :loading-traffic-summary="loadingTrafficSummary"
              :loading-traffic-events="loadingTrafficEvents"
              :format-date-time="formatDateTime"
              :format-percent="formatPercent"
              :get-traffic-status-group-label="getTrafficStatusGroupLabel"
              :get-traffic-status-group-class="getTrafficStatusGroupClass"
              :get-traffic-team-name="getTrafficTeamName"
              :get-traffic-challenge-title="getTrafficChallengeTitle"
              :get-traffic-source-label="getTrafficSourceLabel"
              @apply-traffic-filters="emit('applyTrafficFilters', $event)"
              @change-traffic-page="emit('changeTrafficPage', $event)"
              @reset-traffic-filters="emit('resetTrafficFilters')"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.awd-inspector-workbench { display: flex; flex-direction: column; height: 100%; }

.awd-stats-hud {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
  padding: 1.5rem 0;
  background: transparent;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  padding: 1.25rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 1rem;
  transition: all 0.2s ease;
}

.stat-card:hover {
  border-color: var(--color-border-default);
  box-shadow: 0 4px 12px color-mix(in srgb, var(--color-text-primary) 4%, transparent);
}

.stat-card__icon {
  width: 2.75rem;
  height: 2.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
}

.stat-card__icon--blue { background: color-mix(in srgb, var(--color-primary) 10%, transparent); color: var(--color-primary); }
.stat-card__icon--green { background: color-mix(in srgb, var(--color-success) 10%, transparent); color: var(--color-success); }
.stat-card__icon--orange { background: color-mix(in srgb, var(--color-warning) 10%, transparent); color: var(--color-warning); }
.stat-card__icon--purple { background: color-mix(in srgb, var(--color-secondary) 10%, transparent); color: var(--color-secondary); }

.stat-card__content { flex: 1; display: flex; flex-direction: column; gap: 0.25rem; }

.unit-label { font-size: 10px; font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); letter-spacing: 0.05em; }
.unit-value { font-size: 1.25rem; font-weight: 900; color: var(--color-text-primary); line-height: 1; }
.unit-value small { font-size: 11px; opacity: 0.5; margin-left: 2px; }
.unit-helper { font-size: 11px; font-weight: 600; color: var(--color-text-secondary); }
.unit-helper--success { color: var(--color-success); }
.unit-value--warning { color: var(--color-warning); }

.awd-detail-canvas { flex: 1; display: flex; flex-direction: column; background: transparent; min-height: 0; }
.canvas-tabs-header {
  height: 3.5rem; padding: 0; border-bottom: 1px solid var(--color-border-default);
  display: flex; justify-content: space-between; align-items: center; background: transparent;
}
.sub-tabs { display: flex; gap: 2rem; height: 100%; }
.sub-tab {
  display: flex; align-items: center; gap: 0.5rem; padding: 0 0.25rem;
  font-size: 13px; font-weight: 800; color: var(--color-text-secondary);
  border-bottom: 2px solid transparent; cursor: pointer; transition: all 0.2s ease;
}
.sub-tab:hover { color: var(--color-text-primary); }
.sub-tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }

.canvas-content { flex: 1; overflow-y: auto; padding: 2rem 0; position: relative; }
.canvas-loading-overlay { position: absolute; inset: 0; background: color-mix(in srgb, var(--color-bg-base) 70%, transparent); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 10; }

.alert-banner { display: flex; align-items: center; gap: 1.5rem; padding: 0.75rem 1.25rem; background: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface)); border: 1px solid color-mix(in srgb, var(--color-warning) 20%, transparent); border-radius: 0.75rem; }
.banner-tag { font-size: 10px; font-weight: 800; color: var(--color-warning); text-transform: uppercase; }
.alert-pills { display: flex; gap: 0.5rem; }
.alert-pill { padding: 0.25rem 0.75rem; border-radius: 6px; font-size: 11px; font-weight: 700; cursor: pointer; border: 1px solid var(--awd-service-alert-border, color-mix(in srgb, var(--color-warning) 30%, transparent)); background: var(--awd-service-alert-bg, transparent); color: var(--awd-service-alert-color, var(--color-warning)); transition: all 0.2s ease; }
.alert-pill:hover { background: var(--color-bg-elevated); }
.alert-pill.is-active { background: var(--awd-service-alert-active-bg, var(--color-warning)); color: var(--color-text-inverse); border-color: var(--awd-service-alert-active-bg, var(--color-warning)); }

.awd-service-alert--danger {
  --awd-service-alert-bg: color-mix(in srgb, var(--color-danger) 10%, var(--color-bg-surface));
  --awd-service-alert-border: color-mix(in srgb, var(--color-danger) 20%, transparent);
  --awd-service-alert-color: var(--color-danger);
  --awd-service-alert-active-bg: var(--color-danger);
}

.awd-service-alert--warning {
  --awd-service-alert-bg: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface));
  --awd-service-alert-border: color-mix(in srgb, var(--color-warning) 20%, transparent);
  --awd-service-alert-color: var(--color-warning);
  --awd-service-alert-active-bg: var(--color-warning);
}

.awd-service-alert--neutral {
  --awd-service-alert-bg: color-mix(in srgb, var(--color-text-muted) 10%, var(--color-bg-surface));
  --awd-service-alert-border: color-mix(in srgb, var(--color-text-muted) 20%, transparent);
  --awd-service-alert-color: var(--color-text-primary);
  --awd-service-alert-active-bg: var(--color-text-secondary);
}

.ops-btn {
  display: inline-flex; align-items: center; justify-content: center;
  height: 2.25rem; padding: 0 1rem; border-radius: 0.75rem;
  font-size: 12px; font-weight: 700; background: var(--color-bg-surface); border: 1px solid var(--color-border-default); color: var(--color-text-secondary); cursor: pointer;
}
</style>

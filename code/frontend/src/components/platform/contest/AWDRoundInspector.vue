<script setup lang="ts">
import { toRef } from 'vue'
import type { AWDTeamServiceData } from '@/api/contracts'
import AWDAttackLogPanel from '@/components/platform/contest/AWDAttackLogPanel.vue'
import AWDRoundHeaderPanel from '@/components/platform/contest/AWDRoundHeaderPanel.vue'
import AWDScoreboardSummaryPanel from '@/components/platform/contest/AWDScoreboardSummaryPanel.vue'
import AWDRoundSelectionPanel from '@/components/platform/contest/AWDRoundSelectionPanel.vue'
import AWDServiceStatusPanel from '@/components/platform/contest/AWDServiceStatusPanel.vue'
import AWDTrafficPanel from '@/components/platform/contest/AWDTrafficPanel.vue'
import type {
  AWDRoundInspectorEmits,
  AWDRoundInspectorProps,
} from '@/components/platform/contest/awdInspector.types'
import SectionCard from '@/components/common/SectionCard.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useAwdInspectorCoreState } from '@/composables/useAwdInspectorCoreState'
import { useAwdInspectorDerivedData } from '@/composables/useAwdInspectorDerivedData'
import { useAwdInspectorExports } from '@/composables/useAwdInspectorExports'
import { useAwdInspectorFormatting } from '@/composables/useAwdInspectorFormatting'
const props = defineProps<AWDRoundInspectorProps>()
const emit = defineEmits<AWDRoundInspectorEmits>()
const {
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  resetFilters,
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
  defenseSuccessCount,
  manualCheckCount,
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
  manualCheckCount,
})

const {
  getCheckSourceLabel,
  getCheckerTypeLabel,
  getCheckStatusLabel,
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
  getServiceAlertReason,
  getServiceAlertSubtitle,
  getServiceAlertClass,
  getServiceAlertLabel,
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
  getCheckStatusLabel,
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
  getServiceAlertLabel,
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
  <div class="space-y-6">
    <AWDRoundHeaderPanel
      :contest="contest"
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
      @refresh="emit('refresh')"
      @open-create-round-dialog="emit('openCreateRoundDialog')"
      @open-service-check-dialog="emit('openServiceCheckDialog')"
      @open-attack-log-dialog="emit('openAttackLogDialog')"
      @run-selected-round-check="emit('runSelectedRoundCheck')"
    />

    <section class="awd-round-workspace-grid grid gap-6">
      <AWDRoundSelectionPanel
        :rounds="rounds"
        :selected-round-id="selectedRoundId"
        :selected-round="selectedRound"
        :loading-rounds="loadingRounds"
        :compromised-count="compromisedCount"
        :down-count="downCount"
        :total-service-count="totalServiceCount"
        :format-date-time="formatDateTime"
        :get-round-status-label="getRoundStatusLabel"
        @update:selected-round-id="emit('update:selectedRoundId', $event)"
      />

      <SectionCard title="回合态势" subtitle="排行榜、服务明细与攻击流水共用当前选中轮次。">
        <div v-if="loadingRoundDetail" class="flex justify-center py-12">
          <AppLoading>正在同步当前轮次数据...</AppLoading>
        </div>

        <div v-else-if="!selectedRound" class="py-4">
          <AppEmpty
            title="请选择一个 AWD 轮次"
            description="轮次选定后即可查看服务健康、攻击记录和本轮汇总。"
            icon="FileChartColumnIncreasing"
          />
        </div>

        <div v-else class="space-y-6">
          <div class="flex justify-end">
            <button
              id="awd-export-review-package"
              type="button"
              class="ui-btn ui-btn--secondary awd-round-toolbar__button"
              :disabled="!selectedRound"
              @click="exportReviewPackage"
            >
              导出当前轮复盘包
            </button>
          </div>

          <div class="grid gap-3 md:grid-cols-2 2xl:grid-cols-4">
            <AppCard
              variant="metric"
              accent="primary"
              eyebrow="服务总览"
              :title="String(totalServiceCount)"
              :subtitle="`正常 ${upCount} / 下线 ${downCount} / 失陷 ${compromisedCount}`"
            />
            <AppCard
              variant="metric"
              accent="success"
              eyebrow="攻击流量"
              :title="String(totalAttackCount)"
              :subtitle="`成功 ${successfulAttackCount} / 失败 ${failedAttackCount}`"
            />
            <AppCard
              variant="metric"
              accent="warning"
              eyebrow="防守成功"
              :title="String(defenseSuccessCount)"
              :subtitle="`受攻击服务 ${attackedServiceCount}`"
            />
            <AppCard
              variant="metric"
              accent="neutral"
              eyebrow="来源构成"
              :title="getSourceOverviewLabel()"
              :subtitle="getSourceOverviewDescription()"
            />
          </div>

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

          <div v-if="serviceAlerts.length > 0" class="grid gap-3 md:grid-cols-2 2xl:grid-cols-3">
            <button
              v-for="alert in serviceAlerts"
              :key="alert.key"
              type="button"
              class="text-left"
              @click="applyServiceAlertFilter(alert.key)"
            >
              <AppCard
                variant="action"
                accent="warning"
                eyebrow="重点告警"
                :title="alert.label"
                :subtitle="getServiceAlertSubtitle(alert)"
              >
                <template #default>
                  <div
                    class="rounded-2xl border px-3 py-3 text-sm transition"
                    :class="[
                      getServiceAlertClass(alert.key),
                      serviceAlertReasonFilter === alert.key ? 'ring-2 ring-primary/60' : '',
                    ]"
                  >
                    <div class="font-medium">受影响服务 {{ alert.count }} 个</div>
                    <div class="mt-2 space-y-1 text-xs">
                      <div
                        v-for="sample in alert.samples"
                        :key="`${alert.key}-${sample.service_id || sample.team_name}-${sample.challenge_title}`"
                      >
                        {{ sample.team_name }} · {{ sample.challenge_title }}
                      </div>
                    </div>
                    <div class="awd-alert-hint mt-2 text-xs">
                      {{
                        serviceAlertReasonFilter === alert.key
                          ? '再次点击可取消筛选'
                          : '点击筛选同类异常'
                      }}
                    </div>
                  </div>
                </template>
              </AppCard>
            </button>
          </div>

          <AWDScoreboardSummaryPanel
            :scoreboard-rows="scoreboardRows"
            :scoreboard-frozen="scoreboardFrozen"
            :summary="summary"
            :format-score="formatScore"
            :format-date-time="formatDateTime"
          />

          <div class="grid gap-6 xl:grid-cols-2">
            <AWDServiceStatusPanel
              :services="services"
              :filtered-services="filteredServices"
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
        </div>
      </SectionCard>
    </section>
  </div>
</template>

<style scoped>
.awd-round-toolbar__button {
  white-space: nowrap;
}

.awd-alert-hint {
  color: color-mix(in srgb, var(--color-text-secondary) 80%, transparent);
}

@media (min-width: 1280px) {
  .awd-round-workspace-grid {
    grid-template-columns: 0.85fr 1.15fr;
  }
}
</style>

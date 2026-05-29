<script setup lang="ts">
import { toRef } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { AWDRoundInspector } from '@/features/awd-inspector'
import { AWDReadinessSummary } from '@/features/awd-readiness'
import { usePlatformContestAwd } from '@/features/contest-awd-admin'

import AWDContestSelectorField from './AWDContestSelectorField.vue'
import AWDOperationsDialogHub from './AWDOperationsDialogHub.vue'
import AWDInstanceOrchestrationPanel from './AWDInstanceOrchestrationPanel.vue'
import AWDOperationsPreRuntimeStage from './AWDOperationsPreRuntimeStage.vue'
import AWDOperationsRuntimeStage from './AWDOperationsRuntimeStage.vue'
import AWDRuntimePendingState from './AWDRuntimePendingState.vue'
import './awdOperationsPanel.css'
import { useAwdOperationsDialogAvailability } from './useAwdOperationsDialogAvailability'
import { useAwdOperationsDialogState } from './useAwdOperationsDialogState'
import { useAwdOperationsPanelViewState } from './useAwdOperationsPanelViewState'

const props = defineProps<{
  contests: ContestDetailData[]
  selectedContestId: string | null
  hideContestSelector?: boolean
  hideStudioLink?: boolean
  hideReadinessActions?: boolean
  hideOperationTabs?: boolean
  operationPanel?: 'inspector' | 'instances'
  runtimeContent?: 'all' | 'readiness' | 'round-inspector' | 'instances'
  initialTab?: 'matrix' | 'attacks' | 'traffic' | 'scoreboard'
}>()

const emit = defineEmits<{
  'update:selectedContestId': [contestId: string]
  'open:awd-config': [challengeId: string]
  'open:contest-edit': []
}>()

const {
  selectedContest,
  shouldShowContestSelector,
  runtimeStageReady,
  activePanel,
  visibleOperationTabs,
  shouldShowOperationTabs,
  runtimeContent,
  shouldShowRuntimeReadiness,
  shouldShowRoundInspector,
  shouldShowInstanceOrchestration,
  selectPanel,
  registerTabButton,
  handlePanelTabKeydown,
} = useAwdOperationsPanelViewState({
  contests: toRef(props, 'contests'),
  selectedContestId: toRef(props, 'selectedContestId'),
  operationPanel: toRef(props, 'operationPanel'),
  hideContestSelector: toRef(props, 'hideContestSelector'),
  hideOperationTabs: toRef(props, 'hideOperationTabs'),
  runtimeContent: toRef(props, 'runtimeContent'),
})

const {
  rounds,
  selectedRoundId,
  services,
  attacks,
  summary,
  trafficSummary,
  trafficEvents,
  trafficEventsTotal,
  trafficFilters,
  scoreboardRows,
  scoreboardFrozen,
  teams,
  challengeLinks,
  instanceOrchestration,
  readiness,
  loadingRounds,
  loadingRoundDetail,
  loadingTrafficSummary,
  loadingTrafficEvents,
  loadingInstanceOrchestration,
  loadingReadiness,
  checking,
  creatingRound,
  savingServiceCheck,
  savingAttackLog,
  startingInstanceKey,
  shouldAutoRefresh,
  overrideDialogState,
  refresh,
  refreshInstanceOrchestration,
  applyTrafficFilters,
  setTrafficPage,
  resetTrafficFilters,
  runSelectedRoundCheck,
  startTeamServiceInstance,
  startTeamAllServices,
  startAllTeamServices,
  confirmOverrideAction,
  closeOverrideDialog,
  createRound,
  createServiceCheck,
  createAttackLog,
} = usePlatformContestAwd(selectedContest)

function updateSelectedRoundId(value: string) {
  selectedRoundId.value = value
}

const {
  roundDialogOpen,
  serviceCheckDialogOpen,
  attackLogDialogOpen,
  nextRoundNumber,
  openRoundDialog,
  updateRoundDialogOpen,
  openServiceCheckDialog,
  updateServiceCheckDialogOpen,
  openAttackLogDialog,
  updateAttackLogDialogOpen,
  handleCreateRound,
  handleCreateServiceCheck,
  handleCreateAttackLog,
  handleOverrideDialogOpenChange,
} = useAwdOperationsDialogState({
  runtimeStageReady,
  rounds,
  createRound,
  createServiceCheck,
  createAttackLog,
  closeOverrideDialog,
})

const {
  canRecordServiceChecks,
  canRecordAttackLogs,
  serviceCheckHint,
  attackLogHint,
} = useAwdOperationsDialogAvailability({
  teams,
  challengeLinks,
})
</script>

<template>
  <div class="studio-ops-shell">
    <AWDContestSelectorField
      v-if="shouldShowContestSelector"
      :contests="contests"
      :selected-contest-id="selectedContestId"
      @update:selected-contest-id="emit('update:selectedContestId', $event)"
    />

    <AppEmpty
      v-if="contests.length === 0"
      title="暂无 AWD 赛事"
      description="当前页没有 AWD 赛事，可先创建或切换到包含 AWD 赛事的页码。"
      icon="Flag"
      class="py-20"
    />

    <AppEmpty
      v-else-if="!selectedContest"
      title="暂无 AWD 赛事"
      description="请先选择一个 AWD 赛事以进入运维面板。"
      icon="Flag"
      class="py-20"
    />

    <div
      v-else
      class="studio-ops-content"
    >
      <AWDOperationsPreRuntimeStage
        v-if="!runtimeStageReady"
        :show-tabs="shouldShowOperationTabs"
        :tabs="visibleOperationTabs"
        :active-panel="activePanel"
        :register-tab-button="registerTabButton"
        :contest-title="selectedContest.title"
        :hide-studio-link="Boolean(hideStudioLink)"
        :should-show-runtime-readiness="shouldShowRuntimeReadiness"
        :runtime-content="runtimeContent"
        :should-show-instance-orchestration="shouldShowInstanceOrchestration"
        @select-panel="selectPanel"
        @tab-keydown="handlePanelTabKeydown"
        @open-contest-edit="emit('open:contest-edit')"
      >
        <template #pending>
          <AWDRuntimePendingState />
        </template>
        <template #readiness>
          <AWDReadinessSummary
            :readiness="readiness"
            :loading="loadingReadiness"
            :hide-actions="hideReadinessActions"
            @edit-config="emit('open:awd-config', $event)"
          />
        </template>
        <template #instances>
          <AWDInstanceOrchestrationPanel
            id="awd-ops-panel-instances"
            role="tabpanel"
            aria-labelledby="awd-ops-tab-instances"
            :orchestration="instanceOrchestration"
            :loading="loadingInstanceOrchestration"
            :starting-key="startingInstanceKey"
            @refresh="refreshInstanceOrchestration"
            @start-cell="startTeamServiceInstance"
            @start-team="startTeamAllServices"
            @start-all="startAllTeamServices"
          />
        </template>
      </AWDOperationsPreRuntimeStage>

      <AWDOperationsRuntimeStage
        v-else
        :show-tabs="shouldShowOperationTabs"
        :tabs="visibleOperationTabs"
        :active-panel="activePanel"
        :register-tab-button="registerTabButton"
        :should-show-runtime-readiness="shouldShowRuntimeReadiness"
        :should-show-round-inspector="shouldShowRoundInspector"
        :should-show-instance-orchestration="shouldShowInstanceOrchestration"
        @select-panel="selectPanel"
        @tab-keydown="handlePanelTabKeydown"
      >
        <template #readiness>
          <AWDReadinessSummary
            :readiness="readiness"
            :loading="loadingReadiness"
            :hide-actions="hideReadinessActions"
            class="runtime-readiness-strip"
            @edit-config="emit('open:awd-config', $event)"
          />
        </template>
        <template #inspector>
          <AWDRoundInspector
            id="awd-ops-panel-inspector"
            role="tabpanel"
            aria-labelledby="awd-ops-tab-inspector"
            :contest="selectedContest"
            :rounds="rounds"
            :selected-round-id="selectedRoundId"
            :services="services"
            :attacks="attacks"
            :challenge-links="challengeLinks"
            :summary="summary"
            :traffic-summary="trafficSummary"
            :traffic-events="trafficEvents"
            :traffic-events-total="trafficEventsTotal"
            :traffic-filters="trafficFilters"
            :scoreboard-rows="scoreboardRows"
            :scoreboard-frozen="scoreboardFrozen"
            :loading-rounds="loadingRounds"
            :loading-round-detail="loadingRoundDetail"
            :loading-traffic-summary="loadingTrafficSummary"
            :loading-traffic-events="loadingTrafficEvents"
            :checking="checking"
            :should-auto-refresh="shouldAutoRefresh"
            :can-record-service-checks="canRecordServiceChecks"
            :can-record-attack-logs="canRecordAttackLogs"
            :service-check-hint="serviceCheckHint"
            :attack-log-hint="attackLogHint"
            :initial-tab="initialTab"
            :hide-studio-link="hideStudioLink"
            @refresh="refresh"
            @apply-traffic-filters="applyTrafficFilters"
            @change-traffic-page="setTrafficPage"
            @reset-traffic-filters="resetTrafficFilters"
            @open-create-round-dialog="openRoundDialog"
            @open-service-check-dialog="openServiceCheckDialog"
            @open-attack-log-dialog="openAttackLogDialog"
            @run-selected-round-check="runSelectedRoundCheck"
            @update:selected-round-id="updateSelectedRoundId"
            @open:contest-edit="emit('open:contest-edit')"
          >
            <template #service-alerts="slotProps">
              <slot
                name="service-alerts"
                v-bind="slotProps"
              />
            </template>
          </AWDRoundInspector>
        </template>
        <template #instances>
          <AWDInstanceOrchestrationPanel
            id="awd-ops-panel-instances"
            role="tabpanel"
            aria-labelledby="awd-ops-tab-instances"
            :orchestration="instanceOrchestration"
            :loading="loadingInstanceOrchestration"
            :starting-key="startingInstanceKey"
            @refresh="refreshInstanceOrchestration"
            @start-cell="startTeamServiceInstance"
            @start-team="startTeamAllServices"
            @start-all="startAllTeamServices"
          />
        </template>
      </AWDOperationsRuntimeStage>
    </div>

    <AWDOperationsDialogHub
      :round-dialog-open="roundDialogOpen"
      :next-round-number="nextRoundNumber"
      :creating-round="creatingRound"
      :service-check-dialog-open="serviceCheckDialogOpen"
      :teams="teams"
      :challenge-links="challengeLinks"
      :saving-service-check="savingServiceCheck"
      :attack-log-dialog-open="attackLogDialogOpen"
      :saving-attack-log="savingAttackLog"
      :override-dialog-state="overrideDialogState"
      @update:round-dialog-open="updateRoundDialogOpen"
      @save-round="handleCreateRound"
      @update:service-check-dialog-open="updateServiceCheckDialogOpen"
      @save-service-check="handleCreateServiceCheck"
      @update:attack-log-dialog-open="updateAttackLogDialogOpen"
      @save-attack-log="handleCreateAttackLog"
      @update:override-dialog-open="handleOverrideDialogOpenChange"
      @confirm-override="confirmOverrideAction"
    />
  </div>
</template>

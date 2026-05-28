<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { AWDTrafficStatusGroup, ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { AWDRoundInspector } from '@/features/awd-inspector'
import { AWDReadinessSummary } from '@/features/awd-readiness'
import { usePlatformContestAwd } from '@/features/contest-awd-admin'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'

import AWDContestSelectorField from './AWDContestSelectorField.vue'
import AWDOperationsDialogHub from './AWDOperationsDialogHub.vue'
import AWDInstanceOrchestrationPanel from './AWDInstanceOrchestrationPanel.vue'
import AWDOperationsPreRuntimeStage from './AWDOperationsPreRuntimeStage.vue'
import AWDOperationsRuntimeStage from './AWDOperationsRuntimeStage.vue'
import type { AWDOperationsPanelKey, AWDOperationsTabItem } from './awdOperations.types'
import AWDRuntimePendingState from './AWDRuntimePendingState.vue'

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

const selectedContest = computed(
  () => props.contests.find((item) => item.id === props.selectedContestId) || null
)
const shouldShowContestSelector = computed(() => !props.hideContestSelector)
const runtimeStageReady = computed(
  () =>
    selectedContest.value?.status === 'running' ||
    selectedContest.value?.status === 'frozen' ||
    selectedContest.value?.status === 'ended'
)
const roundDialogOpen = ref(false)
const serviceCheckDialogOpen = ref(false)
const attackLogDialogOpen = ref(false)

const operationTabs: readonly AWDOperationsTabItem[] = [
  {
    key: 'inspector',
    label: '轮次态势',
    tabId: 'awd-ops-tab-inspector',
    panelId: 'awd-ops-panel-inspector',
  },
  {
    key: 'instances',
    label: '实例编排',
    tabId: 'awd-ops-tab-instances',
    panelId: 'awd-ops-panel-instances',
  },
] as const

const operationTabOrder = operationTabs.map((tab) => tab.key) as AWDOperationsPanelKey[]
const activePanel = ref<AWDOperationsPanelKey>(props.operationPanel ?? 'inspector')
const visibleOperationTabs = computed(() =>
  runtimeStageReady.value ? operationTabs : operationTabs.filter((tab) => tab.key === 'inspector')
)
const shouldShowOperationTabs = computed(() => !props.hideOperationTabs)
const runtimeContent = computed(() => props.runtimeContent ?? 'all')
const shouldShowRuntimeReadiness = computed(
  () => runtimeContent.value === 'all' || runtimeContent.value === 'readiness'
)
const shouldShowRoundInspector = computed(
  () =>
    activePanel.value === 'inspector' &&
    (runtimeContent.value === 'all' || runtimeContent.value === 'round-inspector')
)
const shouldShowInstanceOrchestration = computed(
  () =>
    activePanel.value === 'instances' &&
    (runtimeContent.value === 'all' || runtimeContent.value === 'instances')
)

watch(
  () => props.operationPanel,
  (panel) => {
    if (panel) {
      activePanel.value = panel
    }
  }
)

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

const nextRoundNumber = computed(() =>
  rounds.value.length === 0 ? 1 : Math.max(...rounds.value.map((item) => item.round_number)) + 1
)
const canRecordServiceChecks = computed(
  () => teams.value.length > 0 && challengeLinks.value.length > 0
)
const canRecordAttackLogs = computed(
  () => teams.value.length >= 2 && challengeLinks.value.length > 0
)
const serviceCheckHint = computed(() => {
  if (teams.value.length === 0 && challengeLinks.value.length === 0) {
    return '当前赛事还没有队伍和题目，无法录入服务检查。'
  }
  if (teams.value.length === 0) {
    return '当前赛事还没有队伍，无法录入服务检查。'
  }
  if (challengeLinks.value.length === 0) {
    return '当前赛事还没有关联题目，无法录入服务检查。'
  }
  return ''
})
const attackLogHint = computed(() => {
  if (teams.value.length < 2 && challengeLinks.value.length === 0) {
    return '至少需要 2 支队伍且已关联题目后，才能补录攻击日志。'
  }
  if (teams.value.length < 2) {
    return '至少需要 2 支队伍后，才能补录攻击日志。'
  }
  if (challengeLinks.value.length === 0) {
    return '当前赛事还没有关联题目，无法补录攻击日志。'
  }
  return ''
})

function updateSelectedContestId(value: string) {
  emit('update:selectedContestId', value)
}

function updateSelectedRoundId(value: string) {
  selectedRoundId.value = value
}

function openRoundDialog() {
  if (!runtimeStageReady.value) {
    return
  }
  roundDialogOpen.value = true
}

function updateRoundDialogOpen(value: boolean) {
  roundDialogOpen.value = value
}

function openServiceCheckDialog() {
  if (!runtimeStageReady.value) {
    return
  }
  serviceCheckDialogOpen.value = true
}

function updateServiceCheckDialogOpen(value: boolean) {
  serviceCheckDialogOpen.value = value
}

function openAttackLogDialog() {
  if (!runtimeStageReady.value) {
    return
  }
  attackLogDialogOpen.value = true
}

function updateAttackLogDialogOpen(value: boolean) {
  attackLogDialogOpen.value = value
}

function selectPanel(panel: AWDOperationsPanelKey) {
  if (props.operationPanel) {
    return
  }
  activePanel.value = panel
}

function registerTabButton(key: AWDOperationsPanelKey, element: HTMLButtonElement | null) {
  setTabButtonRef(key, element)
}

function handlePanelTabKeydown(event: KeyboardEvent, index: number) {
  handleTabKeydown(event, index)
}

const { setTabButtonRef, handleTabKeydown } = useTabKeyboardNavigation<AWDOperationsPanelKey>({
  orderedTabs: operationTabOrder,
  selectTab: selectPanel,
})

async function handleCreateRound(payload: {
  round_number: number
  status: 'pending' | 'running' | 'finished'
  attack_score: number
  defense_score: number
}) {
  await createRound(payload)
  roundDialogOpen.value = false
}

async function handleCreateServiceCheck(payload: {
  team_id: number
  service_id: number
  service_status: 'up' | 'down' | 'compromised'
  check_result?: Record<string, unknown>
}) {
  await createServiceCheck(payload)
  serviceCheckDialogOpen.value = false
}

async function handleCreateAttackLog(payload: {
  attacker_team_id: number
  victim_team_id: number
  service_id: number
  attack_type: 'flag_capture' | 'service_exploit'
  submitted_flag?: string
  is_success: boolean
}) {
  await createAttackLog(payload)
  attackLogDialogOpen.value = false
}

async function handleApplyTrafficFilters(payload: {
  attacker_team_id?: string
  victim_team_id?: string
  service_id?: string
  awd_challenge_id?: string
  status_group?: 'all' | AWDTrafficStatusGroup
  path_keyword?: string
}) {
  await applyTrafficFilters(payload)
}

async function handleTrafficPageChange(page: number) {
  await setTrafficPage(page)
}

async function handleResetTrafficFilters() {
  await resetTrafficFilters()
}

function handleEditReadinessConfig(challengeId: string) {
  emit('open:awd-config', challengeId)
}

async function handleStartTeamServiceInstance(teamId: string, serviceId: string) {
  await startTeamServiceInstance(teamId, serviceId)
}

async function handleStartTeamAllServices(teamId: string) {
  await startTeamAllServices(teamId)
}

async function handleStartAllTeamServices() {
  await startAllTeamServices()
}

function handleOverrideDialogOpenChange(value: boolean) {
  if (!value) {
    closeOverrideDialog()
  }
}
</script>

<template>
  <div class="studio-ops-shell">
    <AWDContestSelectorField
      v-if="shouldShowContestSelector"
      :contests="contests"
      :selected-contest-id="selectedContestId"
      @update:selected-contest-id="updateSelectedContestId"
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
            @edit-config="handleEditReadinessConfig"
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
            @start-cell="handleStartTeamServiceInstance"
            @start-team="handleStartTeamAllServices"
            @start-all="handleStartAllTeamServices"
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
            @edit-config="handleEditReadinessConfig"
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
            @apply-traffic-filters="handleApplyTrafficFilters"
            @change-traffic-page="handleTrafficPageChange"
            @reset-traffic-filters="handleResetTrafficFilters"
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
            @start-cell="handleStartTeamServiceInstance"
            @start-team="handleStartTeamAllServices"
            @start-all="handleStartAllTeamServices"
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

<style scoped>
.studio-ops-shell {
  min-height: 100%;
  background: transparent;
}

.studio-ops-content {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>

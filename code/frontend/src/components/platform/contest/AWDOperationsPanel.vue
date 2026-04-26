<script setup lang="ts">
import { computed, ref } from 'vue'

import type { AWDTrafficStatusGroup, ContestDetailData } from '@/api/contracts'
import AWDAttackLogDialog from './AWDAttackLogDialog.vue'
import AWDContestSelectorField from './AWDContestSelectorField.vue'
import AWDReadinessOverrideDialog from './AWDReadinessOverrideDialog.vue'
import AWDReadinessSummary from './AWDReadinessSummary.vue'
import AWDRuntimePendingState from './AWDRuntimePendingState.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { usePlatformContestAwd } from '@/composables/usePlatformContestAwd'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'

import AWDRoundCreateDialog from './AWDRoundCreateDialog.vue'
import AWDRoundInspector from './AWDRoundInspector.vue'
import AWDServiceCheckDialog from './AWDServiceCheckDialog.vue'

const props = defineProps<{
  contests: ContestDetailData[]
  selectedContestId: string | null
  hideContestSelector?: boolean
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

const operationTabs = [
  {
    key: 'inspector',
    label: '轮次态势',
    tabId: 'awd-ops-tab-inspector',
    panelId: 'awd-ops-panel-inspector',
  },
] as const

type AWDOperationsPanelKey = (typeof operationTabs)[number]['key']
const operationTabOrder = operationTabs.map((tab) => tab.key) as AWDOperationsPanelKey[]
const activePanel = ref<AWDOperationsPanelKey>('inspector')

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
  readiness,
  loadingRounds,
  loadingRoundDetail,
  loadingTrafficSummary,
  loadingTrafficEvents,
  loadingReadiness,
  checking,
  creatingRound,
  savingServiceCheck,
  savingAttackLog,
  shouldAutoRefresh,
  overrideDialogState,
  refresh,
  applyTrafficFilters,
  setTrafficPage,
  resetTrafficFilters,
  runSelectedRoundCheck,
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
  activePanel.value = panel
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
  challenge_id?: string
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
      <!-- 1. Pre-runtime Readiness (shown if not running) -->
      <section
        v-if="!runtimeStageReady"
        class="studio-ops-section"
      >
        <header class="section-header">
          <div class="section-identity">
            <div class="section-overline">Command Center / Pre-flight</div>
            <h2 class="section-title">
              {{ selectedContest.title }}
            </h2>
          </div>
          <div class="section-actions">
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="emit('open:contest-edit')"
            >
              进入竞赛工作室
            </button>
          </div>
        </header>
        <div class="readiness-wrap">
          <AWDRuntimePendingState />
          <AWDReadinessSummary
            :readiness="readiness"
            :loading="loadingReadiness"
            @edit-config="handleEditReadinessConfig"
          />
        </div>
      </section>

      <!-- 2. Runtime Workspace -->
      <section
        v-else
        class="studio-ops-section"
      >
        <!-- Dashboard Navigation (Integrated) -->
        <nav
          v-if="operationTabs.length > 1"
          class="studio-ops-tabs"
        >
          <button
            v-for="tab in operationTabs"
            :key="tab.key"
            class="tab-item"
            :class="{ active: activePanel === tab.key }"
            @click="selectPanel(tab.key)"
          >
            {{ tab.label }}
          </button>
        </nav>

        <div class="inspector-wrap">
          <AWDRoundInspector
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
          />
        </div>
      </section>
    </div>

    <div
      v-if="overrideDialogState.open"
      class="sr-only"
      aria-live="assertive"
    >
      {{ overrideDialogState.title }} 强制继续
    </div>

    <AWDRoundCreateDialog
      :open="roundDialogOpen"
      :next-round-number="nextRoundNumber"
      :saving="creatingRound"
      @update:open="updateRoundDialogOpen"
      @save="handleCreateRound"
    />

    <AWDServiceCheckDialog
      :open="serviceCheckDialogOpen"
      :teams="teams"
      :challenge-links="challengeLinks"
      :saving="savingServiceCheck"
      @update:open="updateServiceCheckDialogOpen"
      @save="handleCreateServiceCheck"
    />

    <AWDAttackLogDialog
      :open="attackLogDialogOpen"
      :teams="teams"
      :challenge-links="challengeLinks"
      :saving="savingAttackLog"
      @update:open="updateAttackLogDialogOpen"
      @save="handleCreateAttackLog"
    />

    <AWDReadinessOverrideDialog
      :open="overrideDialogState.open"
      :title="overrideDialogState.title"
      :readiness="overrideDialogState.readiness"
      :confirm-loading="overrideDialogState.confirmLoading"
      @update:open="handleOverrideDialogOpenChange"
      @confirm="confirmOverrideAction"
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

.studio-ops-section {
  display: flex;
  flex-direction: column;
}

.section-header {
  margin-bottom: var(--space-8);
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-4);
}

.section-overline {
  font-size: var(--font-size-10);
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-1-5);
}

.section-title { font-size: var(--font-size-1-25); font-weight: 900; color: var(--color-text-primary); margin: 0; letter-spacing: -0.01em; }

.section-actions {
  display: flex;
  gap: var(--space-3);
}

.studio-ops-tabs { display: flex; gap: var(--space-8); border-bottom: 1px solid var(--color-border-default); margin-bottom: var(--space-6); }
.tab-item { padding: var(--space-3) var(--space-1); font-size: var(--font-size-13); font-weight: 800; color: var(--color-text-secondary); border-bottom: 2px solid transparent; cursor: pointer; transition: all 0.2s ease; }
.tab-item:hover { color: var(--color-text-primary); }
.tab-item.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }

.runtime-readiness-strip { margin-bottom: var(--space-6); }

.inspector-wrap { min-width: 0; }
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { AdminContestChallengeData, AWDTrafficStatusGroup, ContestDetailData } from '@/api/contracts'
import AWDAttackLogDialog from './AWDAttackLogDialog.vue'
import AWDChallengeConfigDialog from './AWDChallengeConfigDialog.vue'
import AWDChallengeConfigPanel from './AWDChallengeConfigPanel.vue'
import AWDReadinessOverrideDialog from './AWDReadinessOverrideDialog.vue'
import AWDReadinessSummary from './AWDReadinessSummary.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAdminContestAWD } from '@/composables/useAdminContestAWD'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'

import AWDRoundCreateDialog from './AWDRoundCreateDialog.vue'
import AWDRoundInspector from './AWDRoundInspector.vue'
import AWDServiceCheckDialog from './AWDServiceCheckDialog.vue'

const props = defineProps<{
  contests: ContestDetailData[]
  selectedContestId: string | null
  hideContestSelector?: boolean
}>()

const emit = defineEmits<{
  'update:selectedContestId': [contestId: string]
}>()

const selectedContest = computed(
  () => props.contests.find((item) => item.id === props.selectedContestId) || null
)
const shouldShowContestSelector = computed(() => !props.hideContestSelector)
const runtimeStageReady = computed(
  () => selectedContest.value?.status === 'running' || selectedContest.value?.status === 'frozen'
)
const roundDialogOpen = ref(false)
const serviceCheckDialogOpen = ref(false)
const attackLogDialogOpen = ref(false)
const challengeConfigDialogOpen = ref(false)
const challengeConfigMode = ref<'create' | 'edit'>('create')
const editingChallengeLink = ref<AdminContestChallengeData | null>(null)

const operationTabs = [
  {
    key: 'inspector',
    label: '轮次态势',
    tabId: 'awd-ops-tab-inspector',
    panelId: 'awd-ops-panel-inspector',
  },
  {
    key: 'challenges',
    label: '题目配置',
    tabId: 'awd-ops-tab-challenges',
    panelId: 'awd-ops-panel-challenges',
  },
] as const

type AWDOperationsPanelKey = (typeof operationTabs)[number]['key']
const operationTabOrder = operationTabs.map((tab) => tab.key) as AWDOperationsPanelKey[]

function getOperationsPanelStorageKey(contestId: string): string {
  return `ctf_admin_awd_ops_panel:${contestId}`
}

function loadStoredOperationsPanel(contestId: string): AWDOperationsPanelKey {
  if (typeof window === 'undefined') {
    return 'inspector'
  }
  const value = window.sessionStorage.getItem(getOperationsPanelStorageKey(contestId))
  return value === 'challenges' ? 'challenges' : 'inspector'
}

function persistOperationsPanel(contestId: string, value: AWDOperationsPanelKey): void {
  if (typeof window === 'undefined') {
    return
  }
  window.sessionStorage.setItem(getOperationsPanelStorageKey(contestId), value)
}

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
  challengeCatalog,
  readiness,
  loadingRounds,
  loadingRoundDetail,
  loadingTrafficSummary,
  loadingTrafficEvents,
  loadingChallengeCatalog,
  loadingReadiness,
  checking,
  creatingRound,
  savingServiceCheck,
  savingAttackLog,
  savingChallengeConfig,
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
  loadChallengeCatalog,
  createChallengeLink,
  updateChallengeLink,
} = useAdminContestAWD(selectedContest)

const nextRoundNumber = computed(() =>
  rounds.value.length === 0
    ? 1
    : Math.max(...rounds.value.map((item) => item.round_number)) + 1
)
const canRecordServiceChecks = computed(() => teams.value.length > 0 && challengeLinks.value.length > 0)
const canRecordAttackLogs = computed(() => teams.value.length >= 2 && challengeLinks.value.length > 0)
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
const existingChallengeIds = computed(() => challengeLinks.value.map((item) => item.challenge_id))

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

function openChallengeCreateDialog() {
  challengeConfigMode.value = 'create'
  editingChallengeLink.value = null
  challengeConfigDialogOpen.value = true
  void loadChallengeCatalog()
}

function openChallengeEditDialog(challenge: AdminContestChallengeData) {
  challengeConfigMode.value = 'edit'
  editingChallengeLink.value = challenge
  challengeConfigDialogOpen.value = true
  activePanel.value = 'challenges'
}

function updateChallengeConfigDialogOpen(value: boolean) {
  challengeConfigDialogOpen.value = value
  if (!value) {
    editingChallengeLink.value = null
  }
}

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
  challenge_id: number
  service_status: 'up' | 'down' | 'compromised'
  check_result?: Record<string, unknown>
}) {
  await createServiceCheck(payload)
  serviceCheckDialogOpen.value = false
}

async function handleCreateAttackLog(payload: {
  attacker_team_id: number
  victim_team_id: number
  challenge_id: number
  attack_type: 'flag_capture' | 'service_exploit'
  submitted_flag?: string
  is_success: boolean
}) {
  await createAttackLog(payload)
  attackLogDialogOpen.value = false
}

async function handleSaveChallengeConfig(payload: {
  challenge_id: number
  points: number
  order: number
  is_visible: boolean
  awd_checker_type: 'legacy_probe' | 'http_standard'
  awd_checker_config: Record<string, unknown>
  awd_sla_score: number
  awd_defense_score: number
  awd_checker_preview_token?: string
}) {
  if (challengeConfigMode.value === 'create') {
    await createChallengeLink(payload)
  } else if (editingChallengeLink.value) {
    const { challenge_id: _challengeId, ...updatePayload } = payload
    await updateChallengeLink(editingChallengeLink.value.challenge_id, updatePayload)
  }
  challengeConfigDialogOpen.value = false
  editingChallengeLink.value = null
}

async function handleApplyTrafficFilters(payload: {
  attacker_team_id?: string
  victim_team_id?: string
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
  const matchedChallenge = challengeLinks.value.find((item) => item.challenge_id === challengeId)
  activePanel.value = 'challenges'
  if (matchedChallenge) {
    openChallengeEditDialog(matchedChallenge)
    return
  }
  void loadChallengeCatalog()
}

function handleOverrideDialogOpenChange(value: boolean) {
  if (!value) {
    closeOverrideDialog()
  }
}

watch(
  () => selectedContest.value?.id || null,
  (contestId) => {
    activePanel.value = contestId ? loadStoredOperationsPanel(contestId) : 'inspector'
  },
  { immediate: true }
)

watch(
  () => [selectedContest.value?.id || null, activePanel.value] as const,
  ([contestId, panel]) => {
    if (!contestId) {
      return
    }
    persistOperationsPanel(contestId, panel)
    if (panel === 'challenges' && challengeCatalog.value.length === 0) {
      void loadChallengeCatalog()
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="space-y-6">
    <label v-if="shouldShowContestSelector" class="space-y-2">
      <span class="text-sm text-[var(--color-text-secondary)]">选择 AWD 赛事</span>
      <select
        id="awd-contest-selector"
        :value="selectedContestId || ''"
        class="w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
        :disabled="contests.length === 0"
        @change="updateSelectedContestId(($event.target as HTMLSelectElement).value)"
      >
        <option v-if="contests.length === 0" value="" disabled>暂无 AWD 赛事</option>
        <option v-for="contest in contests" :key="contest.id" :value="contest.id">
          {{ contest.title }}
        </option>
      </select>
    </label>

    <AppEmpty
      v-if="contests.length === 0"
      title="当前页没有 AWD 赛事"
      description="先切到包含 AWD 比赛的筛选结果或分页，这里才会展示可操作的攻防运营面板。"
      icon="Flag"
    />

    <AppEmpty
      v-else-if="!selectedContest"
      title="暂无 AWD 赛事"
      description="当前没有可用的 AWD 赛事选择，请先返回赛事列表确认筛选结果。"
      icon="Flag"
    />

    <section v-else class="space-y-6">
      <AWDReadinessSummary
        :readiness="readiness"
        :loading="loadingReadiness"
        @edit-config="handleEditReadinessConfig"
      />

      <nav class="top-tabs awd-ops-tabs" role="tablist" aria-label="AWD 运维工作区切换">
        <button
          v-for="(tab, index) in operationTabs"
          :id="tab.tabId"
          :key="tab.key"
          :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
          type="button"
          role="tab"
          class="top-tab"
          :class="{ active: activePanel === tab.key }"
          :aria-selected="activePanel === tab.key ? 'true' : 'false'"
          :aria-controls="tab.panelId"
          :tabindex="activePanel === tab.key ? 0 : -1"
          @click="selectPanel(tab.key)"
          @keydown="handleTabKeydown($event, index)"
        >
          {{ tab.label }}
        </button>
      </nav>

      <section
        id="awd-ops-panel-inspector"
        class="awd-ops-tab-panel"
        role="tabpanel"
        aria-labelledby="awd-ops-tab-inspector"
        :aria-hidden="activePanel === 'inspector' ? 'false' : 'true'"
        v-show="activePanel === 'inspector'"
      >
        <AWDRoundInspector
          v-if="runtimeStageReady"
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
          @refresh="refresh"
          @apply-traffic-filters="handleApplyTrafficFilters"
          @change-traffic-page="handleTrafficPageChange"
          @reset-traffic-filters="handleResetTrafficFilters"
          @open-create-round-dialog="openRoundDialog"
          @open-service-check-dialog="openServiceCheckDialog"
          @open-attack-log-dialog="openAttackLogDialog"
          @run-selected-round-check="runSelectedRoundCheck"
          @update:selected-round-id="updateSelectedRoundId"
        />

        <section
          v-else
          class="awd-runtime-shell rounded-[28px] border p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
        >
          <div
            class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-primary-hover)]/75"
          >
            <span>Operations</span>
            <span class="awd-runtime-shell-chip rounded-full px-2 py-1">待开赛</span>
          </div>
          <div class="mt-3 grid gap-4">
            <div>
              <h2 class="text-3xl font-semibold tracking-tight text-white">尚未进入运行阶段</h2>
              <p class="mt-3 text-sm leading-7 text-[var(--color-text-secondary)]/90">
                当前赛事还不能进入轮次运行。需先通过赛前检查并开赛，随后才会接管创建轮次、服务巡检、攻击补录和当前轮态势。
              </p>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <button
                id="awd-runtime-shell-create-round"
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition disabled:cursor-not-allowed disabled:opacity-60"
                disabled
              >
                创建轮次
              </button>
              <button
                id="awd-runtime-shell-record-service"
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition disabled:cursor-not-allowed disabled:opacity-60"
                disabled
              >
                录入服务检查
              </button>
              <button
                id="awd-runtime-shell-record-attack"
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition disabled:cursor-not-allowed disabled:opacity-60"
                disabled
              >
                补录攻击日志
              </button>
              <button
                id="awd-runtime-shell-run-check"
                type="button"
                class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition disabled:cursor-not-allowed disabled:opacity-60"
                disabled
              >
                立即巡检当前轮
              </button>
            </div>

            <p class="text-xs text-[var(--color-primary-hover)]/75">需先通过赛前检查并开赛</p>
          </div>
        </section>
      </section>

      <section
        id="awd-ops-panel-challenges"
        class="awd-ops-tab-panel"
        role="tabpanel"
        aria-labelledby="awd-ops-tab-challenges"
        :aria-hidden="activePanel === 'challenges' ? 'false' : 'true'"
        v-show="activePanel === 'challenges'"
      >
        <AWDChallengeConfigPanel
          :challenge-links="challengeLinks"
          @create="openChallengeCreateDialog"
          @edit="openChallengeEditDialog"
        />
      </section>
    </section>

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

    <AWDChallengeConfigDialog
      :contest-id="selectedContest?.id || null"
      :open="challengeConfigDialogOpen"
      :mode="challengeConfigMode"
      :challenge-options="challengeCatalog"
      :existing-challenge-ids="existingChallengeIds"
      :draft="editingChallengeLink"
      :loading-challenge-catalog="loadingChallengeCatalog"
      :saving="savingChallengeConfig"
      @update:open="updateChallengeConfigDialogOpen"
      @save="handleSaveChallengeConfig"
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
.awd-ops-tabs {
  margin-top: 0.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
}

.awd-ops-tab-panel {
  min-width: 0;
}

.awd-runtime-shell {
  background:
    linear-gradient(145deg, color-mix(in srgb, var(--color-surface-panel) 94%, white 6%), var(--color-surface-panel)),
    radial-gradient(circle at top right, color-mix(in srgb, var(--color-primary) 18%, transparent), transparent 52%);
}

.awd-runtime-shell-chip {
  background: color-mix(in srgb, var(--color-warning) 18%, transparent);
  color: color-mix(in srgb, var(--color-warning) 74%, white 26%);
}
</style>

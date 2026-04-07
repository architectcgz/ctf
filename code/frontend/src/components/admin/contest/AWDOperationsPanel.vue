<script setup lang="ts">
import { computed, ref } from 'vue'

import type { AWDTrafficStatusGroup, ContestDetailData } from '@/api/contracts'
import AWDAttackLogDialog from './AWDAttackLogDialog.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAdminContestAWD } from '@/composables/useAdminContestAWD'

import AWDRoundCreateDialog from './AWDRoundCreateDialog.vue'
import AWDRoundInspector from './AWDRoundInspector.vue'
import AWDServiceCheckDialog from './AWDServiceCheckDialog.vue'

const props = defineProps<{
  contests: ContestDetailData[]
  selectedContestId: string | null
}>()

const emit = defineEmits<{
  'update:selectedContestId': [contestId: string]
}>()

const selectedContest = computed(
  () => props.contests.find((item) => item.id === props.selectedContestId) || null
)
const roundDialogOpen = ref(false)
const serviceCheckDialogOpen = ref(false)
const attackLogDialogOpen = ref(false)

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
  loadingRounds,
  loadingRoundDetail,
  loadingTrafficSummary,
  loadingTrafficEvents,
  checking,
  creatingRound,
  savingServiceCheck,
  savingAttackLog,
  shouldAutoRefresh,
  refresh,
  applyTrafficFilters,
  setTrafficPage,
  resetTrafficFilters,
  runSelectedRoundCheck,
  createRound,
  createServiceCheck,
  createAttackLog,
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

function updateSelectedContestId(value: string) {
  emit('update:selectedContestId', value)
}

function updateSelectedRoundId(value: string) {
  selectedRoundId.value = value
}

function openRoundDialog() {
  roundDialogOpen.value = true
}

function updateRoundDialogOpen(value: boolean) {
  roundDialogOpen.value = value
}

function openServiceCheckDialog() {
  serviceCheckDialogOpen.value = true
}

function updateServiceCheckDialogOpen(value: boolean) {
  serviceCheckDialogOpen.value = value
}

function openAttackLogDialog() {
  attackLogDialogOpen.value = true
}

function updateAttackLogDialogOpen(value: boolean) {
  attackLogDialogOpen.value = value
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
</script>

<template>
  <div class="space-y-6">
    <label class="space-y-2">
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

    <AWDRoundInspector
      v-else
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
  </div>
</template>

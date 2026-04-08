<script setup lang="ts">
import { computed, toRef, watch } from 'vue'
import { RefreshCw, Search, ShieldAlert, ShieldCheck, Sword, TimerReset } from 'lucide-vue-next'

import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'
import { useAwdInspectorDerivedData } from '@/composables/useAwdInspectorDerivedData'
import { useAwdInspectorFilters } from '@/composables/useAwdInspectorFilters'
import { useAwdInspectorFormatting } from '@/composables/useAwdInspectorFormatting'
import { useAwdTrafficPanel, type AWDTrafficFilters } from '@/composables/useAwdTrafficPanel'
import { downloadCSVFile, downloadJSONFile } from '@/utils/csv'

const props = defineProps<{
  contest: ContestDetailData
  rounds: AWDRoundData[]
  selectedRoundId: string | null
  services: AWDTeamServiceData[]
  attacks: AWDAttackLogData[]
  challengeLinks: AdminContestChallengeData[]
  summary: AWDRoundSummaryData | null
  trafficSummary: AWDTrafficSummaryData | null
  trafficEvents: AWDTrafficEventData[]
  trafficEventsTotal: number
  trafficFilters: AWDTrafficFilters
  scoreboardRows: ScoreboardRow[]
  scoreboardFrozen: boolean
  loadingRounds: boolean
  loadingRoundDetail: boolean
  loadingTrafficSummary: boolean
  loadingTrafficEvents: boolean
  checking: boolean
  shouldAutoRefresh: boolean
  canRecordServiceChecks: boolean
  canRecordAttackLogs: boolean
  serviceCheckHint?: string
  attackLogHint?: string
}>()

const emit = defineEmits<{
  refresh: []
  openCreateRoundDialog: []
  openServiceCheckDialog: []
  openAttackLogDialog: []
  runSelectedRoundCheck: []
  applyTrafficFilters: [filters: Partial<AWDTrafficFilters>]
  changeTrafficPage: [page: number]
  resetTrafficFilters: []
  'update:selectedRoundId': [roundId: string]
}>()

const selectedRound = computed(
  () => props.rounds.find((item) => item.id === props.selectedRoundId) || null
)
const summaryMetrics = computed(() => props.summary?.metrics || null)
const totalServiceCount = computed(
  () => summaryMetrics.value?.total_service_count ?? props.services.length
)
const totalAttackCount = computed(
  () => summaryMetrics.value?.total_attack_count ?? props.attacks.length
)
const upCount = computed(
  () =>
    summaryMetrics.value?.service_up_count ??
    props.services.filter((item) => item.service_status === 'up').length
)
const compromisedCount = computed(
  () =>
    summaryMetrics.value?.service_compromised_count ??
    props.services.filter((item) => item.service_status === 'compromised').length
)
const downCount = computed(
  () =>
    summaryMetrics.value?.service_down_count ??
    props.services.filter((item) => item.service_status === 'down').length
)
const successfulAttackCount = computed(
  () =>
    summaryMetrics.value?.successful_attack_count ??
    props.attacks.filter((item) => item.is_success).length
)
const failedAttackCount = computed(
  () =>
    summaryMetrics.value?.failed_attack_count ??
    props.attacks.filter((item) => !item.is_success).length
)
const attackedServiceCount = computed(
  () =>
    summaryMetrics.value?.attacked_service_count ??
    props.services.filter((item) => item.attack_received > 0).length
)
const defenseSuccessCount = computed(
  () =>
    summaryMetrics.value?.defense_success_count ??
    props.services.filter((item) => item.attack_received > 0 && item.service_status === 'up').length
)
const manualCheckCount = computed(() => {
  if (summaryMetrics.value) {
    return (
      summaryMetrics.value.manual_current_round_check_count +
      summaryMetrics.value.manual_selected_round_check_count +
      summaryMetrics.value.manual_service_check_count
    )
  }
  return props.services.filter((item) => {
    const source = getServiceCheckSourceValue(item.check_result)
    return (
      source === 'manual_current_round' ||
      source === 'manual_selected_round' ||
      source === 'manual_service_check'
    )
  }).length
})
const contestId = computed(() => props.contest.id)
const selectedRoundIdRef = computed(() => props.selectedRoundId)
const {
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  resetFilters,
} = useAwdInspectorFilters({
  contestId,
  selectedRoundId: selectedRoundIdRef,
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
  trafficPathKeywordInput,
  trafficTotalPages,
  trafficTrendRows,
  trafficSummaryStats,
  trafficTrendNarrative,
  trafficStatusGroupOptions,
  applyTrafficKeywordFilter,
  onTrafficStatusGroupChange,
  clearTrafficKeywordFilter,
  applyTrafficFilterPatch,
  handleTrafficPageChange,
} = useAwdTrafficPanel({
  trafficSummary: toRef(props, 'trafficSummary'),
  trafficEventsTotal: toRef(props, 'trafficEventsTotal'),
  trafficFilters: toRef(props, 'trafficFilters'),
  loadingTrafficEvents: toRef(props, 'loadingTrafficEvents'),
  formatDateTime,
  formatPercent,
  applyTrafficFilters: (patch) => emit('applyTrafficFilters', patch),
  changeTrafficPage: (page) => emit('changeTrafficPage', page),
})

const {
  getCheckSourceLabel,
  getCheckStatusLabel,
  summarizeCheckResult,
  getCheckTargets,
  getTargetProbeSummary,
  getProbeStatusText,
  formatLatency,
} = useAwdCheckResultPresentation({
  formatDateTime,
})

function getTrafficTeamName(teamId: string, teamName?: string): string {
  if (teamName && teamName.trim() !== '') {
    return teamName
  }
  return trafficTeamOptions.value.find((item) => item.id === teamId)?.name || `Team #${teamId}`
}

function getTrafficChallengeTitle(challengeId: string, fallbackTitle?: string): string {
  if (fallbackTitle && fallbackTitle.trim() !== '') {
    return fallbackTitle
  }
  return getChallengeTitle(challengeId)
}

function getTrafficSourceLabel(source: string): string {
  if (source === 'proxy_audit' || source === 'runtime_proxy') {
    return '平台代理'
  }
  return source || '未标记'
}

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

function exportFilteredServices() {
  if (filteredServices.value.length === 0) {
    return
  }
  downloadCSVFile(
    buildExportFilename('services'),
    filteredServices.value.map((item) => ({
      赛事: props.contest.title,
      轮次: getSelectedRoundLabel(),
      筛选队伍: serviceTeamFilter.value
        ? serviceTeamOptions.value.find((team) => team.team_id === serviceTeamFilter.value)
            ?.team_name || serviceTeamFilter.value
        : '全部队伍',
      筛选状态:
        serviceStatusFilter.value === 'all'
          ? '全部状态'
          : getServiceStatusLabel(serviceStatusFilter.value),
      筛选来源: serviceCheckSourceFilter.value
        ? getCheckSourceLabel(serviceCheckSourceFilter.value) || serviceCheckSourceFilter.value
        : '全部来源',
      筛选告警: serviceAlertReasonFilter.value
        ? getServiceAlertLabel(serviceAlertReasonFilter.value)
        : '全部告警',
      队伍: item.team_name,
      靶题: getChallengeTitle(item.challenge_id),
      服务状态: getServiceStatusLabel(item.service_status),
      巡检来源: getCheckSourceLabel(item.check_result.check_source) || '',
      检查摘要: summarizeCheckResult(item.check_result),
      防守得分: item.defense_score,
      受攻击次数: item.attack_received,
      更新时间: formatDateTime(item.updated_at),
    }))
  )
}

function exportFilteredAttacks() {
  if (filteredAttacks.value.length === 0) {
    return
  }
  downloadCSVFile(
    buildExportFilename('attacks'),
    filteredAttacks.value.map((item) => ({
      赛事: props.contest.title,
      轮次: getSelectedRoundLabel(),
      筛选队伍: attackTeamFilter.value
        ? attackTeamOptions.value.find((team) => team.id === attackTeamFilter.value)?.name ||
          attackTeamFilter.value
        : '全部队伍',
      筛选结果:
        attackResultFilter.value === 'all'
          ? '全部结果'
          : attackResultFilter.value === 'success'
            ? '仅成功'
            : '仅失败',
      筛选来源:
        attackSourceFilter.value === 'all'
          ? '全部来源'
          : getAttackSourceLabel(attackSourceFilter.value),
      时间: formatDateTime(item.created_at),
      攻击方: item.attacker_team,
      受害方: item.victim_team,
      靶题: getChallengeTitle(item.challenge_id),
      攻击类型: getAttackTypeLabel(item.attack_type),
      记录来源: getAttackSourceLabel(item.source),
      攻击结果: item.is_success ? '成功' : '失败',
      得分: item.score_gained,
      提交Flag: item.submitted_flag || '',
    }))
  )
}

function exportReviewPackage() {
  if (!selectedRound.value) {
    return
  }

  downloadJSONFile(buildExportFilename('review-package').replace(/\.csv$/, '.json'), {
    exported_at: new Date().toISOString(),
    contest: {
      id: props.contest.id,
      title: props.contest.title,
      status: props.contest.status,
      mode: props.contest.mode,
    },
    round: {
      id: selectedRound.value.id,
      round_number: selectedRound.value.round_number,
      status: selectedRound.value.status,
      attack_score: selectedRound.value.attack_score,
      defense_score: selectedRound.value.defense_score,
      started_at: selectedRound.value.started_at || null,
      ended_at: selectedRound.value.ended_at || null,
      updated_at: selectedRound.value.updated_at,
    },
    filters: {
      service: {
        team_id: serviceTeamFilter.value || null,
        team_name:
          serviceTeamOptions.value.find((team) => team.team_id === serviceTeamFilter.value)
            ?.team_name || null,
        status: serviceStatusFilter.value,
        check_source: serviceCheckSourceFilter.value || null,
        check_source_label: serviceCheckSourceFilter.value
          ? getCheckSourceLabel(serviceCheckSourceFilter.value) || serviceCheckSourceFilter.value
          : '全部来源',
        alert_reason: serviceAlertReasonFilter.value || null,
        alert_reason_label: serviceAlertReasonFilter.value
          ? getServiceAlertLabel(serviceAlertReasonFilter.value)
          : '全部告警',
      },
      attack: {
        team_id: attackTeamFilter.value || null,
        team_name:
          attackTeamOptions.value.find((team) => team.id === attackTeamFilter.value)?.name || null,
        result: attackResultFilter.value,
        source: attackSourceFilter.value,
        source_label:
          attackSourceFilter.value === 'all'
            ? '全部来源'
            : getAttackSourceLabel(attackSourceFilter.value),
      },
    },
    summary: {
      round: props.summary?.round || null,
      metrics: props.summary?.metrics || null,
      items: props.summary?.items || [],
      service_alerts: serviceAlerts.value.map((alert) => ({
        key: alert.key,
        label: alert.label,
        count: alert.count,
        affected_teams: alert.affected_teams,
        samples: alert.samples,
      })),
    },
    scoreboard: {
      frozen: props.scoreboardFrozen,
      rows: props.scoreboardRows,
    },
    services: filteredServices.value.map((item) => ({
      id: item.id,
      team_id: item.team_id,
      team_name: item.team_name,
      challenge_id: item.challenge_id,
      challenge_title: getChallengeTitle(item.challenge_id),
      service_status: item.service_status,
      service_status_label: getServiceStatusLabel(item.service_status),
      check_source: getServiceCheckSourceValue(item.check_result) || null,
      check_source_label: getCheckSourceLabel(item.check_result.check_source) || '',
      check_result: item.check_result,
      defense_score: item.defense_score,
      attack_received: item.attack_received,
      attack_score: item.attack_score,
      updated_at: item.updated_at,
    })),
    attacks: filteredAttacks.value.map((item) => ({
      id: item.id,
      attacker_team_id: item.attacker_team_id,
      attacker_team: item.attacker_team,
      victim_team_id: item.victim_team_id,
      victim_team: item.victim_team,
      challenge_id: item.challenge_id,
      challenge_title: getChallengeTitle(item.challenge_id),
      attack_type: item.attack_type,
      attack_type_label: getAttackTypeLabel(item.attack_type),
      source: item.source,
      source_label: getAttackSourceLabel(item.source),
      is_success: item.is_success,
      score_gained: item.score_gained,
      submitted_flag: item.submitted_flag || null,
      created_at: item.created_at,
    })),
  })
}

watch(
  () => props.trafficFilters.path_keyword,
  (keyword) => {
    trafficPathKeywordInput.value = keyword
  },
  { immediate: true }
)

const checkButtonLabel = computed(() => {
  if (props.checking) {
    return '执行巡检中...'
  }
  if (!selectedRound.value) {
    return '立即巡检所选轮次'
  }
  if (selectedRound.value.status === 'running') {
    return '立即巡检当前轮'
  }
  return `重跑第 ${selectedRound.value.round_number} 轮巡检`
})
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <div
        class="rounded-[28px] border border-[var(--color-primary)]/20 bg-[linear-gradient(145deg,rgba(8,145,178,0.15),rgba(15,23,42,0.94))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
      >
        <div
          class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-primary-hover)]/75"
        >
          <span>AWD Operations</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">真实接口</span>
        </div>
        <div class="mt-3 flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-3xl font-semibold tracking-tight text-white">{{ contest.title }}</h2>
            <p class="mt-3 text-sm leading-7 text-[var(--color-text-secondary)]/90">
              针对当前 AWD 赛事查看轮次态势、服务健康、攻击记录，并支持立即触发当前轮巡检。
            </p>
          </div>
          <span
            v-if="selectedRound"
            class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
            :class="getRoundStatusClass(selectedRound.status)"
          >
            第 {{ selectedRound.round_number }} 轮 · {{ getRoundStatusLabel(selectedRound.status) }}
          </span>
        </div>

        <div class="mt-6 flex flex-wrap items-center gap-3">
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary"
            :disabled="loadingRounds || loadingRoundDetail"
            @click="emit('refresh')"
          >
            <RefreshCw class="h-4 w-4" />
            刷新 AWD 数据
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary"
            @click="emit('openCreateRoundDialog')"
          >
            <TimerReset class="h-4 w-4" />
            创建轮次
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!selectedRoundId || !canRecordServiceChecks"
            @click="emit('openServiceCheckDialog')"
          >
            <ShieldCheck class="h-4 w-4" />
            录入服务检查
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!selectedRoundId || !canRecordAttackLogs"
            @click="emit('openAttackLogDialog')"
          >
            <Sword class="h-4 w-4" />
            补录攻击日志
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="checking || !selectedRoundId"
            @click="emit('runSelectedRoundCheck')"
          >
            <TimerReset class="h-4 w-4" />
            {{ checkButtonLabel }}
          </button>
        </div>
        <p v-if="shouldAutoRefresh" class="mt-3 text-xs text-[var(--color-primary-hover)]/70">
          当前正在跟随 live 轮次，面板会每 15 秒自动刷新一次。
        </p>
        <p
          v-if="selectedRoundId && !canRecordServiceChecks && serviceCheckHint"
          class="mt-1 text-xs text-[var(--color-primary-hover)]/70"
        >
          {{ serviceCheckHint }}
        </p>
        <p
          v-if="selectedRoundId && !canRecordAttackLogs && attackLogHint"
          class="mt-1 text-xs text-[var(--color-primary-hover)]/70"
        >
          {{ attackLogHint }}
        </p>
      </div>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <AppCard
          variant="metric"
          accent="primary"
          eyebrow="轮次数量"
          :title="String(rounds.length)"
          subtitle="当前赛事已创建的 AWD 轮次。"
        >
          <template #header>
            <div
              class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary"
            >
              <TimerReset class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="metric"
          accent="warning"
          eyebrow="失陷服务"
          :title="String(compromisedCount)"
          subtitle="当前所选轮次中已被攻破的服务数。"
        >
          <template #header>
            <div
              class="flex h-11 w-11 items-center justify-center rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 text-[var(--color-danger)]"
            >
              <ShieldAlert class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="metric"
          accent="success"
          eyebrow="攻击流量"
          :title="String(totalAttackCount)"
          :subtitle="`成功 ${successfulAttackCount} / 失败 ${failedAttackCount}`"
        >
          <template #header>
            <div
              class="flex h-11 w-11 items-center justify-center rounded-2xl border border-[var(--color-success)]/20 bg-[var(--color-success)]/10 text-[var(--color-success)]"
            >
              <Sword class="h-5 w-5" />
            </div>
          </template>
        </AppCard>
      </div>
    </section>

    <section class="grid gap-6 xl:grid-cols-[0.85fr_1.15fr]">
      <SectionCard title="轮次切换" subtitle="查看当前轮的基础参数与状态。">
        <div class="space-y-4">
          <label class="space-y-2">
            <span class="text-sm text-[var(--color-text-secondary)]">选择轮次</span>
            <select
              id="awd-round-selector"
              :value="selectedRoundId || ''"
              class="w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
              :disabled="loadingRounds || rounds.length === 0"
              @change="emit('update:selectedRoundId', ($event.target as HTMLSelectElement).value)"
            >
              <option v-for="round in rounds" :key="round.id" :value="round.id">
                第 {{ round.round_number }} 轮 · {{ getRoundStatusLabel(round.status) }}
              </option>
            </select>
          </label>

          <div v-if="loadingRounds" class="flex justify-center py-8">
            <AppLoading>正在同步轮次...</AppLoading>
          </div>

          <AppEmpty
            v-else-if="rounds.length === 0"
            title="当前赛事还没有 AWD 轮次"
            description="先让后台调度创建轮次，随后这里会展示服务状态和攻击数据。"
            icon="Flag"
          />

          <div v-else-if="selectedRound" class="grid gap-3">
            <AppCard
              variant="action"
              accent="neutral"
              eyebrow="轮次状态"
              :subtitle="getRoundStatusLabel(selectedRound.status)"
            >
              <template #default>
                <div class="text-sm text-[var(--color-text-secondary)]">
                  <p>攻击分值：{{ selectedRound.attack_score }}</p>
                  <p class="mt-1">防守分值：{{ selectedRound.defense_score }}</p>
                </div>
              </template>
            </AppCard>

            <AppCard
              variant="action"
              accent="neutral"
              eyebrow="时间窗口"
              :subtitle="formatDateTime(selectedRound.started_at)"
            >
              <template #default>
                <div class="text-sm text-[var(--color-text-secondary)]">
                  <p>开始：{{ formatDateTime(selectedRound.started_at) }}</p>
                  <p class="mt-1">结束：{{ formatDateTime(selectedRound.ended_at) }}</p>
                </div>
              </template>
            </AppCard>

            <AppCard
              variant="action"
              accent="warning"
              eyebrow="异常速览"
              :subtitle="`下线 ${downCount} · 失陷 ${compromisedCount}`"
            >
              <template #default>
                <div class="text-sm text-[var(--color-text-secondary)]">
                  <p>服务总数：{{ totalServiceCount }}</p>
                  <p class="mt-1">最后巡检：{{ formatDateTime(selectedRound.updated_at) }}</p>
                </div>
              </template>
            </AppCard>
          </div>
        </div>
      </SectionCard>

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
              class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
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

          <section class="space-y-4 border-t border-border pt-6">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h3 class="text-base font-semibold text-[var(--color-text-primary)]">
                  攻击流量态势
                </h3>
                <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                  代理请求摘要，不等同于已确认攻破结果。
                </p>
              </div>
              <span class="text-xs text-[var(--color-text-muted)]"
                >最近更新时间：{{ formatDateTime(selectedRound.updated_at) }}</span
              >
            </div>

            <div
              v-if="loadingTrafficSummary"
              class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
            >
              正在加载攻击流量摘要...
            </div>
            <div
              v-else-if="!trafficSummary"
              class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
            >
              当前轮次暂未返回攻击流量摘要。
            </div>
            <template v-else>
              <div
                class="border-y border-border/80 bg-[linear-gradient(180deg,rgba(8,145,178,0.06),rgba(8,145,178,0))]"
              >
                <div class="grid gap-0 md:grid-cols-2 xl:grid-cols-5">
                  <div
                    v-for="item in trafficSummaryStats"
                    :key="item.key"
                    class="border-b border-border/70 px-4 py-4 last:border-b-0 md:[&:nth-last-child(-n+2)]:border-b-0 xl:border-b-0 xl:border-r xl:last:border-r-0"
                  >
                    <p
                      class="text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >
                      {{ item.label }}
                    </p>
                    <p
                      class="mt-3 text-2xl font-semibold tracking-tight text-[var(--color-text-primary)]"
                    >
                      {{ item.value }}
                    </p>
                    <p class="mt-2 text-xs leading-6 text-[var(--color-text-muted)]">
                      {{ item.hint }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="grid gap-4 xl:grid-cols-[1.3fr_0.7fr]">
                <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                  <div class="rounded-xl border border-border/80">
                    <div
                      class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]"
                    >
                      热点攻击队
                    </div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_attackers.slice(0, 5)"
                        :key="`traffic-attacker-${item.team_id}`"
                        class="flex items-center justify-between px-3 py-2 text-sm"
                      >
                        <span class="text-[var(--color-text-primary)]">{{ item.team_name }}</span>
                        <span class="font-medium text-[var(--color-text-secondary)]">{{
                          item.request_count
                        }}</span>
                      </li>
                      <li
                        v-if="trafficSummary.top_attackers.length === 0"
                        class="px-3 py-3 text-xs text-[var(--color-text-muted)]"
                      >
                        暂无攻击队热点数据
                      </li>
                    </ol>
                  </div>
                  <div class="rounded-xl border border-border/80">
                    <div
                      class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]"
                    >
                      热点受害队
                    </div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_victims.slice(0, 5)"
                        :key="`traffic-victim-${item.team_id}`"
                        class="flex items-center justify-between px-3 py-2 text-sm"
                      >
                        <span class="text-[var(--color-text-primary)]">{{ item.team_name }}</span>
                        <span class="font-medium text-[var(--color-text-secondary)]">{{
                          item.request_count
                        }}</span>
                      </li>
                      <li
                        v-if="trafficSummary.top_victims.length === 0"
                        class="px-3 py-3 text-xs text-[var(--color-text-muted)]"
                      >
                        暂无目标热点数据
                      </li>
                    </ol>
                  </div>
                  <div class="rounded-xl border border-border/80">
                    <div
                      class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]"
                    >
                      热点题目
                    </div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_challenges.slice(0, 5)"
                        :key="`traffic-challenge-${item.challenge_id}`"
                        class="flex items-center justify-between gap-3 px-3 py-2 text-sm"
                      >
                        <span class="truncate text-[var(--color-text-primary)]">{{
                          getTrafficChallengeTitle(item.challenge_id, item.challenge_title)
                        }}</span>
                        <span class="shrink-0 font-medium text-[var(--color-text-secondary)]">{{
                          item.request_count
                        }}</span>
                      </li>
                      <li
                        v-if="trafficSummary.top_challenges.length === 0"
                        class="px-3 py-3 text-xs text-[var(--color-text-muted)]"
                      >
                        暂无题目热点数据
                      </li>
                    </ol>
                  </div>
                  <div class="rounded-xl border border-border/80">
                    <div
                      class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]"
                    >
                      异常路径
                    </div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_error_paths.slice(0, 5)"
                        :key="`traffic-path-${item.path}`"
                        class="px-3 py-2 text-sm"
                      >
                        <p class="truncate font-mono text-[var(--color-text-primary)]">
                          {{ item.path }}
                        </p>
                        <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                          请求 {{ item.request_count }} / 错误 {{ item.error_count }}
                        </p>
                      </li>
                      <li
                        v-if="trafficSummary.top_error_paths.length === 0"
                        class="px-3 py-3 text-xs text-[var(--color-text-muted)]"
                      >
                        暂无异常路径数据
                      </li>
                    </ol>
                  </div>
                </div>

                <div class="rounded-xl border border-border/80">
                  <div class="border-b border-border bg-surface-alt/40 px-3 py-2">
                    <p
                      class="text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]"
                    >
                      趋势摘要（最近 12 桶）
                    </p>
                    <p class="mt-2 text-xs leading-6 text-[var(--color-text-muted)]">
                      {{ trafficTrendNarrative }}
                    </p>
                  </div>
                  <div class="space-y-2 px-3 py-3">
                    <div
                      v-for="bucket in trafficTrendRows"
                      :key="bucket.bucket_start_at"
                      class="space-y-1"
                    >
                      <div
                        class="flex items-center justify-between text-xs text-[var(--color-text-muted)]"
                      >
                        <span>{{ bucket.label }}</span>
                        <span>请求 {{ bucket.request_count }} / 错误 {{ bucket.error_count }}</span>
                      </div>
                      <div
                        class="h-1.5 overflow-hidden rounded-full bg-[var(--color-text-muted)]/20"
                      >
                        <div
                          class="h-full rounded-full bg-[var(--color-primary)]/70"
                          :style="{ width: `${bucket.ratio}%` }"
                        />
                      </div>
                    </div>
                    <p
                      v-if="trafficTrendRows.length === 0"
                      class="text-xs text-[var(--color-text-muted)]"
                    >
                      当前没有趋势桶数据。
                    </p>
                  </div>
                </div>
              </div>
            </template>

            <div class="overflow-hidden rounded-xl border border-border">
              <div
                class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/60 px-4 py-3"
              >
                <div>
                  <p class="text-sm font-semibold text-[var(--color-text-primary)]">流量明细表</p>
                  <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                    按攻击方、受害方、题目、状态分桶和路径关键字筛选。
                  </p>
                </div>
                <button
                  id="awd-traffic-reset-filters"
                  type="button"
                  class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-[var(--color-text-primary)] transition hover:border-primary"
                  @click="emit('resetTrafficFilters')"
                >
                  重置筛选
                </button>
              </div>

              <div
                class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-5"
              >
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >攻击方</span
                  >
                  <select
                    id="awd-traffic-filter-attacker"
                    :value="trafficFilters.attacker_team_id"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="
                      applyTrafficFilterPatch({
                        attacker_team_id: ($event.target as HTMLSelectElement).value,
                      })
                    "
                  >
                    <option value="">全部攻击方</option>
                    <option
                      v-for="team in trafficTeamOptions"
                      :key="`traffic-attacker-option-${team.id}`"
                      :value="team.id"
                    >
                      {{ team.name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >受害方</span
                  >
                  <select
                    id="awd-traffic-filter-victim"
                    :value="trafficFilters.victim_team_id"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="
                      applyTrafficFilterPatch({
                        victim_team_id: ($event.target as HTMLSelectElement).value,
                      })
                    "
                  >
                    <option value="">全部受害方</option>
                    <option
                      v-for="team in trafficTeamOptions"
                      :key="`traffic-victim-option-${team.id}`"
                      :value="team.id"
                    >
                      {{ team.name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >题目</span
                  >
                  <select
                    id="awd-traffic-filter-challenge"
                    :value="trafficFilters.challenge_id"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="
                      applyTrafficFilterPatch({
                        challenge_id: ($event.target as HTMLSelectElement).value,
                      })
                    "
                  >
                    <option value="">全部题目</option>
                    <option
                      v-for="challenge in challengeLinks"
                      :key="challenge.id"
                      :value="challenge.challenge_id"
                    >
                      {{ challenge.title || `Challenge #${challenge.challenge_id}` }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >状态分桶</span
                  >
                  <select
                    id="awd-traffic-filter-status-group"
                    :value="trafficFilters.status_group"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="onTrafficStatusGroupChange(($event.target as HTMLSelectElement).value)"
                  >
                    <option
                      v-for="item in trafficStatusGroupOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >路径关键字</span
                  >
                  <div class="flex items-center gap-2">
                    <input
                      id="awd-traffic-filter-path"
                      v-model="trafficPathKeywordInput"
                      type="text"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                      placeholder="/api/..."
                      @keydown.enter.prevent="applyTrafficKeywordFilter"
                    />
                    <button
                      id="awd-traffic-filter-search"
                      type="button"
                      class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border text-[var(--color-text-secondary)] transition hover:border-primary hover:text-[var(--color-text-primary)]"
                      @click="applyTrafficKeywordFilter"
                    >
                      <Search class="h-4 w-4" />
                    </button>
                  </div>
                  <button
                    v-if="trafficFilters.path_keyword"
                    type="button"
                    class="text-xs text-[var(--color-text-muted)] underline-offset-2 transition hover:text-[var(--color-text-primary)] hover:underline"
                    @click="clearTrafficKeywordFilter"
                  >
                    清除路径关键字
                  </button>
                </label>
              </div>

              <table class="min-w-full divide-y divide-border">
                <thead
                  class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                >
                  <tr>
                    <th class="px-4 py-3">时间</th>
                    <th class="px-4 py-3">攻击方 / 受害方</th>
                    <th class="px-4 py-3">靶题</th>
                    <th class="px-4 py-3">请求</th>
                    <th class="px-4 py-3">状态</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border bg-surface/70">
                  <tr v-if="loadingTrafficEvents">
                    <td
                      colspan="5"
                      class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
                    >
                      正在加载流量明细...
                    </td>
                  </tr>
                  <tr
                    v-for="event in trafficEvents"
                    :key="`${event.occurred_at}-${event.attacker_team_id}-${event.victim_team_id}-${event.challenge_id}-${event.method}-${event.path}`"
                  >
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      {{ formatDateTime(event.occurred_at) }}
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">
                      <p>
                        {{ getTrafficTeamName(event.attacker_team_id, event.attacker_team_name) }}
                      </p>
                      <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                        → {{ getTrafficTeamName(event.victim_team_id, event.victim_team_name) }}
                      </p>
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      {{ getTrafficChallengeTitle(event.challenge_id, event.challenge_title) }}
                    </td>
                    <td class="px-4 py-4 text-sm">
                      <p class="font-mono text-[var(--color-text-primary)]">
                        {{ event.method.toUpperCase() }} {{ event.path }}
                      </p>
                      <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                        HTTP {{ event.status_code }}
                      </p>
                    </td>
                    <td class="px-4 py-4 text-sm">
                      <span
                        class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                        :class="getTrafficStatusGroupClass(event.status_group)"
                      >
                        {{ getTrafficStatusGroupLabel(event.status_group) }}
                      </span>
                      <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                        {{ getTrafficSourceLabel(event.source) }}
                      </p>
                    </td>
                  </tr>
                  <tr v-if="!loadingTrafficEvents && trafficEvents.length === 0">
                    <td
                      colspan="5"
                      class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
                    >
                      当前筛选条件下没有流量事件。
                    </td>
                  </tr>
                </tbody>
              </table>

              <div
                class="border-t border-border bg-surface-alt/20 px-4 py-3 text-xs text-[var(--color-text-muted)]"
              >
                <AdminPaginationControls
                  :page="trafficFilters.page"
                  :total-pages="trafficTotalPages"
                  :total="trafficEventsTotal"
                  :disabled="loadingTrafficEvents"
                  :total-label="`共 ${trafficEventsTotal} 条流量事件`"
                  prev-button-id="awd-traffic-page-prev"
                  next-button-id="awd-traffic-page-next"
                  @change-page="handleTrafficPageChange"
                />
              </div>
            </div>
          </section>

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
                        :key="`${alert.key}-${sample.service_id}`"
                      >
                        {{ sample.team_name }} · {{ sample.challenge_title }}
                      </div>
                    </div>
                    <div class="mt-2 text-xs text-[var(--color-text-secondary)]/80">
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

          <div class="overflow-hidden rounded-2xl border border-border">
            <div
              class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3"
            >
              <div class="text-sm font-semibold text-[var(--color-text-primary)]">实时排行榜</div>
              <span
                v-if="scoreboardFrozen"
                class="inline-flex rounded-full border border-[var(--color-warning)]/20 bg-[var(--color-warning)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-warning)]"
              >
                排行榜已冻结
              </span>
            </div>
            <table class="min-w-full divide-y divide-border">
              <thead
                class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
              >
                <tr>
                  <th class="px-4 py-3">排名</th>
                  <th class="px-4 py-3">队伍</th>
                  <th class="px-4 py-3">得分</th>
                  <th class="px-4 py-3">解题数</th>
                  <th class="px-4 py-3">最近得分</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border bg-surface/70">
                <tr v-for="item in scoreboardRows" :key="item.team_id">
                  <td class="px-4 py-4 text-sm font-semibold text-[var(--color-text-primary)]">
                    #{{ item.rank }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">
                    {{ item.team_name }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">
                    {{ formatScore(item.score) }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">
                    {{ item.solved_count }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">
                    {{ formatDateTime(item.last_submission_at) }}
                  </td>
                </tr>
                <tr v-if="scoreboardRows.length === 0">
                  <td
                    colspan="5"
                    class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
                  >
                    当前赛事还没有排行榜数据。
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="overflow-hidden rounded-2xl border border-border">
            <div
              class="border-b border-border bg-surface-alt/70 px-4 py-3 text-sm font-semibold text-[var(--color-text-primary)]"
            >
              本轮汇总
            </div>
            <table class="min-w-full divide-y divide-border">
              <thead
                class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
              >
                <tr>
                  <th class="px-4 py-3">队伍</th>
                  <th class="px-4 py-3">总分</th>
                  <th class="px-4 py-3">攻击 / 防守</th>
                  <th class="px-4 py-3">服务状态</th>
                  <th class="px-4 py-3">被攻击情况</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border bg-surface/70">
                <tr v-for="item in summary?.items || []" :key="item.team_id">
                  <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">
                    {{ item.team_name }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">
                    {{ item.total_score }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                    {{ item.attack_score }} / {{ item.defense_score }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                    正常 {{ item.service_up_count }} / 下线 {{ item.service_down_count }} / 失陷
                    {{ item.service_compromised_count }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                    攻破 {{ item.successful_breach_count }} 次，攻击方
                    {{ item.unique_attackers_against }} 支
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="grid gap-6 xl:grid-cols-2">
            <div class="overflow-hidden rounded-2xl border border-border">
              <div
                class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3"
              >
                <div class="text-sm font-semibold text-[var(--color-text-primary)]">服务状态表</div>
                <button
                  id="awd-export-services"
                  type="button"
                  class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="filteredServices.length === 0"
                  @click="exportFilteredServices"
                >
                  导出当前筛选
                </button>
              </div>
              <div
                class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-4"
              >
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >队伍</span
                  >
                  <select
                    id="awd-service-filter-team"
                    v-model="serviceTeamFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部队伍</option>
                    <option
                      v-for="team in serviceTeamOptions"
                      :key="team.team_id"
                      :value="team.team_id"
                    >
                      {{ team.team_name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >状态</span
                  >
                  <select
                    id="awd-service-filter-status"
                    v-model="serviceStatusFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="all">全部状态</option>
                    <option value="up">正常</option>
                    <option value="down">下线</option>
                    <option value="compromised">已失陷</option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >巡检来源</span
                  >
                  <select
                    id="awd-service-filter-source"
                    v-model="serviceCheckSourceFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部来源</option>
                    <option
                      v-for="source in serviceCheckSourceOptions"
                      :key="source"
                      :value="source"
                    >
                      {{ getCheckSourceLabel(source) || source }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >告警类型</span
                  >
                  <select
                    id="awd-service-filter-alert"
                    v-model="serviceAlertReasonFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部告警</option>
                    <option v-for="alert in serviceAlerts" :key="alert.key" :value="alert.key">
                      {{ alert.label }}
                    </option>
                  </select>
                </label>
              </div>
              <table class="min-w-full divide-y divide-border">
                <thead
                  class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                >
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
                    <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">
                      {{ service.team_name }}
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      {{ getChallengeTitle(service.challenge_id) }}
                    </td>
                    <td class="px-4 py-4">
                      <span
                        class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                        :class="getServiceStatusClass(service.service_status)"
                      >
                        {{ getServiceStatusLabel(service.service_status) }}
                      </span>
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      防守 {{ service.defense_score }} / 受攻击 {{ service.attack_received }}
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">
                      <div>{{ summarizeCheckResult(service.check_result) }}</div>
                      <div
                        v-if="getTargetProbeSummary(service.check_result)"
                        class="mt-2 text-xs text-[var(--color-text-muted)]"
                      >
                        {{ getTargetProbeSummary(service.check_result) }}
                      </div>
                      <details
                        v-if="getCheckTargets(service.check_result).length > 0"
                        class="mt-2 rounded-xl border border-border/80 bg-surface-alt/40 p-3 text-xs text-[var(--color-text-secondary)]"
                      >
                        <summary
                          class="cursor-pointer select-none text-[var(--color-text-primary)]"
                        >
                          查看探测明细
                        </summary>
                        <div class="mt-3 space-y-3">
                          <div
                            v-for="(target, targetIndex) in getCheckTargets(service.check_result)"
                            :key="`${service.id}-target-${targetIndex}`"
                            class="rounded-xl border border-border/70 bg-surface/70 p-3"
                          >
                            <div class="font-medium text-[var(--color-text-primary)]">
                              {{ target.access_url || `Target #${targetIndex + 1}` }}
                            </div>
                            <div class="mt-1 text-[var(--color-text-muted)]">
                              {{
                                getProbeStatusText(target.healthy, target.error_code, target.error)
                              }}
                              <span v-if="target.probe"> · {{ target.probe.toUpperCase() }}</span>
                              <span v-if="formatLatency(target.latency_ms)">
                                · {{ formatLatency(target.latency_ms) }}</span
                              >
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
                                {{
                                  getProbeStatusText(
                                    attempt.healthy,
                                    attempt.error_code,
                                    attempt.error
                                  )
                                }}
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
                    <td
                      colspan="5"
                      class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
                    >
                      {{
                        services.length === 0
                          ? '当前轮次还没有服务巡检记录。'
                          : '当前筛选条件下没有服务记录。'
                      }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="overflow-hidden rounded-2xl border border-border">
              <div
                class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3"
              >
                <div class="text-sm font-semibold text-[var(--color-text-primary)]">攻击日志</div>
                <button
                  id="awd-export-attacks"
                  type="button"
                  class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="filteredAttacks.length === 0"
                  @click="exportFilteredAttacks"
                >
                  导出当前筛选
                </button>
              </div>
              <div
                class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-3"
              >
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >队伍</span
                  >
                  <select
                    id="awd-attack-filter-team"
                    v-model="attackTeamFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部队伍</option>
                    <option v-for="team in attackTeamOptions" :key="team.id" :value="team.id">
                      {{ team.name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >结果</span
                  >
                  <select
                    id="awd-attack-filter-result"
                    v-model="attackResultFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="all">全部结果</option>
                    <option value="success">仅成功</option>
                    <option value="failed">仅失败</option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span
                    class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                    >记录来源</span
                  >
                  <select
                    id="awd-attack-filter-source"
                    v-model="attackSourceFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="all">全部来源</option>
                    <option v-for="source in attackSourceOptions" :key="source" :value="source">
                      {{ getAttackSourceLabel(source) }}
                    </option>
                  </select>
                </label>
              </div>
              <table class="min-w-full divide-y divide-border">
                <thead
                  class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
                >
                  <tr>
                    <th class="px-4 py-3">时间</th>
                    <th class="px-4 py-3">攻击方</th>
                    <th class="px-4 py-3">受害方</th>
                    <th class="px-4 py-3">类型</th>
                    <th class="px-4 py-3">结果</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border bg-surface/70">
                  <tr v-for="attack in filteredAttacks" :key="attack.id">
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      {{ formatDateTime(attack.created_at) }}
                    </td>
                    <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">
                      {{ attack.attacker_team }}
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      {{ attack.victim_team }}
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      <div>{{ getAttackTypeLabel(attack.attack_type) }}</div>
                      <div class="mt-1 text-xs text-[var(--color-text-muted)]">
                        {{ getChallengeTitle(attack.challenge_id) }}
                      </div>
                      <div class="mt-1 text-xs text-[var(--color-text-muted)]">
                        {{ getAttackSourceLabel(attack.source) }}
                      </div>
                    </td>
                    <td class="px-4 py-4 text-sm">
                      <span
                        class="inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold"
                        :class="
                          attack.is_success
                            ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]'
                            : 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-secondary)]'
                        "
                      >
                        <ShieldCheck v-if="attack.is_success" class="h-3.5 w-3.5" />
                        {{ attack.is_success ? `成功 +${attack.score_gained}` : '失败' }}
                      </span>
                    </td>
                  </tr>
                  <tr v-if="filteredAttacks.length === 0">
                    <td
                      colspan="5"
                      class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
                    >
                      {{
                        attacks.length === 0
                          ? '当前轮次还没有攻击记录。'
                          : '当前筛选条件下没有攻击记录。'
                      }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </SectionCard>
    </section>
  </div>
</template>

<script setup lang="ts">
import { toRef } from 'vue'
import { RefreshCw, ShieldAlert, ShieldCheck, Sword, TimerReset } from 'lucide-vue-next'
import type { AWDTeamServiceData } from '@/api/contracts'
import AWDTrafficPanel from '@/components/admin/contest/AWDTrafficPanel.vue'
import type {
  AWDRoundInspectorEmits,
  AWDRoundInspectorProps,
} from '@/components/admin/contest/awdInspector.types'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
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
    <section class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <div
        class="awd-round-hero rounded-[28px] border p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
      >
        <div
          class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-primary-hover)]/75"
        >
          <span>AWD Operations</span>
          <span class="awd-round-hero-chip rounded-full px-2 py-1">真实接口</span>
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

        <div class="mt-6 flex flex-wrap items-center gap-3 awd-round-toolbar">
          <button
            type="button"
            class="ui-btn ui-btn--secondary awd-round-toolbar__button"
            :disabled="loadingRounds || loadingRoundDetail"
            @click="emit('refresh')"
          >
            <RefreshCw class="h-4 w-4" />
            刷新 AWD 数据
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--secondary awd-round-toolbar__button"
            @click="emit('openCreateRoundDialog')"
          >
            <TimerReset class="h-4 w-4" />
            创建轮次
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--secondary awd-round-toolbar__button"
            :disabled="!selectedRoundId || !canRecordServiceChecks"
            @click="emit('openServiceCheckDialog')"
          >
            <ShieldCheck class="h-4 w-4" />
            录入服务检查
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--secondary awd-round-toolbar__button"
            :disabled="!selectedRoundId || !canRecordAttackLogs"
            @click="emit('openAttackLogDialog')"
          >
            <Sword class="h-4 w-4" />
            补录攻击日志
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--primary awd-round-toolbar__button"
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
          <label class="ui-field awd-round-filter-field">
            <span class="ui-field__label">选择轮次</span>
            <span
              class="ui-control-wrap awd-round-filter-control"
              :class="{ 'is-disabled': loadingRounds || rounds.length === 0 }"
            >
              <select
                id="awd-round-selector"
                :value="selectedRoundId || ''"
                class="ui-control"
                :disabled="loadingRounds || rounds.length === 0"
                @change="emit('update:selectedRoundId', ($event.target as HTMLSelectElement).value)"
              >
                <option v-for="round in rounds" :key="round.id" :value="round.id">
                  第 {{ round.round_number }} 轮 · {{ getRoundStatusLabel(round.status) }}
                </option>
              </select>
            </span>
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
                  <th class="px-4 py-3">SLA / 攻击 / 防守</th>
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
                    SLA {{ item.sla_score ?? 0 }} / 攻击 {{ item.attack_score }} / 防守
                    {{ item.defense_score }}
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
                  class="ui-btn ui-btn--secondary awd-round-toolbar__button"
                  :disabled="filteredServices.length === 0"
                  @click="exportFilteredServices"
                >
                  导出当前筛选
                </button>
              </div>
              <div
                class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-4"
              >
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">队伍</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-service-filter-team"
                      v-model="serviceTeamFilter"
                      class="ui-control"
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
                  </span>
                </label>
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">状态</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-service-filter-status"
                      v-model="serviceStatusFilter"
                      class="ui-control"
                    >
                      <option value="all">全部状态</option>
                      <option value="up">正常</option>
                      <option value="down">下线</option>
                      <option value="compromised">已失陷</option>
                    </select>
                  </span>
                </label>
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">巡检来源</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-service-filter-source"
                      v-model="serviceCheckSourceFilter"
                      class="ui-control"
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
                  </span>
                </label>
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">告警类型</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-service-filter-alert"
                      v-model="serviceAlertReasonFilter"
                      class="ui-control"
                    >
                      <option value="">全部告警</option>
                      <option v-for="alert in serviceAlerts" :key="alert.key" :value="alert.key">
                        {{ alert.label }}
                      </option>
                    </select>
                  </span>
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
                      <div>
                        SLA {{ service.sla_score ?? 0 }} / 防守 {{ service.defense_score }} / 攻击
                        {{ service.attack_score }}
                      </div>
                      <div class="mt-1 text-xs text-[var(--color-text-muted)]">
                        受攻击 {{ service.attack_received }}
                      </div>
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">
                      <div>
                        {{ summarizeCheckResult(getServiceCheckPresentationResult(service)) }}
                      </div>
                      <div
                        v-if="getCheckActions(service.check_result).length > 0"
                        class="mt-2 flex flex-wrap gap-2 text-xs text-[var(--color-text-secondary)]"
                      >
                        <span
                          v-for="action in getCheckActions(service.check_result)"
                          :key="`${service.id}-action-${action.key}`"
                          class="rounded-full border border-border/70 bg-surface-alt/40 px-2 py-1"
                        >
                          {{ action.label }} ·
                          {{ getProbeStatusText(action.healthy, action.error_code, action.error) }}
                          <span v-if="action.method || action.path">
                            ·
                            {{
                              [action.method?.toUpperCase(), action.path].filter(Boolean).join(' ')
                            }}
                          </span>
                        </span>
                      </div>
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
                          查看检查明细
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
                              v-if="getTargetActions(target).length > 0"
                              class="mt-2 space-y-1 border-t border-border/60 pt-2"
                            >
                              <div
                                v-for="action in getTargetActions(target)"
                                :key="`${service.id}-target-${targetIndex}-action-${action.key}`"
                              >
                                {{ action.label }} ·
                                {{
                                  getProbeStatusText(
                                    action.healthy,
                                    action.error_code,
                                    action.error
                                  )
                                }}
                                <span v-if="action.method || action.path">
                                  ·
                                  {{
                                    [action.method?.toUpperCase(), action.path]
                                      .filter(Boolean)
                                      .join(' ')
                                  }}
                                </span>
                              </div>
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
                  class="ui-btn ui-btn--secondary awd-round-toolbar__button"
                  :disabled="filteredAttacks.length === 0"
                  @click="exportFilteredAttacks"
                >
                  导出当前筛选
                </button>
              </div>
              <div
                class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-3"
              >
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">队伍</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-attack-filter-team"
                      v-model="attackTeamFilter"
                      class="ui-control"
                    >
                      <option value="">全部队伍</option>
                      <option v-for="team in attackTeamOptions" :key="team.id" :value="team.id">
                        {{ team.name }}
                      </option>
                    </select>
                  </span>
                </label>
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">结果</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-attack-filter-result"
                      v-model="attackResultFilter"
                      class="ui-control"
                    >
                      <option value="all">全部结果</option>
                      <option value="success">仅成功</option>
                      <option value="failed">仅失败</option>
                    </select>
                  </span>
                </label>
                <label class="ui-field awd-round-filter-field">
                  <span class="ui-field__label">记录来源</span>
                  <span class="ui-control-wrap awd-round-filter-control">
                    <select
                      id="awd-attack-filter-source"
                      v-model="attackSourceFilter"
                      class="ui-control"
                    >
                      <option value="all">全部来源</option>
                      <option v-for="source in attackSourceOptions" :key="source" :value="source">
                        {{ getAttackSourceLabel(source) }}
                      </option>
                    </select>
                  </span>
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

<style scoped>
.awd-round-hero {
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 15%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base))
  );
}

.awd-round-hero-chip {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 24%, transparent);
}

.awd-round-toolbar__button {
  white-space: nowrap;
}

.awd-round-filter-field {
  --ui-field-gap: var(--space-2);
  --ui-field-label-size: var(--font-size-11);
  --ui-field-label-weight: 700;
  --ui-field-label-color: var(--color-text-muted);
  min-width: 0;
}

.awd-round-filter-field .ui-field__label {
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.awd-round-filter-control {
  width: 100%;
}
</style>

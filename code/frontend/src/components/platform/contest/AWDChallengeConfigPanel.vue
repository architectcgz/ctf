<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight, Edit, Plus } from 'lucide-vue-next'

import type { AdminContestChallengeViewData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'

const props = withDefaults(
  defineProps<{
    challengeLinks: AdminContestChallengeViewData[]
    activeChallengeId?: string | null
    focusSource?: 'pool' | 'preflight' | null
    canNavigatePrevious?: boolean
    canNavigateNext?: boolean
  }>(),
  {
    activeChallengeId: null,
    focusSource: null,
    canNavigatePrevious: false,
    canNavigateNext: false,
  }
)

const emit = defineEmits<{
  create: []
  edit: [challenge: AdminContestChallengeViewData]
  previous: []
  next: []
}>()

const sortedChallengeLinks = computed(() =>
  [...props.challengeLinks].sort(
    (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
  )
)
const activeChallenge = computed(
  () =>
    sortedChallengeLinks.value.find((item) => item.challenge_id === props.activeChallengeId) || null
)

const summaryItems = computed(() => [
  {
    key: 'total',
    label: '已关联题目',
    value: String(sortedChallengeLinks.value.length),
    hint: '当前 AWD 赛事中可参与攻防的服务题目数量',
  },
  {
    key: 'configured',
    label: '已配 Checker',
    value: String(
      sortedChallengeLinks.value.filter(
        (item) =>
          Boolean(item.awd_checker_type) || Object.keys(item.awd_checker_config || {}).length > 0
      ).length
    ),
    hint: '已写入 checker 类型或 checker 配置的题目数',
  },
  {
    key: 'http-standard',
    label: 'HTTP Standard',
    value: String(
      sortedChallengeLinks.value.filter((item) => item.awd_checker_type === 'http_standard').length
    ),
    hint: '已切到 HTTP 标准 Checker 的题目数',
  },
  {
    key: 'hidden',
    label: '隐藏题目',
    value: String(sortedChallengeLinks.value.filter((item) => !item.is_visible).length),
    hint: '当前不会直接对选手展示的赛事题目数',
  },
  {
    key: 'service-linked',
    label: '已建服务关联',
    value: String(sortedChallengeLinks.value.filter((item) => Boolean(item.awd_service_id)).length),
    hint: '已落入赛事级服务关联表的题目数',
  },
])

function formatValidationDateTime(value?: string): string {
  if (!value) {
    return '未记录'
  }
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const { getPrimaryAccessURL, getValidationStateLabel } = useAwdCheckResultPresentation({
  formatDateTime: formatValidationDateTime,
})

function getCheckerTypeLabel(value?: string): string {
  switch (value) {
    case 'legacy_probe':
      return '基础探活'
    case 'http_standard':
      return 'HTTP 标准 Checker'
    default:
      return '未配置'
  }
}

function getConfigSummary(item: AdminContestChallengeViewData): string {
  const config = item.awd_checker_config || {}
  const putFlag = readActionSummary(config.put_flag, 'PUT')
  const getFlag = readActionSummary(config.get_flag, 'GET')
  const havoc = readActionSummary(config.havoc, 'Havoc')
  const healthPath =
    typeof config.health_path === 'string' && config.health_path.trim() !== ''
      ? `Health ${config.health_path.trim()}`
      : ''

  return [putFlag, getFlag, havoc, healthPath].filter(Boolean).join(' · ') || '未配置动作摘要'
}

function readActionSummary(value: unknown, label: string): string {
  if (!value || typeof value !== 'object') {
    return ''
  }
  const item = value as Record<string, unknown>
  const path = typeof item.path === 'string' ? item.path : ''
  if (!path) {
    return label
  }
  return `${label} ${path}`
}

function getChallengeTitle(item: AdminContestChallengeViewData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function buildPresentationResult(item: AdminContestChallengeViewData): Record<string, unknown> {
  const preview = item.awd_checker_last_preview_result
  if (!preview) {
    return {}
  }
  return {
    ...preview.check_result,
    preview_context: preview.preview_context,
  }
}

function getValidationStateText(item: AdminContestChallengeViewData): string {
  return getValidationStateLabel(item.awd_checker_validation_state) || '未验证'
}

function getValidationHint(item: AdminContestChallengeViewData): string {
  const previewAccessURL = getPrimaryAccessURL(buildPresentationResult(item))
  const entries = [
    item.awd_checker_last_preview_at
      ? `最近校验 ${formatValidationDateTime(item.awd_checker_last_preview_at)}`
      : '',
    previewAccessURL ? `目标 ${previewAccessURL}` : '',
  ].filter(Boolean)

  if (entries.length > 0) {
    return entries.join(' · ')
  }

  switch (item.awd_checker_validation_state) {
    case 'stale':
      return 'Checker 草稿已变化，需要重新试跑。'
    case 'failed':
      return '最近一次保存的试跑结果未通过。'
    case 'passed':
      return '最近一次保存的试跑结果已通过。'
    case 'pending':
    default:
      return '保存后可通过试跑绑定最近一次校验结果。'
  }
}

function isActiveChallenge(item: AdminContestChallengeViewData): boolean {
  return item.challenge_id === props.activeChallengeId
}
</script>

<template>
  <div class="studio-awd-config">
    <!-- 1. Header with Global Metrics -->
    <header class="studio-pane-header">
      <div class="header-main">
        <h1 class="pane-title">
          AWD 服务配置
        </h1>
        <p class="pane-description">
          针对每道题目深度定义 Checker 裁判逻辑、分值权重及就绪状态验证。
        </p>
      </div>

      <div class="studio-metric-band">
        <div
          v-for="item in summaryItems"
          :key="item.key"
          class="metric-pill"
        >
          <span class="metric-pill__label">{{ item.label }}</span>
          <span class="metric-pill__value">{{ item.value }}</span>
        </div>
      </div>
    </header>

    <!-- 3. Challenge Asset Directory -->
    <section class="studio-asset-directory">
      <header class="directory-header">
        <h3 class="directory-title">
          服务资源清单
        </h3>
        <button
          id="awd-challenge-config-create"
          class="ops-btn ops-btn--primary"
          @click="emit('create')"
        >
          <Plus class="h-3.5 w-3.5" /> 关联新资源
        </button>
      </header>

      <AppEmpty
        v-if="sortedChallengeLinks.length === 0"
        title="暂无关联服务"
        description="请先在题目池中关联题目，或点击右侧新增。"
        icon="Layers"
        class="py-20"
      />

      <div
        v-else
        class="studio-table-wrap"
      >
        <table class="studio-table">
          <thead>
            <tr>
              <th class="col-identity">
                服务身份
              </th>
              <th class="col-meta">
                裁判引擎
              </th>
              <th class="col-meta">
                分值权重
              </th>
              <th class="col-meta">
                规则摘要
              </th>
              <th class="col-status">
                就绪验证
              </th>
              <th class="col-actions">
                操作
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in sortedChallengeLinks"
              :key="item.id"
              class="studio-row"
              :class="{ 'is-active': isActiveChallenge(item) }"
            >
              <td class="col-identity">
                <div class="challenge-identity">
                  <div class="challenge-title">
                    {{ getChallengeTitle(item) }}
                  </div>
                  <div class="challenge-subtitle">
                    {{ item.category }} · RANK {{ item.order }}
                  </div>
                </div>
              </td>
              <td class="col-meta">
                <div class="engine-tag">
                  {{ getCheckerTypeLabel(item.awd_checker_type) }}
                </div>
              </td>
              <td class="col-meta">
                <div class="score-stack">
                  <span class="score-main">{{ item.points }} pts</span>
                  <span class="score-sub">SLA:{{ item.awd_sla_score }} / D:{{ item.awd_defense_score }}</span>
                </div>
              </td>
              <td class="col-meta">
                <div
                  class="rules-summary"
                  :title="getConfigSummary(item)"
                >
                  {{ getConfigSummary(item) }}
                </div>
              </td>
              <td class="col-status">
                <div class="validation-block">
                  <span
                    class="validation-pill"
                    :class="item.awd_checker_validation_state"
                  >
                    {{ getValidationStateText(item) }}
                  </span>
                  <span class="validation-time">{{ getValidationHint(item).split(' · ')[0] }}</span>
                </div>
              </td>
              <td class="col-actions">
                <button
                  :id="`awd-challenge-config-edit-${item.challenge_id}`"
                  class="action-btn"
                  @click="emit('edit', item)"
                >
                  <Edit class="h-3.5 w-3.5" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<style scoped>
.studio-awd-config { display: flex; flex-direction: column; gap: 2rem; padding: 1.5rem 2rem; background: var(--color-bg-base); }
.studio-pane-header { display: flex; justify-content: space-between; align-items: flex-end; }
.pane-title { font-size: 1.25rem; font-weight: 900; color: var(--color-text-primary); margin: 0; }
.pane-description { font-size: 13px; color: var(--color-text-secondary); margin: 0.5rem 0 0; }

/* Metric Band */
.studio-metric-band { display: flex; gap: 0.5rem; }
.metric-pill { background: var(--color-bg-surface); padding: 0.45rem 1rem; border-radius: 0.75rem; display: flex; align-items: baseline; gap: 0.75rem; border: 1px solid var(--color-border-default); }
.metric-pill__label { font-size: 8px; font-weight: 800; text-transform: uppercase; color: var(--color-text-secondary); letter-spacing: 0.05em; }
.metric-pill__value { font-size: 13px; font-weight: 900; color: var(--color-text-primary); font-family: var(--font-family-mono); }

/* Directory Header */
.directory-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.directory-title { font-size: 14px; font-weight: 900; color: var(--color-text-primary); text-transform: uppercase; letter-spacing: 0.1em; }

/* Table Styles */
.studio-table-wrap { border: 1px solid var(--color-border-default); border-radius: 1rem; background: var(--color-bg-surface); overflow: hidden; }
.studio-table { width: 100%; border-collapse: collapse; }
.studio-table th { background: var(--color-bg-elevated); padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); border-bottom: 1px solid var(--color-border-default); }
.studio-table td { padding: 1.25rem 1rem; border-bottom: 1px solid var(--color-border-subtle); }

.studio-row.is-active { background: var(--color-primary-soft); }
.studio-row.is-active .challenge-title { color: var(--color-primary); }

.challenge-title { font-size: 14px; font-weight: 800; color: var(--color-text-primary); }
.challenge-subtitle { font-size: 11px; color: var(--color-text-muted); margin-top: 0.2rem; }

.engine-tag { font-size: 11px; font-weight: 700; color: var(--color-text-secondary); }

.score-stack { display: flex; flex-direction: column; }
.score-main { font-size: 13px; font-weight: 900; color: var(--color-text-primary); }
.score-sub { font-size: 10px; font-weight: 600; color: var(--color-text-muted); }

.rules-summary { font-size: 11px; color: var(--color-text-secondary); max-width: 12rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.validation-block { display: flex; flex-direction: column; gap: 0.25rem; }
.validation-pill { font-size: 9px; font-weight: 800; padding: 0.15rem 0.5rem; border-radius: 99px; width: fit-content; }
.validation-pill.passed { background: var(--color-success); color: white; }
.validation-pill.failed { background: var(--color-danger); color: white; }
.validation-pill.pending, .validation-pill.stale { background: var(--color-warning); color: white; }
.validation-time { font-size: 10px; color: var(--color-text-muted); }

.action-btn { width: 2.25rem; height: 2.25rem; border-radius: 0.75rem; border: 1px solid var(--color-border-default); display: flex; align-items: center; justify-content: center; color: var(--color-text-secondary); cursor: pointer; transition: all 0.2s ease; background: var(--color-bg-surface); }
.action-btn:hover { background: var(--color-bg-elevated); color: var(--color-primary); border-color: var(--color-primary); }

.ops-btn { display: inline-flex; align-items: center; gap: 0.5rem; height: 2.5rem; padding: 0 1.25rem; border-radius: 0.85rem; font-size: 13px; font-weight: 700; cursor: pointer; transition: all 0.2s ease; }
.ops-btn--neutral { background: var(--color-bg-surface); border: 1px solid var(--color-border-default); color: var(--color-text-secondary); }
.ops-btn--primary { background: var(--color-primary); color: white; border: none; }
</style>
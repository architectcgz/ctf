<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { Edit } from 'lucide-vue-next'

import type { AdminContestChallengeViewData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAwdCheckResultPresentation } from '@/features/awd-inspector'

const props = withDefaults(
  defineProps<{
    challengeLinks: AdminContestChallengeViewData[]
  }>(),
  {}
)

const emit = defineEmits<{
  edit: [challenge: AdminContestChallengeViewData]
}>()

const sortedChallengeLinks = computed(() =>
  [...props.challengeLinks].sort(
    (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
  )
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
    key: 'standard-checker',
    label: '标准 Checker',
    value: String(
      sortedChallengeLinks.value.filter(
        (item) =>
          item.awd_checker_type === 'http_standard' || item.awd_checker_type === 'tcp_standard'
      ).length
    ),
    hint: '已切到 HTTP / TCP 标准 Checker 的题目数',
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
    case 'tcp_standard':
      return 'TCP 标准 Checker'
    case 'script_checker':
      return '脚本 Checker'
    default:
      return '未配置'
  }
}

function getConfigSummary(item: AdminContestChallengeViewData): string {
  const config = item.awd_checker_config || {}
  if (item.awd_checker_type === 'tcp_standard') {
    const steps = Array.isArray(config.steps) ? config.steps.length : 0
    const timeout = typeof config.timeout_ms === 'number' ? `${config.timeout_ms}ms` : ''
    return [`TCP ${steps} steps`, timeout].filter(Boolean).join(' · ') || '未配置 TCP 步骤'
  }
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

function getChallengePreviewRoute(item: AdminContestChallengeViewData) {
  return {
    name: 'PlatformChallengeDetail',
    params: { id: item.challenge_id },
  }
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

</script>

<template>
  <div class="studio-awd-config">
    <header class="studio-pane-header">
      <div class="header-main">
        <div class="workspace-overline">
          AWD Service Config
        </div>
        <h1 class="pane-title">
          AWD 编排
        </h1>
        <p class="pane-description">
          维护 Checker、SLA / 防守权重、试跑结果和就绪状态。
        </p>
      </div>

      <div class="progress-strip metric-panel-grid metric-panel-default-surface">
        <article
          v-for="item in summaryItems"
          :key="item.key"
          class="journal-note progress-card metric-panel-card"
        >
          <div class="journal-note-label progress-card-label metric-panel-label">
            {{ item.label }}
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ item.value }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            {{ item.hint }}
          </div>
        </article>
      </div>
    </header>

    <section class="workspace-directory-section awd-config-directory">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Challenge Directory
          </div>
          <h3 class="list-heading__title">
            题目目录
          </h3>
        </div>
      </header>

      <AppEmpty
        v-if="sortedChallengeLinks.length === 0"
        title="暂无关联服务"
        description="请先在题目编排中关联题目。"
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
                SLA / 防守权重
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
            >
              <td class="col-identity">
                <div class="challenge-identity">
                  <RouterLink
                    :id="`awd-challenge-preview-${item.challenge_id}`"
                    class="challenge-title challenge-title-link"
                    :to="getChallengePreviewRoute(item)"
                    :title="`打开题目预览：${getChallengeTitle(item)}`"
                  >
                    {{ getChallengeTitle(item) }}
                  </RouterLink>
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
                  <span class="score-main">SLA {{ item.awd_sla_score }}</span>
                  <span class="score-sub">Defense {{ item.awd_defense_score }}</span>
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
                <div class="ui-row-actions config-row__actions">
                  <RouterLink
                    class="ui-btn ui-btn--secondary"
                    :to="getChallengePreviewRoute(item)"
                  >
                    预览
                  </RouterLink>
                  <button
                    :id="`awd-challenge-config-edit-${item.challenge_id}`"
                    class="ui-btn ui-btn--primary"
                    @click="emit('edit', item)"
                  >
                    <Edit class="h-3.5 w-3.5" />
                    编辑
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<style scoped>
.studio-awd-config {
  display: flex;
  flex-direction: column;
  gap: var(--space-section-gap);
  background: transparent;
  padding: var(--space-6) var(--space-8);
}

.studio-pane-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-section-gap-compact);
}

.header-main {
  display: flex;
  min-width: 0;
  flex-direction: column;
  max-width: var(--ui-selector-width-lg);
}

.studio-pane-header > .progress-strip {
  --metric-panel-columns: repeat(
    auto-fit,
    minmax(min(100%, var(--ui-selector-control-min-width)), 1fr)
  );
}

.pane-title {
  margin: 0;
  font-size: var(--font-size-20);
  font-weight: 900;
  color: var(--color-text-primary);
}

.pane-description {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
}

.awd-config-directory {
  --workspace-directory-section-gap: var(--space-section-gap-compact);
  --workspace-directory-heading-gap: var(--space-4);
  --workspace-directory-title-margin-top: var(--space-1-5);
}

.studio-table-wrap {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--workspace-line-soft) 86%, transparent);
  border-radius: var(--ui-control-radius-lg);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base))
    );
  box-shadow: 0 var(--space-2) var(--space-5)
    color-mix(in srgb, var(--color-shadow-soft) 24%, transparent);
}

.studio-table {
  width: 100%;
  border-collapse: collapse;
}

.studio-table th {
  border-bottom: 1px solid color-mix(in srgb, var(--workspace-line-soft) 86%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  padding: var(--space-4);
  text-align: left;
  font-size: var(--font-size-11);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.studio-table td {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: var(--space-5) var(--space-4);
}

.studio-table tbody tr:last-child td {
  border-bottom: 0;
}

.studio-row {
  transition: background var(--ui-motion-fast);
}

.challenge-title {
  display: inline-block;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--color-text-primary);
}

.challenge-title-link {
  text-decoration: none;
  transition:
    color var(--ui-motion-fast),
    text-decoration-color var(--ui-motion-fast);
}

.challenge-title-link:hover {
  color: var(--color-primary);
  text-decoration: underline;
  text-decoration-thickness: var(--ui-focus-ring-width);
  text-underline-offset: var(--space-1);
}

.challenge-title-link:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 72%, transparent);
  outline-offset: var(--space-1);
  border-radius: var(--ui-control-radius-sm);
}

.challenge-subtitle {
  margin-top: var(--space-1);
  font-size: var(--font-size-13);
  color: var(--color-text-muted);
}

.engine-tag {
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-text-secondary);
}

.score-stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.score-main {
  font-size: var(--font-size-15);
  font-weight: 900;
  color: var(--color-text-primary);
}

.score-sub {
  font-size: var(--font-size-12);
  font-weight: 600;
  color: var(--color-text-muted);
}

.rules-summary {
  overflow: hidden;
  max-width: var(--ui-selector-control-min-width);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}

.validation-block {
  display: flex;
  flex-direction: column;
  gap: var(--space-1-5);
}

.validation-pill {
  width: fit-content;
  border-radius: var(--ui-badge-radius-pill);
  padding: var(--space-1) var(--space-2-5);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.validation-pill.passed {
  background: var(--color-success);
  color: var(--color-bg-base);
}

.validation-pill.failed {
  background: var(--color-danger);
  color: var(--color-bg-base);
}

.validation-pill.pending,
.validation-pill.stale {
  background: var(--color-warning);
  color: var(--color-bg-base);
}

.validation-time {
  font-size: var(--font-size-12);
  color: var(--color-text-muted);
}

.col-actions {
  text-align: right;
}

.config-row__actions {
  justify-content: flex-end;
}
</style>

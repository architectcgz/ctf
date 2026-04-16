<script setup lang="ts">
import { computed } from 'vue'

import type { AdminContestChallengeData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAwdCheckResultPresentation } from '@/composables/useAwdCheckResultPresentation'

const props = withDefaults(
  defineProps<{
    challengeLinks: AdminContestChallengeData[]
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
  edit: [challenge: AdminContestChallengeData]
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
const activeChallengeHeading = computed(() =>
  activeChallenge.value ? getChallengeTitle(activeChallenge.value) : ''
)
const activeChallengeContext = computed(() => {
  if (!activeChallenge.value) {
    return ''
  }
  if (props.focusSource === 'preflight') {
    return '这是赛前检查中仍需处理的题目，修正后可以回到赛前检查继续复核。'
  }
  if (props.focusSource === 'pool') {
    return '这是你刚从题目池带回来的当前题，可以继续补齐 checker、分值和验证状态。'
  }
  return '当前题目会保持高亮，便于连续逐题整理配置。'
})

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

function getConfigSummary(item: AdminContestChallengeData): string {
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

function getChallengeTitle(item: AdminContestChallengeData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function buildPresentationResult(item: AdminContestChallengeData): Record<string, unknown> {
  const preview = item.awd_checker_last_preview_result
  if (!preview) {
    return {}
  }
  return {
    ...preview.check_result,
    preview_context: preview.preview_context,
  }
}

function getValidationStateText(item: AdminContestChallengeData): string {
  return getValidationStateLabel(item.awd_checker_validation_state) || '未验证'
}

function getValidationStateClass(item: AdminContestChallengeData): string {
  const state = item.awd_checker_validation_state || 'pending'
  return `config-validation-chip config-validation-chip--${state}`
}

function getValidationHint(item: AdminContestChallengeData): string {
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

function isActiveChallenge(item: AdminContestChallengeData): boolean {
  return item.challenge_id === props.activeChallengeId
}
</script>

<template>
  <section class="space-y-6">
    <header class="panel-head panel-head--config">
      <div class="panel-copy workspace-tab-heading__main">
        <div class="journal-eyebrow">AWD Service Config</div>
        <h2 class="workspace-tab-heading__title">题目配置</h2>
        <p class="admin-page-copy">
          在这里管理赛事题目的 Checker 类型、动作配置和每轮 SLA / 防守分。
        </p>
      </div>

      <div class="progress-strip metric-panel-grid metric-panel-default-surface">
        <article
          v-for="item in summaryItems"
          :key="item.key"
          class="journal-note progress-card metric-panel-card"
        >
          <div class="journal-note-label progress-card-label metric-panel-label">{{ item.label }}</div>
          <div class="journal-note-value progress-card-value metric-panel-value">{{ item.value }}</div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">{{ item.hint }}</div>
        </article>
      </div>
    </header>

    <section class="workspace-directory-section">
      <header class="list-heading config-list-head">
        <div>
          <div class="journal-note-label">Contest Challenges</div>
          <h3 class="list-heading__title">题目目录</h3>
        </div>
        <div class="config-list-actions">
          <div class="config-list-meta">共 {{ sortedChallengeLinks.length }} 道题目</div>
          <button
            id="awd-challenge-config-create"
            type="button"
            class="ui-btn ui-btn--primary"
            @click="emit('create')"
          >
            新增题目
          </button>
        </div>
      </header>

      <section
        v-if="activeChallenge"
        class="workspace-directory-section config-focus-card"
      >
        <header class="list-heading config-focus-card__head">
          <div>
            <div class="journal-note-label">Current Focus</div>
            <h3 class="list-heading__title">当前聚焦题目</h3>
            <p class="config-focus-card__title">{{ activeChallengeHeading }}</p>
            <p class="config-focus-card__copy">{{ activeChallengeContext }}</p>
          </div>
        </header>
        <div class="ui-row-actions config-focus-card__actions">
          <button
            id="awd-challenge-config-prev"
            type="button"
            class="ui-btn ui-btn--secondary"
            :disabled="!canNavigatePrevious"
            @click="emit('previous')"
          >
            上一题
          </button>
          <button
            id="awd-challenge-config-next"
            type="button"
            class="ui-btn ui-btn--secondary"
            :disabled="!canNavigateNext"
            @click="emit('next')"
          >
            下一题
          </button>
        </div>
      </section>

      <AppEmpty
        v-if="sortedChallengeLinks.length === 0"
        title="当前赛事还没有关联题目"
        description="先关联赛事题目，再为每道服务配置 checker 与分值。"
        icon="FileChartColumnIncreasing"
      />

      <template v-else>
        <div class="config-directory-head" aria-hidden="true">
          <span>题目</span>
          <span>可见性</span>
          <span>分值</span>
          <span>Checker</span>
          <span>规则摘要</span>
          <span class="config-directory-head__actions">操作</span>
        </div>

        <article
          v-for="item in sortedChallengeLinks"
          :key="item.id"
          class="config-row"
          :class="{ 'config-row--active': isActiveChallenge(item) }"
        >
          <div class="config-row__identity">
            <h4 class="config-row__title">{{ getChallengeTitle(item) }}</h4>
            <p class="config-row__meta">
              {{ item.category || '未分类' }} · {{ item.difficulty || '未标记难度' }} · 顺序
              {{ item.order }}
            </p>
          </div>
          <div class="config-row__visibility">
            {{ item.is_visible ? '可见' : '隐藏' }}
          </div>
          <div class="config-row__scores">
            <p>{{ item.points }} 分</p>
            <p class="config-row__scores-sub">
              SLA {{ item.awd_sla_score ?? 0 }} / 防守 {{ item.awd_defense_score ?? 0 }}
            </p>
          </div>
          <div class="config-row__checker">
            <div class="config-row__checker-main">
              {{ getCheckerTypeLabel(item.awd_checker_type) }}
            </div>
            <span :class="getValidationStateClass(item)">
              {{ getValidationStateText(item) }}
            </span>
          </div>
          <div class="config-row__summary">
            <p class="config-row__summary-main">{{ getConfigSummary(item) }}</p>
            <p class="config-row__summary-sub">{{ getValidationHint(item) }}</p>
          </div>
          <div
            class="ui-row-actions config-row__actions"
            role="group"
            :aria-label="`题目 ${getChallengeTitle(item)} 操作`"
          >
            <button
              :id="`awd-challenge-config-edit-${item.id}`"
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="emit('edit', item)"
            >
              编辑配置
            </button>
          </div>
        </article>
      </template>
    </section>
  </section>
</template>

<style scoped>
.panel-head--config {
  display: grid;
  gap: 1.5rem;
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.config-list-head {
  margin-bottom: 1.25rem;
}

.config-focus-card {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.25rem;
  padding: 1.25rem 1.35rem;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
}

.config-focus-card__head {
  flex: 1 1 22rem;
  min-width: min(100%, 22rem);
}

.config-focus-card__title,
.config-focus-card__copy {
  margin: 0.4rem 0 0;
}

.config-focus-card__title {
  color: var(--journal-ink);
  font-size: 1rem;
  font-weight: 600;
}

.config-focus-card__copy {
  max-width: 44rem;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.config-focus-card__actions {
  gap: 0.75rem;
}

.config-list-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.85rem;
}

.config-list-meta {
  color: var(--journal-muted);
  font-size: 0.82rem;
}

.config-directory-head,
.config-row {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) 0.7fr 0.9fr 0.9fr minmax(0, 1.2fr) auto;
  gap: 1rem;
  align-items: center;
}

.config-directory-head {
  padding: 0 0 0.9rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 85%, transparent);
  color: var(--journal-muted);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.config-directory-head__actions {
  text-align: right;
}

.config-row {
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
}

.config-row--active {
  padding-inline: 0.85rem;
  margin-inline: -0.85rem;
  border-radius: 1.15rem;
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
}

.config-row:last-child {
  border-bottom: none;
}

.config-row__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.98rem;
  font-weight: 600;
}

.config-row__meta,
.config-row__scores-sub {
  margin: 0.35rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.config-row__visibility,
.config-row__scores,
.config-row__summary {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.config-row__scores p {
  margin: 0;
}

.config-row__checker,
.config-row__summary {
  display: grid;
  gap: 0.45rem;
}

.config-row__checker-main {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.config-row__summary {
  line-height: 1.6;
}

.config-row__summary-main,
.config-row__summary-sub {
  margin: 0;
}

.config-row__summary-sub {
  color: var(--color-text-muted);
  font-size: 0.78rem;
}

.config-validation-chip {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  justify-content: center;
  min-height: 1.9rem;
  padding: 0.2rem 0.75rem;
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: 0.78rem;
  font-weight: 600;
}

.config-validation-chip--pending {
  border-color: color-mix(in srgb, var(--journal-border) 82%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-surface-elevated));
  color: var(--color-text-secondary);
}

.config-validation-chip--passed {
  border-color: color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.config-validation-chip--failed {
  border-color: color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: var(--color-danger);
}

.config-validation-chip--stale {
  border-color: color-mix(in srgb, var(--color-warning) 28%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: color-mix(in srgb, var(--color-warning) 82%, var(--color-text-primary));
}

.config-row__actions {
  justify-content: flex-end;
}

@media (max-width: 1100px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .config-focus-card {
    align-items: stretch;
  }

  .config-directory-head {
    display: none;
  }

  .config-row {
    grid-template-columns: 1fr;
    gap: 0.6rem;
    padding: 1rem 0;
  }

  .config-row__actions {
    justify-content: flex-start;
  }
}
</style>

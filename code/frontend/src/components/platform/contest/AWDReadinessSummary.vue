<script setup lang="ts">
import { computed } from 'vue'

import type { AWDReadinessData, AWDReadinessItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { 
  ShieldCheck, 
  AlertCircle 
} from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    readiness: AWDReadinessData | null
    loading: boolean
    actionLabel?: string
  }>(),
  {
    actionLabel: '编辑配置',
  }
)

const emit = defineEmits<{
  editConfig: [challengeId: string]
}>()

const readinessDecision = computed(() => {
  const readiness = props.readiness
  const hasGlobalBlockers = (readiness?.global_blocking_reasons?.length ?? 0) > 0
  const hasChallengeBlockers = (readiness?.blocking_count ?? 0) > 0 || (readiness?.items?.length ?? 0) > 0

  if (readiness?.ready) {
    return {
      label: '可开赛',
      helper: '题目侧 checker 与系统级门禁均已满足开赛要求。',
      tone: 'ready',
    }
  }

  if (!hasGlobalBlockers && (readiness?.total_challenges ?? 0) > 0 && hasChallengeBlockers) {
    return {
      label: '可强制开赛',
      helper: '当前没有系统级门禁，但题目侧仍有阻塞，需确认风险后强制继续。',
      tone: 'force',
    }
  }

  return {
    label: '不可开赛',
    helper: hasGlobalBlockers
      ? '系统级阻塞仍会拦截开赛关键动作。'
      : '当前仍有未完成的赛前条件，暂不建议开赛。',
    tone: 'blocked',
  }
})

const summaryItems = computed(() => {
  const readiness = props.readiness
  return [
    { key: 'passed', label: '最近通过', value: readiness?.passed_challenges ?? 0 },
    { key: 'pending', label: '未验证', value: readiness?.pending_challenges ?? 0 },
    { key: 'failed', label: '最近失败', value: readiness?.failed_challenges ?? 0 },
    { key: 'stale', label: '待重新验证', value: readiness?.stale_challenges ?? 0 },
    { key: 'missing', label: '未配 Checker', value: readiness?.missing_checker_challenges ?? 0 },
  ]
})

const hasGlobalBlockingReasons = computed(() => (props.readiness?.global_blocking_reasons?.length ?? 0) > 0)
const globalBlockingReasons = computed(() => props.readiness?.global_blocking_reasons || [])

const blockingItems = computed(() => props.readiness?.items || [])
const blockingEmptyState = computed(() => {
  const readiness = props.readiness

  if ((readiness?.total_challenges ?? 0) === 0 || hasGlobalBlockingReasons.value) {
    return {
      title: '题目侧暂无可审计阻塞',
      description: '系统级阻塞仍会拦截开赛关键动作。',
    }
  }

  if (readiness?.ready) {
    return {
      title: '题目校验通过',
      description: '题目侧的 checker 校验已经满足开赛关键动作要求。',
    }
  }

  return {
    title: '题目级别暂无直接阻塞',
    description: '题目级别暂无直接阻塞，请检查系统级配置。',
  }
})

function getGlobalReasonCopy(reason: string): string {
  switch (reason) {
    case 'no_challenges': return '当前赛事还没有关联题目，无法执行开赛关键动作。'
    case 'missing_teams': return '竞赛中尚未加入任何参赛队伍。'
    case 'missing_challenges': return '题目池为空，至少需要关联一道题目。'
    case 'invalid_schedule': return '赛程时间设置有误或尚未开始。'
    default: return reason
  }
}

function getValidationStateLabel(item: AWDReadinessItemData): string {
  switch (item.validation_state) {
    case 'passed': return '最近通过'
    case 'failed': return '最近失败'
    case 'stale': return '待重新验证'
    default: return '未验证'
  }
}

function getValidationStateBadgeClass(value: AWDReadinessItemData['validation_state']): string {
  switch (value) {
    case 'passed': return 'readiness-status-chip--passed'
    case 'failed': return 'readiness-status-chip--failed'
    case 'pending':
    case 'stale':
      return 'readiness-status-chip--warning'
    default:
      return 'readiness-status-chip--neutral'
  }
}

function getBlockingReasonLabel(item: AWDReadinessItemData): string {
  switch (item.blocking_reason) {
    case 'missing_checker': return '未配置 Checker'
    case 'invalid_checker_config': return 'Checker 配置不可用'
    case 'pending_validation': return '还没有试跑结果'
    case 'last_preview_failed': return '最近一次试跑失败'
    case 'validation_stale': return '配置变更后待重新验证'
    default: return item.blocking_reason
  }
}

function formatDateTime(value?: string): string {
  if (!value) return '未记录'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit',
  })
}
</script>

<template>
  <div class="studio-readiness-flow">
    <header class="list-heading readiness-decision__head">
      <div>
        <div class="workspace-overline">AWD Readiness</div>
        <h2 class="list-heading__title">开赛就绪摘要</h2>
      </div>
    </header>

    <section
      v-if="readiness"
      class="workspace-directory-section readiness-decision-card"
      :class="`readiness-decision-card--${readinessDecision.tone}`"
    >
      <div class="journal-note-label readiness-decision-card__label">就绪决策</div>
      <div class="readiness-decision-card__value">{{ readinessDecision.label }}</div>
      <p class="readiness-decision-card__helper">{{ readinessDecision.helper }}</p>
    </section>

    <div
      v-if="readiness"
      class="progress-strip metric-panel-grid metric-panel-default-surface readiness-summary-grid"
    >
      <article
        v-for="item in summaryItems"
        :key="item.key"
        class="journal-note progress-card metric-panel-card"
      >
        <div class="journal-note-label progress-card-label metric-panel-label">{{ item.label }}</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ item.value }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">题目就绪统计概览</div>
      </article>
    </div>

    <!-- 2. Global Blockers -->
    <section
      v-if="hasGlobalBlockingReasons"
      class="global-blockers"
    >
      <header class="list-heading">
        <div>
          <div class="journal-note-label">Global Blocking</div>
          <h3 class="list-heading__title">系统级阻塞</h3>
        </div>
      </header>
      <div class="blocker-list">
        <div
          v-for="reason in globalBlockingReasons"
          :key="reason"
          class="blocker-item"
        >
          <AlertCircle class="readiness-blocker-icon" />
          <span>{{ getGlobalReasonCopy(reason) }}</span>
        </div>
      </div>
    </section>

    <!-- 3. Challenge Blockers Directory -->
    <section class="challenge-blockers">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">Blocking Shortlist</div>
          <h3 class="list-heading__title">阻塞短名单</h3>
        </div>
        <div class="directory-meta">
          发现 {{ readiness?.blocking_count ?? 0 }} 个阻塞点
        </div>
      </header>

      <AppEmpty
        v-if="blockingItems.length === 0"
        :title="blockingEmptyState.title"
        :description="blockingEmptyState.description"
        icon="ShieldCheck"
        class="py-12"
      />

      <div
        v-else
        class="studio-table-wrap"
      >
        <table class="studio-table">
          <thead>
            <tr>
              <th class="col-identity">
                题目资源
              </th>
              <th class="col-status">
                当前状态
              </th>
              <th class="col-reason">
                阻塞原因
              </th>
              <th class="col-meta">
                最近校验
              </th>
              <th class="col-actions">
                操作
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in blockingItems"
              :key="item.challenge_id"
              class="studio-row"
            >
              <td class="col-identity">
                <div class="challenge-identity">
                  <div class="challenge-title">
                    {{ item.title }}
                  </div>
                  <div class="challenge-subtitle">
                    {{ item.checker_type === 'http_standard' ? 'HTTP Standard' : '基础探活' }}
                  </div>
                </div>
              </td>
              <td class="col-status">
                <span
                  class="ui-badge readiness-status-chip ui-badge--pill ui-badge--soft"
                  :class="getValidationStateBadgeClass(item.validation_state)"
                >{{ getValidationStateLabel(item) }}</span>
              </td>
              <td class="col-reason">
                <div class="reason-text">
                  {{ getBlockingReasonLabel(item) }}
                </div>
              </td>
              <td class="col-meta readiness-meta">
                {{ formatDateTime(item.last_preview_at) }}
              </td>
              <td class="col-actions">
                <div class="ui-row-actions readiness-row__actions">
                  <button
                  :id="`awd-readiness-edit-${item.challenge_id}`"
                  class="ui-btn ui-btn--sm ui-btn--secondary"
                  @click="emit('editConfig', item.challenge_id)"
                >
                  {{ props.actionLabel }}
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
.studio-readiness-flow { display: flex; flex-direction: column; gap: 2rem; }

.readiness-decision-card {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-6);
}

.readiness-decision-card--ready {
  border-color: color-mix(in srgb, var(--color-success) 22%, var(--journal-border));
}

.readiness-decision-card--force {
  border-color: color-mix(in srgb, var(--color-warning) 24%, var(--journal-border));
}

.readiness-decision-card--blocked {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--journal-border));
}

.readiness-decision-card__label {
  color: var(--journal-muted);
}

.readiness-decision-card__value {
  font-size: var(--font-size-1-45);
  font-weight: 900;
  color: var(--journal-ink);
}

.readiness-decision-card__helper {
  margin: 0;
  color: var(--journal-muted);
  line-height: 1.7;
}

/* Metric Band */
.studio-metric-band { display: flex; gap: 0.5rem; background: var(--color-bg-elevated); padding: 1rem; border-radius: 1rem; border: 1px solid var(--color-border-default); }
.metric-pill { background: var(--color-bg-surface); border: 1px solid var(--color-border-default); padding: 0.45rem 1rem; border-radius: 0.75rem; display: flex; align-items: baseline; gap: 0.75rem; }
.metric-pill__label { font-size: 8px; font-weight: 800; text-transform: uppercase; color: var(--color-text-secondary); letter-spacing: 0.05em; }
.metric-pill__value { font-size: 13px; font-weight: 900; color: var(--color-text-primary); font-family: var(--font-family-mono); }

/* Global Blockers */
.global-blockers {
  background: var(--color-bg-elevated);
  border-radius: 1rem;
  padding: 1.5rem;
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, var(--color-border-default));
}
.section-title { font-size: var(--font-size-13); font-weight: 900; color: var(--color-danger); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 1rem; }
.blocker-list { display: flex; flex-direction: column; gap: 0.75rem; }
.blocker-item { display: flex; align-items: center; gap: 0.75rem; font-size: var(--font-size-13); font-weight: 700; color: var(--color-danger); }
.readiness-blocker-icon { width: 1rem; height: 1rem; flex: none; color: var(--color-danger); }

/* Directory */
.directory-header { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1.25rem; }
.directory-title { font-size: var(--font-size-13); font-weight: 900; color: var(--color-text-primary); text-transform: uppercase; letter-spacing: 0.1em; }
.directory-meta { font-size: var(--font-size-11); font-weight: 600; color: var(--color-text-muted); }

/* Table Styles */
.studio-table-wrap { border: 1px solid var(--color-border-default); border-radius: 1rem; background: var(--color-bg-surface); overflow: hidden; }
.studio-table { width: 100%; border-collapse: collapse; }
.studio-table th { background: var(--color-bg-elevated); padding: 0.75rem 1rem; text-align: left; font-size: var(--font-size-11); font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); border-bottom: 1px solid var(--color-border-default); }
.studio-table td { padding: 1.15rem 1rem; border-bottom: 1px solid var(--color-border-subtle); }

.challenge-title { font-size: var(--font-size-14); font-weight: 800; color: var(--color-text-primary); }
.challenge-subtitle { font-size: var(--font-size-12); color: var(--color-text-muted); margin-top: 0.25rem; }

.readiness-status-chip {
  --ui-badge-padding: 0.2rem 0.6rem;
  --ui-badge-size: var(--font-size-11);
  font-weight: 800;
}
.readiness-status-chip--passed { --ui-badge-tone: var(--color-success); }
.readiness-status-chip--failed { --ui-badge-tone: var(--color-danger); }
.readiness-status-chip--warning { --ui-badge-tone: var(--color-warning); }
.readiness-status-chip--neutral { --ui-badge-tone: var(--color-text-muted); }

.readiness-meta {
  font-size: var(--font-size-11);
  color: var(--color-text-muted);
}

.reason-text { font-size: var(--font-size-13); font-weight: 700; color: var(--color-text-secondary); }

.action-btn { font-size: var(--font-size-11); font-weight: 800; color: var(--color-primary); background: var(--color-primary-soft); padding: 0.35rem 0.75rem; border-radius: 0.5rem; cursor: pointer; transition: all 0.2s ease; border: none; }
.action-btn:hover { background: color-mix(in srgb, var(--color-primary) 20%, var(--color-bg-surface)); }
</style>

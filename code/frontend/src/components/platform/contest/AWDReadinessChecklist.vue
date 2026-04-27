<script setup lang="ts">
import { computed } from 'vue'
import { AlertCircle, ShieldCheck } from 'lucide-vue-next'

import type { AWDReadinessData, AWDReadinessItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = withDefaults(
  defineProps<{
    readiness: AWDReadinessData | null
    actionLabel?: string
    hideActions?: boolean
  }>(),
  {
    actionLabel: '编辑配置',
    hideActions: false,
  }
)

const emit = defineEmits<{
  editConfig: [challengeId: string]
}>()

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
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="readiness-checklist">
    <div
      v-if="readiness"
      class="progress-strip metric-panel-grid metric-panel-default-surface readiness-summary-grid"
    >
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
          题目就绪统计概览
        </div>
      </article>
    </div>

    <section
      v-if="hasGlobalBlockingReasons"
      class="global-blockers"
    >
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Global Blocking
          </div>
          <h3 class="list-heading__title">
            系统级阻塞
          </h3>
        </div>
      </header>
      <div class="blocker-list">
        <div
          v-for="reason in globalBlockingReasons"
          :key="reason"
          class="blocker-item"
        >
          <AlertCircle class="blocker-item__icon h-4 w-4" />
          <span>{{ getGlobalReasonCopy(reason) }}</span>
        </div>
      </div>
    </section>

    <section class="challenge-blockers">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Blocking Shortlist
          </div>
          <h3 class="list-heading__title">
            阻塞短名单
          </h3>
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
              <th
                v-if="!hideActions"
                class="col-actions"
              >
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
                  class="ui-badge readiness-status-chip"
                  :class="item.validation_state"
                >
                  {{ getValidationStateLabel(item) }}
                </span>
              </td>
              <td class="col-reason">
                <div class="reason-text">
                  {{ getBlockingReasonLabel(item) }}
                </div>
              </td>
              <td class="col-meta readiness-meta-cell">
                {{ formatDateTime(item.last_preview_at) }}
              </td>
              <td
                v-if="!hideActions"
                class="col-actions"
              >
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
.readiness-checklist {
  display: flex;
  flex-direction: column;
  gap: var(--space-section-gap);
}

.global-blockers {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, var(--color-border-default));
  border-radius: var(--ui-control-radius-lg);
  background: var(--color-bg-elevated);
  padding: var(--space-6);
}

.challenge-blockers {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.blocker-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.blocker-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-danger);
}

.blocker-item__icon {
  color: var(--color-danger);
}

.directory-meta {
  font-size: var(--font-size-11);
  font-weight: 600;
  color: var(--color-text-muted);
}

.studio-table-wrap {
  overflow: hidden;
  border: 1px solid var(--color-border-default);
  border-radius: var(--ui-control-radius-lg);
  background: var(--color-bg-surface);
}

.studio-table {
  width: 100%;
  border-collapse: collapse;
}

.studio-table th {
  border-bottom: 1px solid var(--color-border-default);
  background: var(--color-bg-elevated);
  padding: var(--space-3) var(--space-4);
  text-align: left;
  font-size: var(--font-size-11);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.studio-table td {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: var(--space-4) var(--space-4);
}

.studio-table tbody tr:last-child td {
  border-bottom: 0;
}

.challenge-title {
  font-size: var(--font-size-14);
  font-weight: 800;
  color: var(--color-text-primary);
}

.challenge-subtitle {
  margin-top: var(--space-1);
  font-size: var(--font-size-12);
  color: var(--color-text-muted);
}

.readiness-meta-cell {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
}

.reason-text {
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-text-secondary);
}
</style>

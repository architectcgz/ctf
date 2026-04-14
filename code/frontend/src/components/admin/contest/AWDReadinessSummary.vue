<script setup lang="ts">
import { computed } from 'vue'

import type { AWDReadinessData, AWDReadinessItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  readiness: AWDReadinessData | null
  loading: boolean
}>()

const emit = defineEmits<{
  editConfig: [challengeId: string]
}>()

const summaryItems = computed(() => {
  const readiness = props.readiness
  return [
    {
      key: 'passed',
      label: '最近通过',
      value: String(readiness?.passed_challenges ?? 0),
      hint: '最近一次试跑已通过的题目数',
    },
    {
      key: 'pending',
      label: '未验证',
      value: String(readiness?.pending_challenges ?? 0),
      hint: '还没有可用试跑结果的题目数',
    },
    {
      key: 'failed',
      label: '最近失败',
      value: String(readiness?.failed_challenges ?? 0),
      hint: '最近一次试跑未通过的题目数',
    },
    {
      key: 'stale',
      label: '待重新验证',
      value: String(readiness?.stale_challenges ?? 0),
      hint: '配置变更后尚未重新试跑的题目数',
    },
    {
      key: 'missing',
      label: '未配 Checker',
      value: String(readiness?.missing_checker_challenges ?? 0),
      hint: '还没有可执行 checker 的题目数',
    },
  ]
})

const blockingItems = computed(() => props.readiness?.items || [])
const globalBlockingReasons = computed(() => props.readiness?.global_blocking_reasons ?? [])
const hasGlobalBlockingReasons = computed(() => (props.readiness?.global_blocking_reasons?.length ?? 0) > 0)
const blockingActionLabels = computed(() =>
  (props.readiness?.blocking_actions || []).map((action) => getBlockingActionLabel(action))
)
const blockingEmptyDescription = computed(() =>
  hasGlobalBlockingReasons.value
    ? '当前没有题目级阻塞项，系统级阻塞仍会拦截开赛关键动作。'
    : '题目侧的 checker 校验已经满足开赛关键动作要求。'
)

function getBlockingActionLabel(action: string): string {
  switch (action) {
    case 'create_round':
      return '创建轮次'
    case 'run_current_round_check':
      return '立即巡检当前轮'
    case 'start_contest':
      return '启动赛事'
    default:
      return action
  }
}

function getGlobalReasonCopy(reason: string): string {
  switch (reason) {
    case 'no_challenges':
      return '当前赛事还没有关联题目，无法执行开赛关键动作。'
    default:
      return reason
  }
}

function getValidationStateLabel(item: AWDReadinessItemData): string {
  switch (item.validation_state) {
    case 'passed':
      return '最近通过'
    case 'failed':
      return '最近失败'
    case 'stale':
      return '待重新验证'
    case 'pending':
    default:
      return '未验证'
  }
}

function getBlockingReasonLabel(item: AWDReadinessItemData): string {
  switch (item.blocking_reason) {
    case 'missing_checker':
      return '未配置 Checker'
    case 'invalid_checker_config':
      return 'Checker 配置不可用'
    case 'pending_validation':
      return '还没有试跑结果'
    case 'last_preview_failed':
      return '最近一次试跑失败'
    case 'validation_stale':
      return '配置变更后待重新验证'
    default:
      return item.blocking_reason
  }
}

function formatDateTime(value?: string): string {
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
</script>

<template>
  <section class="space-y-6">
    <header class="panel-head panel-head--readiness">
      <div class="panel-copy workspace-tab-heading__main">
        <div class="journal-eyebrow">AWD Readiness</div>
        <h2 class="workspace-tab-heading__title">开赛就绪摘要</h2>
        <p class="admin-page-copy">
          这里汇总当前赛事的 checker 校验状态，并标记会阻塞创建轮次、当前轮巡检和启动赛事的风险项。
        </p>
      </div>

      <div class="metric-panel-grid metric-panel-default-surface readiness-summary-grid">
        <article v-for="item in summaryItems" :key="item.key" class="journal-note metric-panel-card">
          <div class="journal-note-label metric-panel-label">{{ item.label }}</div>
          <div class="journal-note-value metric-panel-value">{{ item.value }}</div>
          <div class="journal-note-helper metric-panel-helper">{{ item.hint }}</div>
        </article>
      </div>
    </header>

    <section v-if="loading" class="workspace-directory-section readiness-section">
      <div class="readiness-loading">正在同步开赛就绪状态...</div>
    </section>

    <template v-else>
      <section
        v-if="hasGlobalBlockingReasons"
        class="workspace-directory-section readiness-alert"
      >
        <header class="list-heading">
          <div>
            <div class="journal-note-label">Global Blocking</div>
            <h3 class="list-heading__title">系统级阻塞</h3>
          </div>
        </header>
        <ul class="readiness-alert-list">
          <li
            v-for="reason in globalBlockingReasons"
            :key="reason"
            class="readiness-alert-item"
          >
            {{ getGlobalReasonCopy(reason) }}
          </li>
        </ul>
      </section>

      <section class="workspace-directory-section readiness-section">
        <header class="list-heading readiness-list-head">
          <div>
            <div class="journal-note-label">Blocking Items</div>
            <h3 class="list-heading__title">阻塞短名单</h3>
          </div>
          <div class="readiness-list-head__meta">
            <span class="readiness-count">阻塞 {{ readiness?.blocking_count ?? 0 }} 项</span>
            <div v-if="blockingActionLabels.length > 0" class="readiness-action-list" aria-label="阻塞动作">
              <span v-for="label in blockingActionLabels" :key="label" class="readiness-action-chip">
                {{ label }}
              </span>
            </div>
          </div>
        </header>

        <AppEmpty
          v-if="blockingItems.length === 0"
          title="当前没有题目级阻塞项"
          :description="blockingEmptyDescription"
          icon="ShieldCheck"
        />

        <template v-else>
          <div class="readiness-directory-head" aria-hidden="true">
            <span>题目</span>
            <span>当前状态</span>
            <span>阻塞原因</span>
            <span>最近校验</span>
            <span>目标地址</span>
            <span class="readiness-directory-head__actions">操作</span>
          </div>

          <article v-for="item in blockingItems" :key="item.challenge_id" class="readiness-row">
            <div class="readiness-row__identity">
              <h4 class="readiness-row__title">{{ item.title }}</h4>
              <p class="readiness-row__meta">
                {{ item.checker_type === 'http_standard' ? 'HTTP Standard' : 'Checker 未配置' }}
              </p>
            </div>
            <div class="readiness-row__status">
              <span class="readiness-status-chip">
                {{ getValidationStateLabel(item) }}
              </span>
            </div>
            <div class="readiness-row__reason">{{ getBlockingReasonLabel(item) }}</div>
            <div class="readiness-row__time">{{ formatDateTime(item.last_preview_at) }}</div>
            <div class="readiness-row__target">{{ item.last_access_url || '无目标地址' }}</div>
            <div class="readiness-row__actions" role="group" :aria-label="`题目 ${item.title} 操作`">
              <button
                :id="`awd-readiness-edit-${item.challenge_id}`"
                type="button"
                class="rounded-xl border border-border px-3 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary"
                @click="emit('editConfig', item.challenge_id)"
              >
                编辑配置
              </button>
            </div>
          </article>
        </template>
      </section>
    </template>
  </section>
</template>

<style scoped>
.panel-head--readiness {
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

.readiness-summary-grid {
  --metric-panel-columns: repeat(5, minmax(0, 1fr));
}

.readiness-section,
.readiness-alert {
  padding: 1.5rem;
}

.readiness-loading {
  color: var(--journal-muted);
  font-size: 0.95rem;
}

.readiness-alert-list {
  margin: 0;
  padding-left: 1.1rem;
  color: var(--journal-ink);
  display: grid;
  gap: 0.65rem;
}

.readiness-alert-item {
  line-height: 1.6;
}

.readiness-list-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
}

.readiness-list-head__meta {
  display: grid;
  gap: 0.65rem;
  justify-items: end;
}

.readiness-count {
  color: var(--journal-muted);
  font-size: 0.875rem;
}

.readiness-action-list {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
}

.readiness-action-chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 0.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: 0.8rem;
  font-weight: 600;
}

.readiness-directory-head,
.readiness-row {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(140px, 0.8fr) minmax(180px, 0.9fr) minmax(170px, 0.8fr) minmax(180px, 1fr) minmax(112px, 0.6fr);
  gap: 1rem;
  align-items: center;
}

.readiness-directory-head {
  padding: 1rem 0;
  color: var(--journal-muted);
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.readiness-directory-head__actions {
  text-align: right;
}

.readiness-row {
  padding: 1.1rem 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 78%, transparent);
}

.readiness-row__identity,
.readiness-row__status,
.readiness-row__reason,
.readiness-row__time,
.readiness-row__target,
.readiness-row__actions {
  min-width: 0;
}

.readiness-row__title {
  margin: 0;
  color: var(--journal-ink);
  font-size: 1rem;
  font-weight: 600;
}

.readiness-row__meta {
  margin: 0.35rem 0 0;
  color: var(--journal-muted);
  font-size: 0.82rem;
}

.readiness-row__reason,
.readiness-row__time,
.readiness-row__target {
  color: var(--journal-ink);
  font-size: 0.9rem;
  line-height: 1.5;
  word-break: break-word;
}

.readiness-row__actions {
  display: flex;
  justify-content: flex-end;
}

.readiness-status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 0.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
  font-size: 0.8rem;
  font-weight: 600;
}

@media (max-width: 1100px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .readiness-summary-grid {
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }

  .readiness-directory-head,
  .readiness-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .readiness-directory-head {
    display: none;
  }

  .readiness-row {
    gap: 0.75rem;
  }

  .readiness-row__actions {
    justify-content: flex-start;
  }

  .readiness-list-head,
  .readiness-list-head__meta {
    justify-items: start;
  }
}
</style>

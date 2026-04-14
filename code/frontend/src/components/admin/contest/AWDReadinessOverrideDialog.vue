<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { AWDReadinessData, AWDReadinessItemData } from '@/api/contracts'

const props = defineProps<{
  open: boolean
  title: string
  readiness: AWDReadinessData | null
  confirmLoading?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: [reason: string]
}>()

const reason = ref('')
const fieldError = ref('')

const summaryItems = computed(() => {
  const readiness = props.readiness
  return [
    { key: 'blocking', label: '阻塞项', value: String(readiness?.blocking_count ?? 0), hint: '当前会拦截关键动作的风险数量' },
    { key: 'failed', label: '最近失败', value: String(readiness?.failed_challenges ?? 0), hint: '最近一次试跑未通过的题目数' },
    { key: 'pending', label: '未验证', value: String(readiness?.pending_challenges ?? 0), hint: '还没有可用试跑结果的题目数' },
    { key: 'stale', label: '待重新验证', value: String(readiness?.stale_challenges ?? 0), hint: '配置变更后尚未重新试跑的题目数' },
  ]
})

watch(
  () => props.open,
  (open) => {
    if (!open) {
      reason.value = ''
      fieldError.value = ''
      return
    }
    fieldError.value = ''
  }
)

function getGlobalReasonCopy(reasonCode: string): string {
  switch (reasonCode) {
    case 'no_challenges':
      return '当前赛事还没有关联题目，无法执行开赛关键动作。'
    default:
      return reasonCode
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

function getValidationLabel(item: AWDReadinessItemData): string {
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

function closeDialog() {
  emit('update:open', false)
}

function handleSubmit() {
  const normalizedReason = reason.value.trim()
  if (!normalizedReason) {
    fieldError.value = '请填写本次放行原因'
    return
  }
  fieldError.value = ''
  emit('confirm', normalizedReason)
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="title"
    width="760px"
    @close="closeDialog"
    @update:model-value="emit('update:open', $event)"
  >
    <div class="readiness-override-dialog">
      <p class="readiness-override-lead">
        填写本次放行原因。当前操作只会跳过这一次门禁，不会改写题目的 checker 校验状态。
      </p>

      <div class="metric-panel-grid metric-panel-default-surface readiness-override-summary">
        <article v-for="item in summaryItems" :key="item.key" class="journal-note metric-panel-card">
          <div class="journal-note-label metric-panel-label">{{ item.label }}</div>
          <div class="journal-note-value metric-panel-value">{{ item.value }}</div>
          <div class="journal-note-helper metric-panel-helper">{{ item.hint }}</div>
        </article>
      </div>

      <section
        v-if="readiness?.global_blocking_reasons?.length"
        class="workspace-directory-section readiness-override-section"
      >
        <header class="list-heading">
          <div>
            <div class="journal-note-label">Global Blocking</div>
            <h3 class="list-heading__title">系统级阻塞</h3>
          </div>
        </header>
        <ul class="readiness-override-list">
          <li v-for="reasonCode in readiness.global_blocking_reasons" :key="reasonCode">
            {{ getGlobalReasonCopy(reasonCode) }}
          </li>
        </ul>
      </section>

      <section class="workspace-directory-section readiness-override-section">
        <header class="list-heading">
          <div>
            <div class="journal-note-label">Blocking Items</div>
            <h3 class="list-heading__title">阻塞题目</h3>
          </div>
        </header>

        <div v-if="readiness?.items?.length" class="readiness-override-rows">
          <article v-for="item in readiness.items" :key="item.challenge_id" class="readiness-override-row">
            <div class="readiness-override-row__title">
              <strong>{{ item.title }}</strong>
              <span>{{ getValidationLabel(item) }}</span>
            </div>
            <p class="readiness-override-row__detail">
              {{ getBlockingReasonLabel(item) }} · 最近校验 {{ formatDateTime(item.last_preview_at) }}
            </p>
            <p v-if="item.last_access_url" class="readiness-override-row__detail">
              目标地址 {{ item.last_access_url }}
            </p>
          </article>
        </div>
        <p v-else class="readiness-override-empty">当前没有题目级阻塞项。</p>
      </section>

      <section class="workspace-directory-section readiness-override-section">
        <header class="list-heading">
          <div>
            <div class="journal-note-label">Override Reason</div>
            <h3 class="list-heading__title">填写本次放行原因</h3>
          </div>
        </header>

        <label class="readiness-override-form" for="awd-readiness-override-reason">
          <span class="readiness-override-form__label">原因会附带到审计日志，用于说明本次为什么仍要继续。</span>
          <textarea
            id="awd-readiness-override-reason"
            v-model="reason"
            rows="4"
            class="readiness-override-textarea"
            placeholder="例如：赛前演练，允许临时绕过当前 checker 阻塞。"
          />
          <span v-if="fieldError" class="readiness-override-error">{{ fieldError }}</span>
        </label>
      </section>
    </div>

    <template #footer>
      <div class="readiness-override-footer">
        <button
          id="awd-readiness-override-cancel"
          type="button"
          class="rounded-xl border border-border px-4 py-2 text-sm text-[var(--color-text-primary)] transition hover:border-primary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-readiness-override-submit"
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="confirmLoading"
          @click="handleSubmit"
        >
          {{ confirmLoading ? '强制继续中...' : '强制继续' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>

<style scoped>
.readiness-override-dialog {
  display: grid;
  gap: 1rem;
}

.readiness-override-lead {
  margin: 0;
  color: var(--journal-ink);
  line-height: 1.7;
}

.readiness-override-summary {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.readiness-override-section {
  padding: 1.25rem 1.35rem;
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

.readiness-override-list {
  margin: 0;
  padding-left: 1.1rem;
  display: grid;
  gap: 0.65rem;
  color: var(--journal-ink);
}

.readiness-override-rows {
  display: grid;
  gap: 0.85rem;
}

.readiness-override-row {
  padding: 0.95rem 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 78%, transparent);
}

.readiness-override-row:first-child {
  border-top: none;
  padding-top: 0;
}

.readiness-override-row__title {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  color: var(--journal-ink);
}

.readiness-override-row__detail {
  margin: 0.35rem 0 0;
  color: var(--journal-muted);
  line-height: 1.6;
  word-break: break-word;
}

.readiness-override-empty {
  margin: 0;
  color: var(--journal-muted);
}

.readiness-override-form {
  display: grid;
  gap: 0.75rem;
}

.readiness-override-form__label {
  color: var(--journal-muted);
  line-height: 1.6;
}

.readiness-override-textarea {
  width: 100%;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, #fff 6%);
  padding: 0.95rem 1rem;
  color: var(--journal-ink);
  outline: none;
  resize: vertical;
  min-height: 132px;
}

.readiness-override-textarea:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
}

.readiness-override-error {
  color: var(--color-danger);
  font-size: 0.85rem;
}

.readiness-override-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

@media (max-width: 900px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .readiness-override-summary {
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>

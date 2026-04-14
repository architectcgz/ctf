<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type { ContestFieldLocks, ContestFormDraft } from '@/composables/useAdminContests'
import { getStatusLabel } from '@/utils/contest'

const props = withDefaults(
  defineProps<{
    mode: 'create' | 'edit'
    draft: ContestFormDraft
    saving: boolean
    statusOptions?: Array<{ label: string; value: ContestFormDraft['status'] }>
    fieldLocks: ContestFieldLocks
    showCancel?: boolean
    note?: string
  }>(),
  {
    statusOptions: () => [],
    showCancel: true,
    note: '若需下线竞赛，当前仍通过状态流转控制访问窗口。',
  }
)

const emit = defineEmits<{
  cancel: []
  save: [value: ContestFormDraft]
}>()

const localDraft = reactive<ContestFormDraft>({
  title: '',
  description: '',
  mode: 'jeopardy',
  starts_at: '',
  ends_at: '',
  status: 'draft',
})

const fieldErrors = reactive<Partial<Record<keyof ContestFormDraft, string>>>({})

const submitLabel = computed(() => {
  if (props.saving) {
    return '保存中...'
  }
  return props.mode === 'create' ? '创建竞赛' : '保存变更'
})

watch(
  () => props.draft,
  (draft) => {
    Object.assign(localDraft, draft)
    clearErrors()
  },
  { immediate: true, deep: true }
)

function clearErrors() {
  fieldErrors.title = ''
  fieldErrors.starts_at = ''
  fieldErrors.ends_at = ''
}

function validate(): boolean {
  clearErrors()

  if (!localDraft.title.trim()) {
    fieldErrors.title = '请填写竞赛标题'
  }

  if (!localDraft.starts_at) {
    fieldErrors.starts_at = '请填写开始时间'
  }

  if (!localDraft.ends_at) {
    fieldErrors.ends_at = '请填写结束时间'
  }

  if (
    localDraft.starts_at &&
    localDraft.ends_at &&
    new Date(localDraft.ends_at) <= new Date(localDraft.starts_at)
  ) {
    fieldErrors.ends_at = '结束时间必须晚于开始时间'
  }

  return !fieldErrors.title && !fieldErrors.starts_at && !fieldErrors.ends_at
}

function handleSubmit() {
  if (!validate()) {
    return
  }

  emit('save', {
    title: localDraft.title,
    description: localDraft.description,
    mode: localDraft.mode,
    starts_at: localDraft.starts_at,
    ends_at: localDraft.ends_at,
    status: localDraft.status,
  })
}
</script>

<template>
  <form class="contest-form-dialog__form" @submit.prevent="handleSubmit">
    <section class="workspace-directory-section contest-form-section">
      <header class="list-heading contest-form-section__head">
        <div>
          <div class="journal-note-label">Contest Setup</div>
          <h3 class="list-heading__title">基础信息</h3>
        </div>
      </header>

      <div class="contest-form-grid">
        <label class="contest-form-field contest-form-field--wide" for="contest-title">
          <span class="contest-form-label">竞赛标题</span>
          <input
            id="contest-title"
            v-model="localDraft.title"
            type="text"
            class="contest-form-control"
            placeholder="例如：2026 春季校园 CTF"
          />
          <span v-if="fieldErrors.title" class="contest-form-error">{{ fieldErrors.title }}</span>
        </label>

        <label class="contest-form-field" for="contest-mode">
          <span class="contest-form-label">竞赛模式</span>
          <select
            id="contest-mode"
            v-model="localDraft.mode"
            class="contest-form-control"
            :disabled="fieldLocks.mode"
          >
            <option value="jeopardy">Jeopardy</option>
            <option value="awd">AWD</option>
          </select>
          <span v-if="fieldLocks.mode" class="contest-form-hint">
            竞赛进入 draft 之后不再允许修改模式。
          </span>
        </label>

        <label v-if="mode === 'edit'" class="contest-form-field" for="contest-status">
          <span class="contest-form-label">状态</span>
          <select id="contest-status" v-model="localDraft.status" class="contest-form-control">
            <option v-for="option in statusOptions" :key="option.value" :value="option.value">
              {{ getStatusLabel(option.value) }}
            </option>
          </select>
        </label>

        <label class="contest-form-field contest-form-field--wide" for="contest-description">
          <span class="contest-form-label">竞赛描述</span>
          <textarea
            id="contest-description"
            v-model="localDraft.description"
            rows="4"
            class="contest-form-control contest-form-control--textarea"
            placeholder="描述赛制、参赛范围或报名说明。"
          />
        </label>
      </div>
    </section>

    <section class="workspace-directory-section contest-form-section">
      <header class="list-heading contest-form-section__head">
        <div>
          <div class="journal-note-label">Schedule</div>
          <h3 class="list-heading__title">赛制与时间</h3>
        </div>
      </header>

      <div class="contest-form-grid">
        <label class="contest-form-field" for="contest-starts-at">
          <span class="contest-form-label">开始时间</span>
          <input
            id="contest-starts-at"
            v-model="localDraft.starts_at"
            type="datetime-local"
            class="contest-form-control"
            :disabled="fieldLocks.starts_at"
          />
          <span v-if="fieldErrors.starts_at" class="contest-form-error">
            {{ fieldErrors.starts_at }}
          </span>
          <span v-else-if="fieldLocks.starts_at" class="contest-form-hint">
            报名中、进行中、已结束状态禁止修改开始时间。
          </span>
        </label>

        <label class="contest-form-field" for="contest-ends-at">
          <span class="contest-form-label">结束时间</span>
          <input
            id="contest-ends-at"
            v-model="localDraft.ends_at"
            type="datetime-local"
            class="contest-form-control"
            :disabled="fieldLocks.ends_at"
          />
          <span v-if="fieldErrors.ends_at" class="contest-form-error">
            {{ fieldErrors.ends_at }}
          </span>
          <span v-else-if="fieldLocks.ends_at" class="contest-form-hint">
            进行中、已结束状态禁止修改结束时间。
          </span>
        </label>
      </div>
    </section>
  </form>

  <div class="contest-form-dialog__footer">
    <p class="contest-form-dialog__note">{{ note }}</p>
    <div class="contest-form-dialog__actions">
      <button
        v-if="showCancel"
        type="button"
        class="contest-form-button contest-form-button--ghost"
        @click="emit('cancel')"
      >
        取消
      </button>
      <button
        type="button"
        class="contest-form-button contest-form-button--primary"
        :disabled="saving"
        @click="handleSubmit"
      >
        {{ submitLabel }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.contest-form-dialog__form {
  display: grid;
  gap: var(--space-4);
}

.contest-form-section {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) var(--space-5-5);
}

.contest-form-section__head {
  align-items: flex-end;
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
  color: var(--journal-ink, var(--color-text-primary));
}

.contest-form-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-form-field {
  display: grid;
  gap: var(--space-2);
}

.contest-form-field--wide {
  grid-column: 1 / -1;
}

.contest-form-label {
  font-size: var(--font-size-0-875);
  font-weight: 600;
  color: var(--journal-ink, var(--color-text-primary));
}

.contest-form-control {
  width: 100%;
  min-height: 2.875rem;
  border-radius: 1rem;
  border: 1px solid
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 78%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, transparent);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink, var(--color-text-primary));
  outline: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease,
    background-color 150ms ease;
}

.contest-form-control:focus {
  border-color: color-mix(in srgb, var(--journal-accent, var(--color-primary)) 42%, transparent);
  box-shadow: 0 0 0 3px
    color-mix(in srgb, var(--journal-accent, var(--color-primary)) 12%, transparent);
}

.contest-form-control:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.contest-form-control--textarea {
  min-height: 8rem;
  resize: vertical;
}

.contest-form-hint,
.contest-form-dialog__note {
  font-size: var(--font-size-0-78);
  line-height: 1.65;
  color: var(--journal-muted, var(--color-text-secondary));
}

.contest-form-error {
  font-size: var(--font-size-0-78);
  color: var(--color-danger);
}

.contest-form-dialog__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-form-dialog__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.contest-form-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.625rem;
  border-radius: 1rem;
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    color 150ms ease,
    opacity 150ms ease;
}

.contest-form-button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.contest-form-button--ghost {
  border: 1px solid
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 78%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, transparent);
  color: var(--journal-ink, var(--color-text-primary));
}

.contest-form-button--ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent, var(--color-primary)) 28%, transparent);
  color: var(--journal-accent, var(--color-primary));
}

.contest-form-button--primary {
  border: 1px solid transparent;
  background: var(--journal-accent, var(--color-primary));
  color: #fff;
}

.contest-form-button--primary:hover {
  background: var(--color-primary-hover, var(--journal-accent, var(--color-primary)));
}

@media (max-width: 767px) {
  .contest-form-section {
    padding: var(--space-4-5) var(--space-4);
  }

  .contest-form-grid {
    grid-template-columns: 1fr;
  }

  .contest-form-dialog__footer,
  .contest-form-dialog__actions {
    width: 100%;
  }

  .contest-form-dialog__actions {
    justify-content: stretch;
  }

  .contest-form-button {
    flex: 1 1 0;
  }
}
</style>

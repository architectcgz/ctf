<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'

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
  'update:draft': [value: ContestFormDraft]
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
const syncingFromProps = ref(false)

const submitLabel = computed(() => {
  if (props.saving) {
    return '保存中...'
  }
  return props.mode === 'create' ? '创建竞赛' : '保存变更'
})

watch(
  () => props.draft,
  (draft) => {
    syncingFromProps.value = true
    Object.assign(localDraft, draft)
    clearErrors()
    syncingFromProps.value = false
  },
  { immediate: true, deep: true }
)

watch(
  localDraft,
  (draft) => {
    if (syncingFromProps.value) {
      return
    }
    emit('update:draft', {
      title: draft.title,
      description: draft.description,
      mode: draft.mode,
      starts_at: draft.starts_at,
      ends_at: draft.ends_at,
      status: draft.status,
    })
  },
  { deep: true }
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
        <label class="ui-field contest-form-field contest-form-field--wide" for="contest-title">
          <span class="ui-field__label contest-form-label">竞赛标题</span>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.title }">
            <input
              id="contest-title"
              v-model="localDraft.title"
              type="text"
              class="ui-control"
              placeholder="例如：2026 春季校园 CTF"
            />
          </span>
          <span v-if="fieldErrors.title" class="ui-field__error contest-form-error">
            {{ fieldErrors.title }}
          </span>
        </label>

        <label class="ui-field contest-form-field" for="contest-mode">
          <span class="ui-field__label contest-form-label">竞赛模式</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-disabled': fieldLocks.mode }"
          >
            <select
              id="contest-mode"
              v-model="localDraft.mode"
              class="ui-control"
              :disabled="fieldLocks.mode"
            >
              <option value="jeopardy">Jeopardy</option>
              <option value="awd">AWD</option>
            </select>
          </span>
          <span v-if="fieldLocks.mode" class="ui-field__hint contest-form-hint">
            竞赛进入 draft 之后不再允许修改模式。
          </span>
        </label>

        <label v-if="mode === 'edit'" class="ui-field contest-form-field" for="contest-status">
          <span class="ui-field__label contest-form-label">状态</span>
          <span class="ui-control-wrap">
            <select id="contest-status" v-model="localDraft.status" class="ui-control">
              <option v-for="option in statusOptions" :key="option.value" :value="option.value">
                {{ getStatusLabel(option.value) }}
              </option>
            </select>
          </span>
        </label>

        <label class="ui-field contest-form-field contest-form-field--wide" for="contest-description">
          <span class="ui-field__label contest-form-label">竞赛描述</span>
          <span class="ui-control-wrap">
            <textarea
              id="contest-description"
              v-model="localDraft.description"
              rows="4"
              class="ui-control contest-form-control--textarea"
              placeholder="描述赛制、参赛范围或报名说明。"
            />
          </span>
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
        <label class="ui-field contest-form-field" for="contest-starts-at">
          <span class="ui-field__label contest-form-label">开始时间</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.starts_at, 'is-disabled': fieldLocks.starts_at }"
          >
            <input
              id="contest-starts-at"
              v-model="localDraft.starts_at"
              type="datetime-local"
              class="ui-control"
              :disabled="fieldLocks.starts_at"
            />
          </span>
          <span v-if="fieldErrors.starts_at" class="ui-field__error contest-form-error">
            {{ fieldErrors.starts_at }}
          </span>
          <span v-else-if="fieldLocks.starts_at" class="ui-field__hint contest-form-hint">
            报名中、进行中、已结束状态禁止修改开始时间。
          </span>
        </label>

        <label class="ui-field contest-form-field" for="contest-ends-at">
          <span class="ui-field__label contest-form-label">结束时间</span>
          <span
            class="ui-control-wrap"
            :class="{ 'is-error': !!fieldErrors.ends_at, 'is-disabled': fieldLocks.ends_at }"
          >
            <input
              id="contest-ends-at"
              v-model="localDraft.ends_at"
              type="datetime-local"
              class="ui-control"
              :disabled="fieldLocks.ends_at"
            />
          </span>
          <span v-if="fieldErrors.ends_at" class="ui-field__error contest-form-error">
            {{ fieldErrors.ends_at }}
          </span>
          <span v-else-if="fieldLocks.ends_at" class="ui-field__hint contest-form-hint">
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
        class="ui-btn ui-btn--secondary contest-form-button contest-form-button--ghost"
        @click="emit('cancel')"
      >
        取消
      </button>
      <button
        type="button"
        class="ui-btn ui-btn--primary contest-form-button contest-form-button--primary"
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
  color: var(--journal-ink, var(--color-text-primary));
}

.contest-form-field .ui-control-wrap {
  width: 100%;
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
  min-width: 7.5rem;
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

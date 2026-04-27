<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { FileText, Settings, Clock, Swords, Trophy } from 'lucide-vue-next'

import type { ContestFieldLocks, ContestFormDraft } from '@/composables/usePlatformContests'
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
    note: '',
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

watch(
  () => props.draft,
  (draft) => {
    syncingFromProps.value = true
    Object.assign(localDraft, draft)
    fieldErrors.title = ''
    fieldErrors.starts_at = ''
    fieldErrors.ends_at = ''
    syncingFromProps.value = false
  },
  { immediate: true, deep: true }
)

watch(
  localDraft,
  (draft) => {
    if (syncingFromProps.value) return
    emit('update:draft', { ...draft })
  },
  { deep: true }
)

function validate(): boolean {
  fieldErrors.title = ''
  fieldErrors.starts_at = ''
  fieldErrors.ends_at = ''

  if (!localDraft.title.trim()) fieldErrors.title = '请填写竞赛标题'
  if (!localDraft.starts_at) fieldErrors.starts_at = '请填写开始时间'
  if (!localDraft.ends_at) fieldErrors.ends_at = '请填写结束时间'

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
  if (!validate()) return
  emit('save', { ...localDraft })
}
</script>

<template>
  <form
    class="studio-settings-layout"
    @submit.prevent="handleSubmit"
  >
    <section class="settings-group">
      <div class="settings-group__info">
        <div class="info-icon">
          <FileText class="info-icon__glyph" />
        </div>
        <h3 class="list-heading__title">
          基础信息
        </h3>
        <p class="info-desc">
          定义竞赛在平台展示的基础信息与访问权限。
        </p>
      </div>

      <div class="settings-group__content">
        <div class="ui-field contest-form-field settings-row">
          <label class="row-label">竞赛标题</label>
          <div class="row-control">
            <div
              class="ui-control-wrap control-wrap"
              :class="{ 'is-error': !!fieldErrors.title }"
            >
              <input
                id="contest-title"
                v-model="localDraft.title"
                type="text"
                class="ui-control"
                placeholder="输入竞赛标题..."
              >
            </div>
            <p
              v-if="fieldErrors.title"
              class="field-error"
            >
              {{ fieldErrors.title }}
            </p>
            <p class="field-hint">
              请控制在 40 个字以内，建议包含年份与赛季信息。
            </p>
          </div>
        </div>

        <div class="ui-field contest-form-field settings-row">
          <label class="row-label">竞赛描述</label>
          <div class="row-control">
            <div class="ui-control-wrap control-wrap">
              <textarea
                id="contest-description"
                v-model="localDraft.description"
                rows="4"
                class="ui-control studio-textarea"
                placeholder="描述竞赛的背景、赛制及对参赛者的要求..."
              />
            </div>
            <p class="field-hint">
              支持 Markdown 语法，将展示在竞赛详情页。
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="settings-group">
      <div class="settings-group__info">
        <div class="info-icon">
          <Settings class="info-icon__glyph" />
        </div>
        <h3 class="list-heading__title">
          赛制与状态
        </h3>
        <p class="info-desc">
          控制竞赛的底层逻辑模式与全平台生命周期。
        </p>
      </div>

      <div class="settings-group__content">
        <div class="ui-field contest-form-field settings-row">
          <label class="row-label">竞技模式</label>
          <div class="row-control">
            <div class="mode-options">
              <button
                type="button"
                class="mode-card"
                :class="{ active: localDraft.mode === 'jeopardy', disabled: fieldLocks.mode }"
                :disabled="fieldLocks.mode"
                @click="!fieldLocks.mode && (localDraft.mode = 'jeopardy')"
              >
                <Trophy class="mode-card__icon" />
                <span class="mode-label">Jeopardy</span>
                <span class="mode-desc">经典夺旗解题赛</span>
              </button>
              <button
                type="button"
                class="mode-card"
                :class="{ active: localDraft.mode === 'awd', disabled: fieldLocks.mode }"
                :disabled="fieldLocks.mode"
                @click="!fieldLocks.mode && (localDraft.mode = 'awd')"
              >
                <Swords class="mode-card__icon" />
                <span class="mode-label">AWD</span>
                <span class="mode-desc">实时攻防对抗赛</span>
              </button>
            </div>
            <p
              v-if="fieldLocks.mode"
              class="field-hint field-hint--warning field-hint--strong"
            >
              竞赛已生效，模式锁定不可更改。
            </p>
          </div>
        </div>

        <div
          v-if="mode === 'edit'"
          class="ui-field contest-form-field settings-row"
        >
          <label class="row-label">运行阶段</label>
          <div class="row-control">
            <div class="ui-control-wrap control-wrap">
              <select
                id="contest-status"
                v-model="localDraft.status"
                class="ui-control studio-select"
              >
                <option
                  v-for="option in statusOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getStatusLabel(option.value) }}
                </option>
              </select>
            </div>
            <p class="field-hint">
              手动控制竞赛在前端的可见性与交互状态。
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="settings-group">
      <div class="settings-group__info">
        <div class="info-icon">
          <Clock class="info-icon__glyph" />
        </div>
        <h3 class="list-heading__title">
          赛制与时间
        </h3>
        <p class="info-desc">
          精确配置比赛的启停节点，系统将按此时钟自动调度。
        </p>
      </div>

      <div class="settings-group__content">
        <div class="settings-row">
          <label class="row-label">赛程时间轴</label>
          <div class="row-control">
            <div class="timeline-fields">
              <div class="timeline-field">
                <div
                  class="ui-control-wrap control-wrap"
                  :class="{ 'is-disabled': fieldLocks.starts_at }"
                >
                  <input
                    id="contest-starts-at"
                    v-model="localDraft.starts_at"
                    type="datetime-local"
                    class="ui-control studio-input"
                    :disabled="fieldLocks.starts_at"
                  >
                </div>
                <p class="field-hint field-hint--compact">
                  开始时间
                </p>
              </div>
              <div class="timeline-divider">
                ——
              </div>
              <div class="timeline-field">
                <div
                  class="ui-control-wrap control-wrap"
                  :class="{ 'is-disabled': fieldLocks.ends_at }"
                >
                  <input
                    id="contest-ends-at"
                    v-model="localDraft.ends_at"
                    type="datetime-local"
                    class="ui-control studio-input"
                    :disabled="fieldLocks.ends_at"
                  >
                </div>
                <p class="field-hint field-hint--compact">
                  结束时间
                </p>
              </div>
            </div>
            <p
              v-if="fieldErrors.starts_at || fieldErrors.ends_at"
              class="field-error field-error--spaced"
            >
              {{ fieldErrors.starts_at || fieldErrors.ends_at }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <div class="contest-form-actions">
      <button
        v-if="showCancel"
        type="button"
        class="ui-btn ui-btn--secondary contest-form-button contest-form-button--secondary"
        @click="emit('cancel')"
      >
        取消
      </button>
      <button
        type="submit"
        class="ui-btn ui-btn--primary contest-form-button contest-form-button--primary"
        :disabled="saving"
      >
        {{ saving ? '保存中...' : mode === 'create' ? '创建竞赛' : '保存变更' }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.studio-settings-layout {
  --contest-form-sidebar-width: var(--ui-dialog-sidebar-width);
  --contest-form-control-width: var(--ui-selector-control-width);
  --contest-form-section-gap: var(--space-section-gap);
  --contest-form-section-gap-compact: var(--space-section-gap-compact);
  --contest-form-inline-gap: var(--space-4);
  --contest-form-control-padding-y: var(--space-3);

  width: 100%;
  max-width: var(--ui-dialog-wide-width);
  margin: 0 auto;
  padding: var(--space-7) var(--space-8);
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.settings-group {
  display: grid;
  grid-template-columns: minmax(0, var(--contest-form-sidebar-width)) minmax(0, 1fr);
  gap: var(--contest-form-section-gap);
  align-items: start;
}

.settings-group__info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.info-icon {
  width: var(--ui-control-height-md);
  height: var(--ui-control-height-md);
  border-radius: var(--ui-control-radius-md);
  background: var(--color-primary-soft);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--space-4);
}

.info-icon__glyph {
  width: var(--space-4);
  height: var(--space-4);
}

.info-desc {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: var(--space-2) 0 0;
}

.settings-group__content {
  display: flex;
  flex-direction: column;
  gap: var(--contest-form-section-gap);
  min-width: 0;
}

.settings-row {
  display: flex;
  flex-direction: column;
  gap: var(--ui-field-gap);
}

.row-label {
  font-size: var(--font-size-13);
  font-weight: 800;
  color: var(--color-text-primary);
}

.row-control {
  width: 100%;
  max-width: var(--contest-form-control-width);
  min-width: 0;
}

.control-wrap {
  width: 100%;
  border-radius: var(--ui-control-radius-md);
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  transition:
    border-color var(--ui-motion-fast),
    background var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
  overflow: hidden;
}

.control-wrap:focus-within {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 var(--ui-focus-ring-width)
    color-mix(in srgb, var(--color-primary) 18%, transparent);
}

.control-wrap.is-error {
  border-color: var(--color-danger);
}

.control-wrap.is-disabled {
  background: var(--color-bg-elevated);
  opacity: 0.7;
  cursor: not-allowed;
}

.ui-control,
.studio-input,
.studio-select,
.studio-textarea {
  width: 100%;
  border: none;
  background: transparent;
  padding: var(--contest-form-control-padding-y) var(--ui-control-padding-x-md);
  font-size: var(--font-size-14);
  font-weight: 600;
  color: var(--color-text-primary);
  outline: none;
}

.studio-textarea {
  min-height: calc(var(--ui-control-height-md) * 3);
  resize: vertical;
  line-height: 1.6;
}

.field-hint {
  font-size: var(--font-size-12);
  color: var(--color-text-muted);
  margin: var(--space-2) 0 0;
  font-weight: 500;
}

.field-hint--warning,
.timeline-divider {
  color: var(--color-warning);
}

.field-hint--strong {
  font-weight: 700;
}

.field-hint--compact {
  margin-top: var(--space-1);
}

.field-error {
  font-size: var(--font-size-12);
  color: var(--color-danger);
  font-weight: 700;
}

.field-error--spaced {
  margin-top: var(--space-2);
}

.contest-form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--ui-action-gap);
  padding-top: var(--space-2);
}

.mode-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--contest-form-inline-gap);
}

.mode-card {
  min-width: 0;
  padding: var(--space-5);
  border-radius: var(--ui-control-radius-lg);
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  transition:
    border-color var(--ui-motion-fast),
    background var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast),
    color var(--ui-motion-fast);
  color: var(--color-text-secondary);
}

.mode-card:hover:not(:disabled) {
  border-color: var(--color-primary);
  background: var(--color-bg-elevated);
}

.mode-card.active {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
  box-shadow: 0 var(--space-1) var(--space-3)
    color-mix(in srgb, var(--color-primary) 15%, transparent);
}

.mode-card.active .mode-label {
  color: var(--color-text-primary);
}

.mode-card:disabled,
.mode-card.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.mode-card__icon {
  width: var(--space-5);
  height: var(--space-5);
  margin-bottom: var(--space-2);
}

.mode-label {
  font-size: var(--font-size-14);
  font-weight: 900;
  margin-bottom: var(--space-1);
}

.mode-desc {
  font-size: var(--font-size-11);
  opacity: 0.8;
}

.timeline-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: start;
  gap: var(--contest-form-inline-gap);
}

.timeline-field {
  min-width: 0;
}

.timeline-divider {
  padding-top: var(--contest-form-control-padding-y);
  font-weight: 800;
}

@media (max-width: 1024px) {
  .studio-settings-layout {
    padding: var(--space-6);
    gap: var(--space-8);
  }

  .settings-group {
    grid-template-columns: 1fr;
    gap: var(--contest-form-section-gap-compact);
  }
}

@media (max-width: 640px) {
  .studio-settings-layout {
    padding: var(--space-5);
  }

  .mode-options,
  .timeline-fields {
    grid-template-columns: 1fr;
  }

  .timeline-divider {
    display: none;
  }
}
</style>

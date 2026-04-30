<script setup lang="ts">
import { AlertTriangle, CheckCircle2 } from 'lucide-vue-next'
import type { AWDCheckerPreviewData } from '@/api/contracts'

defineProps<{
  checkerConfigJson: string
  previewing: boolean
  previewResult: AWDCheckerPreviewData | null
  previewError: string
  previewAccessUrl: string
  previewSummary: string
  previewFlag: string
  accessUrl: string
  getCheckStatusLabel: (value: string) => string
}>()

const emit = defineEmits<{
  'update:previewFlag': [value: string]
  'update:accessUrl': [value: string]
}>()
</script>

<template>
  <section class="awd-config-form-section awd-config-debug-station">
    <header class="list-heading awd-config-section-head">
      <div>
        <div class="journal-note-label">Debug Station</div>
        <h3 class="list-heading__title">配置预览与试跑</h3>
      </div>
      <span class="awd-config-section-tag">调试区</span>
    </header>
    <div class="awd-config-debug-grid">
      <section class="awd-config-debug-panel awd-config-debug-panel--preview">
        <header class="awd-config-debug-panel__head">
          <div>
            <div class="journal-note-label">JSON Snapshot</div>
            <h4 class="awd-config-debug-panel__title">当前配置快照</h4>
          </div>
        </header>
        <pre class="checker-json-preview" id="awd-config-json-preview">{{ checkerConfigJson }}</pre>
      </section>

      <section class="awd-config-debug-panel awd-config-debug-panel--trial">
        <header class="awd-config-debug-panel__head">
          <div>
            <div class="journal-note-label">Trial Run</div>
            <h4 class="awd-config-debug-panel__title">试跑控制台</h4>
          </div>
          <div class="awd-config-debug-status">
            <span class="awd-config-debug-status__dot" :class="{ 'is-active': previewing }"></span>
            <span>{{ previewing ? '运行中' : previewResult ? '已完成' : '待运行' }}</span>
          </div>
        </header>

        <div class="checker-action-grid checker-action-grid--preview">
          <label class="ui-field checker-field checker-field--path">
            <span class="ui-field__label">目标访问地址</span>
            <span class="ui-control-wrap">
              <input
                :value="accessUrl"
                type="text"
                class="ui-control"
                placeholder="留空时由平台启动预览实例"
                @input="emit('update:accessUrl', ($event.target as HTMLInputElement).value)"
              />
            </span>
          </label>
          <label class="ui-field">
            <span class="ui-field__label">预览 Flag</span>
            <span class="ui-control-wrap">
              <input
                :value="previewFlag"
                type="text"
                class="ui-control awd-config-control--mono"
                @input="emit('update:previewFlag', ($event.target as HTMLInputElement).value)"
              />
            </span>
          </label>
        </div>

        <div class="awd-config-debug-console">
          <div v-if="previewError" class="awd-config-alert awd-config-alert--console">
            <AlertTriangle class="h-4 w-4" />
            <span>{{ previewError }}</span>
          </div>
          <div v-else-if="previewResult" class="preview-result preview-result--console">
            <CheckCircle2 class="h-4 w-4" />
            <span>{{ previewSummary || getCheckStatusLabel(String(previewResult.service_status)) || previewResult.service_status }}</span>
            <small v-if="previewAccessUrl">{{ previewAccessUrl }}</small>
          </div>
          <div v-else class="awd-config-debug-console__placeholder">
            试跑结果会显示在这里，用于确认当前 checker 配置是否可用。
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.awd-config-form-section {
  display: grid;
  gap: var(--space-3);
}

.awd-config-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.awd-config-section-tag {
  flex: none;
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  background: color-mix(in srgb, var(--color-primary-soft) 55%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.awd-config-debug-station {
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--awd-card-border);
  border-radius: var(--awd-card-radius);
  background: var(--awd-card-surface);
  box-shadow: var(--awd-card-shadow);
}

.awd-config-debug-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
  gap: var(--space-4);
  align-items: start;
}

.awd-config-debug-panel {
  min-width: 0;
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: calc(var(--awd-card-radius) - 0.125rem);
  background: var(--awd-card-subtle);
}

.awd-config-debug-panel__head {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: var(--space-3);
}

.awd-config-debug-panel__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: 700;
}

.awd-config-debug-status,
.preview-result,
.awd-config-alert {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-13);
}

.awd-config-debug-status {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.awd-config-debug-status__dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-text-secondary) 45%, transparent);
}

.awd-config-debug-status__dot.is-active {
  background: var(--color-warning);
  box-shadow: 0 0 0 0.2rem color-mix(in srgb, var(--color-warning) 18%, transparent);
}

.checker-action-grid {
  display: grid;
  gap: var(--space-3);
}

.checker-action-grid--preview {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-field--path {
  min-width: 0;
}

.awd-config-control--mono,
.checker-json-preview {
  font-family: var(--font-family-mono);
}

.checker-json-preview {
  margin: 0;
  overflow: auto;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: calc(var(--awd-card-radius) - 0.125rem);
  background: color-mix(in srgb, var(--color-bg-base) 88%, var(--color-bg-elevated));
  padding: var(--space-4);
  color: color-mix(in srgb, var(--color-text-primary) 82%, white 18%);
  font-size: var(--font-size-12);
  line-height: 1.5;
  min-height: 16rem;
  max-height: 24rem;
}

.preview-result {
  width: 100%;
  align-items: start;
  border-radius: var(--ui-control-radius);
  background: var(--color-success-soft);
  padding: var(--space-2) var(--space-3);
  color: var(--color-success);
}

.preview-result small {
  color: var(--color-text-secondary);
}

.awd-config-alert {
  margin-top: 0;
  border-radius: var(--ui-control-radius);
  background: var(--color-warning-soft);
  padding: var(--space-2) var(--space-3);
  color: var(--color-warning);
}

.awd-config-debug-console {
  min-height: 9rem;
  display: grid;
  align-content: start;
  gap: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: calc(var(--awd-card-radius) - 0.125rem);
  background: color-mix(in srgb, var(--color-bg-base) 88%, var(--color-bg-elevated));
  padding: var(--space-3);
}

.awd-config-debug-console__placeholder {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  line-height: 1.6;
}

:deep(.ui-field) {
  gap: var(--space-1);
}

:deep(.ui-field__label) {
  font-size: var(--font-size-12);
}

:deep(.ui-control) {
  min-height: 2.5rem;
}

@media (max-width: 1023px) {
  .checker-action-grid--preview,
  .awd-config-debug-grid {
    grid-template-columns: 1fr;
  }
}
</style>

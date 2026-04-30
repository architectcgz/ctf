<script setup lang="ts">
import { Play, Save } from 'lucide-vue-next'
import type { AWDCheckerPreviewData } from '@/api/contracts'

defineProps<{
  previewing: boolean
  saving: boolean
  previewError: string
  previewResult: AWDCheckerPreviewData | null
  canAttachPreviewToken: boolean
}>()

const emit = defineEmits<{
  preview: []
  save: []
}>()
</script>

<template>
  <footer class="awd-config-page__footer">
    <div class="awd-config-page__footer-status">
      <span
        class="awd-config-page__footer-status-dot"
        :class="{ 'is-running': previewing, 'is-ready': previewResult && !previewError }"
      ></span>
      <span class="awd-config-page__footer-status-text">
        {{
          previewing
            ? 'Checker 试跑中'
            : previewError
              ? '上次试跑失败'
              : previewResult
                ? '上次试跑已完成'
                : '尚未试跑'
        }}
      </span>
    </div>
    <div class="awd-config-page__footer-actions">
      <button type="button" class="ui-btn ui-btn--secondary" :disabled="previewing || saving" @click="emit('preview')">
        <Play class="h-4 w-4" />
        {{ previewing ? '试跑中...' : '试跑 Checker' }}
      </button>
      <button type="button" class="ui-btn ui-btn--primary" :disabled="saving || previewing" @click="emit('save')">
        <Save class="h-4 w-4" />
        {{ saving ? '保存中...' : canAttachPreviewToken ? '保存并写入试跑结果' : '保存配置' }}
      </button>
    </div>
  </footer>
</template>

<style scoped>
.awd-config-page__footer {
  position: sticky;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  margin-top: var(--space-2);
  padding: var(--space-3) var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  border-radius: var(--awd-card-radius);
  background: var(--awd-card-surface);
  box-shadow: 0 -0.35rem 1.2rem color-mix(in srgb, var(--color-shadow-soft) 48%, transparent);
}

.awd-config-page__footer-status {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.awd-config-page__footer-status-dot {
  width: 0.55rem;
  height: 0.55rem;
  flex: none;
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-text-muted) 72%, transparent);
}

.awd-config-page__footer-status-dot.is-running {
  background: var(--color-warning);
  box-shadow: 0 0 0 0.25rem color-mix(in srgb, var(--color-warning) 16%, transparent);
}

.awd-config-page__footer-status-dot.is-ready {
  background: var(--color-success);
  box-shadow: 0 0 0 0.25rem color-mix(in srgb, var(--color-success) 14%, transparent);
}

.awd-config-page__footer-status-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.awd-config-page__footer-actions {
  flex: none;
  display: inline-flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

@media (max-width: 767px) {
  .awd-config-page__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .awd-config-page__footer-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .awd-config-page__footer-actions .ui-btn {
    flex: 1 1 10rem;
  }
}
</style>

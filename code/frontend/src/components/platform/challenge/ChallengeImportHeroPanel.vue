<script setup lang="ts">
import { ArrowLeft, FileSearch, Plus } from 'lucide-vue-next'

const emit = defineEmits<{
  back: []
  openGuide: []
}>()

function handleBack(): void {
  emit('back')
}

function handleOpenGuide(): void {
  emit('openGuide')
}
</script>

<template>
  <div class="workspace-tab-heading challenge-import-actions">
    <div class="workspace-tab-heading__main">
      <div class="workspace-overline">
        Challenge Import
      </div>
      <h1 class="workspace-page-title">
        导入资源包
      </h1>
      <p class="workspace-page-copy">
        在独立导入页完成上传、核对题目包规范，并处理待确认导入任务。
      </p>
    </div>
    <div class="challenge-import-hero-actions">
      <button
        type="button"
        class="ui-btn ui-btn--secondary challenge-import-action"
        @click="handleBack"
      >
        <ArrowLeft class="mr-1.5 h-3.5 w-3.5" />
        返回题目目录
      </button>
      <button
        type="button"
        class="ui-btn ui-btn--secondary challenge-import-action"
        @click="handleOpenGuide"
      >
        <FileSearch class="mr-1.5 h-3.5 w-3.5" />
        题目包规范
      </button>
      <a
        class="ui-btn ui-btn--primary challenge-import-action challenge-import-action--primary"
        href="/downloads/challenge-package-sample-v1.zip"
        download="challenge-package-sample-v1.zip"
      >
        <Plus class="mr-1.5 h-4 w-4" />
        下载示例题目包
      </a>
    </div>
  </div>
</template>

<style scoped>
.challenge-import-actions {
  align-items: flex-end;
}

.challenge-import-hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.challenge-import-action {
  --ui-btn-height: 2.5rem;
  --ui-btn-padding: 0 1.25rem;
  --ui-btn-radius: 12px;
  --ui-btn-font-size: var(--font-size-12);
  --ui-btn-font-weight: 700;
  --ui-btn-hover-transform: translateY(-1px);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
}

.challenge-import-action--primary {
  --ui-btn-primary-border: color-mix(in srgb, var(--workspace-brand) 42%, transparent);
  --ui-btn-primary-background: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  --ui-btn-primary-hover-background: color-mix(
    in srgb,
    var(--workspace-brand-ink) 92%,
    var(--challenge-page-text)
  );
  --ui-btn-primary-hover-border: color-mix(in srgb, var(--workspace-brand-ink) 62%, transparent);
  --ui-btn-primary-hover-shadow: 0 10px 24px color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}
</style>

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
  <div class="workspace-tab-heading challenge-import-heading">
    <div class="workspace-tab-heading__main">
      <div class="workspace-overline">
        Challenge Import
      </div>
      <h1 class="workspace-page-title">
        导入题目
      </h1>
      <p class="workspace-page-copy">
        在独立导入页完成上传、核对题目包规范，并处理待确认导入任务。
      </p>
    </div>
    <div class="header-actions challenge-import-hero-actions">
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="handleBack"
      >
        <ArrowLeft class="h-4 w-4" />
        返回题目目录
      </button>
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="handleOpenGuide"
      >
        <FileSearch class="h-4 w-4" />
        题目包规范
      </button>
      <a
        class="header-btn header-btn--primary"
        href="/downloads/challenge-package-sample-v1.zip"
        download="challenge-package-sample-v1.zip"
      >
        <Plus class="h-4 w-4" />
        下载示例题目包
      </a>
    </div>
  </div>
</template>

<style scoped>
.challenge-import-heading {
  align-items: flex-end;
}

.challenge-import-hero-actions {
  align-items: flex-end;
}
</style>

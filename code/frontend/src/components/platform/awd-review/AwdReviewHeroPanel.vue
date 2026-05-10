<script setup lang="ts">
import { FolderKanban, RefreshCcw, ScanEye, Waypoints } from 'lucide-vue-next'

defineProps<{
  contestCount: number
  runningCount: number
  exportReadyCount: number
}>()

const emit = defineEmits<{
  (event: 'back'): void
  (event: 'refresh'): void
}>()
</script>

<template>
  <header class="workspace-page-header admin-awd-review-shell__hero">
    <div class="admin-awd-review-shell__hero-main">
      <div class="workspace-overline">
        Review Workspace
      </div>
      <h1 class="workspace-page-title">
        AWD复盘
      </h1>
      <p class="workspace-page-copy">
        在平台视角统一查看可进入的 AWD 赛事、当前状态和报告就绪度，并直接进入复盘详情。
      </p>
    </div>

    <div class="header-actions admin-awd-review-shell__hero-actions">
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="emit('back')"
      >
        返回平台概览
      </button>
      <button
        type="button"
        class="header-btn header-btn--primary"
        @click="emit('refresh')"
      >
        <RefreshCcw class="h-4 w-4" />
        刷新目录
      </button>
    </div>
  </header>

  <div
    class="admin-summary-grid admin-awd-review-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
  >
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>赛事数量</span>
        <FolderKanban class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ contestCount.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前可进入复盘的 AWD 赛事
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>进行中</span>
        <ScanEye class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ runningCount.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        仍在持续产出攻防信号的赛事
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>可导出报告</span>
        <Waypoints class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ exportReadyCount.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        已允许导出教师复盘报告的赛事
      </div>
    </article>
  </div>
</template>

<style scoped>
.admin-awd-review-shell__hero-main {
  max-width: 48rem;
}

.admin-awd-review-shell__summary {
  --metric-panel-columns: 3;
  --metric-panel-border: color-mix(in srgb, var(--workspace-brand) 16%, var(--workspace-line-soft));
}

@media (max-width: 900px) {
  .admin-awd-review-shell__hero-actions {
    width: 100%;
  }
}
</style>

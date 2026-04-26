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
  <header class="admin-awd-review-shell__hero">
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

    <div class="admin-awd-review-shell__hero-actions">
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="emit('back')"
      >
        返回平台概览
      </button>
      <button
        type="button"
        class="ui-btn ui-btn--primary"
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
.admin-awd-review-shell__hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.admin-awd-review-shell__hero-main {
  max-width: 48rem;
}

.admin-awd-review-shell__hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.admin-awd-review-shell__hero-actions > .ui-btn {
  --ui-btn-height: 2.75rem;
  --ui-btn-radius: 1rem;
  --ui-btn-padding: var(--space-2-5) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
}

.admin-awd-review-shell__hero-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--admin-control-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
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

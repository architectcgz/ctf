<script setup lang="ts">
import {
  Book,
  Calendar,
  CheckCircle,
  Edit3,
  Plus,
} from 'lucide-vue-next'

defineProps<{
  total: number
  publishedCount: number
  draftCount: number
  archivedCount: number
}>()

const emit = defineEmits<{
  import: []
}>()

function handleImport(): void {
  emit('import')
}
</script>

<template>
  <div class="challenge-manage-hero-panel">
    <section class="workspace-hero">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">
          Challenge Workspace
        </div>
        <h1 class="workspace-page-title">
          Jeopardy题库
        </h1>
        <p class="workspace-page-copy">
          集中管理 Jeopardy 题目目录、发布状态与题库变更。
        </p>
      </div>
      <div class="challenge-manage-hero-actions">
        <div class="header-actions challenge-manage-actions">
          <button
            type="button"
            class="header-btn header-btn--primary challenge-manage-import-button"
            @click="handleImport"
          >
            <Plus class="h-4 w-4" />
            导入题目
          </button>
        </div>
      </div>
    </section>

    <div class="manage-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
      <article class="journal-note progress-card metric-panel-card">
        <div class="challenge-metric-head">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>题目总量</span>
            <Book class="h-4 w-4" />
          </div>
        </div>
        <div class="challenge-metric-value-wrap">
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ total.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            Jeopardy题目总计
          </div>
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="challenge-metric-head">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>已发布</span>
            <CheckCircle class="h-4 w-4" />
          </div>
        </div>
        <div class="challenge-metric-value-wrap">
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ publishedCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            线上公开题目
          </div>
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="challenge-metric-head">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>草稿存量</span>
            <Edit3 class="h-4 w-4" />
          </div>
        </div>
        <div class="challenge-metric-value-wrap">
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ draftCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            导入后仍待发布
          </div>
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="challenge-metric-head">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>已归档</span>
            <Calendar class="h-4 w-4" />
          </div>
        </div>
        <div class="challenge-metric-value-wrap">
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ archivedCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            只读保留题目
          </div>
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.challenge-manage-hero-panel {
  display: grid;
  gap: 0;
}

.manage-summary-grid {
  --metric-panel-columns: 4;
}

.challenge-metric-head {
  margin-bottom: var(--space-2);
}

.challenge-metric-value-wrap .metric-panel-value,
.challenge-metric-value-wrap .metric-panel-helper {
  margin-top: 0;
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.challenge-manage-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: var(--space-2);
}

.challenge-manage-actions {
  display: flex;
  gap: var(--space-3);
}

.challenge-manage-import-button {
  --header-btn-border: color-mix(in srgb, var(--journal-accent) 34%, var(--journal-border));
  --header-btn-background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  --header-btn-color: var(--journal-accent-strong);
  --header-btn-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 18%,
    var(--journal-surface)
  );
  --header-btn-hover-border: color-mix(in srgb, var(--journal-accent-strong) 46%, var(--journal-border));
  --header-btn-hover-color: var(--journal-accent-strong);
  --header-btn-hover-shadow: 0 0.75rem 1.5rem color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --header-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 28%, transparent);
}
</style>

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
        <div class="workspace-overline">Challenge Workspace</div>
        <h1 class="workspace-page-title">
          题目资源管理中心
        </h1>
        <p class="workspace-page-copy">集中查看题目目录、发布状态与题库变更。</p>
      </div>
      <div class="awd-library-hero-actions">
        <div class="quick-actions">
          <button
            type="button"
            class="ui-btn ui-btn--primary"
            @click="handleImport"
          >
            <Plus class="h-4 w-4" />
            导入资源包
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
            题目资源总计
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
  gap: var(--space-section-gap-compact, var(--space-4));
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

.awd-library-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: 0.5rem;
}

.quick-actions {
  display: flex;
  gap: 0.75rem;
}
</style>

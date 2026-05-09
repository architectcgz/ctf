<script setup lang="ts">
import {
  Activity,
  Layers,
  RefreshCw,
  Trophy,
} from 'lucide-vue-next'

defineProps<{
  currentCount: number
  total: number
  totalPages: number
  loading: boolean
}>()

const emit = defineEmits<{
  sync: []
}>()

function handleSync(): void {
  emit('sync')
}
</script>

<template>
  <div class="audit-log-hero-panel">
    <section class="workspace-hero">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">Audit Log</div>
        <h1 class="hero-title">
          审计日志
        </h1>
        <p class="hero-summary">
          追踪全站资源变更、用户行为与系统关键操作，确保平台安全与合规。
        </p>
      </div>

      <div class="awd-library-hero-actions">
        <div class="header-actions quick-actions">
          <button
            type="button"
            class="header-btn header-btn--primary"
            :disabled="loading"
            @click="handleSync"
          >
            <RefreshCw class="h-4 w-4" />
            同步日志
          </button>
        </div>
      </div>
    </section>

    <div class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>当前页加载</span>
          <Activity class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ currentCount.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          本页已加载的日志条数
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>全站总记录</span>
          <Trophy class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ total.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          审计数据库中的累计总量
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>总分页范围</span>
          <Layers class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ totalPages.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前条件下的分页总数
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.audit-log-hero-panel {
  display: grid;
  gap: 0;
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft, var(--color-border-subtle));
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--color-text-primary);
}

.hero-summary {
  max-width: 47.5rem;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--color-text-secondary);
}

.quick-actions {
  display: flex;
  gap: var(--space-3);
  align-items: flex-end;
  height: 100%;
  padding-bottom: 0.5rem;
}
</style>

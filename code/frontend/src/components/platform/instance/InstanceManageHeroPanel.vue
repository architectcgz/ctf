<script setup lang="ts">
import { Activity, AlertTriangle, RefreshCw, Server } from 'lucide-vue-next'

defineProps<{
  runningCount: number
  total: number
  warningCount: number
}>()

const emit = defineEmits<{
  back: []
  refresh: []
}>()

function handleBack(): void {
  emit('back')
}

function handleRefresh(): void {
  emit('refresh')
}
</script>

<template>
  <section class="workspace-hero">
    <div class="workspace-tab-heading__main">
      <div class="workspace-overline">
        Instance Workspace
      </div>
      <h1 class="hero-title">
        实例管理
      </h1>
      <p class="hero-summary">
        在后台视角查看实例状态、到期节奏与访问地址，并快速销毁异常环境。
      </p>
    </div>

    <div class="awd-library-hero-actions">
      <div class="quick-actions">
        <button
          type="button"
          class="ui-btn ui-btn--ghost"
          @click="handleBack"
        >
          返回概览
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          @click="handleRefresh"
        >
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
      </div>
    </div>
  </section>

  <div
    class="admin-summary-grid admin-instance-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
  >
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>运行中</span>
        <Activity class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ runningCount.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前活跃实例
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>总实例数</span>
        <Server class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ total.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        系统托管总计
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>预警项</span>
        <AlertTriangle class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ warningCount.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        即将过期或异常
      </div>
    </article>
  </div>
</template>

<style scoped>
.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--journal-ink);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
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

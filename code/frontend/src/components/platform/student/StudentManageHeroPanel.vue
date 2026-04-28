<script setup lang="ts">
import { Activity, GraduationCap, RefreshCw, Users } from 'lucide-vue-next'

defineProps<{
  total: number
  activeStudents: number
  assignedClassCount: number
}>()

const emit = defineEmits<{
  refresh: []
}>()

function handleRefresh(): void {
  emit('refresh')
}
</script>

<template>
  <section class="workspace-hero">
    <div class="workspace-tab-heading__main">
      <div class="workspace-overline">
        Student Workspace
      </div>
      <h1 class="hero-title">
        学生管理
      </h1>
      <p class="hero-summary">
        在后台视角查看学生目录、班级归属与学习表现，并快速进入学员分析。
      </p>
    </div>

    <div class="awd-library-hero-actions">
      <div class="quick-actions">
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          @click="handleRefresh"
        >
          <RefreshCw class="h-4 w-4" />
          刷新目录
        </button>
      </div>
    </div>
  </section>

  <div
    class="admin-summary-grid admin-student-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
  >
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>学生总量</span>
        <Users class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ total.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        平台注册学员总数
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>活跃学员</span>
        <Activity class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ activeStudents.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前结果中最近 30 天有登录记录
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>正式班级</span>
        <GraduationCap class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ assignedClassCount.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        已分配学生的班级数量
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

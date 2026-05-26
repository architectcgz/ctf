<script setup lang="ts">
interface TrendSignal {
  key: string
  label: string
  value: string
  copy: string
}

interface FocusClass {
  class_name: string
  active_rate: number
  recent_event_count: number
  risk_student_count: number
  dominant_weak_dimension?: string | null
}

defineProps<{
  trendSignals: TrendSignal[]
  focusClasses: FocusClass[]
}>()
</script>

<template>
  <section class="overview-panel workspace-directory-section teacher-directory-section">
    <header class="list-heading">
      <div>
        <h2 class="list-heading__title">趋势复盘</h2>
      </div>
    </header>

    <div class="teacher-dashboard-panel-body trend-grid">
      <div
        v-if="trendSignals.some((item) => item.value !== '--')"
        class="summary-grid progress-strip metric-panel-grid metric-panel-default-surface"
      >
        <article
          v-for="item in trendSignals"
          :key="item.key"
          class="summary-note progress-card metric-panel-card"
        >
          <div class="summary-note-label progress-card-label metric-panel-label">
            {{ item.label }}
          </div>
          <div class="summary-note-value progress-card-value metric-panel-value">
            {{ item.value }}
          </div>
          <div class="summary-note-copy progress-card-hint metric-panel-helper">
            {{ item.copy }}
          </div>
        </article>
      </div>
      <div v-else class="workspace-directory-empty portrait-empty">暂无可复盘的趋势数据</div>

      <div v-if="focusClasses.length > 0" class="focus-class-list workspace-directory-list">
        <article
          v-for="(item, index) in focusClasses.slice(0, 4)"
          :key="item.class_name"
          class="focus-class-row"
        >
          <div class="weak-rank">
            {{ `${index + 1}`.padStart(2, '0') }}
          </div>
          <div class="focus-class-row__main">
            <h3 class="focus-class-row__title">{{ item.class_name }}</h3>
            <p class="focus-class-row__detail">
              {{
                item.dominant_weak_dimension
                  ? `当前主要薄弱维度为 ${item.dominant_weak_dimension}，共有 ${item.risk_student_count} 名待跟进学生。`
                  : `当前共有 ${item.risk_student_count} 名待跟进学生，薄弱维度仍在形成中。`
              }}
            </p>
            <div class="focus-class-row__chips">
              <span class="workspace-directory-status-pill workspace-directory-status-pill--muted">
                活跃率 {{ Math.round(item.active_rate) }}%
              </span>
              <span class="workspace-directory-status-pill workspace-directory-status-pill--muted">
                近 7 天 {{ item.recent_event_count }} 次事件
              </span>
            </div>
          </div>
        </article>
      </div>
      <div v-else class="workspace-directory-empty portrait-empty">暂无重点班级趋势摘要</div>
    </div>
  </section>
</template>

<style scoped>
.overview-panel {
  --workspace-directory-section-padding: 0;
  --workspace-directory-section-gap: var(--space-5);
  --workspace-directory-shell-radius: 16px;
  --workspace-directory-shell-padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.overview-panel > .list-heading {
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.teacher-dashboard-panel-body,
.trend-grid,
.focus-class-list {
  display: grid;
  gap: var(--space-5);
}

.summary-grid {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.summary-note {
  min-height: 7.25rem;
}

.summary-note-copy {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.portrait-empty {
  padding: var(--space-5);
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

.focus-class-list {
  overflow: hidden;
}

.focus-class-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-4-5) var(--space-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
}

.focus-class-row:last-child {
  border-bottom: 0;
}

.weak-rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.125rem;
  height: 2.125rem;
  border-radius: 12px;
  background: var(--workspace-brand-soft);
  font: 700 var(--font-size-13) / 1 var(--font-family-mono);
  color: var(--journal-accent-strong);
}

.focus-class-row__main {
  min-width: 0;
}

.focus-class-row__title {
  margin: 0;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--journal-ink);
}

.focus-class-row__detail {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.focus-class-row__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-3);
}

@media (max-width: 760px) {
  .summary-grid {
    --metric-panel-columns: 1fr;
  }
}
</style>

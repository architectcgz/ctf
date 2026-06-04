<script setup lang="ts">
interface PortraitSummaryNote {
  key: string
  label: string
  value: string
  copy?: string
}

interface WeakDimensionStat {
  dimension: string
  count: number
  width: string
}

defineProps<{
  portraitSummaryNotes: PortraitSummaryNote[]
  weakDimensionStats: WeakDimensionStat[]
}>()
</script>

<template>
  <section class="overview-panel overview-panel--wide workspace-directory-section teacher-directory-section">
    <header class="list-heading">
      <div>
        <h2 class="list-heading__title">能力画像与薄弱维度</h2>
      </div>
    </header>

    <div class="teacher-dashboard-panel-body portrait-grid">
      <div class="portrait-summary-block">
        <div class="teacher-dashboard-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
          <article
            v-for="item in portraitSummaryNotes"
            :key="item.key"
            class="teacher-dashboard-summary-card progress-card metric-panel-card"
          >
            <div class="summary-note-label progress-card-label metric-panel-label">
              {{ item.label }}
            </div>
            <div class="summary-note-value progress-card-value metric-panel-value">
              {{ item.value }}
            </div>
            <div class="teacher-dashboard-summary-copy progress-card-hint metric-panel-helper">
              {{ item.copy || '画像摘要' }}
            </div>
          </article>
        </div>

        <div class="portrait-guidance">
          <div class="portrait-guidance__label">使用方式</div>
          <div class="portrait-guidance__copy">
            先看影响学生最多的能力方向，再结合复盘结论安排题单或课堂讲解。
          </div>
        </div>
      </div>

      <div class="portrait-dimension-block">
        <div class="panel-header-row">
          <h3 class="panel-title">优先补强方向</h3>
          <span class="panel-badge">按学生数排序</span>
        </div>

        <div v-if="weakDimensionStats.length > 0" class="weak-list workspace-directory-list">
          <article
            v-for="(item, index) in weakDimensionStats.slice(0, 5)"
            :key="item.dimension"
            class="weak-item"
          >
            <div class="weak-rank">
              {{ `${index + 1}`.padStart(2, '0') }}
            </div>
            <div class="weak-content">
              <div class="weak-name" :title="item.dimension">
                {{ item.dimension }}
              </div>
              <div class="weak-copy">{{ item.count }} 名学生当前在该方向暴露弱项。</div>
              <div class="weak-bar">
                <span :style="{ width: item.width }" />
              </div>
            </div>
            <div class="weak-score">{{ item.count }} 人</div>
          </article>
        </div>
        <div v-else class="workspace-directory-empty portrait-empty">
          暂无可排序的薄弱维度
        </div>
      </div>
    </div>
  </section>
</template>

<style src="@/features/teacher/dashboard/ui/teacherDashboardSummary.css"></style>

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

.overview-panel--wide {
  grid-column: 1 / -1;
}

.overview-panel > .list-heading {
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.teacher-dashboard-panel-body {
  min-width: 0;
}

.portrait-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.25fr);
  gap: var(--space-5);
  align-items: start;
}

.portrait-summary-block,
.portrait-dimension-block {
  display: grid;
  gap: var(--space-5);
  min-width: 0;
}

.portrait-guidance {
  border-left: 3px solid color-mix(in srgb, var(--journal-accent) 58%, transparent);
  padding: var(--space-3) var(--space-4);
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
}

.portrait-guidance__label {
  font-size: var(--font-size-12);
  font-weight: 800;
  color: var(--journal-accent-strong);
}

.portrait-guidance__copy {
  margin-top: var(--space-1);
  font-size: var(--font-size-13);
  line-height: 1.65;
  color: var(--journal-muted);
}

.panel-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.panel-title {
  margin: 0;
  font-size: var(--font-size-17);
  font-weight: 800;
  color: var(--journal-ink);
}

.panel-badge {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  border-radius: 999px;
  background: var(--workspace-brand-soft);
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.weak-list {
  --workspace-directory-shell-background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  display: grid;
  overflow: hidden;
}

.weak-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: center;
  padding: var(--space-4) var(--space-4-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
  background: transparent;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.weak-item:last-child {
  border-bottom: 0;
}

.weak-item:hover {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
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

.weak-content {
  min-width: 0;
}

.weak-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-15);
  font-weight: 800;
  color: var(--journal-ink);
}

.weak-copy {
  margin-top: var(--space-1);
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.weak-score {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-13);
  font-weight: 800;
  color: var(--journal-accent-strong);
}

.weak-bar {
  height: 0.375rem;
  margin-top: var(--space-2);
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--teacher-card-border) 66%, transparent);
}

.weak-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: color-mix(in srgb, var(--journal-accent) 72%, var(--journal-accent-strong));
}

.portrait-empty {
  padding: var(--space-5);
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

@media (max-width: 1180px) {
  .portrait-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .weak-item {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .weak-score {
    grid-column: 2;
  }
}
</style>

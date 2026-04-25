<script setup lang="ts">
import { toRef } from 'vue'
import { AlertTriangle } from 'lucide-vue-next'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'
import { useTeacherDashboardMetrics } from '@/composables/useTeacherDashboardMetrics'

const props = defineProps<{
  classes: TeacherClassItem[]
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedClass: TeacherClassItem | null
  review: TeacherClassReviewData | null
  summary: TeacherClassSummaryData | null
  trend: TeacherClassTrendData | null
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
}>()

const {
  activeRateText,
  riskStudentCount,
  overviewDescription,
  metaPills,
  overviewMetrics,
  studentInsightRows,
  portraitSummaryNotes,
  weakDimensionStats,
} = useTeacherDashboardMetrics({
  students: toRef(props, 'students'),
  selectedClassName: toRef(props, 'selectedClassName'),
  selectedClass: toRef(props, 'selectedClass'),
  review: toRef(props, 'review'),
  summary: toRef(props, 'summary'),
  trend: toRef(props, 'trend'),
})
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface teacher-dashboard-shell flex min-h-full flex-1 flex-col">
    <div class="workspace-grid">
      <main class="content-pane teacher-dashboard-content">
        <section
          id="overview"
          class="workspace-hero teacher-dashboard-hero"
        >
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Progress Signal
            </div>
            <h1 class="hero-title">
              教学介入台
            </h1>
            <p class="hero-summary">
              {{ overviewDescription }}
            </p>

            <div class="meta-strip">
              <span
                v-for="(pill, index) in metaPills"
                :key="pill"
                class="meta-pill"
                :class="{ brand: index === 0 }"
              >
                {{ pill }}
              </span>
            </div>

            <div class="teacher-overview-summary progress-strip metric-panel-grid metric-panel-default-surface">
              <article
                v-for="item in overviewMetrics"
                :key="item.key"
                class="teacher-overview-card progress-card metric-panel-card"
              >
                <div class="progress-card-label metric-panel-label">
                  {{ item.label }}
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ item.value }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  {{ item.hint }}
                </div>
              </article>
            </div>

            <div
              v-if="error"
              class="workspace-alert"
            >
              <div class="workspace-alert-title-row">
                <AlertTriangle class="workspace-alert-icon h-4 w-4" />
                <div class="workspace-alert-title">
                  加载失败
                </div>
              </div>
              <div class="workspace-alert-copy">
                {{ error }}
              </div>
              <div class="workspace-alert-actions">
                <button
                  type="button"
                  class="ui-btn ui-btn--primary ui-btn--sm"
                  @click="emit('retry')"
                >
                  重试加载
                </button>
              </div>
            </div>
          </div>

          <aside class="hero-rail workspace-subpanel">
            <div class="rail-label">
              Class Pulse
            </div>
            <div class="rail-score">
              {{ activeRateText.replace('%', '') }}
              <small v-if="activeRateText !== '--'">% active</small>
            </div>
            <div class="rail-copy">
              {{
                riskStudentCount > 0
                  ? `当前仍有 ${riskStudentCount} 名学生需要优先回流训练节奏。`
                  : '班级整体节奏稳定，建议关注薄弱维度补强。'
              }}
            </div>
          </aside>

          <div class="overview-workbench">
            <section class="overview-panel overview-panel--wide workspace-directory-section teacher-directory-section">
              <header class="list-heading">
                <div>
                  <div class="workspace-overline">
                    Skill Portrait
                  </div>
                  <h2 class="list-heading__title">
                    能力画像与薄弱维度
                  </h2>
                </div>
              </header>

              <div class="teacher-dashboard-panel-body portrait-grid">
                <div class="portrait-summary-block">
                  <div class="summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
                    <article
                      v-for="item in portraitSummaryNotes"
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
                        {{ item.copy || '画像摘要' }}
                      </div>
                    </article>
                  </div>

                  <div class="portrait-guidance">
                    <div class="portrait-guidance__label">
                      使用方式
                    </div>
                    <div class="portrait-guidance__copy">
                      先看影响学生最多的能力方向，再结合复盘结论安排题单或课堂讲解。
                    </div>
                  </div>
                </div>

                <div class="portrait-dimension-block">
                  <div class="panel-header-row">
                    <h3 class="panel-title">
                      优先补强方向
                    </h3>
                    <span class="panel-badge">按学生数排序</span>
                  </div>

                  <div
                    v-if="weakDimensionStats.length > 0"
                    class="weak-list workspace-directory-list"
                  >
                    <article
                      v-for="(item, index) in weakDimensionStats.slice(0, 5)"
                      :key="item.dimension"
                      class="weak-item"
                    >
                      <div class="weak-rank">
                        {{ `${index + 1}`.padStart(2, '0') }}
                      </div>
                      <div class="weak-content">
                        <div
                          class="weak-name"
                          :title="item.dimension"
                        >
                          {{ item.dimension }}
                        </div>
                        <div class="weak-copy">
                          {{ item.count }} 名学生当前在该方向暴露弱项。
                        </div>
                        <div class="weak-bar">
                          <span :style="{ width: item.width }" />
                        </div>
                      </div>
                      <div class="weak-score">
                        {{ item.count }} 人
                      </div>
                    </article>
                  </div>
                  <div
                    v-else
                    class="workspace-directory-empty portrait-empty"
                  >
                    暂无可排序的薄弱维度
                  </div>
                </div>
              </div>
            </section>

            <section class="overview-panel workspace-directory-section teacher-directory-section">
              <header class="list-heading">
                <div>
                  <div class="workspace-overline">
                    Student Insight
                  </div>
                  <h2 class="list-heading__title">
                    学生洞察
                  </h2>
                </div>
              </header>

              <div class="teacher-dashboard-panel-body">
                <div class="student-insight-list workspace-directory-list">
                  <article
                    v-for="row in studentInsightRows"
                    :key="row.key"
                    class="student-insight-row"
                    :class="`student-insight-row--${row.tone}`"
                  >
                    <div class="student-insight-row__status">
                      {{ row.status }}
                    </div>
                    <div class="student-insight-row__main">
                      <h3
                        class="student-insight-row__title"
                        :title="row.title"
                      >
                        {{ row.title }}
                      </h3>
                      <p class="student-insight-row__detail">
                        {{ row.detail }}
                      </p>
                      <div class="student-insight-row__chips">
                        <span
                          v-for="chip in row.chips"
                          :key="chip"
                          class="student-insight-chip"
                        >
                          {{ chip }}
                        </span>
                      </div>
                    </div>
                  </article>
                </div>
              </div>
            </section>

            <section class="overview-panel workspace-directory-section teacher-directory-section">
              <header class="list-heading">
                <div>
                  <div class="workspace-overline">
                    Trend Review
                  </div>
                  <h2 class="list-heading__title">
                    趋势复盘
                  </h2>
                </div>
              </header>

              <div class="teacher-dashboard-panel-body workspace-subpanel workspace-subpanel--flat">
                <TeacherClassTrendPanel
                  :trend="trend"
                  title="班级近 7 天训练趋势"
                  bare
                />
              </div>
            </section>

            <section class="overview-panel workspace-directory-section teacher-directory-section">
              <header class="list-heading">
                <div>
                  <div class="workspace-overline">
                    Review
                  </div>
                  <h2 class="list-heading__title">
                    教学复盘结论
                  </h2>
                </div>
              </header>

              <div class="teacher-dashboard-panel-body workspace-subpanel workspace-subpanel--flat">
                <TeacherClassReviewPanel
                  :review="review"
                  :class-name="selectedClassName"
                  bare
                />
              </div>
            </section>

            <section class="overview-panel workspace-directory-section teacher-directory-section">
              <header class="list-heading">
                <div>
                  <div class="workspace-overline">
                    Intervention
                  </div>
                  <h2 class="list-heading__title">
                    优先介入学生
                  </h2>
                </div>
              </header>

              <div class="teacher-dashboard-panel-body workspace-subpanel workspace-subpanel--flat">
                <TeacherInterventionPanel
                  :students="students"
                  :class-name="selectedClassName"
                  bare
                />
              </div>
            </section>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
@import '../teacher-workspace-subpanel.css';

.teacher-dashboard-shell {
  --journal-ink: var(--color-text-primary);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--color-border-default) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--color-border-default) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --workspace-line-soft: var(--color-border-subtle);
  --workspace-panel: var(--color-bg-surface);
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: var(--journal-accent-strong);
  --workspace-brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.teacher-btn {
  border: 1px solid var(--teacher-control-border);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-dashboard-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.teacher-dashboard-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(17rem, 0.34fr);
  gap: var(--space-7);
  align-items: stretch;
}

.overview-workbench {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-5);
  padding-top: var(--space-6);
  border-top: 1px solid var(--workspace-line-soft);
}

.overview-panel {
  --workspace-directory-section-padding: var(--space-5) var(--space-5-5) var(--space-5-5);
  --workspace-directory-section-gap: var(--space-5);
  --workspace-directory-shell-radius: 16px;
  --workspace-directory-shell-padding: 0;
  border: 1px solid var(--teacher-card-border);
  border-radius: 22px;
  background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
}

.overview-panel--wide {
  grid-column: 1 / -1;
}

.overview-panel > .list-heading {
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-4);
  font-size: var(--font-size-15);
  line-height: 1.8;
  color: var(--journal-muted);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-5);
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 var(--space-3);
  border: 1px solid var(--teacher-control-border);
  border-radius: 8px;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: var(--font-size-12);
  color: var(--journal-muted);
}
.meta-pill.brand {
  border-color: color-mix(in srgb, var(--journal-accent) 34%, transparent);
  background: var(--workspace-brand-soft);
  color: var(--journal-accent-strong);
}

.teacher-overview-summary {
  margin-top: var(--space-5);
}

.summary-grid {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.teacher-overview-card {
  min-height: 7.75rem;
}

.hero-rail {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  min-height: 100%;
  padding: var(--space-5);
  border-color: color-mix(in srgb, var(--journal-accent) 18%, var(--teacher-card-border));
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.rail-label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--journal-accent) 68%, var(--journal-muted));
}

.rail-score {
  margin-top: var(--space-3);
  font: 900 var(--font-size-38, 2.375rem) / 1 var(--font-family-mono);
  color: var(--journal-ink);
}

.rail-score small {
  margin-left: var(--space-1);
  font: 700 var(--font-size-12) / 1 var(--font-family-sans);
  color: var(--journal-muted);
}

.rail-copy {
  margin-top: var(--space-4);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
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
  --workspace-directory-shell-background: color-mix(
    in srgb,
    var(--journal-surface) 96%,
    transparent
  );
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
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease;
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

.summary-note {
  min-height: 7.25rem;
}

.summary-note-copy {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.teacher-dashboard-panel-body.workspace-subpanel--flat {
  display: flex;
  flex-direction: column;
  padding: 0;
  border: 0;
  background: transparent;
  box-shadow: none;
}

.teacher-dashboard-panel-body.workspace-subpanel--flat :deep(.teacher-panel) {
  width: 100%;
}

.teacher-dashboard-panel-body.workspace-subpanel--flat :deep(.review-item),
.teacher-dashboard-panel-body.workspace-subpanel--flat :deep(.intervention-item) {
  border-width: 0 0 1px;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.student-insight-list {
  display: grid;
}

.student-insight-row {
  display: grid;
  grid-template-columns: minmax(7rem, 0.18fr) minmax(0, 1fr);
  gap: var(--space-5);
  padding: var(--space-4-5) var(--space-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
}

.student-insight-row:last-child {
  border-bottom: 0;
}

.student-insight-row__status {
  align-self: start;
  justify-self: start;
  display: inline-flex;
  align-items: center;
  min-height: 1.875rem;
  padding: 0 var(--space-3);
  border: 1px solid var(--teacher-card-border);
  border-radius: 999px;
  font-size: var(--font-size-12);
  font-weight: 800;
  color: var(--journal-muted);
}

.student-insight-row--ready .student-insight-row__status {
  border-color: color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  color: color-mix(in srgb, var(--color-success) 78%, var(--journal-ink));
}

.student-insight-row--warning .student-insight-row__status {
  border-color: color-mix(in srgb, var(--color-warning) 32%, transparent);
  background: color-mix(in srgb, var(--color-warning) 9%, transparent);
  color: color-mix(in srgb, var(--color-warning) 78%, var(--journal-ink));
}

.student-insight-row--danger .student-insight-row__status {
  border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  color: color-mix(in srgb, var(--color-danger) 78%, var(--journal-ink));
}

.student-insight-row__main {
  min-width: 0;
}

.student-insight-row__title {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--journal-ink);
}

.student-insight-row__detail {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.student-insight-row__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-3);
}

.student-insight-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.625rem;
  padding: 0 var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 86%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface) 78%, transparent);
  font-size: var(--font-size-12);
  color: var(--journal-muted);
}

@media (max-width: 1180px) {
  .teacher-dashboard-hero,
  .overview-workbench,
  .portrait-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .teacher-dashboard-shell {
    --metric-panel-columns: 1fr;
  }

  .summary-grid {
    --metric-panel-columns: 1fr;
  }

  .weak-item {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .weak-score {
    grid-column: 2;
  }

  .student-insight-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}
</style>

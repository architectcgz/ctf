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
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'
import { useTeacherDashboardMetrics } from '@/composables/useTeacherDashboardMetrics'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

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

type WorkspaceTab = 'overview' | 'portrait' | 'trend' | 'insight' | 'advice' | 'action'

interface WorkspaceTabItem {
  key: WorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

const workspaceTabs: WorkspaceTabItem[] = [
  { key: 'overview', label: '进度总览', buttonId: 'top-tab-overview', panelId: 'overview' },
  { key: 'portrait', label: '能力画像', buttonId: 'top-tab-portrait', panelId: 'portrait' },
  { key: 'trend', label: '趋势复盘', buttonId: 'top-tab-trend', panelId: 'trend' },
  { key: 'insight', label: '学生洞察', buttonId: 'top-tab-insight', panelId: 'insight' },
  { key: 'advice', label: '今日教学建议', buttonId: 'top-tab-advice', panelId: 'advice' },
  { key: 'action', label: '介入建议', buttonId: 'top-tab-action', panelId: 'action' },
]

const workspaceTabOrder = workspaceTabs.map((tab) => tab.key) as WorkspaceTab[]
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } = useUrlSyncedTabs<WorkspaceTab>(
  {
    orderedTabs: workspaceTabOrder,
    defaultTab: 'overview',
  }
)

const {
  averageSolvedText,
  activeRateText,
  studentCountText,
  recentEventCountText,
  recentTrendPoints,
  weakDimensionStats,
  dominantWeakDimension,
  riskStudentCount,
  activeStudentValue,
  topStudent,
  strongestDimensionCount,
  overviewDescription,
  metaPills,
  overviewMetrics,
  interventionTips,
  teachingAdvice,
  studentInsightRows,
  portraitSummaryNotes,
  trendSignals,
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
  <div class="workspace-shell teacher-management-shell teacher-surface">
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Teaching Workspace</span>
        <span class="class-chip">{{ selectedClassName || '未选择班级' }}</span>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="教学概览标签页">
      <button
        v-for="(tab, index) in workspaceTabs"
        :id="tab.buttonId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        class="top-tab"
        :class="{ active: activeTab === tab.key }"
        type="button"
        role="tab"
        :tabindex="activeTab === tab.key ? 0 : -1"
        :aria-selected="activeTab === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        @click="selectTab(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <div class="workspace-grid">
      <main class="content-pane">
        <section
          id="overview"
          class="workspace-hero tab-panel"
          :class="{ active: activeTab === 'overview' }"
          role="tabpanel"
          aria-labelledby="top-tab-overview"
          :aria-hidden="activeTab === 'overview' ? 'false' : 'true'"
          v-show="activeTab === 'overview'"
        >
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Progress Signal</div>
            <h1 class="hero-title">教学介入台</h1>
            <p class="hero-summary">{{ overviewDescription }}</p>

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

            <div class="progress-strip metric-panel-grid">
              <article v-for="item in overviewMetrics" :key="item.key" class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">{{ item.label }}</div>
                <div class="progress-card-value metric-panel-value">{{ item.value }}</div>
                <div class="progress-card-hint metric-panel-helper">{{ item.hint }}</div>
              </article>
            </div>

            <div v-if="error" class="workspace-alert" role="alert" aria-live="polite">
              <div class="workspace-alert-title-row">
                <AlertTriangle class="workspace-alert-icon" />
                <div class="workspace-alert-title">教师概览加载失败</div>
              </div>
              <div class="workspace-alert-copy">{{ error }}</div>
              <div class="workspace-alert-copy">
                可先重试刷新数据，再继续查看趋势与复盘信息；若持续失败，可先进入班级管理确认当前班级与权限状态。
              </div>
              <div class="workspace-alert-actions">
                <button
                  type="button"
                  class="quick-action quick-action--compact"
                  @click="emit('retry')"
                >
                  <span>重试加载</span><span>→</span>
                </button>
                <button
                  type="button"
                  class="quick-action quick-action--compact"
                  @click="emit('openClassManagement')"
                >
                  <span>班级管理</span><span>→</span>
                </button>
              </div>
            </div>
          </div>

          <aside class="hero-rail">
            <div class="rail-label">Class Pulse</div>
            <div class="rail-score">
              {{ activeRateText.replace('%', '') }}
              <small v-if="activeRateText !== '--'">% active</small>
            </div>
            <div class="rail-copy">
              {{
                riskStudentCount > 0
                  ? `当前仍有 ${riskStudentCount} 名学生需要优先回流训练节奏，建议先投放低门槛补训题，再通过复盘结论安排课堂干预。`
                : '班级整体节奏稳定，可以把更多注意力放在薄弱维度补强和中段学生进阶上。'
              }}
            </div>
          </aside>
        </section>

        <section
          id="portrait"
          class="section tab-panel"
          :class="{ active: activeTab === 'portrait' }"
          role="tabpanel"
          aria-labelledby="top-tab-portrait"
          :aria-hidden="activeTab === 'portrait' ? 'false' : 'true'"
          v-show="activeTab === 'portrait'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Skill Portrait</div>
              <h2 class="workspace-tab-heading__title">能力画像与薄弱维度</h2>
            </div>
          </div>

          <div class="portrait-grid">
            <div class="portrait-summary-block">
              <h3 class="panel-title">优先补强方向</h3>

              <div v-if="weakDimensionStats.length > 0" class="weak-list">
                <article
                  v-for="(item, index) in weakDimensionStats.slice(0, 3)"
                  :key="item.dimension"
                  class="weak-item"
                >
                  <div class="weak-rank">{{ `${index + 1}`.padStart(2, '0') }}</div>
                  <div>
                    <div class="weak-name">{{ item.dimension }}</div>
                    <div class="weak-copy">
                      {{
                        item.count
                      }}
                      名学生当前在该方向暴露弱项，建议优先投放基础题并安排一次路径梳理。
                    </div>
                  </div>
                  <div class="weak-score">{{ item.count }} 人</div>
                </article>
              </div>
              <div v-else class="empty-inline">当前班级还没有足够的能力画像数据。</div>

              <div class="summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
                <article
                  v-for="item in portraitSummaryNotes"
                  :key="item.key"
                  class="summary-note progress-card metric-panel-card"
                >
                  <div class="summary-note-label progress-card-label metric-panel-label">{{ item.label }}</div>
                  <div class="summary-note-value progress-card-value metric-panel-value">{{ item.value }}</div>
                  <div class="summary-note-copy progress-card-hint metric-panel-helper">{{ item.copy }}</div>
                </article>
              </div>
            </div>

            <div class="workspace-subpanel">
              <TeacherClassInsightsPanel
                :students="students"
                :class-name="selectedClassName"
                split-cards
              />
            </div>
          </div>
        </section>

        <section
          id="trend"
          class="section tab-panel"
          :class="{ active: activeTab === 'trend' }"
          role="tabpanel"
          aria-labelledby="top-tab-trend"
          :aria-hidden="activeTab === 'trend' ? 'false' : 'true'"
          v-show="activeTab === 'trend'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Trend Review</div>
              <h2 class="workspace-tab-heading__title">近 7 天训练趋势</h2>
            </div>
          </div>

          <div class="trend-layout">
            <div class="workspace-subpanel">
              <TeacherClassTrendPanel :trend="trend" title="班级近 7 天训练趋势" subtitle="" bare />
            </div>

            <aside class="trend-side">
              <article v-for="item in trendSignals" :key="item.key" class="trend-signal">
                <div class="trend-signal-label">{{ item.label }}</div>
                <div class="trend-signal-value">{{ item.value }}</div>
                <div class="trend-signal-copy">{{ item.copy }}</div>
              </article>
            </aside>
          </div>
        </section>

        <section
          id="insight"
          class="section tab-panel"
          :class="{ active: activeTab === 'insight' }"
          role="tabpanel"
          aria-labelledby="top-tab-insight"
          :aria-hidden="activeTab === 'insight' ? 'false' : 'true'"
          v-show="activeTab === 'insight'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Student Insight</div>
              <h2 class="workspace-tab-heading__title">学生洞察</h2>
            </div>
          </div>

          <article class="panel panel-pad">
            <div class="insight-list">
              <div v-for="item in studentInsightRows" :key="item.key" class="insight-item">
                <div>
                  <strong>{{ item.title }}</strong>
                  <div class="insight-meta">
                    <span v-for="chip in item.chips" :key="chip" class="chip" :class="item.tone">
                      {{ chip }}
                    </span>
                  </div>
                  <div class="item-copy">{{ item.detail }}</div>
                </div>
                <div class="status-pill" :class="item.tone">{{ item.status }}</div>
              </div>
            </div>
          </article>
        </section>

        <section
          id="advice"
          class="section tab-panel"
          :class="{ active: activeTab === 'advice' }"
          role="tabpanel"
          aria-labelledby="top-tab-advice"
          :aria-hidden="activeTab === 'advice' ? 'false' : 'true'"
          v-show="activeTab === 'advice'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Today Focus</div>
              <h2 class="workspace-tab-heading__title">今日教学建议</h2>
            </div>
          </div>

          <article class="panel panel-pad">
            <div class="advice-lines">
              <div v-for="(item, index) in teachingAdvice" :key="item.title" class="hint-line">
                <div class="hint-index">{{ index + 1 }}</div>
                <div>
                  <div class="hint-label">{{ item.title }}</div>
                  <div class="hint-copy">{{ item.detail }}</div>
                </div>
              </div>
            </div>
          </article>

          <div class="workspace-subpanel section-stack">
            <TeacherClassReviewPanel :review="review" :class-name="selectedClassName" />
          </div>
        </section>

        <section
          id="action"
          class="section tab-panel"
          :class="{ active: activeTab === 'action' }"
          role="tabpanel"
          aria-labelledby="top-tab-action"
          :aria-hidden="activeTab === 'action' ? 'false' : 'true'"
          v-show="activeTab === 'action'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Intervention Board</div>
              <h2 class="workspace-tab-heading__title">介入建议</h2>
            </div>
          </div>

          <div class="workspace-subpanel section-stack">
            <TeacherInterventionPanel :students="students" :class-name="selectedClassName" />
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-line-strong: color-mix(in srgb, var(--color-text-primary) 16%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-success: var(--color-success);
  --workspace-warning: var(--color-warning);
  --workspace-danger: var(--color-danger);
  --workspace-shadow-shell: 0 24px 84px
    color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-xl: 28px;
  --workspace-radius-lg: 18px;
  --workspace-radius-md: 14px;
  --workspace-font-sans: var(--font-family-sans);
  --workspace-font-mono: var(--font-family-mono);
}

.tab-panel.section {
  padding-top: 0;
  border-top: 0;
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 244px;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.tab-panel.workspace-hero.active {
  display: grid;
}

.hero-title {
  max-width: 11ch;
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-4-5);
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 var(--space-2-5);
  border: 1px solid var(--workspace-line-soft);
  border-radius: 8px;
  background: color-mix(in srgb, var(--workspace-panel) 72%, transparent);
  font-size: var(--font-size-12);
  color: var(--journal-muted);
}

.meta-pill.brand {
  border-color: color-mix(in srgb, var(--workspace-brand) 20%, transparent);
  background: var(--workspace-brand-soft);
  color: var(--workspace-brand-ink);
}

.progress-strip {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
  --metric-panel-grid-gap: var(--space-3);
  margin-top: var(--space-5-5);
}

.panel,
.trend-signal {
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
}

.trend-signal-label {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.trend-signal-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.quick-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2-5);
  min-height: 52px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--workspace-line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--workspace-panel) 82%, transparent);
  color: var(--journal-ink);
  text-decoration: none;
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.quick-action span:last-child {
  color: var(--workspace-faint);
}

.quick-action:hover,
.quick-action:focus-visible {
  border-color: color-mix(in srgb, var(--workspace-brand) 34%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 8%, var(--workspace-panel));
  color: var(--workspace-brand-ink);
  outline: none;
}

.quick-action--compact {
  min-height: 42px;
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

.hero-rail {
  padding-left: var(--space-6);
  border-left: 1px solid var(--workspace-line-soft);
}

.rail-label {
  font-size: var(--font-size-11);
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.rail-score {
  margin-top: var(--space-2-5);
  font: 700 38px/1 var(--workspace-font-mono);
  color: var(--journal-ink);
}

.rail-score small {
  margin-left: var(--space-1);
  font-size: var(--font-size-15);
  color: var(--workspace-faint);
}

.rail-copy {
  margin-top: var(--space-3-5);
  padding-top: var(--space-3-5);
  border-top: 1px solid var(--workspace-line-soft);
  font-size: var(--font-size-14);
  line-height: 1.78;
  color: var(--journal-muted);
}

.panel-title {
  margin: 0;
  font-size: var(--font-size-18);
  line-height: 1.2;
  color: var(--journal-ink);
}

.section {
  padding-top: var(--space-6);
  border-top: 1px solid var(--workspace-line-soft);
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.section-kicker {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--workspace-brand) 60%, var(--workspace-faint));
}

.section-title {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-22);
  line-height: 1.12;
  color: var(--journal-ink);
}

.portrait-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-5);
}

.portrait-summary-block {
  min-width: 0;
}

.weak-list {
  display: grid;
  gap: var(--space-3);
  margin-top: var(--space-4-5);
}

.weak-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: var(--space-3-5);
  align-items: start;
  padding: var(--space-3-5) 0;
  border-top: 1px dashed var(--workspace-line-soft);
}

.weak-item:first-child {
  padding-top: 0;
  border-top: 0;
}

.weak-rank,
.hint-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 12px;
  background: var(--workspace-brand-soft);
  font: 700 13px/1 var(--workspace-font-mono);
  color: var(--workspace-brand-ink);
}

.weak-name {
  font-size: var(--font-size-15);
  font-weight: 700;
  color: var(--journal-ink);
}

.weak-copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--journal-muted);
}

.weak-score {
  font: 600 13px/1 var(--workspace-font-mono);
  color: var(--workspace-faint);
}

.summary-grid {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
  --metric-panel-grid-gap: var(--space-3);
  margin-top: var(--space-4-5);
}

.trend-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 220px;
  gap: var(--space-4-5);
}

.trend-side {
  display: grid;
  gap: var(--space-3);
}

.trend-signal {
  padding: var(--space-4);
}

.trend-signal-value {
  margin-top: var(--space-2-5);
  font-size: var(--font-size-24);
  letter-spacing: -0.03em;
  color: var(--journal-ink);
}

.insight-list,
.action-list {
  display: grid;
  border-top: 1px solid var(--workspace-line-soft);
}

.insight-item,
.action-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4-5);
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--workspace-line-soft);
}

.insight-item strong,
.action-item strong {
  display: block;
  font-size: var(--font-size-15);
  color: var(--journal-ink);
}

.insight-meta,
.action-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.chip,
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 var(--space-2);
  border-radius: 7px;
  border: 1px solid var(--workspace-line-soft);
  font-size: var(--font-size-11-5);
  font-weight: 600;
  letter-spacing: 0.01em;
  color: var(--journal-muted);
}

.chip.ready,
.status-pill.ready {
  border-color: color-mix(in srgb, var(--workspace-success) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-success) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-success) 82%, var(--journal-ink));
}

.chip.warning,
.status-pill.warning {
  border-color: color-mix(in srgb, var(--workspace-warning) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-warning) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-warning) 86%, var(--journal-ink));
}

.chip.danger,
.status-pill.danger {
  border-color: color-mix(in srgb, var(--workspace-danger) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-danger) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.item-copy {
  margin-top: var(--space-2-5);
  font-size: var(--font-size-14);
  line-height: 1.8;
  color: var(--journal-muted);
}

.status-pill {
  min-height: 30px;
  min-width: 78px;
  border-radius: 8px;
}

.advice-lines {
  margin-top: var(--space-0-5);
}

.hint-line {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  padding: var(--space-3-5) 0;
  border-top: 1px dashed var(--workspace-line-soft);
}

.hint-line:first-of-type {
  padding-top: 0;
  border-top: 0;
}

.hint-index {
  width: 28px;
  height: 28px;
  border-radius: 10px;
  font-size: var(--font-size-12);
}

.hint-label {
  font-size: var(--font-size-14);
  font-weight: 600;
  color: var(--journal-ink);
}

.hint-copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--journal-muted);
}

.section-stack {
  margin-top: var(--space-4-5);
}

.workspace-alert {
  margin-top: var(--space-4-5);
  padding: var(--space-4) var(--space-4-5);
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
}

.workspace-alert-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2-5);
}

.workspace-alert-icon {
  width: 18px;
  height: 18px;
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.workspace-alert-title {
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-3-5);
}

.workspace-subpanel :deep(.teacher-panel) {
  border: 1px solid var(--workspace-line-soft);
  border-radius: 22px;
  background: color-mix(in srgb, var(--workspace-panel) 90%, transparent);
  box-shadow: var(--workspace-shadow-panel);
  padding: var(--space-5);
}

.workspace-subpanel :deep(.teacher-panel.teacher-panel--shellless) {
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}

.workspace-subpanel :deep(.teacher-panel__header),
.workspace-subpanel :deep(.teacher-subsection__header) {
  margin-bottom: var(--space-4);
}

.workspace-subpanel :deep(.journal-eyebrow) {
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--workspace-brand) 60%, var(--workspace-faint));
}

.workspace-subpanel :deep(.teacher-panel__title) {
  margin-top: var(--space-2-5);
  font-size: var(--font-size-22);
  line-height: 1.15;
  color: var(--journal-ink);
}

.workspace-subpanel :deep(.teacher-subsection + .teacher-subsection) {
  border-top-color: var(--workspace-line-soft);
}

.workspace-subpanel :deep(.top-student-item),
.workspace-subpanel :deep(.dimension-item),
.workspace-subpanel :deep(.review-item__recommendation),
.workspace-subpanel :deep(.review-item),
.workspace-subpanel :deep(.intervention-item) {
  border-color: var(--workspace-line-soft);
}

.workspace-subpanel :deep(.teacher-panel__chart) {
  border-color: var(--workspace-line-soft);
  background: color-mix(in srgb, var(--workspace-panel-soft) 82%, transparent);
}

.workspace-subpanel :deep(.review-item) {
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-panel-soft) 86%, transparent);
}

.workspace-subpanel :deep(.top-student-item__rank),
.workspace-subpanel :deep(.teacher-tip-index) {
  font-family: var(--workspace-font-mono);
}

.empty-inline {
  margin-top: var(--space-4-5);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--workspace-faint);
}

@media (max-width: 1180px) {
  .workspace-hero,
  .portrait-grid,
  .trend-layout {
    grid-template-columns: 1fr;
  }

  .hero-rail {
    padding-top: var(--space-5);
    padding-left: 0;
    border-top: 1px solid var(--workspace-line-soft);
    border-left: 0;
  }
}

@media (max-width: 860px) {
  .progress-strip,
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .section-head {
    display: block;
  }
}

@media (max-width: 640px) {
  .workspace-topbar,
  .top-tabs,
  .content-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }

  .workspace-topbar {
    display: block;
  }

  .top-note {
    justify-content: flex-start;
    margin-top: var(--space-3);
  }

  .progress-strip,
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>

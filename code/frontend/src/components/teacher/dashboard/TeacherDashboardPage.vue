<script setup lang="ts">
import { toRef } from 'vue'
import { AlertTriangle, Activity } from 'lucide-vue-next'

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
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } = useUrlSyncedTabs<WorkspaceTab>({
  orderedTabs: workspaceTabOrder,
  defaultTab: 'overview',
})

const {
  activeRateText,
  riskStudentCount,
  overviewDescription,
  metaPills,
  overviewMetrics,
  teachingAdvice,
  studentInsightRows,
  portraitSummaryNotes,
  trendSignals,
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
    <nav
      class="workspace-tabbar top-tabs teacher-dashboard-tabs"
      role="tablist"
      aria-label="教学概览标签页"
    >
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
      <main class="content-pane teacher-dashboard-content">
        <section
          v-show="activeTab === 'overview'"
          id="overview"
          class="workspace-hero teacher-dashboard-hero tab-panel active"
          role="tabpanel"
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
        </section>

        <section
          v-show="activeTab === 'portrait'"
          id="portrait"
          class="section tab-panel active"
          role="tabpanel"
        >
          <div class="section-head">
            <div class="teacher-heading">
              <div class="section-kicker">
                Skill Portrait
              </div>
              <h2 class="section-title">
                能力画像与薄弱维度
              </h2>
            </div>
          </div>

          <div class="portrait-grid">
            <div class="portrait-summary-block">
              <h3 class="panel-title">
                优先补强方向
              </h3>

              <div
                v-if="weakDimensionStats.length > 0"
                class="weak-list"
              >
                <article
                  v-for="(item, index) in weakDimensionStats.slice(0, 3)"
                  :key="item.dimension"
                  class="weak-item"
                >
                  <div class="weak-rank">
                    {{ `${index + 1}`.padStart(2, '0') }}
                  </div>
                  <div>
                    <div class="weak-name">
                      {{ item.dimension }}
                    </div>
                    <div class="weak-copy">
                      {{ item.count }} 名学生当前在该方向暴露弱项。
                    </div>
                  </div>
                  <div class="weak-score">
                    {{ item.count }} 人
                  </div>
                </article>
              </div>

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
                    当前画像摘要
                  </div>
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
          v-show="activeTab === 'trend'"
          id="trend"
          class="section tab-panel active"
          role="tabpanel"
        >
          <div class="workspace-subpanel">
            <header class="workspace-tab-heading">
              <div class="workspace-overline">
                Trend Review
              </div>
              <h2 class="workspace-tab-heading__title">
                趋势复盘
              </h2>
            </header>
            <TeacherClassTrendPanel
              :trend="trend"
              title="班级近 7 天训练趋势"
              bare
            />
          </div>
        </section>

        <section
          v-show="activeTab === 'insight'"
          id="insight"
          class="section tab-panel active"
          role="tabpanel"
        >
          <div class="workspace-subpanel">
            <header class="workspace-tab-heading">
              <div class="workspace-overline">
                Student Insight
              </div>
              <h2 class="workspace-tab-heading__title">
                学生洞察
              </h2>
            </header>
            <TeacherClassInsightsPanel
              :students="students"
              :class-name="selectedClassName"
              split-cards
            />
          </div>
        </section>

        <section
          v-show="activeTab === 'advice'"
          id="advice"
          class="section tab-panel active"
          role="tabpanel"
        >
          <div class="workspace-subpanel">
            <header class="workspace-tab-heading">
              <div class="workspace-overline">
                Teaching Advice
              </div>
              <h2 class="workspace-tab-heading__title">
                今日教学建议
              </h2>
            </header>
            <TeacherClassReviewPanel
              :review="review"
              :class-name="selectedClassName"
            />
          </div>
        </section>

        <section
          v-show="activeTab === 'action'"
          id="action"
          class="section tab-panel active"
          role="tabpanel"
        >
          <div class="workspace-subpanel">
            <header class="workspace-tab-heading">
              <div class="workspace-overline">
                Intervention
              </div>
              <h2 class="workspace-tab-heading__title">
                介入建议
              </h2>
            </header>
            <TeacherInterventionPanel
              :students="students"
              :class-name="selectedClassName"
            />
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
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: 0;
  --page-top-tabs-padding: 0 var(--space-workspace-side-padding);
  --page-top-tabs-border: var(--workspace-line-soft);
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

.section {
  padding-top: var(--space-8);
  border-top: 1px solid var(--workspace-line-soft);
}

.section-head {
  margin-bottom: var(--space-5);
}

.section-kicker {
  font-size: var(--font-size-11);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--journal-muted);
  letter-spacing: 0.1em;
}

.section-title {
  margin-top: var(--space-1);
  font-size: var(--font-size-22);
  font-weight: 900;
  color: var(--journal-ink);
}

.portrait-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.35fr);
  gap: var(--space-5);
  align-items: start;
}

.portrait-summary-block {
  display: grid;
  gap: var(--space-4);
}

.panel-title {
  margin: 0;
  font-size: var(--font-size-17);
  font-weight: 800;
  color: var(--journal-ink);
}

.weak-list {
  display: grid;
  gap: var(--space-3);
}

.weak-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: var(--space-3);
  align-items: center;
  padding: var(--space-4);
  border: 1px solid var(--teacher-card-border);
  border-radius: var(--workspace-radius-md, 14px);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.weak-rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 12px;
  background: var(--workspace-brand-soft);
  font: 700 var(--font-size-13) / 1 var(--font-family-mono);
  color: var(--journal-accent-strong);
}
.weak-name {
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

.quick-action {
  display: flex; align-items: center; justify-content: space-between; gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  border: 1px solid var(--teacher-card-border);
  border-radius: 12px;
  background: var(--journal-surface);
  color: var(--journal-ink);
  cursor: pointer;
  transition: all 0.2s ease;
}
.quick-action:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 34%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, var(--journal-surface));
}

@media (max-width: 1180px) {
  .teacher-dashboard-hero,
  .portrait-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .teacher-dashboard-shell {
    --page-top-tabs-gap: var(--space-5);
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
}
</style>

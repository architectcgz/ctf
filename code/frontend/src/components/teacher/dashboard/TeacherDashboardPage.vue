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
  <div class="workspace-shell teacher-management-shell teacher-surface">
    <nav
      class="top-tabs"
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
      <main class="content-pane">
        <section
          v-show="activeTab === 'overview'"
          id="overview"
          class="workspace-hero tab-panel active"
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

            <div class="progress-strip metric-panel-grid">
              <article
                v-for="item in overviewMetrics"
                :key="item.key"
                class="progress-card metric-panel-card"
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

          <aside class="hero-rail">
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

              <div
                class="summary-grid progress-strip metric-panel-grid"
              >
                <article
                  v-for="item in portraitSummaryNotes"
                  :key="item.key"
                  class="progress-card metric-panel-card"
                >
                  <div class="metric-panel-label">
                    {{ item.label }}
                  </div>
                  <div class="metric-panel-value">
                    {{ item.value }}
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

        <!-- Other tabs handled by subpanels similarly... -->
      </main>
    </div>
  </div>
</template>

<style scoped>
@import '../teacher-workspace-subpanel.css';

.workspace-shell {
  --workspace-line-soft: var(--color-border-subtle);
  --workspace-panel: var(--color-bg-surface);
}

.hero-summary {
  max-width: 760px; margin-top: var(--space-4); font-size: var(--font-size-15); line-height: 1.8; color: var(--color-text-secondary);
}

.meta-pill {
  display: inline-flex; align-items: center; min-height: 28px; padding: 0 var(--space-3);
  border: 1px solid var(--color-border-default); border-radius: 8px;
  background: var(--color-bg-elevated); font-size: var(--font-size-12); color: var(--color-text-secondary);
}
.meta-pill.brand {
  border-color: var(--color-primary); background: var(--color-primary-soft); color: var(--color-primary);
}

.hero-rail { padding-left: var(--space-8); border-left: 1px solid var(--color-border-subtle); }
.rail-score { margin-top: var(--space-3); font: 900 38px/1 var(--font-family-mono); color: var(--color-text-primary); }

.section { padding-top: var(--space-8); border-top: 1px solid var(--color-border-subtle); }
.section-kicker { font-size: var(--font-size-11); font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); letter-spacing: 0.1em; }
.section-title { margin-top: 0.25rem; font-size: var(--font-size-22); font-weight: 900; color: var(--color-text-primary); }

.weak-rank {
  display: inline-flex; align-items: center; justify-content: center; width: 34px; height: 34px; border-radius: 12px;
  background: var(--color-primary-soft); font: 700 13px/1 var(--font-family-mono); color: var(--color-primary);
}
.weak-name { font-size: var(--font-size-15); font-weight: 800; color: var(--color-text-primary); }

.quick-action {
  display: flex; align-items: center; justify-content: space-between; gap: var(--space-4);
  padding: 1rem 1.25rem; border: 1px solid var(--color-border-default); border-radius: 12px;
  background: var(--color-bg-surface); color: var(--color-text-primary); cursor: pointer; transition: all 0.2s ease;
}
.quick-action:hover {
  border-color: var(--color-primary); background: var(--color-bg-elevated);
}

@media (max-width: 1180px) {
  .workspace-hero { grid-template-columns: 1fr; }
  .hero-rail { padding-left: 0; padding-top: var(--space-6); border-left: 0; border-top: 1px solid var(--color-border-subtle); }
}
</style>
<script setup lang="ts">
import { toRef } from 'vue'
import { AlertTriangle } from 'lucide-vue-next'

import type { TeacherOverviewData } from '@/api/contracts'
import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'
import { useDashboardMetrics } from '../model/useDashboardMetrics'
import type { TeacherDashboardPanelKey } from '../model'
import TeacherDashboardInterventionPanel from './TeacherDashboardInterventionPanel.vue'
import TeacherDashboardPortraitPanel from './TeacherDashboardPortraitPanel.vue'
import TeacherDashboardReviewPanel from './TeacherDashboardReviewPanel.vue'
import TeacherDashboardStudentInsightPanel from './TeacherDashboardStudentInsightPanel.vue'
import TeacherDashboardTrendPanel from './TeacherDashboardTrendPanel.vue'

interface DashboardRouteTarget {
  name: string
}

const props = defineProps<{
  overview: TeacherOverviewData | null
  error: string | null
  classManagementRoute: DashboardRouteTarget
  activePanel: TeacherDashboardPanelKey
}>()

const emit = defineEmits<{
  retry: []
  switchPanel: [panel: TeacherDashboardPanelKey]
}>()

const dashboardTabs: Array<{
  key: TeacherDashboardPanelKey
  label: string
  buttonId: string
  panelId: string
}> = [
  { key: 'overview', label: '进度总览', buttonId: 'dashboard-tab-overview', panelId: 'overview' },
  { key: 'portrait', label: '能力画像', buttonId: 'dashboard-tab-portrait', panelId: 'portrait' },
  { key: 'insight', label: '学生洞察', buttonId: 'dashboard-tab-insight', panelId: 'insight' },
  { key: 'trend', label: '趋势复盘', buttonId: 'dashboard-tab-trend', panelId: 'trend' },
  { key: 'review', label: '教学复盘', buttonId: 'dashboard-tab-review', panelId: 'review' },
  {
    key: 'intervention',
    label: '介入建议',
    buttonId: 'dashboard-tab-intervention',
    panelId: 'intervention',
  },
]

const dashboardTabOrder = dashboardTabs.map((tab) => tab.key) as TeacherDashboardPanelKey[]
const { setTabButtonRef, handleTabKeydown } = useTabKeyboardNavigation<TeacherDashboardPanelKey>({
  orderedTabs: dashboardTabOrder,
  selectTab: (tab) => emit('switchPanel', tab),
})

const {
  overviewDescription,
  metaPills,
  overviewMetrics,
  studentInsightRows,
  portraitSummaryNotes,
  weakDimensionStats,
  focusClasses,
  trendSignals,
  reviewHighlights,
  interventionTargets,
} = useDashboardMetrics({
  overview: toRef(props, 'overview'),
})
</script>

<template>
  <div
    class="workspace-shell teacher-management-shell teacher-surface teacher-dashboard-shell flex min-h-full flex-1 flex-col"
  >
    <nav class="workspace-tabbar top-tabs" role="tablist" aria-label="教学概览标签页">
      <button
        v-for="(tab, index) in dashboardTabs"
        :id="tab.buttonId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        class="workspace-tab top-tab"
        :class="{ active: activePanel === tab.key }"
        type="button"
        role="tab"
        :tabindex="activePanel === tab.key ? 0 : -1"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        @click="emit('switchPanel', tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane teacher-dashboard-content">
      <section
        v-show="activePanel === 'overview'"
        id="overview"
        class="tab-panel teacher-dashboard-overview"
        :class="{ active: activePanel === 'overview' }"
        role="tabpanel"
        aria-labelledby="dashboard-tab-overview"
        :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
      >
        <header class="workspace-panel-header teacher-dashboard-overview-head">
          <div class="workspace-panel-header__intro">
            <div class="workspace-overline">
              Teaching Overview
            </div>
            <h1 class="hero-title workspace-page-title">教学介入台</h1>
            <p class="workspace-page-copy hero-summary">
              {{ overviewDescription }}
            </p>
          </div>

          <div class="workspace-panel-header__meta meta-strip">
            <span
              v-for="(pill, index) in metaPills"
              :key="pill"
              class="meta-pill"
              :class="{ brand: index === 0 }"
            >
              {{ pill }}
            </span>
          </div>

          <div class="workspace-panel-header__actions header-actions">
            <AppRouteLink :to="classManagementRoute" class="header-btn header-btn--primary">
              班级管理
            </AppRouteLink>
          </div>

          <div
            class="workspace-panel-header__summary teacher-overview-summary progress-strip metric-panel-grid metric-panel-default-surface"
          >
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
        </header>

        <div v-if="error" class="workspace-alert">
          <div class="workspace-alert-title-row">
            <AlertTriangle class="workspace-alert-icon h-4 w-4" />
            <div class="workspace-alert-title">加载失败</div>
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
      </section>

      <section
        v-show="activePanel === 'portrait'"
        id="portrait"
        class="tab-panel"
        :class="{ active: activePanel === 'portrait' }"
        role="tabpanel"
        aria-labelledby="dashboard-tab-portrait"
        :aria-hidden="activePanel === 'portrait' ? 'false' : 'true'"
      >
        <TeacherDashboardPortraitPanel
          :portrait-summary-notes="portraitSummaryNotes"
          :weak-dimension-stats="weakDimensionStats"
        />
      </section>

      <section
        v-show="activePanel === 'insight'"
        id="insight"
        class="tab-panel"
        :class="{ active: activePanel === 'insight' }"
        role="tabpanel"
        aria-labelledby="dashboard-tab-insight"
        :aria-hidden="activePanel === 'insight' ? 'false' : 'true'"
      >
        <TeacherDashboardStudentInsightPanel :student-insight-rows="studentInsightRows" />
      </section>

      <section
        v-show="activePanel === 'trend'"
        id="trend"
        class="tab-panel"
        :class="{ active: activePanel === 'trend' }"
        role="tabpanel"
        aria-labelledby="dashboard-tab-trend"
        :aria-hidden="activePanel === 'trend' ? 'false' : 'true'"
      >
        <TeacherDashboardTrendPanel
          :trend-signals="trendSignals"
          :focus-classes="focusClasses"
        />
      </section>

      <section
        v-show="activePanel === 'review'"
        id="review"
        class="tab-panel"
        :class="{ active: activePanel === 'review' }"
        role="tabpanel"
        aria-labelledby="dashboard-tab-review"
        :aria-hidden="activePanel === 'review' ? 'false' : 'true'"
      >
        <TeacherDashboardReviewPanel :review-highlights="reviewHighlights" />
      </section>

      <section
        v-show="activePanel === 'intervention'"
        id="intervention"
        class="tab-panel"
        :class="{ active: activePanel === 'intervention' }"
        role="tabpanel"
        aria-labelledby="dashboard-tab-intervention"
        :aria-hidden="activePanel === 'intervention' ? 'false' : 'true'"
      >
        <TeacherDashboardInterventionPanel :intervention-targets="interventionTargets" />
      </section>
    </main>
  </div>
</template>

<style scoped>
@import '../../../assets/styles/teacher-workspace-subpanel.css';

.teacher-dashboard-shell {
  --journal-ink: var(--color-text-primary);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--color-border-default) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--color-border-default) 78%, transparent);
  --header-control-border: var(--teacher-control-border);
  --teacher-divider: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --workspace-line-soft: var(--color-border-subtle);
  --workspace-panel: var(--color-bg-surface);
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: var(--journal-accent-strong);
  --workspace-brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
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

.teacher-dashboard-overview {
  display: grid;
  gap: var(--space-5);
}

.teacher-dashboard-overview-head {
  --workspace-panel-header-block-gap: var(--space-5);
}

.hero-summary {
  max-width: 760px;
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
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

.summary-grid {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.teacher-overview-card {
  min-height: 7.75rem;
}

@media (max-width: 760px) {
  .teacher-dashboard-shell {
    --metric-panel-columns: 1fr;
  }
}
</style>

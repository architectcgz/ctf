<template>
  <div class="workspace-shell teacher-management-shell teacher-surface">
    <nav class="top-tabs" role="tablist" aria-label="班级详情标签页">
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

    <main class="content-pane">
      <ClassStudentsInsightWindowPanel
        :insight-window-from-date="props.insightWindowFromDate"
        :insight-window-to-date="props.insightWindowToDate"
        :insight-window-error="props.insightWindowError"
        :insight-window-label="props.insightWindowLabel"
        :can-apply-insight-window="props.canApplyInsightWindow"
        :can-reset-insight-window="props.canResetInsightWindow"
        @update-insight-window-from-date="emit('updateInsightWindowFromDate', $event)"
        @update-insight-window-to-date="emit('updateInsightWindowToDate', $event)"
        @apply-insight-window="emit('applyInsightWindow')"
        @reset-insight-window="emit('resetInsightWindow')"
      />

      <section
        v-show="activeTab === 'overview'"
        id="class-overview"
        class="tab-panel section active"
        :class="{ active: activeTab === 'overview' }"
        role="tabpanel"
        aria-labelledby="class-tab-overview"
        :aria-hidden="activeTab === 'overview' ? 'false' : 'true'"
      >
        <ClassStudentsOverviewPanel
          :selected-class-name="selectedClassName"
          :students="students"
          :summary="summary"
          :error="error"
          @retry="emit('retry')"
          @open-class-management="emit('openClassManagement')"
          @open-dashboard="emit('openDashboard')"
          @open-report-export="emit('openReportExport')"
        />
      </section>

      <section
        v-show="activeTab === 'students'"
        id="class-students"
        class="tab-panel section"
        :class="{ active: activeTab === 'students' }"
        role="tabpanel"
        aria-labelledby="class-tab-students"
        :aria-hidden="activeTab === 'students' ? 'false' : 'true'"
      >
        <ClassStudentsDirectoryPanel
          :students="students"
          :student-no-query="studentNoQuery"
          :loading-students="loadingStudents"
          @update-student-no-query="emit('updateStudentNoQuery', $event)"
          @open-student="emit('openStudent', $event)"
        />
      </section>

      <section
        v-for="tab in panelWorkspaceTabs"
        v-show="activeTab === tab.key"
        :id="tab.panelId"
        :key="tab.panelId"
        class="tab-panel section active"
        :class="{ active: activeTab === tab.key }"
        role="tabpanel"
      >
        <div :class="resolveWorkspacePanelWrapperClass(tab.key)">
          <component
            :is="resolveWorkspacePanelComponent(tab.key)"
            v-bind="resolveWorkspacePanelProps(tab.key)"
            @open-student="emit('openStudent', $event)"
          />
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
@import '../../../assets/styles/teacher-workspace-subpanel.css';

.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --header-control-border: var(--teacher-control-border);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-directory-section {
  margin-top: var(--workspace-directory-page-block-gap, var(--space-5));
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

@media (max-width: 1080px) {
  .teacher-page {
    min-height: auto;
  }
}
</style>

<script setup lang="ts">
import { type Component } from 'vue'

import type {
  ClassInsightReviewData,
  ClassInsightTrendData,
  StudentDirectoryItem,
} from '@/api/contracts'
import ClassInsightsPanel from './ClassInsightsPanel.vue'
import ClassReviewPanel from './ClassReviewPanel.vue'
import ClassTrendPanel from './ClassTrendPanel.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import { InterventionPanel } from '@/features/teaching/student-analysis-review'
import ClassStudentsDirectoryPanel from './ClassStudentsDirectoryPanel.vue'
import ClassStudentsInsightWindowPanel from './ClassStudentsInsightWindowPanel.vue'
import ClassStudentsOverviewPanel from './ClassStudentsOverviewPanel.vue'

const props = defineProps<{
  selectedClassName: string
  students: StudentDirectoryItem[]
  review: ClassInsightReviewData | null
  summary: import('@/api/contracts').ClassInsightSummaryData | null
  trend: ClassInsightTrendData | null
  studentNoQuery: string
  loadingStudents: boolean
  error: string | null
  insightWindowFromDate: string
  insightWindowToDate: string
  insightWindowError: string | null
  insightWindowLabel: string
  canApplyInsightWindow: boolean
  canResetInsightWindow: boolean
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openDashboard: []
  openReportExport: []
  updateStudentNoQuery: [value: string]
  updateInsightWindowFromDate: [value: string]
  updateInsightWindowToDate: [value: string]
  applyInsightWindow: []
  resetInsightWindow: []
  openStudent: [studentId: string]
}>()

type WorkspaceTab = 'overview' | 'trend' | 'students' | 'review' | 'insight' | 'action'
type WorkspacePanelTab = Exclude<WorkspaceTab, 'overview' | 'students'>

interface WorkspaceTabItem {
  key: WorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

interface WorkspacePanelTabItem extends WorkspaceTabItem {
  key: WorkspacePanelTab
}

const workspaceTabs: WorkspaceTabItem[] = [
  { key: 'overview', label: '主看板', buttonId: 'class-tab-overview', panelId: 'class-overview' },
  { key: 'trend', label: '趋势复盘', buttonId: 'class-tab-trend', panelId: 'class-trend' },
  { key: 'students', label: '学生列表', buttonId: 'class-tab-students', panelId: 'class-students' },
  { key: 'review', label: '复盘结论', buttonId: 'class-tab-review', panelId: 'class-review' },
  { key: 'insight', label: '学生洞察', buttonId: 'class-tab-insight', panelId: 'class-insight' },
  { key: 'action', label: '介入建议', buttonId: 'class-tab-action', panelId: 'class-action' },
]

const workspaceTabOrder = workspaceTabs.map((tab) => tab.key) as WorkspaceTab[]
const panelWorkspaceTabs = workspaceTabs.filter(
  (tab): tab is WorkspacePanelTabItem => tab.key !== 'overview' && tab.key !== 'students'
)
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } =
  useUrlSyncedTabs<WorkspaceTab>({
    orderedTabs: workspaceTabOrder,
    defaultTab: 'overview',
  })

function resolveWorkspacePanelComponent(tabKey: WorkspacePanelTab): Component {
  switch (tabKey) {
    case 'trend':
      return ClassTrendPanel
    case 'review':
      return ClassReviewPanel
    case 'insight':
      return ClassInsightsPanel
    case 'action':
      return InterventionPanel
  }
}

function resolveWorkspacePanelProps(tabKey: WorkspacePanelTab): Record<string, unknown> {
  switch (tabKey) {
    case 'trend':
      return {
        trend: props.trend,
        title: '班级训练趋势',
        subtitle: `当前窗口：${props.insightWindowLabel}`,
      }
    case 'review':
      return {
        review: props.review,
        className: props.selectedClassName,
      }
    case 'insight':
      return {
        students: props.students,
        className: props.selectedClassName,
        splitCards: true,
      }
    case 'action':
      return {
        students: props.students,
        className: props.selectedClassName,
      }
  }
}

function resolveWorkspacePanelWrapperClass(tabKey: WorkspacePanelTab): string[] {
  return tabKey === 'insight'
    ? ['workspace-subpanel', 'workspace-subpanel--flat', 'workspace-subpanel--insight']
    : ['workspace-subpanel', 'workspace-subpanel--flat']
}
</script>

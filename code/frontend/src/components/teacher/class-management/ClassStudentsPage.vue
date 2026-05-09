<script setup lang="ts">
import { Activity, ArrowRight, ChevronLeft, Search, Target, Users } from 'lucide-vue-next'
import { computed, type Component } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  students: TeacherStudentItem[]
  review: TeacherClassReviewData | null
  summary: TeacherClassSummaryData | null
  trend: TeacherClassTrendData | null
  studentNoQuery: string
  loadingStudents: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openDashboard: []
  openReportExport: []
  selectClass: [className: string]
  updateStudentNoQuery: [value: string]
  openStudent: [studentId: string]
}>()

const averageSolvedText = computed(() => {
  if (!props.summary) return '--'
  return props.summary.average_solved.toFixed(1)
})

const activeRateText = computed(() => {
  if (!props.summary) return '--'
  return `${Math.round(props.summary.active_rate)}%`
})

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
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } = useUrlSyncedTabs<WorkspaceTab>({
  orderedTabs: workspaceTabOrder,
  defaultTab: 'overview',
})

function resolveWorkspacePanelComponent(tabKey: WorkspacePanelTab): Component {
  switch (tabKey) {
    case 'trend':
      return TeacherClassTrendPanel
    case 'review':
      return TeacherClassReviewPanel
    case 'insight':
      return TeacherClassInsightsPanel
    case 'action':
      return TeacherInterventionPanel
  }
}

function resolveWorkspacePanelProps(tabKey: WorkspacePanelTab): Record<string, unknown> {
  switch (tabKey) {
    case 'trend':
      return {
        trend: props.trend,
        title: '班级近 7 天训练趋势',
        subtitle: '先看整体节奏，再下钻到具体学生。',
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

    <div class="workspace-grid">
      <main class="content-pane">
        <section
          v-show="activeTab === 'overview'"
          id="class-overview"
          class="tab-panel section active"
          :class="{ active: activeTab === 'overview' }"
          role="tabpanel"
          aria-labelledby="class-tab-overview"
          :aria-hidden="activeTab === 'overview' ? 'false' : 'true'"
        >
          <header class="teacher-topbar">
            <div class="teacher-heading">
              <section class="teacher-summary">
                <div class="teacher-summary-title">
                  <span>Class Snapshot</span>
                </div>
                <div
                  class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"
                >
                  <div class="progress-card metric-panel-card">
                    <div class="progress-card-label metric-panel-label">
                      <span>班级人数</span>
                      <Users class="h-4 w-4" />
                    </div>
                    <div class="progress-card-value metric-panel-value">
                      {{ props.summary?.student_count ?? students.length }}
                    </div>
                  </div>
                  <div class="progress-card metric-panel-card">
                    <div class="progress-card-label metric-panel-label">
                      <span>平均解题</span>
                      <Target class="h-4 w-4" />
                    </div>
                    <div class="progress-card-value metric-panel-value">
                      {{ averageSolvedText }}
                    </div>
                  </div>
                  <div class="progress-card metric-panel-card">
                    <div class="progress-card-label metric-panel-label">
                      <span>近 7 天活跃率</span>
                      <Activity class="h-4 w-4" />
                    </div>
                    <div class="progress-card-value metric-panel-value">
                      {{ activeRateText }}
                    </div>
                  </div>
                </div>
              </section>
            </div>

            <div class="teacher-actions">
              <button
                type="button"
                class="ui-btn ui-btn--secondary"
                @click="emit('openClassManagement')"
              >
                返回
              </button>
              <button type="button" class="ui-btn ui-btn--secondary" @click="emit('openDashboard')">
                概览
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="emit('openReportExport')"
              >
                导出班级报告
              </button>
            </div>
          </header>

          <div v-if="error" class="workspace-alert" role="alert">
            <div class="workspace-alert-title">加载失败</div>
            <div class="workspace-alert-copy">
              {{ error }}
            </div>
            <div class="workspace-alert-actions">
              <button type="button" class="ui-btn ui-btn--primary" @click="emit('retry')">
                重试加载
              </button>
            </div>
          </div>
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
          <section class="teacher-student-list-section">
            <div class="teacher-section-head">
              <div class="teacher-heading">
                <div class="workspace-overline">Class Workspace</div>
                <h3 class="teacher-section-title">学生列表</h3>
              </div>
              <button
                type="button"
                class="ui-btn ui-btn--ghost ui-btn--sm"
                @click="emit('openClassManagement')"
              >
                <ChevronLeft class="h-4 w-4" />
                返回列表
              </button>
            </div>

            <section class="teacher-controls teacher-student-controls">
              <div class="teacher-filter-grid">
                <label class="teacher-field teacher-field--class-switch">
                  <span class="teacher-field-label">切换班级</span>
                  <div
                    class="teacher-field-control teacher-filter-control teacher-filter-control--select"
                  >
                    <select
                      :value="selectedClassName"
                      aria-label="选择班级"
                      class="teacher-input teacher-select"
                      @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
                    >
                      <option v-for="item in classes" :key="item.name" :value="item.name">
                        {{ item.name }} · {{ item.student_count || 0 }} 人
                      </option>
                    </select>
                  </div>
                </label>

                <label class="teacher-field">
                  <span class="teacher-field-label">学号查询</span>
                  <div class="teacher-field-control teacher-filter-control">
                    <Search class="h-4 w-4 text-text-muted" />
                    <input
                      :value="studentNoQuery"
                      type="text"
                      placeholder="输入学号精确查询"
                      class="teacher-input"
                      @input="
                        emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)
                      "
                    />
                  </div>
                </label>
                <button
                  v-if="studentNoQuery"
                  type="button"
                  class="teacher-filter-reset teacher-filter-clear"
                  @click="emit('updateStudentNoQuery', '')"
                >
                  清空学号
                </button>
              </div>
            </section>

            <div v-if="loadingStudents" class="teacher-skeleton-list">
              <div
                v-for="index in 6"
                :key="index"
                class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-elevated)]"
              />
            </div>

            <AppEmpty
              v-else-if="students.length === 0"
              class="teacher-empty-state"
              icon="Users"
              title="暂无学生"
              description="该班级下还没有可用学生记录。"
            />

            <section v-else class="teacher-directory teacher-table-shell">
              <div class="teacher-directory-head">
                <span>学号</span>
                <span>学生名称</span>
                <span>昵称</span>
                <span>薄弱项</span>
                <span>做题数 / 得分数</span>
                <span>操作</span>
              </div>

              <button
                v-for="student in students"
                :key="student.id"
                type="button"
                class="teacher-directory-row group"
                @click="emit('openStudent', student.id)"
              >
                <div class="teacher-directory-cell">
                  {{ student.student_no || '未设置学号' }}
                </div>

                <div class="teacher-directory-cell">
                  <h4 class="teacher-directory-row-title" :title="student.name || '未设置姓名'">
                    {{ student.name || '未设置姓名' }}
                  </h4>
                </div>

                <div class="teacher-directory-cell">
                  <div class="teacher-directory-row-points" :title="student.username">
                    {{ student.username }}
                  </div>
                </div>

                <div class="teacher-directory-row-tags">
                  <span
                    class="teacher-directory-state-chip teacher-directory-state-chip-empty"
                    :class="'workspace-directory-status-pill workspace-directory-status-pill--muted'"
                  >
                    {{ student.weak_dimension || '暂无薄弱项' }}
                  </span>
                </div>

                <div class="teacher-directory-row-metrics">
                  <span
                    >{{ student.solved_count ?? 0 }} 题 / {{ student.total_score ?? 0 }} 分</span
                  >
                </div>

                <div class="workspace-directory-row-btn teacher-directory-row-cta">
                  <span>分析</span>
                  <ArrowRight class="h-4 w-4" />
                </div>
              </button>
            </section>
          </section>
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
            />
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
@import '../teacher-workspace-subpanel.css';

.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
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

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
  border-radius: var(--workspace-radius-lg, 18px);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: 0 var(--space-5);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 18rem) minmax(0, 1fr);
}
.teacher-select {
  min-height: 1.75rem;
  border: 0;
  appearance: none;
  cursor: pointer;
  background: transparent;
  width: 100%;
  outline: none;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns:
    minmax(7.5rem, 0.7fr) minmax(10rem, 1fr) minmax(10rem, 0.9fr)
    minmax(8rem, 0.8fr) minmax(8rem, 0.8fr) minmax(6.5rem, 0.6fr);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-5) 0;
  border: 0;
  border-bottom: 1px solid var(--color-border-subtle);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: var(--color-primary-soft);
  outline: none;
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-size: var(--font-size-0-98);
  font-weight: 800;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.group:hover .teacher-directory-row-title {
  color: var(--color-primary);
}

.teacher-directory-row-points {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-filter-reset {
  align-self: end;
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-4);
  border: 1px solid var(--teacher-control-border);
  border-radius: var(--ui-control-radius-md);
  background: transparent;
  color: var(--color-primary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.teacher-directory-state-chip-ready {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}
.teacher-directory-state-chip-empty {
  background: var(--color-bg-elevated);
  color: var(--color-text-muted);
}

.teacher-directory-row-cta {
  gap: var(--space-2);
  color: var(--color-primary);
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.2s ease;
}
.teacher-directory-row:hover .teacher-directory-row-cta {
  opacity: 1;
  transform: translateX(0);
}

@media (max-width: 1080px) {
  .teacher-directory-head {
    display: none;
  }
  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }
  .teacher-directory-row-cta {
    opacity: 1;
    transform: none;
  }
}
</style>

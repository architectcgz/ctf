<script setup lang="ts">
import { Activity, ArrowRight, Search, Target, Users } from 'lucide-vue-next'
import { computed, type Component } from 'vue'

import type {
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import { ChallengeCategoryPill, toChallengeCategory } from '@/entities/challenge'

interface ClassStudentDirectoryRow {
  id: string
  student_no: string
  name: string
  username: string
  weak_dimension: string
  metrics: string
  solved_count: number
  total_score: number
  actions: 'open'
}

const props = defineProps<{
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

const rows = computed<ClassStudentDirectoryRow[]>(() =>
  props.students.map((student) => ({
    id: student.id,
    student_no: student.student_no || '未设置学号',
    name: student.name || '未设置姓名',
    username: student.username,
    weak_dimension: student.weak_dimension || '暂无薄弱项',
    metrics: `${student.solved_count ?? 0} 题 / ${student.total_score ?? 0} 分`,
    solved_count: student.solved_count ?? 0,
    total_score: student.total_score ?? 0,
    actions: 'open',
  }))
)

const columns = [
  { key: 'student_no', label: '学号', widthClass: 'w-[14%] min-w-[8rem]' },
  { key: 'name', label: '学生名称', widthClass: 'w-[20%] min-w-[11rem]' },
  { key: 'username', label: '昵称', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'weak_dimension', label: '薄弱项', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'metrics', label: '做题数 / 得分数', widthClass: 'w-[16%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[9rem]', align: 'right' as const },
]

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

function studentWeakCategory(student: { weak_dimension?: string | null }) {
  return toChallengeCategory(student.weak_dimension)
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

            <div class="header-actions">
              <button
                type="button"
                class="header-btn header-btn--ghost"
                @click="emit('openClassManagement')"
              >
                返回
              </button>
              <button
                type="button"
                class="header-btn header-btn--ghost"
                @click="emit('openDashboard')"
              >
                概览
              </button>
              <button
                type="button"
                class="header-btn header-btn--primary"
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
            <section class="teacher-directory-shell workspace-directory-list">
              <section class="teacher-directory-filters" aria-label="学生过滤">
                <div class="teacher-filter-grid">
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
                    class="ui-btn ui-btn--secondary teacher-filter-reset teacher-filter-clear"
                    @click="emit('updateStudentNoQuery', '')"
                  >
                    清空学号
                  </button>
                </div>
              </section>

              <div v-if="loadingStudents" class="workspace-directory-loading">
                <AppLoading>同步学生目录...</AppLoading>
              </div>

              <AppEmpty
                v-else-if="students.length === 0"
                class="teacher-empty-state workspace-directory-empty"
                icon="Users"
                title="暂无学生"
                description="该班级下还没有可用学生记录。"
              />

              <div v-else class="teacher-directory">
                <WorkspaceDataTable
                  class="teacher-student-directory-table"
                  :columns="columns"
                  :rows="rows"
                  row-key="id"
                >
                  <template #cell-student_no="{ row }">
                    <span class="teacher-directory-cell-student-no">
                      {{ (row as ClassStudentDirectoryRow).student_no }}
                    </span>
                  </template>

                  <template #cell-name="{ row }">
                    <div class="teacher-directory-cell-name">
                      <h4
                        class="teacher-directory-row-title"
                        :title="(row as ClassStudentDirectoryRow).name"
                      >
                        {{ (row as ClassStudentDirectoryRow).name }}
                      </h4>
                    </div>
                  </template>

                  <template #cell-username="{ row }">
                    <span
                      class="teacher-directory-row-points"
                      :title="(row as ClassStudentDirectoryRow).username"
                    >
                      {{ (row as ClassStudentDirectoryRow).username }}
                    </span>
                  </template>

                  <template #cell-weak_dimension="{ row }">
                    <ChallengeCategoryPill
                      v-if="studentWeakCategory(row as ClassStudentDirectoryRow)"
                      :category="studentWeakCategory(row as ClassStudentDirectoryRow)!"
                    />
                    <span
                      v-else
                      class="teacher-directory-chip teacher-directory-chip-muted"
                      :class="'workspace-directory-status-pill workspace-directory-status-pill--muted'"
                    >
                      {{ (row as ClassStudentDirectoryRow).weak_dimension }}
                    </span>
                  </template>

                  <template #cell-metrics="{ row }">
                    <span class="teacher-directory-row-metrics">
                      {{ (row as ClassStudentDirectoryRow).metrics }}
                    </span>
                  </template>

                  <template #cell-actions="{ row }">
                    <div class="workspace-directory-row-actions teacher-directory-row-cta">
                      <button
                        type="button"
                        class="ui-btn ui-btn--primary ui-btn--xs"
                        :aria-label="`${(row as ClassStudentDirectoryRow).name}，${(row as ClassStudentDirectoryRow).solved_count} 题，${(row as ClassStudentDirectoryRow).total_score} 分，查看学员分析`"
                        @click="emit('openStudent', (row as ClassStudentDirectoryRow).id)"
                      >
                        学员分析
                        <ArrowRight class="h-4 w-4" />
                      </button>
                    </div>
                  </template>
                </WorkspaceDataTable>
              </div>
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

.teacher-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 20rem) auto;
}

.teacher-student-directory-table {
  --workspace-directory-shell-border: color-mix(
    in srgb,
    var(--teacher-card-border) 86%,
    transparent
  );
}

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-cell-student-no {
  font-size: var(--font-size-0-76);
  font-weight: 800;
  letter-spacing: 0.02em;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
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

.teacher-student-directory-table
  :deep(.workspace-data-table__row:hover)
  .teacher-directory-row-title {
  color: var(--color-primary);
}

.teacher-directory-row-points {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-metrics {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-81);
  font-weight: 800;
  color: var(--color-text-primary);
}

.teacher-filter-reset {
  align-self: end;
}

.teacher-directory-row-cta {
  justify-content: flex-end;
}

@media (max-width: 1080px) {
  .teacher-directory-row-cta {
    justify-content: flex-start;
  }
}
</style>

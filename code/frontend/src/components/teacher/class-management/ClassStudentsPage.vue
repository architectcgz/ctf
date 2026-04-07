<script setup lang="ts">
import { ArrowRight, ChevronLeft, Search } from 'lucide-vue-next'
import { computed, ref } from 'vue'

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

interface WorkspaceTabItem {
  key: WorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

const workspaceTabs: WorkspaceTabItem[] = [
  { key: 'overview', label: '主看板', buttonId: 'class-tab-overview', panelId: 'class-overview' },
  { key: 'trend', label: '趋势复盘', buttonId: 'class-tab-trend', panelId: 'class-trend' },
  { key: 'students', label: '学生列表', buttonId: 'class-tab-students', panelId: 'class-students' },
  { key: 'review', label: '复盘结论', buttonId: 'class-tab-review', panelId: 'class-review' },
  { key: 'insight', label: '学生洞察', buttonId: 'class-tab-insight', panelId: 'class-insight' },
  { key: 'action', label: '介入建议', buttonId: 'class-tab-action', panelId: 'class-action' },
]

const activeTab = ref<WorkspaceTab>('overview')
const tabButtonRefs: Partial<Record<WorkspaceTab, HTMLButtonElement | null>> = {}

function setTabButtonRef(tab: WorkspaceTab, element: HTMLButtonElement | null): void {
  tabButtonRefs[tab] = element
}

function selectTab(tab: WorkspaceTab): void {
  activeTab.value = tab
}

function focusTab(tab: WorkspaceTab): void {
  tabButtonRefs[tab]?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (
    event.key !== 'ArrowRight' &&
    event.key !== 'ArrowLeft' &&
    event.key !== 'Home' &&
    event.key !== 'End'
  ) {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    selectTab(workspaceTabs[0].key)
    focusTab(workspaceTabs[0].key)
    return
  }

  if (event.key === 'End') {
    const lastTab = workspaceTabs[workspaceTabs.length - 1]
    selectTab(lastTab.key)
    focusTab(lastTab.key)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + workspaceTabs.length) % workspaceTabs.length
  const nextTab = workspaceTabs[nextIndex]
  selectTab(nextTab.key)
  focusTab(nextTab.key)
}
</script>

<template>
  <div class="workspace-shell">
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Class Workspace</span>
        <span class="class-chip">{{ selectedClassName || '未选择班级' }}</span>
      </div>
    </header>

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
          id="class-overview"
          class="tab-panel section active"
          :class="{ active: activeTab === 'overview' }"
          role="tabpanel"
          aria-labelledby="class-tab-overview"
          :aria-hidden="activeTab === 'overview' ? 'false' : 'true'"
          v-show="activeTab === 'overview'"
        >
          <header class="teacher-topbar">
            <div class="teacher-heading">
              <div class="teacher-eyebrow-row">
                <div class="teacher-surface-eyebrow journal-eyebrow">Class Students</div>
                <span class="teacher-class-chip teacher-surface-chip">{{
                  selectedClassName || '未选择班级'
                }}</span>
              </div>

              <h1 class="teacher-title">
                {{ selectedClassName ? `${selectedClassName} · 学生列表` : '班级学生' }}
              </h1>
              <p class="teacher-copy">查看当前班级学生名单，并继续进入学员分析。</p>
            </div>

            <div class="teacher-actions">
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost"
                @click="emit('openClassManagement')"
              >
                返回班级管理
              </button>
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost"
                @click="emit('openDashboard')"
              >
                教学概览
              </button>
              <button
                type="button"
                class="teacher-btn teacher-btn--primary"
                @click="emit('openReportExport')"
              >
                导出报告
              </button>
            </div>
          </header>

          <div v-if="error" class="workspace-alert" role="alert" aria-live="polite">
            <div class="workspace-alert-title">班级详情加载失败</div>
            <div class="workspace-alert-copy">{{ error }}</div>
            <div class="workspace-alert-actions">
              <button
                type="button"
                class="quick-action quick-action--compact"
                @click="emit('retry')"
              >
                <span>重试加载</span>
                <span>→</span>
              </button>
            </div>
          </div>

          <div class="teacher-surface-board">
            <div class="teacher-tip-block">
              <section class="teacher-summary">
                <div class="teacher-summary-title">
                  <span>Class Snapshot</span>
                </div>
                <div class="teacher-summary-grid">
                  <div class="teacher-summary-item">
                    <div class="teacher-summary-label">班级人数</div>
                    <div class="teacher-summary-value">
                      {{ props.summary?.student_count ?? students.length }}
                    </div>
                    <div class="teacher-summary-helper">当前班级纳入统计的学生数量</div>
                  </div>
                  <div class="teacher-summary-item">
                    <div class="teacher-summary-label">平均解题</div>
                    <div class="teacher-summary-value">{{ averageSolvedText }}</div>
                    <div class="teacher-summary-helper">班级当前平均完成情况</div>
                  </div>
                  <div class="teacher-summary-item">
                    <div class="teacher-summary-label">近 7 天活跃率</div>
                    <div class="teacher-summary-value">{{ activeRateText }}</div>
                    <div class="teacher-summary-helper">当前班级近 7 天训练参与情况</div>
                  </div>
                </div>
              </section>

              <section class="teacher-controls">
                <div class="teacher-controls-bar">
                  <div class="teacher-controls-heading">
                    <div class="teacher-surface-eyebrow journal-eyebrow">Class Workspace</div>
                    <h3 class="teacher-controls-title">班级详情</h3>
                    <p class="teacher-controls-copy">
                      {{
                        selectedClassName || '当前未选择班级'
                      }}。先看班级趋势与复盘，再按学号快速定位学生。
                    </p>
                  </div>
                </div>

                <div class="teacher-context-card">
                  <div class="teacher-field-label">当前班级</div>
                  <div class="teacher-context-value">{{ selectedClassName || '未选择班级' }}</div>
                  <div class="teacher-context-copy">
                    {{ props.summary?.student_count ?? students.length }} 名学生
                  </div>
                </div>
              </section>

              <div class="teacher-hero-divider" />

              <div class="teacher-tip-label">当前焦点</div>
              <div class="teacher-tip-copy">先看班级趋势和复盘结论，再决定课堂介入动作。</div>
            </div>

            <div class="teacher-summary-cards">
              <article class="teacher-summary-card">
                <div class="teacher-summary-label">近 7 天训练事件</div>
                <div class="teacher-summary-value">
                  {{ props.summary?.recent_event_count ?? '--' }}
                </div>
                <div class="teacher-summary-helper">提交、实例启动与销毁等动作总数</div>
              </article>
              <article class="teacher-summary-card">
                <div class="teacher-summary-label">当前学生记录</div>
                <div class="teacher-summary-value">{{ students.length }}</div>
                <div class="teacher-summary-helper">当前列表内可直接进入分析的学生数量</div>
              </article>
            </div>
          </div>
        </section>

        <section
          id="class-trend"
          class="tab-panel section"
          :class="{ active: activeTab === 'trend' }"
          role="tabpanel"
          aria-labelledby="class-tab-trend"
          :aria-hidden="activeTab === 'trend' ? 'false' : 'true'"
          v-show="activeTab === 'trend'"
        >
          <div class="workspace-subpanel workspace-subpanel--flat">
            <TeacherClassTrendPanel
              :trend="trend"
              title="班级近 7 天训练趋势"
              subtitle="先看整体节奏，再下钻到具体学生。"
            />
          </div>
        </section>

        <section
          id="class-students"
          class="tab-panel section"
          :class="{ active: activeTab === 'students' }"
          role="tabpanel"
          aria-labelledby="class-tab-students"
          :aria-hidden="activeTab === 'students' ? 'false' : 'true'"
          v-show="activeTab === 'students'"
        >
          <section class="teacher-student-list-section">
            <div class="teacher-section-head">
              <div>
                <div class="teacher-surface-eyebrow journal-eyebrow">Students</div>
                <h3 class="teacher-section-title">学生列表</h3>
                <p class="teacher-section-copy">选择学生后进入学员分析。</p>
              </div>
              <button
                type="button"
                class="teacher-inline-link"
                @click="emit('openClassManagement')"
              >
                <ChevronLeft class="h-4 w-4" />
                返回班级列表
              </button>
            </div>

            <section class="teacher-controls teacher-student-controls">
              <div class="teacher-controls-bar">
                <div>
                  <div class="teacher-surface-eyebrow journal-eyebrow">Student Filters</div>
                  <h3 class="teacher-controls-title">学生筛选</h3>
                  <p class="teacher-controls-copy">按学号快速定位学生，并继续进入学员分析。</p>
                </div>
                <div class="teacher-section-meta">共 {{ students.length }} 名学生</div>
              </div>

              <div class="teacher-filter-grid">
                <div class="teacher-context-card">
                  <div class="teacher-field-label">当前班级</div>
                  <div class="teacher-context-value">{{ selectedClassName || '未选择班级' }}</div>
                  <div class="teacher-context-copy">{{ students.length }} 名学生</div>
                </div>

                <label class="teacher-field">
                  <span class="teacher-field-label">按学号查询</span>
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
              </div>
            </section>

            <div v-if="loadingStudents" class="teacher-skeleton-list">
              <div
                v-for="index in 6"
                :key="index"
                class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
              />
            </div>

            <AppEmpty
              v-else-if="students.length === 0"
              class="teacher-empty-state"
              icon="Users"
              title="当前班级暂无学生"
              description="该班级下还没有可用学生记录。"
            />

            <section v-else class="teacher-directory" aria-label="学生目录">
              <div class="teacher-directory-head">
                <span class="teacher-directory-head-cell teacher-directory-head-cell-student-no"
                  >学号</span
                >
                <span class="teacher-directory-head-cell teacher-directory-head-cell-name"
                  >学生名称</span
                >
                <span class="teacher-directory-head-cell teacher-directory-head-cell-alias"
                  >昵称</span
                >
                <span>状态</span>
                <span>数据</span>
                <span>操作</span>
              </div>

              <button
                v-for="student in students"
                :key="student.id"
                type="button"
                class="teacher-directory-row"
                :aria-label="`${student.name || student.username}，${student.solved_count ?? 0} 题，${student.total_score ?? 0} 分，查看学员分析`"
                @click="emit('openStudent', student.id)"
              >
                <div class="teacher-directory-cell teacher-directory-cell-student-no">
                  {{ student.student_no || '未设置学号' }}
                </div>

                <div class="teacher-directory-cell teacher-directory-cell-name">
                  <h4 class="teacher-directory-row-title" :title="student.name || '未设置姓名'">
                    {{ student.name || '未设置姓名' }}
                  </h4>
                </div>

                <div class="teacher-directory-cell teacher-directory-cell-alias">
                  <div class="teacher-directory-row-points" :title="`@${student.username}`">
                    @{{ student.username }}
                  </div>
                </div>

                <div class="teacher-directory-row-status">
                  <span
                    class="teacher-directory-state-chip"
                    :class="
                      (student.solved_count ?? 0) > 0
                        ? 'teacher-directory-state-chip-ready'
                        : 'teacher-directory-state-chip-empty'
                    "
                  >
                    {{ (student.solved_count ?? 0) > 0 ? '已有解题记录' : '暂无解题记录' }}
                  </span>
                </div>

                <div class="teacher-directory-row-metrics">
                  <span>{{ student.solved_count ?? 0 }} 题</span>
                  <span>{{ student.total_score ?? 0 }} 分</span>
                </div>

                <div class="teacher-directory-row-cta">
                  <span>查看学员分析</span>
                  <ArrowRight class="h-4 w-4" />
                </div>
              </button>
            </section>
          </section>
        </section>

        <section
          id="class-review"
          class="tab-panel section"
          :class="{ active: activeTab === 'review' }"
          role="tabpanel"
          aria-labelledby="class-tab-review"
          :aria-hidden="activeTab === 'review' ? 'false' : 'true'"
          v-show="activeTab === 'review'"
        >
          <div class="workspace-subpanel workspace-subpanel--flat">
            <TeacherClassReviewPanel :review="review" :class-name="selectedClassName" />
          </div>
        </section>

        <section
          id="class-insight"
          class="tab-panel section"
          :class="{ active: activeTab === 'insight' }"
          role="tabpanel"
          aria-labelledby="class-tab-insight"
          :aria-hidden="activeTab === 'insight' ? 'false' : 'true'"
          v-show="activeTab === 'insight'"
        >
          <div class="workspace-subpanel workspace-subpanel--flat">
            <TeacherClassInsightsPanel :students="students" :class-name="selectedClassName" />
          </div>
        </section>

        <section
          id="class-action"
          class="tab-panel section"
          :class="{ active: activeTab === 'action' }"
          role="tabpanel"
          aria-labelledby="class-tab-action"
          :aria-hidden="activeTab === 'action' ? 'false' : 'true'"
          v-show="activeTab === 'action'"
        >
          <div class="workspace-subpanel workspace-subpanel--flat">
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
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-shadow-shell: 0 24px 84px
    color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-xl: 28px;
  --workspace-radius-lg: 18px;
  --teacher-summary-columns: repeat(4, minmax(0, 1fr));
  --teacher-directory-columns: var(--teacher-student-directory-columns);
  --teacher-directory-margin-top: 1rem;
  --teacher-student-directory-columns: minmax(7.5rem, 0.7fr) minmax(10rem, 1fr) minmax(10rem, 0.9fr)
    minmax(8rem, 0.8fr) minmax(8rem, 0.8fr) minmax(8.5rem, 0.85fr);
}

.teacher-eyebrow-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.teacher-class-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.85rem;
  padding: 0 0.75rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.teacher-filter-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 18rem) minmax(0, 1fr);
}

.teacher-context-card {
  display: grid;
  gap: 0.45rem;
  min-height: 2.9rem;
  align-content: center;
  border: 1px solid var(--teacher-control-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.72rem 0.95rem;
}

.teacher-context-value {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-context-copy {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.teacher-hero-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed var(--teacher-divider);
}

.workspace-alert {
  margin-bottom: 1.5rem;
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
  padding: 1rem 1.1rem;
}

.workspace-alert-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: 0.45rem;
  font-size: 0.84rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  margin-top: 0.85rem;
}

.quick-action {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 52px;
  padding: 0 14px;
  border: 1px solid var(--workspace-line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--workspace-panel) 82%, transparent);
  color: var(--journal-ink);
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

.tab-panel.section {
  padding-top: 0;
  border-top: 0;
}

.teacher-tip-block {
  display: grid;
  gap: 0.35rem;
}

.teacher-tip-label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-tip-copy {
  font-size: 0.86rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-summary-cards {
  margin-top: 1.25rem;
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.teacher-summary-card {
  padding-top: 0.25rem;
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
}

.workspace-subpanel :deep(.teacher-panel) {
  border: 1px solid color-mix(in srgb, var(--journal-border) 74%, transparent);
  border-radius: 22px;
  background: color-mix(in srgb, var(--journal-surface) 90%, transparent);
  box-shadow: 0 14px 34px var(--color-shadow-soft);
  padding: 20px;
}

.workspace-subpanel :deep(.teacher-panel__header),
.workspace-subpanel :deep(.teacher-subsection__header) {
  margin-bottom: 16px;
}

.workspace-subpanel :deep(.journal-eyebrow) {
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--journal-accent) 60%, var(--journal-muted));
}

.workspace-subpanel :deep(.teacher-panel__title) {
  margin-top: 10px;
  font-size: 22px;
  line-height: 1.15;
  color: var(--journal-ink);
}

.workspace-subpanel :deep(.teacher-subsection + .teacher-subsection) {
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.workspace-subpanel :deep(.top-student-item),
.workspace-subpanel :deep(.dimension-item),
.workspace-subpanel :deep(.review-item__recommendation),
.workspace-subpanel :deep(.review-item),
.workspace-subpanel :deep(.intervention-item) {
  border-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.workspace-subpanel :deep(.teacher-panel__chart) {
  border-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface-subtle) 82%, transparent);
}

.workspace-subpanel--flat :deep(.teacher-panel) {
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}

.workspace-subpanel--flat :deep(.teacher-panel__chart) {
  margin-top: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
  box-shadow: none;
  overflow: visible;
}

.workspace-subpanel :deep(.review-item) {
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 86%, transparent);
}

.workspace-subpanel :deep(.top-student-item__rank),
.workspace-subpanel :deep(.teacher-tip-index) {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

.teacher-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.teacher-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-section-copy {
  margin-top: 0.45rem;
  font-size: 0.86rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-section-meta {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.teacher-inline-link {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border: 0;
  background: transparent;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.teacher-student-toolbar {
  margin: 1rem 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.teacher-student-filter {
  min-width: min(100%, 20rem);
}

.teacher-skeleton-list {
  display: grid;
  gap: 0.75rem;
}

.teacher-empty-state {
  margin-top: 1.5rem;
}

.teacher-student-list-section {
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--teacher-student-directory-columns);
  gap: 1rem;
  align-items: center;
  width: 100%;
  padding: 1.1rem 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 62%, transparent);
  outline: none;
}

.teacher-directory-cell {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-alias .teacher-directory-row-points,
.teacher-directory-row-points {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.teacher-directory-cell-student-no {
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--journal-muted);
  font-variant-numeric: tabular-nums;
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.98rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
}

.teacher-directory-head-cell-student-no,
.teacher-directory-head-cell-name,
.teacher-directory-head-cell-alias,
.teacher-directory-cell-student-no,
.teacher-directory-cell-name,
.teacher-directory-cell-alias {
  justify-self: start;
  width: 100%;
}

.teacher-directory-row-points {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-status {
  display: flex;
  justify-content: flex-start;
}

.teacher-directory-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 0.62rem;
  border-radius: 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.teacher-directory-state-chip-ready {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.teacher-directory-state-chip-empty {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-metrics {
  display: grid;
  gap: 0.25rem;
  font-size: 0.81rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
}

@media (max-width: 1080px) {
  .teacher-topbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid,
  .teacher-filter-grid,
  .teacher-summary-cards {
    grid-template-columns: 1fr;
  }

  .teacher-directory-head {
    display: none;
  }

  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: 0.85rem;
    padding: 1rem 0;
  }
}

@media (max-width: 640px) {
  .workspace-topbar,
  .top-tabs,
  .content-pane {
    padding-left: 18px;
    padding-right: 18px;
  }

  .workspace-topbar {
    display: block;
  }
}
</style>

<script setup lang="ts">
import { ArrowRight, ChevronLeft, Search } from 'lucide-vue-next'
import { computed } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

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
  openWorkspaceSection: [section: WorkspaceEntryKey]
  selectClass: [className: string]
  updateStudentNoQuery: [value: string]
  openStudent: [studentId: string]
}>()

const workspaceTitle = computed(() =>
  props.selectedClassName ? `${props.selectedClassName} 班级工作台` : '班级工作台'
)
const averageSolvedText = computed(() => {
  if (!props.summary) return '--'
  return props.summary.average_solved.toFixed(1)
})
const activeRateText = computed(() => {
  if (!props.summary) return '--'
  return `${Math.round(props.summary.active_rate)}%`
})
const recentEventCountText = computed(() => {
  if (!props.summary) return '--'
  return String(props.summary.recent_event_count ?? 0)
})

type WorkspaceEntryKey = 'trend' | 'review' | 'insights' | 'intervention'

interface WorkspaceEntryItem {
  key: WorkspaceEntryKey
  eyebrow: string
  title: string
  description: string
  detail: string
  actionLabel: string
}

const priorityStudentCount = computed(
  () => props.students.filter((student) => (student.recent_event_count ?? 0) <= 1).length
)

const workspaceEntries = computed<WorkspaceEntryItem[]>(() => [
  {
    key: 'trend',
    eyebrow: 'Trend',
    title: '训练趋势',
    description: '查看班级近 7 天训练节奏、活跃学生与解题走势。',
    detail:
      props.trend?.points?.length && props.trend.points.length > 0
        ? `已采集 ${props.trend.points.length} 个趋势采样点`
        : '最近一周还没有可展示的趋势数据',
    actionLabel: '查看训练趋势',
  },
  {
    key: 'review',
    eyebrow: 'Review',
    title: '教学复盘',
    description: '集中查看当前班级已经形成的复盘结论与跟进建议。',
    detail:
      props.review?.items && props.review.items.length > 0
        ? `当前已有 ${props.review.items.length} 条复盘结论待查看`
        : '当前班级还没有稳定的复盘结论',
    actionLabel: '查看教学复盘',
  },
  {
    key: 'insights',
    eyebrow: 'Insights',
    title: '学生洞察',
    description: '快速进入学生画像与班级能力结构视角。',
    detail: props.students.length > 0 ? `当前纳入 ${props.students.length} 名学生画像` : '暂无学生画像数据',
    actionLabel: '查看学生洞察',
  },
  {
    key: 'intervention',
    eyebrow: 'Intervention',
    title: '介入建议',
    description: '查看当前最值得优先跟进的学生名单与跟进方向。',
    detail:
      priorityStudentCount.value > 0
        ? `当前有 ${priorityStudentCount.value} 名学生需要优先关注`
        : '当前暂无高优先级介入对象',
    actionLabel: '查看介入建议',
  },
])
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="teacher-class-workspace">
          <header class="teacher-topbar">
            <div class="teacher-heading">
              <div class="teacher-eyebrow-row">
                <div class="teacher-surface-eyebrow journal-eyebrow">Class Workspace</div>
                <span v-if="selectedClassName" class="teacher-class-chip">
                  {{ selectedClassName }}
                </span>
              </div>
              <h1 class="teacher-title">{{ workspaceTitle }}</h1>
              <p class="teacher-copy">先看班级概况与学生目录，再按入口进入趋势、复盘、洞察和介入页面。</p>
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
                导出班级报告
              </button>
            </div>
          </header>

          <section class="teacher-summary metric-panel-default-surface">
            <div class="teacher-summary-title">
              <span>Class Snapshot</span>
            </div>
            <div class="teacher-summary-grid progress-strip metric-panel-grid">
              <div class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">班级人数</div>
                <div class="progress-card-value metric-panel-value">
                  {{ props.summary?.student_count ?? students.length }}
                </div>
                <div class="progress-card-hint metric-panel-helper">当前班级纳入统计的学生数量</div>
              </div>
              <div class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">平均解题</div>
                <div class="progress-card-value metric-panel-value">
                  {{ averageSolvedText }}
                </div>
                <div class="progress-card-hint metric-panel-helper">班级当前平均完成情况</div>
              </div>
              <div class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">近 7 天活跃率</div>
                <div class="progress-card-value metric-panel-value">{{ activeRateText }}</div>
                <div class="progress-card-hint metric-panel-helper">当前班级近 7 天训练参与情况</div>
              </div>
              <div class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">近 7 天训练动作</div>
                <div class="progress-card-value metric-panel-value">{{ recentEventCountText }}</div>
                <div class="progress-card-hint metric-panel-helper">最近一周训练事件总量</div>
              </div>
            </div>
          </section>

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

          <section class="teacher-workspace-launchpad" aria-label="班级工作区入口">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">Workspace Entry</div>
                <h2 class="list-heading__title">班级工作区入口</h2>
                <p class="teacher-section-copy">将互补视角拆成独立页面，避免在总览页继续堆叠长内容。</p>
              </div>
            </header>

            <div class="workspace-entry-grid">
              <button
                v-for="entry in workspaceEntries"
                :key="entry.key"
                type="button"
                class="workspace-entry-card"
                @click="emit('openWorkspaceSection', entry.key)"
              >
                <div class="workspace-entry-card__main">
                  <div class="workspace-entry-card__eyebrow">{{ entry.eyebrow }}</div>
                  <h3 class="workspace-entry-card__title">{{ entry.title }}</h3>
                  <p class="workspace-entry-card__description">{{ entry.description }}</p>
                </div>
                <div class="workspace-entry-card__footer">
                  <span class="workspace-entry-card__detail">{{ entry.detail }}</span>
                  <span class="workspace-entry-card__action">
                    <span>{{ entry.actionLabel }}</span>
                    <ArrowRight class="h-4 w-4" />
                  </span>
                </div>
              </button>
            </div>
          </section>

          <section
            class="workspace-directory-section teacher-student-list-section teacher-anchor-section"
            aria-label="学生目录"
          >
            <header class="list-heading">
              <div>
                <div class="workspace-overline">Student Directory</div>
                <h2 class="list-heading__title">学生列表</h2>
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
            </header>

            <section class="teacher-controls teacher-student-controls">
              <div class="teacher-controls-bar">
                <div class="teacher-controls-meta">
                  <div class="teacher-section-meta" aria-live="polite">
                    {{
                      studentNoQuery
                        ? `学号筛选：${studentNoQuery} · 匹配 ${students.length} 名学生`
                        : `共 ${students.length} 名学生`
                    }}
                  </div>
                  <button
                    v-if="studentNoQuery"
                    type="button"
                    class="teacher-filter-reset"
                    @click="emit('updateStudentNoQuery', '')"
                  >
                    清空筛选
                  </button>
                </div>
              </div>

              <div class="teacher-filter-grid">
                <label class="teacher-field teacher-field--class-switch">
                  <span class="teacher-field-label">班级切换</span>
                  <div
                    class="teacher-field-control teacher-filter-control teacher-filter-control--select"
                  >
                    <select
                      :value="selectedClassName"
                      class="teacher-input teacher-select"
                      aria-label="选择班级"
                      @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
                    >
                      <option v-if="classes.length === 0" value="" disabled>暂无可切换班级</option>
                      <option v-for="item in classes" :key="item.name" :value="item.name">
                        {{ item.name }} · {{ item.student_count || 0 }} 人
                      </option>
                    </select>
                  </div>
                </label>

                <label class="teacher-field">
                  <span class="teacher-field-label">按学号查询</span>
                  <div class="teacher-field-control teacher-filter-control">
                    <Search class="h-4 w-4 text-text-muted" />
                    <input
                      :value="studentNoQuery"
                      type="text"
                      placeholder="输入学号精确查询"
                      class="teacher-input"
                      @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
                    />
                    <button
                      v-if="studentNoQuery"
                      type="button"
                      class="teacher-filter-clear"
                      aria-label="清空学号筛选"
                      @click="emit('updateStudentNoQuery', '')"
                    >
                      清空
                    </button>
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

            <section v-else class="teacher-directory" aria-label="学生目录列表">
              <div class="teacher-directory-head">
                <span class="teacher-directory-head-cell teacher-directory-head-cell-student-no">
                  学号
                </span>
                <span class="teacher-directory-head-cell teacher-directory-head-cell-name">
                  学生名称
                </span>
                <span class="teacher-directory-head-cell teacher-directory-head-cell-alias">
                  昵称
                </span>
                <span class="teacher-directory-head-cell teacher-directory-head-cell-status">
                  状态
                </span>
                <span class="teacher-directory-head-cell teacher-directory-head-cell-metrics">
                  数据
                </span>
                <span class="teacher-directory-head-cell teacher-directory-head-cell-action">
                  操作
                </span>
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
  --teacher-directory-margin-top: var(--space-4);
  --teacher-student-directory-columns: minmax(7.5rem, 0.7fr) minmax(10rem, 1fr) minmax(10rem, 0.9fr)
    minmax(8rem, 0.8fr) minmax(8rem, 0.8fr) minmax(8.5rem, 0.85fr);
  --teacher-workspace-panel-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --teacher-workspace-panel-background: color-mix(in srgb, var(--journal-surface) 90%, transparent);
  --teacher-workspace-panel-shadow: 0 14px 34px var(--color-shadow-soft);
  --teacher-workspace-panel-padding: var(--space-8);
  --teacher-workspace-panel-header-gap: var(--space-8);
  --teacher-workspace-eyebrow-color: color-mix(
    in srgb,
    var(--journal-accent) 60%,
    var(--journal-muted)
  );
  --teacher-workspace-line-soft: color-mix(in srgb, var(--journal-border) 88%, transparent);
  --teacher-workspace-chart-background: color-mix(
    in srgb,
    var(--journal-surface-subtle) 82%,
    transparent
  );
  --teacher-workspace-review-background: color-mix(
    in srgb,
    var(--journal-surface-subtle) 86%,
    transparent
  );
  --teacher-workspace-mono-font: var(--font-family-mono);
}

.teacher-class-workspace {
  display: grid;
  gap: var(--space-8);
  min-width: 0;
}

.teacher-eyebrow-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2-5);
}

.teacher-class-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.85rem;
  padding: 0 0.75rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: var(--font-size-0-78);
  font-weight: 600;
  color: var(--journal-muted);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 18rem) minmax(0, 1fr);
}

.teacher-workspace-launchpad {
  display: grid;
  gap: var(--space-5);
}

.workspace-entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.workspace-entry-card {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
  padding: var(--space-5);
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 96%, transparent);
  border-radius: 1.25rem;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 96%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 94%, transparent)
  );
  text-align: left;
  box-shadow: 0 14px 32px color-mix(in srgb, var(--color-shadow-soft) 34%, transparent);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    transform 160ms ease,
    box-shadow 160ms ease;
}

.workspace-entry-card:hover,
.workspace-entry-card:focus-visible {
  border-color: color-mix(in srgb, var(--workspace-brand) 34%, transparent);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--workspace-brand) 7%, var(--journal-surface)),
    color-mix(in srgb, var(--workspace-brand) 5%, var(--journal-surface-subtle))
  );
  box-shadow: 0 18px 40px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  transform: translateY(-1px);
  outline: none;
}

.workspace-entry-card__main {
  display: grid;
  gap: var(--space-2);
}

.workspace-entry-card__eyebrow {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--workspace-brand) 72%, var(--journal-muted));
}

.workspace-entry-card__title {
  margin: 0;
  font-size: var(--font-size-1-05);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-entry-card__description,
.workspace-entry-card__detail {
  margin: 0;
  font-size: var(--font-size-0-84);
  line-height: 1.68;
  color: var(--journal-muted);
}

.workspace-entry-card__footer {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
}

.workspace-entry-card__action {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  flex-shrink: 0;
  font-size: var(--font-size-0-82);
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.workspace-alert {
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
  padding: var(--space-4) var(--space-4-5);
}

.workspace-alert-title {
  font-size: var(--font-size-0-92);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-84);
  line-height: 1.65;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  margin-top: var(--space-3);
}

.quick-action {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2-5);
  min-height: 52px;
  padding: 0 var(--space-3-5);
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

.teacher-tip-block {
  display: grid;
  gap: var(--space-1-5);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

.teacher-section-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-86);
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-section-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.teacher-filter-control--select {
  justify-content: flex-start;
}

.teacher-select {
  min-height: 1.75rem;
  padding-right: var(--space-5);
  border: 0;
  appearance: none;
  cursor: pointer;
}

.teacher-select:focus-visible {
  outline: none;
}

.teacher-controls-meta {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2-5);
}

.teacher-filter-reset {
  display: inline-flex;
  align-items: center;
  min-height: 2rem;
  padding: 0 var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--teacher-control-border) 92%, transparent);
  border-radius: 0.6rem;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  color: var(--journal-muted);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.teacher-filter-reset:hover,
.teacher-filter-reset:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
  color: var(--journal-accent-strong);
  outline: none;
}

.teacher-inline-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  border: 0;
  background: transparent;
  font-size: var(--font-size-0-875);
  font-weight: 600;
  color: var(--journal-accent);
}

.teacher-skeleton-list {
  display: grid;
  gap: var(--space-3);
}

.teacher-empty-state {
  margin-top: var(--space-6);
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
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-4-5) 0;
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
  gap: var(--space-2);
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-alias .teacher-directory-row-points,
.teacher-directory-row-points {
  font-family: var(--font-family-mono);
}

.teacher-directory-cell-student-no {
  font-size: var(--font-size-0-76);
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
  font-size: var(--font-size-0-98);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
}

.teacher-directory-head-cell-student-no,
.teacher-directory-head-cell-name,
.teacher-directory-head-cell-alias,
.teacher-directory-head-cell-status,
.teacher-directory-head-cell-metrics,
.teacher-directory-head-cell-action,
.teacher-directory-cell-student-no,
.teacher-directory-cell-name,
.teacher-directory-cell-alias {
  justify-self: start;
  width: 100%;
}

.teacher-directory-head-cell-action {
  justify-self: end;
  text-align: right;
}

.teacher-directory-row-points {
  font-size: var(--font-size-0-80);
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
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  font-size: var(--font-size-0-75);
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
  gap: var(--space-1);
  font-size: var(--font-size-0-81);
  line-height: 1.5;
  color: var(--journal-muted);
  font-variant-numeric: tabular-nums;
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1-5);
  justify-self: end;
  min-height: 2.1rem;
  padding: 0 var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--teacher-control-border) 92%, transparent);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--journal-surface-subtle) 90%, transparent);
  font-size: var(--font-size-0-82);
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.teacher-directory-row:hover .teacher-directory-row-cta,
.teacher-directory-row:focus-visible .teacher-directory-row-cta {
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
}

.teacher-filter-clear {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  min-height: 1.95rem;
  padding: 0 var(--space-2);
  border: 1px solid color-mix(in srgb, var(--teacher-control-border) 90%, transparent);
  border-radius: 0.55rem;
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 600;
  color: var(--journal-muted);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.teacher-filter-clear:hover,
.teacher-filter-clear:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent-strong);
  outline: none;
}

@media (max-width: 1080px) {
  .teacher-topbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid,
  .workspace-entry-grid,
  .teacher-filter-grid,
  .teacher-summary-cards {
    grid-template-columns: 1fr;
  }

  .teacher-directory-head {
    display: none;
  }

  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }

  .teacher-controls-meta {
    justify-content: flex-start;
  }

  .teacher-directory-row-cta {
    justify-self: start;
  }
}

@media (max-width: 640px) {
  .content-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }
}
</style>

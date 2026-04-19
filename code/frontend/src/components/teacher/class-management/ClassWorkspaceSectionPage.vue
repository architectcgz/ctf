<script setup lang="ts">
import { ChevronLeft } from 'lucide-vue-next'
import { computed, type Component } from 'vue'

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
import type { ClassWorkspaceSectionKey } from '@/utils/teachingWorkspaceRouting'

const props = defineProps<{
  sectionKey: ClassWorkspaceSectionKey
  classes: TeacherClassItem[]
  selectedClassName: string
  students: TeacherStudentItem[]
  review: TeacherClassReviewData | null
  summary: TeacherClassSummaryData | null
  trend: TeacherClassTrendData | null
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassOverview: []
  openClassManagement: []
  openDashboard: []
  openReportExport: []
  selectClass: [className: string]
}>()

interface ActiveSectionDefinition {
  eyebrow: string
  pageTitle: string
  pageDescription: string
  sectionTitle: string
  sectionDescription: string
  component: Component
}

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

const activeSection = computed<ActiveSectionDefinition>(() => {
  switch (props.sectionKey) {
    case 'review':
      return {
        eyebrow: 'Review',
        pageTitle: '班级教学复盘',
        pageDescription: '将复盘结论从总览页拆出，集中处理当前班级的教学判断与跟进建议。',
        sectionTitle: '教学复盘结论',
        sectionDescription: '围绕当前班级已经形成的结论与建议，直接进入复盘处理视角。',
        component: TeacherClassReviewPanel,
      }
    case 'insights':
      return {
        eyebrow: 'Insights',
        pageTitle: '班级学生洞察',
        pageDescription: '将学生画像拆到独立页面，专门查看 Top 学生与薄弱维度分布。',
        sectionTitle: '学生洞察',
        sectionDescription: '面向班级整体画像，查看当前最值得关注的学生结构与能力分布。',
        component: TeacherClassInsightsPanel,
      }
    case 'intervention':
      return {
        eyebrow: 'Intervention',
        pageTitle: '班级介入建议',
        pageDescription: '将需要教师介入的学生名单单独承载，避免总览页继续堆叠跟进内容。',
        sectionTitle: '介入建议',
        sectionDescription: '专门处理当前班级最值得优先跟进的学生名单与建议训练题。',
        component: TeacherInterventionPanel,
      }
    case 'trend':
    default:
      return {
        eyebrow: 'Trend',
        pageTitle: '班级训练趋势',
        pageDescription: '将训练节奏拆到独立页面，先看走势，再决定是否回到学生或复盘视角。',
        sectionTitle: '训练趋势',
        sectionDescription: '围绕近 7 天训练事件、解题走势和活跃学生变化查看班级节奏。',
        component: TeacherClassTrendPanel,
      }
  }
})

const sectionProps = computed<Record<string, unknown>>(() => {
  switch (props.sectionKey) {
    case 'review':
      return {
        review: props.review,
        className: props.selectedClassName,
      }
    case 'insights':
      return {
        students: props.students,
        className: props.selectedClassName,
      }
    case 'intervention':
      return {
        students: props.students,
        className: props.selectedClassName,
      }
    case 'trend':
    default:
      return {
        trend: props.trend,
        title: '班级近 7 天训练趋势',
        subtitle: '先看整体节奏，再决定回到哪一个工作区继续处理。',
      }
  }
})
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="teacher-class-workspace">
          <header class="teacher-topbar">
            <div class="teacher-heading">
              <div class="teacher-eyebrow-row">
                <div class="teacher-surface-eyebrow journal-eyebrow">{{ activeSection.eyebrow }}</div>
                <span v-if="selectedClassName" class="teacher-class-chip">
                  {{ selectedClassName }}
                </span>
              </div>
              <h1 class="teacher-title">{{ activeSection.pageTitle }}</h1>
              <p class="teacher-copy">{{ activeSection.pageDescription }}</p>
            </div>

            <div class="teacher-actions">
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost"
                @click="emit('openClassOverview')"
              >
                返回班级总览
              </button>
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost"
                @click="emit('openClassManagement')"
              >
                返回班级管理
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

          <section class="teacher-workspace-section" :aria-label="activeSection.sectionTitle">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">{{ activeSection.eyebrow }}</div>
                <h2 class="list-heading__title">{{ activeSection.sectionTitle }}</h2>
                <p class="teacher-section-copy">{{ activeSection.sectionDescription }}</p>
              </div>

              <button
                type="button"
                class="teacher-inline-link"
                @click="emit('openClassOverview')"
              >
                <ChevronLeft class="h-4 w-4" />
                返回班级总览
              </button>
            </header>

            <section class="teacher-controls teacher-workspace-controls">
              <div class="teacher-controls-bar">
                <div class="teacher-section-meta" aria-live="polite">
                  当前班级：{{ selectedClassName || '未选择班级' }}
                </div>
              </div>

              <div class="teacher-filter-grid teacher-filter-grid--single">
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
              </div>
            </section>

            <div class="workspace-subpanel workspace-subpanel--flat">
              <component :is="activeSection.component" v-bind="sectionProps" />
            </div>
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

.teacher-section-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-86);
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-workspace-section {
  display: grid;
  gap: var(--space-5);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 18rem);
}

.teacher-filter-grid--single {
  justify-content: start;
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

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid,
  .teacher-filter-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .content-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }
}
</style>

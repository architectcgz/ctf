<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, LifeBuoy } from 'lucide-vue-next'

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
  openReportExport: []
}>()

const averageSolvedText = computed(() => {
  if (!props.summary) return '--'
  return props.summary.average_solved.toFixed(1)
})

const activeRateText = computed(() => {
  if (!props.summary) return '--'
  return `${Math.round(props.summary.active_rate)}%`
})

const studentCountText = computed(
  () => props.summary?.student_count || props.selectedClass?.student_count || props.students.length
)

const activeStudentCountText = computed(() => props.summary?.active_student_count ?? '--')

const recentEventCountText = computed(() => props.summary?.recent_event_count ?? '--')

const reviewItemCount = computed(() => props.review?.items.length ?? 0)

const overviewBadges = computed(() => [
  { key: 'class-count', label: '班级池', value: `${props.classes.length} 个班级` },
  { key: 'student-count', label: '学生样本', value: `${studentCountText.value} 人` },
  { key: 'review-count', label: '复盘结论', value: `${reviewItemCount.value} 条` },
])

const interventionTips = computed(() => {
  const tips: string[] = []

  if (props.summary?.active_rate !== undefined) {
    tips.push(
      props.summary.active_rate < 65
        ? '近 7 天活跃率偏低，优先安排低活跃学生进行补训。'
        : '班级活跃率稳定，建议继续维持节奏并聚焦薄弱维度。'
    )
  }

  if (props.review?.items[0]?.title) {
    tips.push(`优先执行「${props.review.items[0].title}」对应的干预动作。`)
  }

  tips.push('完成本页排查后可直接导出报告，用于课后复盘与沟通。')
  return tips.slice(0, 3)
})

interface OverviewMetricItem {
  key: string
  label: string
  value: string | number
  hint: string
}

const overviewMetrics = computed<OverviewMetricItem[]>(() => [
  {
    key: 'student-count',
    label: '班级人数',
    value: studentCountText.value,
    hint: '当前班级纳入视图的人数',
  },
  {
    key: 'average-solved',
    label: '平均解题',
    value: averageSolvedText.value,
    hint: '当前班级学生的人均解题数',
  },
  {
    key: 'active-rate',
    label: '近 7 天活跃率',
    value: activeRateText.value,
    hint: '近 7 天有训练动作的学生占比',
  },
  {
    key: 'active-student',
    label: '近 7 天活跃学生',
    value: activeStudentCountText.value,
    hint: '至少有一次训练动作的学生数量',
  },
  {
    key: 'recent-event',
    label: '近 7 天训练事件',
    value: recentEventCountText.value,
    hint: '提交、实例启动和销毁等训练动作总数',
  },
])

interface OverviewAnchorItem {
  key: string
  label: string
  href: string
}

const overviewAnchors: OverviewAnchorItem[] = [
  { key: 'trend', label: '趋势', href: '#teacher-trend' },
  { key: 'review', label: '复盘', href: '#teacher-review' },
  { key: 'insight', label: '学生', href: '#teacher-insight' },
  { key: 'intervention', label: '介入', href: '#teacher-intervention' },
]

const overviewDescription = computed(() => {
  if (!props.selectedClassName) return '先选择一个班级。'
  if (!props.summary) return '正在汇总班级数据。'

  return `近 7 天活跃率 ${activeRateText.value}，人均解题 ${averageSolvedText.value}。`
})
</script>

<template>
  <div class="teacher-dashboard space-y-6">
    <section class="teacher-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow-row">
            <div class="journal-eyebrow">Teacher Flight Deck</div>
            <span class="teacher-class-chip">{{ selectedClassName || '未选择班级' }}</span>
          </div>

          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            教学介入台
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            {{ overviewDescription }}
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn" @click="emit('openClassManagement')">
              班级管理
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>

          <nav class="teacher-anchor-nav mt-6">
            <a
              v-for="anchor in overviewAnchors"
              :key="anchor.key"
              :href="anchor.href"
              class="teacher-anchor-link"
            >
              {{ anchor.label }}
            </a>
          </nav>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="text-sm font-medium text-[var(--journal-ink)]">当前班级概况</div>
          <div class="teacher-badge-grid mt-5">
            <div v-for="badge in overviewBadges" :key="badge.key" class="teacher-badge-card">
              <div class="teacher-badge-label">{{ badge.label }}</div>
              <div class="teacher-badge-value">{{ badge.value }}</div>
            </div>
          </div>

          <div class="teacher-tip-block mt-5">
            <div class="teacher-tip-title">今日教学建议</div>
            <ul class="teacher-tip-list mt-3">
              <li v-for="(tip, index) in interventionTips" :key="tip" class="teacher-tip-item">
                <span class="teacher-tip-index">{{ index + 1 }}</span>
                <span>{{ tip }}</span>
              </li>
            </ul>
          </div>
        </article>
      </div>

      <div class="teacher-metric-grid mt-6">
        <article
          v-for="item in overviewMetrics"
          :key="item.key"
          class="journal-metric teacher-metric-card rounded-[20px] border px-4 py-4"
        >
          <div class="teacher-metric-label">{{ item.label }}</div>
          <div class="teacher-metric-value">{{ item.value }}</div>
          <div class="teacher-metric-hint">{{ item.hint }}</div>
        </article>
      </div>

      <div class="teacher-board">
        <div v-if="error" class="teacher-error-card" role="alert" aria-live="polite">
          <div class="teacher-error-header">
            <div class="teacher-error-icon-wrap">
              <AlertTriangle class="teacher-error-icon" />
            </div>
            <div class="min-w-0">
              <div class="teacher-error-title">教师概览加载失败</div>
              <div class="teacher-error-text">
                {{ error }}
              </div>
            </div>
          </div>

          <div class="teacher-error-tips">
            <div class="teacher-error-tip">
              <LifeBuoy class="teacher-error-tip-icon" />
              <span>可先重试刷新数据，再继续查看趋势与复盘信息。</span>
            </div>
            <div class="teacher-error-tip">
              <LifeBuoy class="teacher-error-tip-icon" />
              <span>若持续失败，可先进入班级管理确认当前班级与权限状态。</span>
            </div>
          </div>

          <div class="teacher-error-actions">
            <ElButton type="danger" @click="emit('retry')"> 重试加载 </ElButton>
            <ElButton plain @click="emit('openClassManagement')"> 班级管理 </ElButton>
          </div>
        </div>

        <section id="teacher-trend" class="teacher-anchor-section">
          <TeacherClassTrendPanel
            :trend="trend"
            title="班级近 7 天训练趋势"
            subtitle="把训练事件、成功解题和活跃学生放在同一条时间轴上观察。"
          />
        </section>

        <section id="teacher-review" class="teacher-anchor-section">
          <TeacherClassReviewPanel :review="review" :class-name="selectedClassName" />
        </section>

        <section id="teacher-insight" class="teacher-anchor-section">
          <TeacherClassInsightsPanel :students="students" :class-name="selectedClassName" />
        </section>

        <section id="teacher-intervention" class="teacher-anchor-section">
          <TeacherInterventionPanel :students="students" :class-name="selectedClassName" />
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.teacher-dashboard {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 70%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 56%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.2rem 0.72rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-hero {
  border-color: var(--teacher-card-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 14%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-base))
    );
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  border-color: var(--teacher-card-border);
  background: var(--journal-surface-subtle);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.journal-metric {
  border-color: var(--teacher-card-border);
  background: var(--journal-surface);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
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
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.3rem 0.75rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.5rem;
  border-radius: 0.9rem;
  border: 1px solid var(--teacher-control-border);
  background: var(--journal-surface);
  padding: 0.55rem 1.1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.teacher-btn:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px rgba(79, 70, 229, 0.18);
}

.teacher-btn--primary:hover {
  background: var(--journal-accent-strong);
  border-color: transparent;
}

.teacher-anchor-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
}

.teacher-anchor-link {
  display: inline-flex;
  align-items: center;
  border-bottom: 1px dashed rgba(99, 102, 241, 0.28);
  padding-bottom: 0.06rem;
  font-size: 0.76rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
  transition:
    color 160ms ease,
    border-color 160ms ease;
}

.teacher-anchor-link:hover {
  color: var(--journal-accent);
  border-bottom-color: var(--journal-accent);
}

.teacher-badge-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.teacher-badge-card {
  border-radius: 18px;
  border: 1px solid var(--teacher-card-border);
  background: var(--journal-surface);
  padding: 0.9rem 0.95rem;
}

.teacher-badge-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-badge-value {
  margin-top: 0.55rem;
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
  padding-top: 1rem;
}

.teacher-tip-title {
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-tip-list {
  display: grid;
  gap: 0.6rem;
}

.teacher-tip-item {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  font-size: 0.83rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-tip-index {
  display: inline-flex;
  min-width: 1.2rem;
  justify-content: center;
  margin-top: 0.04rem;
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
    monospace;
  font-size: 0.76rem;
  font-weight: 700;
  color: var(--journal-accent);
}

.teacher-metric-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.teacher-metric-card {
  animation: teacherDeckEnter 0.24s ease both;
}

.teacher-metric-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-metric-value {
  margin-top: 0.55rem;
  font-size: 1.18rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-metric-hint {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-error-card {
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--color-danger) 22%, var(--teacher-card-border));
  background: color-mix(in srgb, var(--color-danger) 6%, transparent);
  padding: 1rem 1rem 1.1rem;
}

.teacher-error-header {
  display: flex;
  align-items: flex-start;
  gap: 0.62rem;
}

.teacher-error-icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.55rem;
  height: 1.55rem;
  flex-shrink: 0;
  border: 1px solid color-mix(in srgb, var(--color-danger) 36%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
}

.teacher-error-icon {
  width: 0.9rem;
  height: 0.9rem;
  color: var(--color-danger);
}

.teacher-error-title {
  font-size: 0.96rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.teacher-error-text {
  margin-top: 0.22rem;
  font-size: 0.83rem;
  line-height: 1.6;
  color: color-mix(in srgb, var(--color-danger) 70%, var(--journal-muted));
}

.teacher-error-tips {
  margin-top: 0.62rem;
  display: grid;
  gap: 0.32rem;
}

.teacher-error-tip {
  display: flex;
  align-items: flex-start;
  gap: 0.38rem;
  font-size: 0.8rem;
  line-height: 1.54;
  color: var(--journal-muted);
}

.teacher-error-tip-icon {
  margin-top: 0.08rem;
  width: 0.78rem;
  height: 0.78rem;
  color: color-mix(in srgb, var(--color-danger) 74%, var(--journal-muted));
  flex-shrink: 0;
}

.teacher-error-actions {
  margin-top: 0.7rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.teacher-board {
  border-top: 1px dashed var(--teacher-divider);
  padding-top: 1.25rem;
}

.teacher-board > * + * {
  margin-top: 1.25rem;
  border-top: 1px dashed var(--teacher-divider);
  padding-top: 1.25rem;
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

@media (min-width: 768px) {
  .teacher-metric-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .teacher-metric-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

@media (max-width: 639px) {
  .teacher-badge-grid,
  .teacher-metric-grid {
    grid-template-columns: 1fr;
  }
}

@keyframes teacherDeckEnter {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

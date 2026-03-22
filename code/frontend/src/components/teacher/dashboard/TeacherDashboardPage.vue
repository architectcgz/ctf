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
import PageHeader from '@/components/common/PageHeader.vue'
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
  if (!props.selectedClassName) return '选择班级后查看当前班级概览。'
  if (!props.summary) return '当前正在汇总班级训练数据，可先查看趋势和复盘结论。'

  return `近 7 天活跃率 ${activeRateText.value}，人均解题 ${averageSolvedText.value}，当前活跃学生 ${activeStudentCountText.value} 人。`
})
</script>

<template>
  <div class="teacher-dashboard space-y-6">
    <PageHeader
      eyebrow="Teacher Flight Deck"
      title="教学介入台"
      description="查看班级状态。"
    >
      <ElButton
        plain
        @click="emit('openClassManagement')"
      >
        班级管理
      </ElButton>
      <ElButton
        type="primary"
        @click="emit('openReportExport')"
      >
        导出报告
      </ElButton>
    </PageHeader>

    <section class="teacher-deck">
      <div class="teacher-deck__hero">
        <div class="teacher-deck__hero-grid">
          <div>
            <div class="teacher-deck__eyebrow">
              <span>Intervention Deck</span>
              <span class="teacher-deck__class-chip">{{ selectedClassName || '未选择班级' }}</span>
            </div>

            <h2 class="teacher-deck__title">
              {{ selectedClassName ? `${selectedClassName} 的教学概览` : '先选择一个班级' }}
            </h2>
            <p class="teacher-deck__description">
              {{ overviewDescription }}
            </p>

            <div class="teacher-deck__badges">
              <div
                v-for="badge in overviewBadges"
                :key="badge.key"
                class="teacher-deck__badge"
              >
                <div class="teacher-deck__badge-label">
                  {{ badge.label }}
                </div>
                <div class="teacher-deck__badge-value">
                  {{ badge.value }}
                </div>
              </div>
            </div>

            <nav class="teacher-deck__anchors">
              <a
                v-for="anchor in overviewAnchors"
                :key="anchor.key"
                :href="anchor.href"
                class="teacher-deck__anchor-link"
              >
                {{ anchor.label }}
              </a>
            </nav>
          </div>

          <aside class="teacher-deck__suggestion">
            <div class="teacher-deck__suggestion-title">
              今日教学建议
            </div>
            <ul class="teacher-deck__suggestion-list">
              <li
                v-for="(tip, index) in interventionTips"
                :key="tip"
                class="teacher-deck__suggestion-item"
              >
                <span class="teacher-deck__suggestion-index">{{ index + 1 }}</span>
                <span>{{ tip }}</span>
              </li>
            </ul>
          </aside>
        </div>

        <div class="teacher-deck__metrics">
          <article
            v-for="(item, index) in overviewMetrics"
            :key="item.key"
            class="teacher-deck__metric"
            :style="{ animationDelay: `${80 + index * 70}ms` }"
          >
            <div class="teacher-deck__metric-label">
              {{ item.label }}
            </div>
            <div class="teacher-deck__metric-value">
              {{ item.value }}
            </div>
            <div class="teacher-deck__metric-hint">
              {{ item.hint }}
            </div>
          </article>
        </div>
      </div>
    </section>

    <div
      v-if="error"
      class="teacher-deck__error"
      role="alert"
      aria-live="polite"
    >
      <div class="teacher-deck__error-header">
        <div class="teacher-deck__error-icon-wrap">
          <AlertTriangle class="teacher-deck__error-icon" />
        </div>
        <div class="min-w-0">
          <div class="teacher-deck__error-title">
            教师概览加载失败
          </div>
          <div class="teacher-deck__error-text">
            {{ error }}
          </div>
        </div>
      </div>

      <div class="teacher-deck__error-tips">
        <div class="teacher-deck__error-tip">
          <LifeBuoy class="teacher-deck__error-tip-icon" />
          <span>可先重试刷新数据，再继续查看趋势与复盘信息。</span>
        </div>
        <div class="teacher-deck__error-tip">
          <LifeBuoy class="teacher-deck__error-tip-icon" />
          <span>若持续失败，可先进入班级管理确认当前班级与权限状态。</span>
        </div>
      </div>

      <div class="teacher-deck__error-actions">
        <ElButton
          type="danger"
          @click="emit('retry')"
        >
          重试加载
        </ElButton>
        <ElButton
          plain
          @click="emit('openClassManagement')"
        >
          班级管理
        </ElButton>
      </div>
    </div>

    <section
      id="teacher-trend"
      class="teacher-anchor-section"
    >
      <TeacherClassTrendPanel
        :trend="trend"
        title="班级近 7 天训练趋势"
        subtitle="把训练事件、成功解题和活跃学生放在同一条时间轴上观察。"
      />
    </section>

    <section
      id="teacher-review"
      class="teacher-anchor-section"
    >
      <TeacherClassReviewPanel
        :review="review"
        :class-name="selectedClassName"
      />
    </section>

    <section
      id="teacher-insight"
      class="teacher-anchor-section"
    >
      <TeacherClassInsightsPanel
        :students="students"
        :class-name="selectedClassName"
      />
    </section>

    <section
      id="teacher-intervention"
      class="teacher-anchor-section"
    >
      <TeacherInterventionPanel
        :students="students"
        :class-name="selectedClassName"
      />
    </section>
  </div>
</template>

<style scoped>
.teacher-dashboard {
  --deck-accent: color-mix(in srgb, var(--color-primary) 44%, var(--color-border-default));
  --deck-accent-soft: color-mix(in srgb, var(--color-primary) 16%, transparent);
}

.teacher-deck__hero {
  border-radius: 14px;
  border-top: 1px solid var(--deck-accent);
  border-left: 1px solid color-mix(in srgb, var(--deck-accent) 56%, transparent);
  border-right: 1px solid color-mix(in srgb, var(--deck-accent) 56%, transparent);
  border-bottom: 1px solid var(--deck-accent);
  padding: 1rem 0;
  background:
    linear-gradient(
      to right,
      color-mix(in srgb, var(--color-primary) 8%, transparent),
      transparent 36%,
      color-mix(in srgb, var(--color-primary) 6%, transparent)
    );
}

.teacher-deck__hero-grid {
  display: grid;
  gap: 1.05rem;
}

.teacher-deck__eyebrow {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--color-primary) 72%, var(--color-text-secondary));
}

.teacher-deck__class-chip {
  display: inline-flex;
  align-items: center;
  border: 1px solid color-mix(in srgb, var(--deck-accent) 72%, transparent);
  background: color-mix(in srgb, var(--deck-accent-soft) 70%, transparent);
  padding: 0.14rem 0.45rem;
  letter-spacing: 0;
  text-transform: none;
  color: var(--color-text-primary);
}

.teacher-deck__title {
  margin-top: 0.72rem;
  font-size: clamp(1.35rem, 2.4vw, 1.9rem);
  font-weight: 700;
  line-height: 1.25;
  color: var(--color-text-primary);
}

.teacher-deck__description {
  margin-top: 0.45rem;
  max-width: 58ch;
  font-size: 0.88rem;
  line-height: 1.68;
  color: var(--color-text-secondary);
}

.teacher-deck__badges {
  margin-top: 0.95rem;
  display: grid;
  gap: 0.45rem;
}

.teacher-deck__badge {
  border-left: 2px solid var(--deck-accent);
  border-bottom: 1px solid var(--color-border-subtle);
  padding: 0.45rem 0 0.52rem 0.62rem;
}

.teacher-deck__badge-label {
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.teacher-deck__badge-value {
  margin-top: 0.2rem;
  font-size: 0.94rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.teacher-deck__anchors {
  margin-top: 0.82rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 0.62rem;
}

.teacher-deck__anchor-link {
  display: inline-flex;
  align-items: center;
  border-bottom: 1px dashed color-mix(in srgb, var(--color-primary) 44%, var(--color-border-default));
  padding-bottom: 0.06rem;
  font-size: 0.76rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--color-primary) 82%, var(--color-text-primary));
  transition: color 160ms ease, border-color 160ms ease;
}

.teacher-deck__anchor-link:hover {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.teacher-deck__suggestion {
  border-left: 2px solid var(--deck-accent);
  border-bottom: 1px solid var(--color-border-subtle);
  padding: 0 0 0.72rem 0.62rem;
}

.teacher-deck__suggestion-title {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.teacher-deck__suggestion-list {
  margin-top: 0.5rem;
  display: grid;
  gap: 0.35rem;
}

.teacher-deck__suggestion-item {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  font-size: 0.82rem;
  line-height: 1.58;
  color: var(--color-text-secondary);
}

.teacher-deck__suggestion-index {
  display: inline-flex;
  min-width: 1rem;
  justify-content: center;
  margin-top: 0.02rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--color-primary);
}

.teacher-deck__metrics {
  margin-top: 0.85rem;
  display: grid;
  gap: 0.56rem;
}

.teacher-deck__metric {
  border-bottom: 1px solid var(--color-border-subtle);
  padding: 0.52rem 0.2rem 0.58rem 0;
  animation: teacherDeckEnter 0.24s ease both;
}

.teacher-deck__metric-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.teacher-deck__metric-value {
  margin-top: 0.22rem;
  font-size: clamp(1.1rem, 2vw, 1.28rem);
  font-weight: 700;
  color: var(--color-text-primary);
}

.teacher-deck__metric-hint {
  margin-top: 0.2rem;
  font-size: 0.78rem;
  line-height: 1.55;
  color: var(--color-text-secondary);
}

.teacher-deck__error {
  border-left: 2px solid color-mix(in srgb, var(--color-danger) 62%, var(--color-border-default));
  border-bottom: 1px solid color-mix(in srgb, var(--color-danger) 38%, var(--color-border-default));
  padding: 0.72rem 0.2rem 0.86rem 0.72rem;
}

.teacher-deck__error-header {
  display: flex;
  align-items: flex-start;
  gap: 0.62rem;
}

.teacher-deck__error-icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.55rem;
  height: 1.55rem;
  flex-shrink: 0;
  border: 1px solid color-mix(in srgb, var(--color-danger) 36%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
}

.teacher-deck__error-icon {
  width: 0.9rem;
  height: 0.9rem;
  color: var(--color-danger);
}

.teacher-deck__error-title {
  font-size: 0.96rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--color-danger) 88%, var(--color-text-primary));
}

.teacher-deck__error-text {
  margin-top: 0.22rem;
  font-size: 0.83rem;
  line-height: 1.6;
  color: color-mix(in srgb, var(--color-danger) 70%, var(--color-text-secondary));
}

.teacher-deck__error-tips {
  margin-top: 0.62rem;
  display: grid;
  gap: 0.32rem;
}

.teacher-deck__error-tip {
  display: flex;
  align-items: flex-start;
  gap: 0.38rem;
  font-size: 0.8rem;
  line-height: 1.54;
  color: var(--color-text-secondary);
}

.teacher-deck__error-tip-icon {
  margin-top: 0.08rem;
  width: 0.78rem;
  height: 0.78rem;
  color: color-mix(in srgb, var(--color-danger) 74%, var(--color-text-secondary));
  flex-shrink: 0;
}

.teacher-deck__error-actions {
  margin-top: 0.7rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

@media (min-width: 768px) {
  .teacher-deck__badges {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .teacher-deck__metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1080px) {
  .teacher-deck__hero-grid {
    grid-template-columns: minmax(0, 1fr) 320px;
    gap: 1.15rem;
  }

  .teacher-deck__metrics {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

@media (max-width: 639px) {
  .teacher-deck__hero {
    border-radius: 10px;
  }

  .teacher-deck__title {
    margin-top: 0.75rem;
  }

  .teacher-deck__hero-grid {
    gap: 0.9rem;
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

<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import ClassInsightsPanel from '@/components/teacher/ClassInsightsPanel.vue'
import ClassReviewPanel from '@/components/teacher/ClassReviewPanel.vue'
import ClassTrendPanel from '@/components/teacher/ClassTrendPanel.vue'
import type {
  ClassInsightReviewData,
  ClassInsightSummaryData,
  ClassInsightTrendData,
  StudentDirectoryItem,
} from '@/api/contracts'

defineProps<{
  previewError: string | null
  previewLoading: boolean
  previewSummary: ClassInsightSummaryData | null
  previewTrend: ClassInsightTrendData | null
  previewReview: ClassInsightReviewData | null
  previewStudents: StudentDirectoryItem[]
  previewClassName: string
  selectedWindowLabel: string
  averageSolvedText: string
  activeRateText: string
}>()
</script>

<template>
  <section class="class-report-section class-report-section--preview">
    <div class="class-report-section__head">
      <div>
        <div class="journal-eyebrow">
          Live Preview
        </div>
        <h4 class="class-report-section__title">
          当前班级报告预览
        </h4>
        <p class="class-report-section__copy">
          不下载也能先看当前时间段内的班级趋势、教学复盘结论和学生洞察。
        </p>
      </div>
    </div>

    <div
      v-if="previewError"
      class="teacher-surface-error class-report-preview-error"
    >
      {{ previewError }}
    </div>

    <div
      v-else-if="previewLoading"
      class="class-report-preview-skeletons"
    >
      <div
        v-for="index in 3"
        :key="index"
        class="class-report-preview-skeleton"
      />
    </div>

    <template v-else-if="previewSummary">
      <section class="metric-panel-grid metric-panel-workspace-surface class-report-kpi-grid">
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            班级人数
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ previewSummary.student_count }}
          </div>
          <div class="progress-card-hint metric-panel-helper">
            当前班级纳入统计的学生数
          </div>
        </article>
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            平均解题
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ averageSolvedText }}
          </div>
          <div class="progress-card-hint metric-panel-helper">
            当前班级学生的人均解题数
          </div>
        </article>
        <article class="progress-card metric-panel-card">
          <div class="progress-card-label metric-panel-label">
            当前窗口活跃率
          </div>
          <div class="progress-card-value metric-panel-value">
            {{ activeRateText }}
          </div>
          <div class="progress-card-hint metric-panel-helper">
            当前时间段至少有一次训练动作的学生占比
          </div>
        </article>
      </section>

      <div class="class-report-preview-stack">
        <ClassTrendPanel
          :trend="previewTrend"
          title="班级训练趋势"
          :subtitle="`当前窗口：${selectedWindowLabel}`"
        />

        <ClassReviewPanel
          :review="previewReview"
          :class-name="previewClassName"
        />

        <ClassInsightsPanel
          :students="previewStudents"
          :class-name="previewClassName"
        />
      </div>
    </template>

    <AppEmpty
      v-else
      title="还没有可用预览"
      description="先填写班级名称并加载一次预览，这里会展示当前报告内容。"
      icon="FileChartColumnIncreasing"
    />
  </section>
</template>

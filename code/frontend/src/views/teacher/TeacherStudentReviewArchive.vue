<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { downloadReport } from '@/api/assessment'
import { exportStudentReviewArchive } from '@/api/teacher'
import AppEmpty from '@/components/common/AppEmpty.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import ReviewArchiveEvidencePanel from '@/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue'
import ReviewArchiveHero from '@/components/teacher/review-archive/ReviewArchiveHero.vue'
import ReviewArchiveObservationStrip from '@/components/teacher/review-archive/ReviewArchiveObservationStrip.vue'
import ReviewArchiveReflectionPanel from '@/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue'
import { useTeacherStudentReviewArchive } from '@/composables/useTeacherStudentReviewArchive'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { formatDate } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { start: startPolling, stop: stopPolling } = useReportStatusPolling()

const className = computed(() => String(route.params.className || ''))
const studentId = computed(() => String(route.params.studentId || ''))
const { archive, loading, error, reload } = useTeacherStudentReviewArchive(studentId)

const exporting = ref(false)
const pendingReportId = ref<string | null>(null)

const solvedRate = computed(() => {
  if (!archive.value?.summary.total_challenges) return 0
  return Math.round((archive.value.summary.total_solved / archive.value.summary.total_challenges) * 100)
})

const formattedLastActivity = computed(() => {
  if (!archive.value?.summary.last_activity_at) return '--'
  return formatDate(archive.value.summary.last_activity_at)
})

const rankedSkillDimensions = computed(() =>
  [...(archive.value?.skill_profile.dimensions ?? [])].sort((left, right) => right.value - left.value)
)

function openStudentAnalysis(): void {
  if (!studentId.value || !className.value) return
  router.push({
    name: 'TeacherStudentAnalysis',
    params: {
      className: className.value,
      studentId: studentId.value,
    },
  })
}

function goBack(): void {
  if (!className.value) {
    router.push({ name: 'TeacherStudentManagement' })
    return
  }
  router.push({
    name: 'TeacherClassStudents',
    params: { className: className.value },
  })
}

async function downloadGeneratedReport(reportId: string): Promise<void> {
  const { blob, filename } = await downloadReport(reportId)
  const objectUrl = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = objectUrl
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(objectUrl)
}

async function handleExportArchive(): Promise<void> {
  if (!studentId.value) return

  exporting.value = true
  try {
    const result = await exportStudentReviewArchive(studentId.value, { format: 'json' })
    if (result.status === 'ready') {
      stopPolling()
      await downloadGeneratedReport(result.report_id)
      toast.success('复盘归档已生成并开始下载')
      return
    }
    if (result.status === 'failed') {
      stopPolling()
      toast.error(result.error_message || '复盘归档生成失败')
      return
    }

    pendingReportId.value = result.report_id
    startPolling(result.report_id, (next) => {
      if (next.report_id !== pendingReportId.value) return
      if (next.status === 'ready') {
        pendingReportId.value = null
        void downloadGeneratedReport(next.report_id)
        toast.success('复盘归档已生成并开始下载')
        return
      }
      if (next.status === 'failed') {
        pendingReportId.value = null
        toast.error(next.error_message || '复盘归档生成失败')
      }
    })
    toast.info('复盘归档开始生成，完成后会自动下载')
  } finally {
    exporting.value = false
  }
}
</script>

<template>
  <div class="review-archive-shell space-y-8">
    <ReviewArchiveHero
      :archive="archive"
      :exporting="exporting"
      @back="goBack"
      @open-analysis="openStudentAnalysis"
      @export-archive="handleExportArchive"
    />

    <div v-if="loading" class="review-archive-loading">
      <div class="review-archive-loading__hero" />
      <div class="review-archive-loading__grid">
        <div class="review-archive-loading__block" />
        <div class="review-archive-loading__block" />
      </div>
    </div>

    <AppEmpty
      v-else-if="error"
      title="复盘归档加载失败"
      :description="error"
      icon="AlertTriangle"
    >
      <template #action>
        <ElButton type="primary" @click="reload">重新加载</ElButton>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="!archive"
      title="暂无复盘归档"
      description="当前学生还没有可展示的复盘归档数据。"
      icon="FileChartColumnIncreasing"
    />

    <template v-else>
      <ReviewArchiveObservationStrip :items="archive.teacher_observations.items" />

      <section class="review-archive-summary-grid">
        <SectionCard title="训练摘要" subtitle="将当前归档的关键指标收束为一页课堂摘要。">
          <div class="summary-grid">
            <article class="summary-card summary-card--primary">
              <div class="summary-card__label">完成率</div>
              <div class="summary-card__value">{{ solvedRate }}%</div>
              <div class="summary-card__hint">已完成 {{ archive.summary.total_solved }} / {{ archive.summary.total_challenges }}</div>
            </article>
            <article class="summary-card summary-card--warning">
              <div class="summary-card__label">提示次数</div>
              <div class="summary-card__value">{{ archive.summary.hint_unlock_count }}</div>
              <div class="summary-card__hint">可用于回看关键转折点</div>
            </article>
            <article class="summary-card summary-card--neutral">
              <div class="summary-card__label">最近活跃</div>
              <div class="summary-card__value summary-card__value--time">{{ formattedLastActivity }}</div>
              <div class="summary-card__hint">归档内最后一条训练活动</div>
            </article>
          </div>
        </SectionCard>

        <SectionCard title="能力画像" subtitle="优先识别当前最强与最弱的训练维度。">
          <div class="skill-bars">
            <article
              v-for="item in rankedSkillDimensions"
              :key="item.key"
              class="skill-bars__item"
            >
              <div class="skill-bars__head">
                <strong>{{ item.name }}</strong>
                <span>{{ item.value }}%</span>
              </div>
              <div class="skill-bars__track">
                <div class="skill-bars__fill" :style="{ width: `${item.value}%` }" />
              </div>
            </article>
          </div>
        </SectionCard>
      </section>

      <ReviewArchiveEvidencePanel
        :timeline="archive.timeline"
        :evidence="archive.evidence"
      />

      <ReviewArchiveReflectionPanel
        :writeups="archive.writeups"
        :manual-reviews="archive.manual_reviews"
      />
    </template>
  </div>
</template>

<style scoped>
.review-archive-shell {
  min-height: 100%;
  padding: 0.25rem 0 2rem;
}

.review-archive-loading__hero,
.review-archive-loading__block {
  border-radius: 26px;
  background: linear-gradient(90deg, rgba(226, 232, 240, 0.8), rgba(241, 245, 249, 0.96));
  animation: review-archive-pulse 1.35s ease-in-out infinite;
}

.review-archive-loading__hero {
  height: 220px;
}

.review-archive-loading__grid {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.review-archive-loading__block {
  height: 180px;
}

.review-archive-summary-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 0.92fr) minmax(0, 1.08fr);
}

.summary-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.summary-card {
  padding: 1rem;
  border-radius: 20px;
  border: 1px solid color-mix(in srgb, #1e40af 10%, var(--color-border-default));
  background: rgba(248, 250, 252, 0.86);
}

.summary-card--primary {
  background: linear-gradient(180deg, rgba(30, 64, 175, 0.1), rgba(255, 255, 255, 0.94));
}

.summary-card--warning {
  background: linear-gradient(180deg, rgba(245, 158, 11, 0.14), rgba(255, 255, 255, 0.94));
}

.summary-card__label {
  font-size: 0.74rem;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  color: #475569;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.summary-card__value {
  margin-top: 0.7rem;
  font-size: 1.8rem;
  font-weight: 700;
  color: #0f172a;
}

.summary-card__value--time {
  font-size: 1.06rem;
  line-height: 1.5;
}

.summary-card__hint {
  margin-top: 0.55rem;
  color: #64748b;
  line-height: 1.65;
}

.skill-bars {
  display: grid;
  gap: 0.9rem;
}

.skill-bars__item {
  padding: 0.2rem 0;
}

.skill-bars__head {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 0.45rem;
  color: #1e293b;
}

.skill-bars__head span {
  color: #475569;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.skill-bars__track {
  height: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.16);
}

.skill-bars__fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #1d4ed8, #60a5fa 58%, #f59e0b);
}

@keyframes review-archive-pulse {
  0%, 100% {
    opacity: 0.58;
  }
  50% {
    opacity: 1;
  }
}

@media (max-width: 1023px) {
  .review-archive-summary-grid,
  .summary-grid,
  .review-archive-loading__grid {
    grid-template-columns: 1fr;
  }
}
</style>

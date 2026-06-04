<template>
  <SectionCard
    class="student-insight-section-card"
    variant="teacher-flat"
    title="人工审核题"
    subtitle="查看当前学员需要教师判定的题解内容。"
  >
    <StudentInsightStateSurface
      class="manual-review-state-surface student-insight-state-surface--spacious"
      :empty="manualReviewSubmissions.length === 0"
      surface="plain"
    >
      <template #empty>
        <AppEmpty
          class="student-insight-empty"
          title="暂无题解审核提交"
          description="当前学员还没有需要教师处理的题解审核内容。"
          icon="ClipboardCheck"
        />
      </template>

      <template #default>
        <div
          class="student-insight-kpi-grid student-insight-kpi-grid--3 progress-strip metric-panel-grid metric-panel-default-surface"
        >
          <article class="insight-kpi-card progress-card metric-panel-card">
            <div class="student-insight-kpi-label progress-card-label metric-panel-label">
              <span>待处理</span>
              <ClipboardList class="h-4 w-4" />
            </div>
            <div class="student-insight-kpi-value progress-card-value metric-panel-value">
              {{ manualReviewSubmissions.length }}
            </div>
            <div class="student-insight-kpi-hint progress-card-hint metric-panel-helper">
              当前分析页展示的题解审核提交数
            </div>
          </article>
          <article class="insight-kpi-card progress-card metric-panel-card">
            <div class="student-insight-kpi-label progress-card-label metric-panel-label">
              <span>待审核</span>
              <Clock3 class="h-4 w-4" />
            </div>
            <div class="student-insight-kpi-value progress-card-value metric-panel-value">
              {{ pendingManualReviewCount }}
            </div>
            <div class="student-insight-kpi-hint progress-card-hint metric-panel-helper">
              尚未给出审核结果的提交
            </div>
          </article>
          <article class="insight-kpi-card progress-card metric-panel-card">
            <div class="student-insight-kpi-label progress-card-label metric-panel-label">
              <span>已通过</span>
              <CheckCircle class="h-4 w-4" />
            </div>
            <div class="student-insight-kpi-value progress-card-value metric-panel-value">
              {{ approvedManualReviewCount }}
            </div>
            <div class="student-insight-kpi-hint progress-card-hint metric-panel-helper">
              已经通过审核的题解提交
            </div>
          </article>
        </div>

        <div class="mt-5 grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
          <div class="grid gap-3">
            <AppCard
              v-for="item in manualReviewSubmissions"
              :key="item.id"
              variant="panel"
              accent="neutral"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div class="text-sm font-semibold text-[var(--color-text-primary)]">
                    {{ item.challenge_title }}
                  </div>
                  <div class="mt-1 text-sm text-[var(--color-text-secondary)]">
                    {{ item.answer_preview || '暂无答案摘要' }}
                  </div>
                </div>
                <span :class="manualReviewStatusClass(item.review_status)">
                  {{ manualReviewStatusLabel(item.review_status) }}
                </span>
              </div>

              <div
                class="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs text-[var(--color-text-secondary)]"
              >
                <span>提交于 {{ formatDateTime(item.submitted_at) }}</span>
                <button
                  type="button"
                  class="inline-flex items-center gap-1 font-medium text-[var(--color-primary)]"
                  @click="openManualReview(item.id)"
                >
                  {{ activeManualReview?.id === item.id ? '刷新详情' : '查看审核' }}
                  <ArrowRight class="h-4 w-4" />
                </button>
              </div>
            </AppCard>
          </div>

          <StudentInsightStateSurface
            class="manual-review-detail-shell student-insight-detail-shell"
            :loading="manualReviewLoading"
            :empty="!manualReviewLoading && !activeManualReview"
            surface="plain"
          >
            <template #loading>
              <div class="manual-review-detail-loading">
                <div class="student-insight-skeleton-line manual-review-detail-loading-title" />
                <div class="student-insight-skeleton-block manual-review-detail-loading-panel" />
                <div class="student-insight-skeleton-block manual-review-detail-loading-panel" />
              </div>
            </template>

            <template #empty>
              <AppEmpty
                class="student-insight-empty"
                title="选择一条题解审核提交"
                description="点击左侧卡片查看完整内容并进行审核。"
                icon="ClipboardList"
              />
            </template>

            <template #default>
              <template v-if="activeManualReview">
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <div class="journal-eyebrow">Writeup Review</div>
                    <h4 class="mt-2 text-lg font-semibold text-[var(--color-text-primary)]">
                      {{ activeManualReview.challenge_title }}
                    </h4>
                    <div class="mt-2 text-sm text-[var(--color-text-secondary)]">
                      {{ activeManualReview.student_name || activeManualReview.student_username }}
                    </div>
                  </div>
                  <span :class="manualReviewStatusClass(activeManualReview.review_status)">
                    {{ manualReviewStatusLabel(activeManualReview.review_status) }}
                  </span>
                </div>

                <div class="insight-answer-panel mt-5 rounded-2xl px-4 py-4">
                  <div class="insight-answer-panel__label">题解内容</div>
                  <p
                    class="mt-3 whitespace-pre-wrap text-sm leading-7 text-[var(--color-text-primary)]"
                  >
                    {{ activeManualReview.answer }}
                  </p>
                </div>

                <label class="mt-5 block">
                  <span class="text-sm font-medium text-[var(--color-text-primary)]">审核意见</span>
                  <textarea
                    v-model="manualReviewComment"
                    rows="5"
                    class="challenge-input insight-manual-input mt-3 w-full rounded-2xl border px-4 py-3 text-sm leading-7 transition-colors focus:outline-none"
                    placeholder="记录你的判定依据、补充建议或要求学员修改的点。"
                  />
                </label>

                <div class="mt-5 flex flex-wrap items-center justify-between gap-3">
                  <div class="text-xs text-[var(--color-text-secondary)]">
                    最近更新：{{ formatDateTime(activeManualReview.updated_at) }}
                  </div>
                  <div class="flex flex-wrap gap-3">
                    <button
                      type="button"
                      class="ui-btn ui-btn--secondary insight-outline-action disabled:cursor-not-allowed disabled:opacity-50"
                      :disabled="manualReviewSaving || activeManualReview.review_status !== 'pending'"
                      @click="submitManualReview('rejected')"
                    >
                      {{ manualReviewSaving ? '提交中...' : '驳回并说明' }}
                    </button>
                    <button
                      type="button"
                      class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
                      :disabled="manualReviewSaving || activeManualReview.review_status !== 'pending'"
                      @click="submitManualReview('approved')"
                    >
                      {{ manualReviewSaving ? '提交中...' : '审核通过' }}
                    </button>
                  </div>
                </div>
              </template>
            </template>
          </StudentInsightStateSurface>
        </div>
      </template>
    </StudentInsightStateSurface>
  </SectionCard>
</template>

<style src="@/features/teaching/student-analysis-shared/ui/studentInsightSections.css"></style>

<style scoped>
.writeup-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.writeup-chip--success {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: color-mix(in srgb, var(--color-success) 82%, var(--journal-ink));
}

.writeup-chip--warning {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: color-mix(in srgb, var(--color-warning) 82%, var(--journal-ink));
}

.writeup-chip--muted {
  background: color-mix(in srgb, var(--journal-border) 36%, transparent);
  color: var(--journal-muted);
}

.manual-review-detail-loading {
  display: grid;
  gap: var(--space-3);
}

.manual-review-detail-loading-title {
  width: var(--space-32);
  height: var(--space-5);
}

.manual-review-detail-loading-panel {
  height: 6rem;
}

.insight-answer-panel {
  border: 0;
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 28%, transparent);
  border-radius: 0;
  background: transparent;
}

.insight-answer-panel__label {
  font-size: var(--font-size-0-72);
  font-weight: 600;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.insight-manual-input {
  border-color: color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  background: transparent;
  color: var(--journal-ink);
}

.insight-manual-input::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 84%, transparent);
}

.insight-manual-input:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 34%, transparent);
}
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowRight, CheckCircle, ClipboardList, Clock3 } from 'lucide-vue-next'

import type {
  ManualReviewSubmissionDetailData,
  ManualReviewSubmissionItemData,
} from '@/api/contracts'
import { StudentInsightStateSurface } from '@/features/teaching/student-analysis-shared/ui'
import AppCard from '@/shared/ui/common/AppCard.vue'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import SectionCard from '@/shared/ui/common/SectionCard.vue'
import {
  formatDateTime,
  manualReviewStatusClass,
  manualReviewStatusLabel,
} from './studentInsightShared'

const props = defineProps<{
  manualReviewSubmissions: ManualReviewSubmissionItemData[]
  activeManualReview: ManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
}>()

const emit = defineEmits<{
  openManualReview: [submissionId: string]
  reviewManualReview: [
    payload: {
      submissionId: string
      reviewStatus: 'approved' | 'rejected'
      reviewComment?: string
    },
  ]
}>()

const manualReviewComment = ref('')
const approvedManualReviewCount = computed(
  () => props.manualReviewSubmissions.filter((item) => item.review_status === 'approved').length
)
const pendingManualReviewCount = computed(
  () => props.manualReviewSubmissions.filter((item) => item.review_status === 'pending').length
)

watch(
  () => props.activeManualReview,
  (value) => {
    manualReviewComment.value = value?.review_comment ?? ''
  },
  { immediate: true }
)

function openManualReview(submissionId: string): void {
  emit('openManualReview', submissionId)
}

function submitManualReview(reviewStatus: 'approved' | 'rejected'): void {
  if (!props.activeManualReview) return
  emit('reviewManualReview', {
    submissionId: props.activeManualReview.id,
    reviewStatus,
    reviewComment: manualReviewComment.value.trim() || undefined,
  })
}
</script>

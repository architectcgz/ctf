<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import type {
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
} from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import {
  formatDateTime,
  manualReviewStatusClass,
  manualReviewStatusLabel,
} from './studentInsightShared'

const props = defineProps<{
  manualReviewSubmissions: TeacherManualReviewSubmissionItemData[]
  activeManualReview: TeacherManualReviewSubmissionDetailData | null
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

<template>
  <SectionCard
    class="insight-tab-section-card"
    title="人工审核题"
    subtitle="查看当前学员需要教师判定的题解内容。"
  >
    <AppEmpty
      v-if="manualReviewSubmissions.length === 0"
      title="暂无题解审核提交"
      description="当前学员还没有需要教师处理的题解审核内容。"
      icon="ClipboardCheck"
    />

    <template v-else>
      <div
        class="insight-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface md:grid-cols-3"
      >
        <article class="insight-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">待处理</div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ manualReviewSubmissions.length }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            当前分析页展示的题解审核提交数
          </div>
        </article>
        <article class="insight-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">待审核</div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ pendingManualReviewCount }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            尚未给出审核结果的提交
          </div>
        </article>
        <article class="insight-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">已通过</div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ approvedManualReviewCount }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
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

        <AppCard variant="panel" accent="neutral">
          <div v-if="manualReviewLoading" class="space-y-3">
            <div class="insight-skeleton-line h-5 w-32 animate-pulse rounded" />
            <div class="insight-skeleton-block h-24 animate-pulse rounded-2xl" />
            <div class="insight-skeleton-block h-24 animate-pulse rounded-2xl" />
          </div>

          <AppEmpty
            v-else-if="!activeManualReview"
            title="选择一条题解审核提交"
            description="点击左侧卡片查看完整内容并进行审核。"
            icon="ClipboardList"
          />

          <template v-else>
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
        </AppCard>
      </div>
    </template>
  </SectionCard>
</template>

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

.insight-kpi-grid {
  --metric-panel-grid-gap: var(--space-3);
  align-items: stretch;
}

.insight-kpi-label {
  --metric-panel-label-size: var(--font-size-0-70);
  --metric-panel-label-spacing: 0.15em;
}

.insight-kpi-value {
  --metric-panel-value-margin-top: var(--space-2);
  --metric-panel-value-size: var(--font-size-1-00);
  --metric-panel-value-line-height: 1.5;
  --metric-panel-value-spacing: 0;
}

.insight-kpi-hint {
  --metric-panel-helper-margin-top: var(--space-2);
  --metric-panel-helper-size: var(--font-size-0-80);
  --metric-panel-helper-line-height: 1.55;
}

.insight-skeleton-line,
.insight-skeleton-block {
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 78%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
}

.insight-answer-panel {
  border: 0;
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 28%, transparent);
  border-radius: 0;
  background: color-mix(in srgb, var(--journal-surface-subtle) 48%, transparent);
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
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-ink);
}

.insight-manual-input::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 84%, transparent);
}

.insight-manual-input:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 34%, transparent);
}
</style>

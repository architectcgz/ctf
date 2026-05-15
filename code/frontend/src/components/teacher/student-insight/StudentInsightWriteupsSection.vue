<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowRight, ClipboardList, FileText, FolderKanban } from 'lucide-vue-next'

import type {
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import {
  visibilityStatusClass,
  visibilityStatusLabel,
  formatDateTime,
  manualReviewStatusClass,
  manualReviewStatusLabel,
} from './studentInsightShared'

const props = defineProps<{
  writeupSubmissions: TeacherSubmissionWriteupItemData[]
  writeupPage: number
  writeupTotal: number
  writeupTotalPages: number
  writeupPaginationLoading: boolean
  manualReviewSubmissions: TeacherManualReviewSubmissionItemData[]
  activeManualReview: TeacherManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  openManualReview: [submissionId: string]
  moderateWriteup: [
    payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' },
  ]
  reviewManualReview: [
    payload: {
      submissionId: string
      reviewStatus: 'approved' | 'rejected'
      reviewComment?: string
    },
  ]
  changeWriteupPage: [page: number]
}>()

const publishedWriteupSubmissions = computed(() =>
  props.writeupSubmissions.filter(
    (item) => item.submission_status === 'published'
  )
)
const publishedChallengeCount = computed(
  () => new Set(publishedWriteupSubmissions.value.map((item) => String(item.challenge_id))).size
)
const manualReviewByChallengeId = computed(() => {
  const items = new Map<string, TeacherManualReviewSubmissionItemData>()
  for (const item of props.manualReviewSubmissions) {
    if (!items.has(String(item.challenge_id))) {
      items.set(String(item.challenge_id), item)
    }
  }
  return items
})
const standaloneManualReviewSubmissions = computed(() => {
  const publishedChallengeIds = new Set(
    publishedWriteupSubmissions.value.map((item) => String(item.challenge_id))
  )
  return props.manualReviewSubmissions.filter(
    (item) => !publishedChallengeIds.has(String(item.challenge_id))
  )
})
const pendingManualReviewCount = computed(
  () => props.manualReviewSubmissions.filter((item) => item.review_status === 'pending').length
)
const hasWriteupRows = computed(
  () => publishedWriteupSubmissions.value.length > 0 || props.manualReviewSubmissions.length > 0
)
const manualReviewComment = ref('')

watch(
  () => props.activeManualReview,
  (value) => {
    manualReviewComment.value = value?.review_comment ?? ''
  },
  { immediate: true }
)

function openChallenge(challengeId: string): void {
  emit('openChallenge', challengeId)
}

function openManualReview(submissionId: string): void {
  emit('openManualReview', submissionId)
}

function changeWriteupPage(page: number): void {
  emit('changeWriteupPage', page)
}

function moderateWriteup(
  submissionId: string,
  action: 'recommend' | 'unrecommend' | 'hide' | 'restore'
): void {
  emit('moderateWriteup', { submissionId, action })
}

function reviewManualReview(reviewStatus: 'approved' | 'rejected'): void {
  if (!props.activeManualReview) return
  emit('reviewManualReview', {
    submissionId: props.activeManualReview.id,
    reviewStatus,
    reviewComment: manualReviewComment.value.trim() || undefined,
  })
}

function findManualReview(challengeId: string): TeacherManualReviewSubmissionItemData | undefined {
  return manualReviewByChallengeId.value.get(String(challengeId))
}
</script>

<template>
  <SectionCard
    class="writeup-section-card insight-tab-section-card"
    title="题解列表"
    subtitle="集中处理发布状态、可见性与人工审核。"
  >
    <AppEmpty
      v-if="!hasWriteupRows"
      title="暂无题解记录"
      description="当前学员还没有发布题解或提交人工审核内容。"
      icon="FileText"
    />

    <template v-else>
      <div class="writeup-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface">
        <article class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">
            <span>已发布题解</span>
            <FileText class="h-4 w-4" />
          </div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ publishedWriteupSubmissions.length }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            当前学员已发布题解
          </div>
        </article>
        <article class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">
            <span>对应题目</span>
            <FolderKanban class="h-4 w-4" />
          </div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ publishedChallengeCount }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            已发布题解覆盖题目
          </div>
        </article>
        <article class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">
            <span>待审核</span>
            <ClipboardList class="h-4 w-4" />
          </div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ pendingManualReviewCount }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            人工审核待处理
          </div>
        </article>
      </div>

      <section class="writeup-directory mt-5">
        <header class="writeup-directory-head">
          <span>题目</span>
          <span>题解</span>
          <span>社区题解状态</span>
          <span>审核状态</span>
          <span>操作</span>
        </header>

        <article
          v-for="item in publishedWriteupSubmissions"
          :key="item.id"
          class="writeup-directory-row"
        >
          <div class="writeup-directory-cell">
            <div class="writeup-directory-challenge">
              {{ item.challenge_title }}
            </div>
          </div>

          <div class="writeup-directory-cell">
            <div class="writeup-directory-title">
              {{ item.title }}
            </div>
            <div class="writeup-directory-preview">
              {{ item.content_preview || '暂无摘要' }}
            </div>
          </div>

          <div class="writeup-directory-cell">
            <div class="writeup-directory-status-label">社区题解状态</div>
            <div class="writeup-directory-status">
              <span class="writeup-chip writeup-chip--muted">已发布</span>
              <span :class="visibilityStatusClass(item.visibility_status)">
                {{ visibilityStatusLabel(item.visibility_status) }}
              </span>
              <span v-if="item.is_recommended" class="writeup-chip writeup-chip--primary">
                推荐题解
              </span>
            </div>
          </div>

          <div class="writeup-directory-cell">
            <div class="writeup-directory-status-label">审核状态</div>
            <div v-if="findManualReview(item.challenge_id)" class="writeup-directory-status">
              <span :class="manualReviewStatusClass(findManualReview(item.challenge_id)!.review_status)">
                {{ manualReviewStatusLabel(findManualReview(item.challenge_id)!.review_status) }}
              </span>
              <span class="writeup-directory-time">
                {{ formatDateTime(findManualReview(item.challenge_id)!.submitted_at) }}
              </span>
            </div>
            <span v-else class="writeup-chip writeup-chip--muted">无人工审核</span>
          </div>

          <div class="writeup-directory-cell writeup-directory-action">
            <div class="writeup-action-stack">
              <button
                type="button"
                class="writeup-open-link inline-flex items-center gap-1 font-medium"
                @click="openChallenge(item.challenge_id)"
              >
                查看题目
                <ArrowRight class="h-4 w-4" />
              </button>
              <button
                v-if="findManualReview(item.challenge_id)"
                type="button"
                class="writeup-action-button writeup-action-button--primary"
                @click="openManualReview(findManualReview(item.challenge_id)!.id)"
              >
                {{ activeManualReview?.id === findManualReview(item.challenge_id)!.id ? '刷新审核' : '查看审核' }}
              </button>
              <button
                type="button"
                class="writeup-action-button"
                @click="moderateWriteup(item.id, item.is_recommended ? 'unrecommend' : 'recommend')"
              >
                {{ item.is_recommended ? '取消推荐' : '推荐题解' }}
              </button>
              <button
                type="button"
                class="writeup-action-button writeup-action-button--warning"
                @click="
                  moderateWriteup(item.id, item.visibility_status === 'hidden' ? 'restore' : 'hide')
                "
              >
                {{ item.visibility_status === 'hidden' ? '恢复公开' : '隐藏题解' }}
              </button>
            </div>
          </div>
        </article>

        <article
          v-for="item in standaloneManualReviewSubmissions"
          :key="`manual-${item.id}`"
          class="writeup-directory-row writeup-directory-row--manual"
        >
          <div class="writeup-directory-cell">
            <div class="writeup-directory-challenge">
              {{ item.challenge_title }}
            </div>
          </div>

          <div class="writeup-directory-cell">
            <div class="writeup-directory-title">
              人工审核提交
            </div>
            <div class="writeup-directory-preview">
              {{ item.answer_preview || '暂无答案摘要' }}
            </div>
          </div>

          <div class="writeup-directory-cell">
            <div class="writeup-directory-status-label">社区题解状态</div>
            <span class="writeup-chip writeup-chip--muted">未发布到题解区</span>
          </div>

          <div class="writeup-directory-cell">
            <div class="writeup-directory-status-label">审核状态</div>
            <div class="writeup-directory-status">
              <span :class="manualReviewStatusClass(item.review_status)">
                {{ manualReviewStatusLabel(item.review_status) }}
              </span>
              <span class="writeup-directory-time">
                {{ formatDateTime(item.submitted_at) }}
              </span>
            </div>
          </div>

          <div class="writeup-directory-cell writeup-directory-action">
            <div class="writeup-action-stack">
              <button
                type="button"
                class="writeup-open-link inline-flex items-center gap-1 font-medium"
                @click="openChallenge(item.challenge_id)"
              >
                查看题目
                <ArrowRight class="h-4 w-4" />
              </button>
              <button
                type="button"
                class="writeup-action-button writeup-action-button--primary"
                @click="openManualReview(item.id)"
              >
                {{ activeManualReview?.id === item.id ? '刷新审核' : '查看审核' }}
              </button>
            </div>
          </div>
        </article>
      </section>

      <div class="writeup-pagination mt-4">
        <PagePaginationControls
          :page="writeupPage"
          :total-pages="writeupTotalPages"
          :total="writeupTotal"
          total-label="发布题解总数"
          :disabled="writeupPaginationLoading"
          show-jump
          @change-page="changeWriteupPage"
        />
      </div>

      <section
        v-if="manualReviewLoading || activeManualReview"
        class="writeup-review-panel"
        aria-live="polite"
      >
        <div v-if="manualReviewLoading" class="space-y-3">
          <div class="insight-skeleton-line h-5 w-32 animate-pulse rounded" />
          <div class="insight-skeleton-block h-24 animate-pulse rounded-2xl" />
        </div>

        <template v-else-if="activeManualReview">
          <div class="writeup-review-panel__head">
            <div>
              <div class="journal-eyebrow">Writeup Review</div>
              <h4 class="writeup-review-panel__title">
                {{ activeManualReview.challenge_title }}
              </h4>
            </div>
            <span :class="manualReviewStatusClass(activeManualReview.review_status)">
              {{ manualReviewStatusLabel(activeManualReview.review_status) }}
            </span>
          </div>

          <div class="insight-answer-panel mt-5 rounded-2xl px-4 py-4">
            <div class="insight-answer-panel__label">题解内容</div>
            <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-[var(--color-text-primary)]">
              {{ activeManualReview.answer }}
            </p>
          </div>

          <label class="mt-5 block">
            <span class="text-sm font-medium text-[var(--color-text-primary)]">审核意见</span>
            <textarea
              v-model="manualReviewComment"
              rows="4"
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
                @click="reviewManualReview('rejected')"
              >
                {{ manualReviewSaving ? '提交中...' : '驳回' }}
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="manualReviewSaving || activeManualReview.review_status !== 'pending'"
                @click="reviewManualReview('approved')"
              >
                {{ manualReviewSaving ? '提交中...' : '审核通过' }}
              </button>
            </div>
          </div>
        </template>
      </section>
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

.writeup-chip--primary {
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent-strong);
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

.writeup-directory {
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 84%, transparent);
}

.writeup-kpi-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.writeup-directory-head,
.writeup-directory-row {
  display: grid;
  grid-template-columns:
    minmax(0, 1.2fr)
    minmax(0, 2fr)
    minmax(0, 1.2fr)
    minmax(0, 1.35fr)
    minmax(108px, 0.9fr);
  gap: var(--space-3-5);
  align-items: start;
}

.writeup-directory-head {
  padding: var(--space-3) var(--space-1-5) var(--space-2-5);
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-directory-row {
  padding: var(--space-3-5) var(--space-1-5);
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 84%, transparent);
}

.writeup-directory-row--manual {
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.writeup-directory-cell {
  min-width: 0;
}

.writeup-directory-challenge,
.writeup-directory-title {
  font-size: var(--font-size-0-86);
  font-weight: 600;
  color: var(--journal-ink);
}

.writeup-directory-preview {
  margin-top: var(--space-1-5);
  line-height: 1.6;
  font-size: var(--font-size-0-80);
  color: var(--journal-muted);
}

.writeup-directory-status {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.writeup-directory-status-label {
  margin-bottom: var(--space-2);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-directory-time {
  font-size: var(--font-size-0-80);
  line-height: 1.6;
  color: var(--journal-muted);
}

.writeup-directory-action {
  display: flex;
  justify-content: flex-end;
}

.writeup-action-stack {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.writeup-pagination {
  padding-top: var(--space-1-5);
}

.writeup-open-link {
  min-height: 34px;
  padding: 0 var(--space-3);
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 28%, var(--teacher-divider));
  color: var(--journal-accent-strong);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease;
}

.writeup-open-link:hover,
.writeup-open-link:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 46%, var(--teacher-divider));
  background: color-mix(in srgb, var(--journal-accent) 16%, transparent);
  color: var(--journal-accent);
  outline: none;
}

.writeup-action-button {
  min-height: 34px;
  padding: 0 var(--space-3);
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 90%, transparent);
  font-size: var(--font-size-0-78);
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease;
}

.writeup-action-button:hover,
.writeup-action-button:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 30%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
  outline: none;
}

.writeup-action-button--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, var(--teacher-divider));
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.writeup-action-button--primary:hover,
.writeup-action-button--primary:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 46%, var(--teacher-divider));
  background: color-mix(in srgb, var(--journal-accent) 16%, transparent);
  color: var(--journal-accent);
}

.writeup-action-button--warning:hover,
.writeup-action-button--warning:focus-visible {
  border-color: color-mix(in srgb, var(--color-warning) 36%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-warning) 86%, var(--journal-ink));
}

.writeup-review-panel {
  margin-top: var(--space-5);
  padding: var(--space-4);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  background: color-mix(in srgb, var(--journal-surface-subtle) 34%, transparent);
}

.writeup-review-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.writeup-review-panel__title {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-1-00);
  font-weight: 650;
  color: var(--journal-ink);
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
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

@media (max-width: 1023px) {
  .writeup-kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .writeup-directory-head {
    display: none;
  }

  .writeup-directory-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  .writeup-directory-action,
  .writeup-action-stack {
    justify-content: flex-start;
  }
}

@media (max-width: 767px) {
  .writeup-kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>

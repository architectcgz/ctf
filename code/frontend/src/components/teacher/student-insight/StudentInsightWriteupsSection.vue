<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import type { TeacherSubmissionWriteupItemData } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import {
  visibilityStatusClass,
  visibilityStatusLabel,
  formatDateTime,
} from './studentInsightShared'

const props = defineProps<{
  writeupSubmissions: TeacherSubmissionWriteupItemData[]
  writeupPage: number
  writeupTotal: number
  writeupTotalPages: number
  writeupPaginationLoading: boolean
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  moderateWriteup: [
    payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' },
  ]
  changeWriteupPage: [page: number]
}>()

const publishedWriteupSubmissions = computed(() =>
  props.writeupSubmissions.filter(
    (item) => item.submission_status === 'published' || item.submission_status === 'submitted'
  )
)
const publishedRecommendedWriteupCount = computed(
  () => publishedWriteupSubmissions.value.filter((item) => item.is_recommended).length
)
const publishedChallengeCount = computed(
  () => new Set(publishedWriteupSubmissions.value.map((item) => String(item.challenge_id))).size
)

function openChallenge(challengeId: string): void {
  emit('openChallenge', challengeId)
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
</script>

<template>
  <SectionCard
    class="writeup-section-card insight-tab-section-card"
    title="发布的题解"
    subtitle="按发布时间查看当前学员已发布题解。"
  >
    <AppEmpty
      v-if="publishedWriteupSubmissions.length === 0"
      title="暂无已发布题解"
      description="当前学员还没有发布到题解区的内容。"
      icon="FileText"
    />

    <template v-else>
      <div class="writeup-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface">
        <article class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">已发布题解</div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ publishedWriteupSubmissions.length }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            当前学员已发布的题解数量
          </div>
        </article>
        <article class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">对应题目</div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ publishedChallengeCount }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            覆盖到的题目总数
          </div>
        </article>
        <article class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card">
          <div class="insight-kpi-label progress-card-label metric-panel-label">推荐中</div>
          <div class="insight-kpi-value progress-card-value metric-panel-value">
            {{ publishedRecommendedWriteupCount }}
          </div>
          <div class="insight-kpi-hint progress-card-hint metric-panel-helper">
            发布题解中被标记为推荐的数量
          </div>
        </article>
      </div>

      <section class="writeup-directory mt-5">
        <header class="writeup-directory-head">
          <span>题目</span>
          <span>题解</span>
          <span>社区题解状态</span>
          <span>发布时间</span>
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

          <div class="writeup-directory-cell writeup-directory-time">
            {{ formatDateTime(item.published_at || item.updated_at) }}
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

.writeup-action-button--warning:hover,
.writeup-action-button--warning:focus-visible {
  border-color: color-mix(in srgb, var(--color-warning) 36%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-warning) 86%, var(--journal-ink));
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

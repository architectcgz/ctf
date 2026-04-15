<script setup lang="ts">
import type {
  ReviewArchiveManualReviewItemData,
  ReviewArchiveWriteupItemData,
} from '@/api/contracts'

function submissionStatusLabel(status: ReviewArchiveWriteupItemData['submission_status']): string {
  return status === 'published' || status === 'submitted' ? '已发布' : '草稿'
}

function visibilityStatusLabel(status: ReviewArchiveWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

defineProps<{
  writeups: ReviewArchiveWriteupItemData[]
  manualReviews: ReviewArchiveManualReviewItemData[]
}>()
</script>

<template>
  <section class="archive-grid">
    <article class="archive-panel teacher-surface-section">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">Writeups</div>
          <h3 class="archive-panel__title">社区题解沉淀</h3>
        </div>
      </header>
      <div v-if="writeups.length === 0" class="archive-panel__empty">暂无 Writeup 记录。</div>
      <div v-else class="reflection-list">
        <article v-for="item in writeups" :key="item.id" class="reflection-item">
          <div class="reflection-item__head">
            <strong>{{ item.title }}</strong>
            <span>{{
              item.is_recommended ? '推荐题解' : visibilityStatusLabel(item.visibility_status)
            }}</span>
          </div>
          <p class="reflection-item__subhead">{{ item.challenge_title }}</p>
          <div class="reflection-item__meta">
            <span>{{ submissionStatusLabel(item.submission_status) }}</span>
            <span>{{ visibilityStatusLabel(item.visibility_status) }}</span>
            <span>{{ item.updated_at }}</span>
          </div>
        </article>
      </div>
    </article>

    <article class="archive-panel teacher-surface-section">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">Manual Reviews</div>
          <h3 class="archive-panel__title">人工审核题</h3>
        </div>
      </header>
      <div v-if="manualReviews.length === 0" class="archive-panel__empty">暂无人工审核记录。</div>
      <div v-else class="reflection-list">
        <article v-for="item in manualReviews" :key="item.id" class="reflection-item">
          <div class="reflection-item__head">
            <strong>{{ item.challenge_title }}</strong>
            <span>{{ item.review_status }}</span>
          </div>
          <p class="reflection-item__body">{{ item.answer }}</p>
          <div class="reflection-item__meta">
            <span>score {{ item.score }}</span>
            <span>{{ item.reviewer_name || '待审核' }}</span>
            <span>{{ item.submitted_at }}</span>
          </div>
        </article>
      </div>
    </article>
  </section>
</template>

<style scoped>
.archive-grid {
  display: grid;
  gap: var(--space-5);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.archive-panel {
  padding: var(--space-4-5) var(--space-4-5) var(--space-5);
}

.archive-panel__header {
  padding-bottom: var(--space-3-5);
  border-bottom: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.archive-panel__eyebrow {
  font-size: var(--font-size-0-72);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
  font-family: var(--font-family-sans);
}

.archive-panel__title {
  margin-top: var(--space-2);
  font-size: var(--font-size-1-18);
  color: var(--journal-ink);
}

.archive-panel__empty {
  padding: var(--space-4) 0;
  color: var(--color-text-secondary);
}

.reflection-list {
  display: grid;
  gap: var(--space-3-5);
  margin-top: var(--space-4);
}

.reflection-item {
  padding: var(--space-4) var(--space-4);
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 18px;
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
}

.reflection-item__head,
.reflection-item__meta {
  display: flex;
  gap: var(--space-3);
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
}

.reflection-item__head span,
.reflection-item__meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.reflection-item__subhead {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-92);
  color: var(--journal-muted);
}

.reflection-item__body {
  margin-top: var(--space-2);
  color: color-mix(in srgb, var(--journal-muted) 80%, var(--journal-ink));
  line-height: 1.72;
}

.reflection-item__meta {
  margin-top: var(--space-3);
}

@media (max-width: 1023px) {
  .archive-grid {
    grid-template-columns: 1fr;
  }
}
</style>

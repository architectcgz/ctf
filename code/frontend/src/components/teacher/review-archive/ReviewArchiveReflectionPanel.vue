<script setup lang="ts">
import type { ReviewArchiveManualReviewItemData, ReviewArchiveWriteupItemData } from '@/api/contracts'

defineProps<{
  writeups: ReviewArchiveWriteupItemData[]
  manualReviews: ReviewArchiveManualReviewItemData[]
}>()
</script>

<template>
  <section class="archive-grid">
    <article class="archive-panel">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">Writeups</div>
          <h3 class="archive-panel__title">Writeup 与评阅</h3>
        </div>
      </header>
      <div v-if="writeups.length === 0" class="archive-panel__empty">暂无 Writeup 记录。</div>
      <div v-else class="reflection-list">
        <article
          v-for="item in writeups"
          :key="item.id"
          class="reflection-item"
        >
          <div class="reflection-item__head">
            <strong>{{ item.title }}</strong>
            <span>{{ item.review_status }}</span>
          </div>
          <p class="reflection-item__subhead">{{ item.challenge_title }}</p>
          <p v-if="item.review_comment" class="reflection-item__body">{{ item.review_comment }}</p>
          <div class="reflection-item__meta">
            <span>{{ item.submission_status }}</span>
            <span>{{ item.reviewer_name || '待评阅' }}</span>
            <span>{{ item.updated_at }}</span>
          </div>
        </article>
      </div>
    </article>

    <article class="archive-panel">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">Manual Reviews</div>
          <h3 class="archive-panel__title">人工审核题</h3>
        </div>
      </header>
      <div v-if="manualReviews.length === 0" class="archive-panel__empty">暂无人工审核记录。</div>
      <div v-else class="reflection-list">
        <article
          v-for="item in manualReviews"
          :key="item.id"
          class="reflection-item"
        >
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
  gap: 1.2rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.archive-panel {
  padding: 1rem 0;
  border-top: 1px solid color-mix(in srgb, #1e40af 18%, var(--color-border-default));
}

.archive-panel__header {
  padding-bottom: 0.9rem;
  border-bottom: 1px solid color-mix(in srgb, #cbd5e1 70%, white);
}

.archive-panel__eyebrow {
  font-size: 0.72rem;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: #1d4ed8;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.archive-panel__title {
  margin-top: 0.45rem;
  font-size: 1.18rem;
  color: #0f172a;
}

.archive-panel__empty {
  padding: 1rem 0;
  color: var(--color-text-secondary);
}

.reflection-list {
  display: grid;
  gap: 0.9rem;
  margin-top: 1rem;
}

.reflection-item {
  padding: 0.95rem 1rem;
  border: 1px solid color-mix(in srgb, #1e40af 10%, var(--color-border-default));
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.9);
}

.reflection-item__head,
.reflection-item__meta {
  display: flex;
  gap: 0.7rem;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
}

.reflection-item__head span,
.reflection-item__meta {
  font-size: 0.82rem;
  color: #64748b;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.reflection-item__subhead {
  margin-top: 0.35rem;
  font-size: 0.92rem;
  color: #475569;
}

.reflection-item__body {
  margin-top: 0.55rem;
  color: #334155;
  line-height: 1.72;
}

.reflection-item__meta {
  margin-top: 0.7rem;
}

@media (max-width: 1023px) {
  .archive-grid {
    grid-template-columns: 1fr;
  }
}
</style>

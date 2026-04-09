<script setup lang="ts">
import type { ReviewArchiveEvidenceItemData, TimelineEvent } from '@/api/contracts'

defineProps<{
  timeline: TimelineEvent[]
  evidence: ReviewArchiveEvidenceItemData[]
}>()
</script>

<template>
  <section class="archive-grid">
    <article class="archive-panel teacher-surface-section">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">Timeline</div>
          <h3 class="archive-panel__title">训练记录</h3>
        </div>
      </header>
      <div v-if="timeline.length === 0" class="archive-panel__empty">暂无训练记录事件。</div>
      <ol v-else class="timeline-list">
        <li v-for="item in timeline" :key="item.id" class="timeline-item">
          <div class="timeline-item__dot" />
          <div class="timeline-item__body">
            <div class="timeline-item__head">
              <strong>{{ item.title }}</strong>
              <span>{{ item.created_at }}</span>
            </div>
            <p>{{ item.detail || item.type }}</p>
          </div>
        </li>
      </ol>
    </article>

    <article class="archive-panel teacher-surface-section">
      <header class="archive-panel__header">
        <div>
          <div class="archive-panel__eyebrow">Evidence</div>
          <h3 class="archive-panel__title">攻防证据链</h3>
        </div>
      </header>
      <div v-if="evidence.length === 0" class="archive-panel__empty">暂无证据链事件。</div>
      <div v-else class="evidence-list">
        <article
          v-for="item in evidence"
          :key="`${item.type}-${item.challenge_id}-${item.timestamp}`"
          class="evidence-item"
        >
          <div class="evidence-item__head">
            <strong>{{ item.title }}</strong>
            <span>{{ item.timestamp }}</span>
          </div>
          <p class="evidence-item__detail">{{ item.detail || item.type }}</p>
          <div class="evidence-item__meta">
            <span>challenge #{{ item.challenge_id }}</span>
            <span>{{ item.type }}</span>
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
  grid-template-columns: minmax(0, 0.94fr) minmax(0, 1.06fr);
}

.archive-panel {
  padding: var(--space-4-5) var(--space-4-5) var(--space-5);
}

.archive-panel__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
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

.timeline-list {
  margin-top: var(--space-4);
}

.timeline-item {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  gap: var(--space-3-5);
  padding-bottom: var(--space-4);
}

.timeline-item__dot {
  width: 10px;
  height: 10px;
  margin-top: var(--space-2);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  box-shadow: 0 0 0 6px color-mix(in srgb, var(--journal-accent) 8%, transparent);
}

.timeline-item__head,
.evidence-item__head,
.evidence-item__meta {
  display: flex;
  gap: var(--space-3);
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
}

.timeline-item__head span,
.evidence-item__head span,
.evidence-item__meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.timeline-item__body p,
.evidence-item__detail {
  margin-top: var(--space-2);
  line-height: 1.75;
  color: color-mix(in srgb, var(--journal-muted) 82%, var(--journal-ink));
}

.evidence-list {
  display: grid;
  gap: var(--space-3-5);
  margin-top: var(--space-4);
}

.evidence-item {
  padding: var(--space-4) var(--space-4);
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 18px;
  background: var(--journal-surface);
}

.evidence-item__meta {
  margin-top: var(--space-2-5);
}

@media (max-width: 1023px) {
  .archive-grid {
    grid-template-columns: 1fr;
  }
}
</style>

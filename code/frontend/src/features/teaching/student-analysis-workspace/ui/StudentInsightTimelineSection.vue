<template>
  <section class="student-insight-timeline-section">
    <StudentInsightLoadingSurface v-if="loading" class="insight-timeline-loading-surface">
      <div class="insight-timeline-loading-hero">
        <span class="student-insight-skeleton-line insight-timeline-loading-eyebrow" />
        <span class="student-insight-skeleton-line insight-timeline-loading-title" />
        <span class="student-insight-skeleton-line insight-timeline-loading-copy" />
        <span class="student-insight-skeleton-line insight-timeline-loading-copy insight-timeline-loading-copy--short" />
      </div>
      <div class="insight-timeline-loading-metrics">
        <span
          v-for="index in 3"
          :key="index"
          class="student-insight-skeleton-panel insight-timeline-loading-metric"
        />
      </div>
      <div class="insight-timeline-loading-list">
        <div
          v-for="index in 3"
          :key="index"
          class="insight-timeline-loading-row"
        >
          <span class="student-insight-skeleton-panel insight-timeline-loading-row-icon" />
          <span class="student-insight-skeleton-line insight-timeline-loading-row-text" />
          <span class="student-insight-skeleton-line insight-timeline-loading-row-meta" />
        </div>
      </div>
    </StudentInsightLoadingSurface>

    <TrainingTimelineContent
      v-else
      :timeline="timeline"
      class="student-insight-timeline-content"
    />
  </section>
</template>

<style scoped>
.student-insight-timeline-section {
  padding-top: var(--space-5);
}

.student-insight-timeline-content {
  min-width: 0;
}

.insight-timeline-loading-surface {
  padding: var(--space-6) var(--space-6);
  display: grid;
  gap: var(--space-5);
}

.insight-timeline-loading-hero {
  display: grid;
  gap: var(--space-2);
}

.insight-timeline-loading-eyebrow {
  width: var(--space-16);
  height: var(--space-2-5);
}

.insight-timeline-loading-title {
  width: min(18rem, 70%);
  height: var(--space-6);
}

.insight-timeline-loading-copy {
  width: min(30rem, 88%);
  height: var(--space-3);
}

.insight-timeline-loading-copy--short {
  width: min(20rem, 62%);
}

.insight-timeline-loading-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
  padding-top: var(--space-2);
}

.insight-timeline-loading-metric {
  height: var(--space-20);
}

.insight-timeline-loading-list {
  display: grid;
  gap: var(--space-2);
  padding-top: var(--space-2);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 68%, transparent);
}

.insight-timeline-loading-row {
  display: grid;
  grid-template-columns: var(--space-10) minmax(0, 1fr) var(--space-16);
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) 0;
}

.insight-timeline-loading-row-icon {
  width: var(--space-10);
  height: var(--space-10);
}

.insight-timeline-loading-row-text {
  width: min(26rem, 78%);
  height: var(--space-3);
}

.insight-timeline-loading-row-meta {
  width: var(--space-16);
  height: var(--space-3);
}

@media (max-width: 767px) {
  .insight-timeline-loading-surface {
    padding: var(--space-5);
  }

  .insight-timeline-loading-metrics {
    grid-template-columns: 1fr;
  }

  .insight-timeline-loading-row-meta {
    display: none;
  }
}
</style>

<script setup lang="ts">
import type { TimelineEvent } from '@/api/contracts'
import { TrainingTimelineContent } from '@/entities/training-timeline'
import { StudentInsightLoadingSurface } from '@/features/teaching/student-analysis-shared/ui'

defineProps<{
  timeline: TimelineEvent[]
  loading?: boolean
}>()
</script>

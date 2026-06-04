<template>
  <template v-if="isSectionVisible('overview')">
    <StudentInsightOverviewSection :profile="profile" :loading="loading" />
  </template>

  <StudentInsightRecommendationsSection
    v-if="isSectionVisible('recommendations')"
    :recommendations="recommendations"
    :loading="recommendationsLoading"
    @open-challenge="emit('openChallenge', $event)"
  />

  <template v-if="isSectionVisible('timeline')">
    <div v-if="loading" class="insight-timeline-glass">
      <div class="insight-timeline-glass__hero">
        <span class="insight-timeline-glass__eyebrow" />
        <span class="insight-timeline-glass__title" />
        <span class="insight-timeline-glass__copy" />
        <span class="insight-timeline-glass__copy insight-timeline-glass__copy--short" />
      </div>
      <div class="insight-timeline-glass__metrics">
        <span
          v-for="index in 3"
          :key="index"
          class="insight-timeline-glass__metric"
        />
      </div>
      <div class="insight-timeline-glass__list">
        <div
          v-for="index in 3"
          :key="index"
          class="insight-timeline-glass__row"
        >
          <span class="insight-timeline-glass__row-icon" />
          <span class="insight-timeline-glass__row-text" />
          <span class="insight-timeline-glass__row-meta" />
        </div>
      </div>
    </div>
    <TrainingTimelinePanel v-else :timeline="timeline" />
  </template>
</template>

<style scoped>
/* ── Timeline glass skeleton ── */

.insight-timeline-glass {
  position: relative;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  border-radius: var(--workspace-radius-lg);
  padding: var(--space-6) var(--space-6);
  background:
    radial-gradient(
      ellipse at top right,
      color-mix(in srgb, var(--journal-accent) 9%, transparent),
      transparent 46%
    ),
    radial-gradient(
      ellipse at bottom left,
      color-mix(in srgb, var(--color-bg-surface) 58%, transparent),
      transparent 52%
    ),
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base))
    );
  box-shadow: var(--workspace-shadow-panel);
  display: grid;
  gap: var(--space-5);
}

.insight-timeline-glass::before {
  position: absolute;
  inset: 1px;
  pointer-events: none;
  content: '';
  border-radius: calc(var(--workspace-radius-lg) - 1px);
  background:
    linear-gradient(
      115deg,
      transparent 0%,
      color-mix(in srgb, var(--color-bg-surface) 34%, transparent) 38%,
      transparent 72%
    );
  opacity: 0.54;
}

.insight-timeline-glass__hero {
  display: grid;
  gap: var(--space-2);
}

.insight-timeline-glass__eyebrow,
.insight-timeline-glass__title,
.insight-timeline-glass__copy,
.insight-timeline-glass__metric,
.insight-timeline-glass__row-icon,
.insight-timeline-glass__row-text,
.insight-timeline-glass__row-meta {
  display: block;
  overflow: hidden;
  border-radius: 999px;
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: insightTimelineSkeletonSweep 1.55s ease-in-out infinite;
}

.insight-timeline-glass__eyebrow {
  width: var(--space-16);
  height: var(--space-2-5);
}

.insight-timeline-glass__title {
  width: min(18rem, 70%);
  height: var(--space-6);
}

.insight-timeline-glass__copy {
  width: min(30rem, 88%);
  height: var(--space-3);
}

.insight-timeline-glass__copy--short {
  width: min(20rem, 62%);
}

.insight-timeline-glass__metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
  padding-top: var(--space-2);
}

.insight-timeline-glass__metric {
  height: var(--space-20);
  border-radius: var(--workspace-radius-lg);
}

.insight-timeline-glass__list {
  display: grid;
  gap: var(--space-2);
  padding-top: var(--space-2);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 68%, transparent);
}

.insight-timeline-glass__row {
  display: grid;
  grid-template-columns: var(--space-10) minmax(0, 1fr) var(--space-16);
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) 0;
}

.insight-timeline-glass__row-icon {
  width: var(--space-10);
  height: var(--space-10);
  border-radius: var(--workspace-radius-lg);
}

.insight-timeline-glass__row-text {
  width: min(26rem, 78%);
  height: var(--space-3);
}

.insight-timeline-glass__row-meta {
  width: var(--space-16);
  height: var(--space-3);
}

@media (max-width: 767px) {
  .insight-timeline-glass {
    padding: var(--space-5);
  }

  .insight-timeline-glass__metrics {
    grid-template-columns: 1fr;
  }

  .insight-timeline-glass__row-meta {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .insight-timeline-glass__eyebrow,
  .insight-timeline-glass__title,
  .insight-timeline-glass__copy,
  .insight-timeline-glass__metric,
  .insight-timeline-glass__row-icon,
  .insight-timeline-glass__row-text,
  .insight-timeline-glass__row-meta {
    animation: none;
  }
}

@keyframes insightTimelineSkeletonSweep {
  0% {
    background-position: 120% 0;
  }

  100% {
    background-position: -120% 0;
  }
}
</style>

<script setup lang="ts">
import type {
  RecommendationItem,
  SkillProfileData,
  TimelineEvent,
} from '@/api/contracts'
import { TrainingTimelinePanel } from '@/entities/training-timeline'
import type { StudentInsightSection } from '@/features/teaching/student-analysis-review'
import StudentInsightOverviewSection from './StudentInsightOverviewSection.vue'
import StudentInsightRecommendationsSection from './StudentInsightRecommendationsSection.vue'

const props = defineProps<{
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  recommendationsLoading?: boolean
  timeline: TimelineEvent[]
  loading?: boolean
  activeSection?: StudentInsightSection
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
}>()

function isSectionVisible(section: Exclude<StudentInsightSection, 'all' | 'writeups' | 'manual-review' | 'evidence'>): boolean {
  return !props.activeSection || props.activeSection === 'all' || props.activeSection === section
}
</script>

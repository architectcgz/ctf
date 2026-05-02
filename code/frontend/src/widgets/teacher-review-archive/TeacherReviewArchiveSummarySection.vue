<script setup lang="ts">
import { computed } from 'vue'

import type { ReviewArchiveData } from '@/api/contracts'
import SectionCard from '@/components/common/SectionCard.vue'
import { formatDate } from '@/utils/format'
import { REVIEW_ARCHIVE_SUMMARY_COPY } from './model/presentation'

const props = defineProps<{
  archive: ReviewArchiveData
}>()

const solvedRate = computed(() => {
  if (!props.archive.summary.total_challenges) return 0
  return Math.round(
    (props.archive.summary.total_solved / props.archive.summary.total_challenges) * 100
  )
})

const formattedLastActivity = computed(() => {
  if (!props.archive.summary.last_activity_at) return '--'
  return formatDate(props.archive.summary.last_activity_at)
})

const rankedSkillDimensions = computed(() =>
  [...props.archive.skill_profile.dimensions].sort((left, right) => right.value - left.value)
)
</script>

<template>
  <section class="review-archive-summary-grid">
    <SectionCard
      :title="REVIEW_ARCHIVE_SUMMARY_COPY.summaryTitle"
      :subtitle="REVIEW_ARCHIVE_SUMMARY_COPY.summarySubtitle"
    >
      <div class="summary-grid metric-panel-grid metric-panel-default-surface">
        <article class="summary-card summary-card--primary metric-panel-card">
          <div class="summary-card__label metric-panel-label">
            {{ REVIEW_ARCHIVE_SUMMARY_COPY.solvedRateLabel }}
          </div>
          <div class="summary-card__value metric-panel-value">
            {{ solvedRate }}%
          </div>
          <div class="summary-card__hint metric-panel-helper">
            {{ REVIEW_ARCHIVE_SUMMARY_COPY.solvedRateHintPrefix }}
            {{ archive.summary.total_solved }} / {{ archive.summary.total_challenges }}
          </div>
        </article>
        <article class="summary-card summary-card--warning metric-panel-card">
          <div class="summary-card__label metric-panel-label">
            {{ REVIEW_ARCHIVE_SUMMARY_COPY.correctSubmissionLabel }}
          </div>
          <div class="summary-card__value metric-panel-value">
            {{ archive.summary.correct_submission_count }}
          </div>
          <div class="summary-card__hint metric-panel-helper">
            {{ REVIEW_ARCHIVE_SUMMARY_COPY.correctSubmissionHint }}
          </div>
        </article>
        <article class="summary-card summary-card--neutral metric-panel-card">
          <div class="summary-card__label metric-panel-label">
            {{ REVIEW_ARCHIVE_SUMMARY_COPY.latestActivityLabel }}
          </div>
          <div class="summary-card__value summary-card__value--time metric-panel-value">
            {{ formattedLastActivity }}
          </div>
          <div class="summary-card__hint metric-panel-helper">
            {{ REVIEW_ARCHIVE_SUMMARY_COPY.latestActivityHint }}
          </div>
        </article>
      </div>
    </SectionCard>

    <SectionCard
      :title="REVIEW_ARCHIVE_SUMMARY_COPY.skillTitle"
      :subtitle="REVIEW_ARCHIVE_SUMMARY_COPY.skillSubtitle"
    >
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
            <div
              class="skill-bars__fill"
              :style="{ width: `${item.value}%` }"
            />
          </div>
        </article>
      </div>
    </SectionCard>
  </section>
</template>

<style scoped>
.review-archive-summary-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 0.92fr) minmax(0, 1.08fr);
}

.summary-grid {
  --metric-panel-grid-gap: var(--space-3-5);
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.summary-card {
  --metric-panel-border: var(--teacher-card-border);
  --metric-panel-radius: 18px;
  --metric-panel-padding: var(--space-4);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.summary-card--primary {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base))
  );
}

.summary-card--warning {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-warning) 14%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base))
  );
}

.summary-card__label {
  font-family: var(--font-family-mono);
}

.summary-card__value {
  --metric-panel-value-margin-top: var(--space-3);
  --metric-panel-value-size: 1.8rem;
}

.summary-card__value--time {
  font-size: var(--font-size-1-06);
  line-height: 1.5;
}

.summary-card__hint {
  --metric-panel-helper-margin-top: var(--space-2);
  --metric-panel-helper-line-height: 1.65;
}

.skill-bars {
  display: grid;
  gap: var(--space-3-5);
}

.skill-bars__item {
  padding: var(--space-1) 0;
}

.skill-bars__head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  align-items: center;
  margin-bottom: var(--space-2);
  color: var(--journal-ink);
}

.skill-bars__head span {
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.skill-bars__track {
  height: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(
    in srgb,
    var(--journal-border, var(--color-border-default)) 34%,
    transparent
  );
}

.skill-bars__fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink)),
    color-mix(in srgb, var(--journal-accent) 48%, white) 58%,
    color-mix(in srgb, var(--color-warning) 84%, var(--journal-accent))
  );
}

@media (max-width: 1023px) {
  .review-archive-summary-grid,
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>

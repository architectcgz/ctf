<script setup lang="ts">
import type { ReviewArchiveData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import ReviewArchiveEvidencePanel from '@/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue'
import ReviewArchiveHero from '@/components/teacher/review-archive/ReviewArchiveHero.vue'
import ReviewArchiveObservationStrip from '@/components/teacher/review-archive/ReviewArchiveObservationStrip.vue'
import ReviewArchiveReflectionPanel from '@/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue'
import TeacherReviewArchiveSummarySection from './TeacherReviewArchiveSummarySection.vue'

const props = defineProps<{
  archive: ReviewArchiveData | null
  loading: boolean
  error: string | null
  exporting: boolean
}>()

const emit = defineEmits<{
  reload: []
  back: []
  openAnalysis: []
  exportArchive: []
}>()
</script>

<template>
  <div class="review-archive-shell teacher-surface space-y-8">
    <ReviewArchiveHero
      :archive="archive"
      :exporting="exporting"
      @back="emit('back')"
      @open-analysis="emit('openAnalysis')"
      @export-archive="emit('exportArchive')"
    />

    <div
      v-if="loading"
      class="review-archive-loading"
    >
      <div class="review-archive-loading__hero" />
      <div class="review-archive-loading__grid">
        <div class="review-archive-loading__block" />
        <div class="review-archive-loading__block" />
      </div>
    </div>

    <AppEmpty
      v-else-if="error"
      title="复盘归档加载失败"
      :description="error"
      icon="AlertTriangle"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          @click="emit('reload')"
        >
          重新加载
        </button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="!archive"
      title="暂无复盘归档"
      description="当前学生还没有可展示的复盘归档数据。"
      icon="FileChartColumnIncreasing"
    />

    <template v-else>
      <ReviewArchiveObservationStrip :items="archive.teacher_observations.items" />

      <TeacherReviewArchiveSummarySection :archive="archive" />

      <ReviewArchiveEvidencePanel
        :timeline="archive.timeline"
        :evidence="archive.evidence"
        :writeups="archive.writeups"
        :manual-reviews="archive.manual_reviews"
      />

      <ReviewArchiveReflectionPanel
        :writeups="archive.writeups"
        :manual-reviews="archive.manual_reviews"
      />
    </template>
  </div>
</template>

<style scoped>
.review-archive-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  min-height: 100%;
  padding: var(--space-1) 0 var(--space-8);
}

:deep(.section-card) {
  border: 1px solid var(--teacher-card-border);
  background: var(--journal-surface-subtle);
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

:deep(.section-card__header) {
  border-bottom: 1px dashed var(--teacher-divider);
}

.review-archive-loading__hero,
.review-archive-loading__block {
  border-radius: 26px;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 80%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  animation: review-archive-pulse 1.35s ease-in-out infinite;
}

.review-archive-loading__hero {
  height: 220px;
}

.review-archive-loading__grid {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.review-archive-loading__block {
  height: 180px;
}

@keyframes review-archive-pulse {
  0%,
  100% {
    opacity: 0.58;
  }
  50% {
    opacity: 1;
  }
}

@media (max-width: 1023px) {
  .review-archive-loading__grid {
    grid-template-columns: 1fr;
  }
}
</style>

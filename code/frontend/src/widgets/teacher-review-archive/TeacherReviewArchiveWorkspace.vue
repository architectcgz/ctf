<script setup lang="ts">
import type { ReviewArchiveData } from '@/api/contracts'
import ReviewArchiveEvidencePanel from '@/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue'
import ReviewArchiveHero from '@/components/teacher/review-archive/ReviewArchiveHero.vue'
import ReviewArchiveObservationStrip from '@/components/teacher/review-archive/ReviewArchiveObservationStrip.vue'
import ReviewArchiveReflectionPanel from '@/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue'
import TeacherReviewArchiveState from './TeacherReviewArchiveState.vue'
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

    <TeacherReviewArchiveState
      :loading="loading"
      :error="error"
      :has-archive="Boolean(archive)"
      @reload="emit('reload')"
    >
      <template v-if="archive">
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
    </TeacherReviewArchiveState>
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
</style>

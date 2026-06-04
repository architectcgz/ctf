<template>
  <div class="review-archive-shell teacher-surface space-y-8">
    <ReviewArchiveHero
      :archive="archive"
      :exporting="exporting"
      :analysis-route="analysisRoute"
      :back-route="backRoute"
      @export-archive="emit('exportArchive')"
    />

    <ReviewArchiveState
      :loading="loading"
      :error="error"
      :has-archive="Boolean(archive)"
      @reload="emit('reload')"
    >
      <template v-if="archive">
        <ReviewArchiveObservationStrip :items="archive.teacher_observations.items" />

        <ReviewArchiveSummarySection :archive="archive" />

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
    </ReviewArchiveState>
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
</style>

<script setup lang="ts">
import type { ReviewArchiveData } from '@/api/contracts'
import ReviewArchiveEvidencePanel from './ReviewArchiveEvidencePanel.vue'
import ReviewArchiveHero from './ReviewArchiveHero.vue'
import ReviewArchiveObservationStrip from './ReviewArchiveObservationStrip.vue'
import ReviewArchiveReflectionPanel from './ReviewArchiveReflectionPanel.vue'
import ReviewArchiveState from './ReviewArchiveState.vue'
import ReviewArchiveSummarySection from './ReviewArchiveSummarySection.vue'

defineProps<{
  archive: ReviewArchiveData | null
  loading: boolean
  error: string | null
  exporting: boolean
  analysisRoute: {
    name: string
    params?: Record<string, string>
  }
  backRoute: {
    name: string
    params?: Record<string, string>
  }
}>()

const emit = defineEmits<{
  reload: []
  exportArchive: []
}>()
</script>

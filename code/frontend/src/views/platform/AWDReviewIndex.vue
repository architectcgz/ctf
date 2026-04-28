<script setup lang="ts">
import { computed } from 'vue'

import AwdReviewHeroPanel from '@/components/platform/awd-review/AwdReviewHeroPanel.vue'
import AwdReviewDirectoryPanel from '@/components/platform/awd-review/AwdReviewDirectoryPanel.vue'
import { useTeacherAwdReviewIndex } from '@/composables/useTeacherAwdReviewIndex'

interface PlatformAwdReviewRow {
  id: string
  title: string
  status: string
  current_round?: number
  round_count: number
  team_count: number
  mode: string
  export_ready: boolean
  latest_evidence_at?: string | null
  contestCode: string
}

const { router, loading, error, contests, filters, hasContests, loadContests, openContest } =
  useTeacherAwdReviewIndex()

const hasActiveFilters = computed(() => Boolean(filters.value.status || filters.value.keyword.trim()))
const runningCount = computed(() => contests.value.filter((item) => item.status === 'running').length)
const exportReadyCount = computed(() => contests.value.filter((item) => item.export_ready).length)
const reviewRows = computed<PlatformAwdReviewRow[]>(() =>
  contests.value.map((contest) => ({
    ...contest,
    contestCode: `AWD-${contest.id}`,
  }))
)

function resetFilters(): void {
  filters.value.status = ''
  filters.value.keyword = ''
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero admin-awd-review-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane admin-awd-review-shell__content">
      <AwdReviewHeroPanel
        :contest-count="contests.length"
        :running-count="runningCount"
        :export-ready-count="exportReadyCount"
        @back="router.push({ name: 'PlatformOverview' })"
        @refresh="loadContests"
      />

      <AwdReviewDirectoryPanel
        :loading="loading"
        :error="error"
        :rows="reviewRows"
        :total="contests.length"
        :has-contests="hasContests"
        :keyword="filters.keyword"
        :status-filter="filters.status"
        :has-active-filters="hasActiveFilters"
        @update:keyword="filters.keyword = $event"
        @update:status-filter="filters.status = $event"
        @reset-filters="resetFilters"
        @retry="loadContests"
        @open-contest="openContest"
      />
    </main>
  </section>
</template>

<style scoped>
.admin-awd-review-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  --awd-review-directory-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --awd-review-directory-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 99%, var(--color-bg-base))
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 10%, transparent),
      transparent 20rem
    );
}

.admin-awd-review-shell__content {
  display: grid;
  gap: var(--workspace-directory-page-block-gap);
}
</style>

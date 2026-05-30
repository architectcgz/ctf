<script setup lang="ts">
import { useAwdReviewIndexPage } from '@/features/awd-review-workspace'
import {
  AwdReviewDirectoryPanel,
  AwdReviewHeroPanel,
} from '@/widgets/awd-review-workspace'

const {
  loading,
  error,
  total,
  page,
  totalPages,
  filters,
  hasContests,
  hasActiveFilters,
  reviewRows,
  contestSummary,
  loadContests,
  changePage,
  resetFilters,
  homeRoute,
  buildContestRoute,
} = useAwdReviewIndexPage('platform')
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero admin-awd-review-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane admin-awd-review-shell__content">
      <AwdReviewHeroPanel
        :contest-count="contestSummary.totalCount"
        :running-count="contestSummary.runningCount"
        :export-ready-count="contestSummary.exportReadyCount"
        :overview-route="homeRoute"
        @refresh="loadContests"
      />

      <AwdReviewDirectoryPanel
        :loading="loading"
        :error="error"
        :rows="reviewRows"
        :total="total"
        :page="page"
        :total-pages="totalPages"
        :has-contests="hasContests"
        :build-contest-route="buildContestRoute"
        :keyword="filters.keyword"
        :status-filter="filters.status"
        :has-active-filters="hasActiveFilters"
        @update:keyword="filters.keyword = $event"
        @update:status-filter="filters.status = $event"
        @reset-filters="resetFilters"
        @retry="loadContests"
        @change-page="changePage"
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

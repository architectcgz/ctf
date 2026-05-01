<script setup lang="ts">
import ChallengeManageDirectoryPanel from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue'
import ChallengeManageHeroPanel from '@/components/platform/challenge/ChallengeManageHeroPanel.vue'
import { useChallengeManagePage } from '@/features/platform-challenges'

const {
  archivedCount,
  categoryFilter,
  changePage,
  clearFilters,
  difficultyFilter,
  draftCount,
  hasActiveFilters,
  hasLoadError,
  keyword,
  loadErrorMessage,
  loading,
  manageEmptyMessage,
  manageEmptyTitle,
  openActionMenuId,
  openChallengeDetail,
  openChallengeTopology,
  openChallengeWriteup,
  openImportWorkspace,
  page,
  publishedCount,
  refresh,
  removeChallenge,
  selectedSortLabel,
  setActionMenuOpen,
  setSort,
  sortOptions,
  sortedChallenges,
  statusFilter,
  submitPublishCheck,
  total,
  totalPages,
} = useChallengeManagePage()
</script>

<template>
  <section class="workspace-shell challenge-manage-shell journal-shell journal-shell-admin journal-notes-card journal-hero">
    <div class="workspace-grid">
      <main class="content-pane challenge-manage-content">
        <div class="challenge-manage-panel">
          <ChallengeManageHeroPanel
            :total="total"
            :published-count="publishedCount"
            :draft-count="draftCount"
            :archived-count="archivedCount"
            @import="void openImportWorkspace()"
          />

          <ChallengeManageDirectoryPanel
            :rows="sortedChallenges"
            :total="total"
            :page="page"
            :total-pages="totalPages"
            :loading="loading"
            :has-load-error="hasLoadError"
            :load-error-message="loadErrorMessage"
            :has-active-filters="hasActiveFilters"
            :manage-empty-title="manageEmptyTitle"
            :manage-empty-message="manageEmptyMessage"
            :keyword="keyword"
            :category-filter="categoryFilter"
            :difficulty-filter="difficultyFilter"
            :status-filter="statusFilter"
            :selected-sort-label="selectedSortLabel"
            :sort-options="sortOptions"
            :open-action-menu-id="openActionMenuId"
            @update:keyword="keyword = $event"
            @update:category-filter="categoryFilter = $event"
            @update:difficulty-filter="difficultyFilter = $event"
            @update:status-filter="statusFilter = $event"
            @select-sort="setSort"
            @reset-filters="clearFilters"
            @retry="void refresh()"
            @change-page="changePage"
            @update-action-menu-open="setActionMenuOpen($event.challengeId, $event.open)"
            @open-detail="openChallengeDetail"
            @open-topology="openChallengeTopology"
            @open-writeup="openChallengeWriteup"
            @submit-publish-check="submitPublishCheck"
            @remove-challenge="removeChallenge"
          />
        </div>
      </main>
    </div>
  </section>
</template>

<style scoped>
.challenge-manage-shell {
  --challenge-page-bg: var(--journal-surface);
  --workspace-shell-bg: var(--challenge-page-bg);
  --workspace-shell-elevated-bg: var(--journal-surface-subtle);
  background: var(--workspace-shell-bg);
}

.challenge-manage-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}

.challenge-manage-panel {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <ScoreboardWorkspaceShell
      :panel-tabs="panelTabs"
      :active-tab="activeTab"
      :set-tab-button-ref="setTabButtonRef"
      :handle-tab-keydown="handleTabKeydown"
      :contest-count="contestCount"
      :running-count="runningCount"
      :frozen-count="frozenCount"
      :ended-count="endedCount"
      :contest-page-start-index="contestPageStartIndex"
      :paginated-sections="paginatedSections"
      :empty-title="emptyTitle"
      :points-empty-title="pointsEmptyTitle"
      :format-contest-window="formatContestWindow"
      :section-accent-style="sectionAccentStyle"
      :get-row-class="getRowClass"
      :get-rank-pill-class="getRankPillClass"
      :contest-page="contestPage"
      :contest-total="contestTotal"
      :contest-total-pages="contestTotalPages"
      :contest-status-filter="contestStatusFilter"
      :contest-mode-filter="contestModeFilter"
      :has-sections="hasSections"
      :has-ranking-rows="hasRankingRows"
      :loading="loading"
      :ranking-hint="rankingHint"
      :ranking-loading="rankingLoading"
      :ranking-rows="rankingRows"
      :selection-hint="selectionHint"
      @select-tab="selectTab"
      @update-status-filter="updateContestStatusFilter"
      @update-mode-filter="updateContestModeFilter"
      @reset-filters="resetContestFilters"
      @refresh="refresh"
      @refresh-practice-ranking="refreshPracticeRanking"
      @change-contest-page="changeContestPage"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
}
</style>

<script setup lang="ts">
import {
  ScoreboardWorkspaceShell,
  useScoreboardContestDirectoryPage,
  useScoreboardRoutePage,
  useScoreboardView,
} from '@/features/scoreboard'

const { panelTabs, activeTab, setTabButtonRef, selectTab, handleTabKeydown } =
  useScoreboardRoutePage()

const {
  contestPage,
  contestPageSize,
  contestSummary,
  contestTotal,
  contestTotalPages,
  changeContestPage,
  contestStatusFilter,
  contestModeFilter,
  updateContestStatusFilter,
  updateContestModeFilter,
  resetContestFilters,
  hasSections,
  hasRankingRows,
  loading,
  rankingError,
  rankingHint,
  rankingLoading,
  rankingRows,
  refresh,
  refreshPracticeRanking,
  sections,
  selectionHint,
} = useScoreboardView()

const {
  contestCount,
  runningCount,
  frozenCount,
  endedCount,
  contestPageStartIndex,
  paginatedSections,
  emptyTitle,
  pointsEmptyTitle,
  formatContestWindow,
  sectionAccentStyle,
  getRowClass,
  getRankPillClass,
} = useScoreboardContestDirectoryPage({
  sections,
  contestSummary,
  contestPage,
  contestPageSize,
  contestTotal,
  selectionHint,
  rankingError,
})
</script>

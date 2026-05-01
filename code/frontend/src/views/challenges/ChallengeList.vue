<script setup lang="ts">
import { LayoutDashboard, Target } from 'lucide-vue-next'

import ChallengeDirectoryPanel from '@/components/challenge/ChallengeDirectoryPanel.vue'
import { useChallengeListPage } from '@/features/challenge-list'

const {
  list,
  total,
  page,
  loading,
  searchQuery,
  categoryFilter,
  difficultyFilter,
  hasActiveFilters,
  hasLoadError,
  errorMessage,
  emptyTitle,
  emptyDescription,
  totalPages,
  summaryStats,
  changePage,
  refresh,
  onSearch,
  onFilterChange,
  resetFilters,
  goToDashboard,
  openSkillProfile,
  goToDetail,
} = useChallengeListPage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="challenge-page">
        <header class="challenge-topbar">
          <div class="challenge-heading">
            <div class="workspace-overline">
              Challenges
            </div>
            <h1 class="workspace-page-title challenge-title">
              靶场训练
            </h1>
          </div>

          <div class="challenge-actions">
            <button
              type="button"
              class="ui-btn ui-btn--primary"
              @click="goToDashboard"
            >
              <LayoutDashboard class="h-4 w-4" />
              返回仪表盘
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--ghost"
              @click="openSkillProfile"
            >
              能力画像
            </button>
          </div>
        </header>

        <section class="challenge-summary metric-panel-default-surface">
          <div class="challenge-summary-title">
            <Target class="h-4 w-4" />
            <span>当前题库概况</span>
          </div>
          <div class="challenge-summary-grid metric-panel-grid">
            <div
              v-for="stat in summaryStats"
              :key="stat.key"
              class="challenge-summary-item metric-panel-card"
            >
              <div class="challenge-summary-label metric-panel-label">
                {{ stat.label }}
              </div>
              <div class="challenge-summary-value metric-panel-value">
                {{ stat.value }}
              </div>
              <div class="challenge-summary-helper metric-panel-helper">
                {{ stat.helper }}
              </div>
            </div>
          </div>
        </section>

        <ChallengeDirectoryPanel
          :list="list"
          :total="total"
          :page="page"
          :total-pages="totalPages"
          :search-query="searchQuery"
          :category-filter="categoryFilter"
          :difficulty-filter="difficultyFilter"
          :loading="loading"
          :has-active-filters="hasActiveFilters"
          :has-load-error="hasLoadError"
          :error-message="errorMessage"
          :empty-title="emptyTitle"
          :empty-description="emptyDescription"
          @update:search-query="searchQuery = $event"
          @update:category-filter="categoryFilter = $event"
          @update:difficulty-filter="difficultyFilter = $event"
          @search="onSearch"
          @filter-change="onFilterChange"
          @reset-filters="resetFilters"
          @refresh="refresh"
          @open-detail="goToDetail"
          @change-page="changePage"
        />
      </div>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 78%,
    var(--color-bg-base)
  );
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --journal-shell-accent-strong: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --challenge-tone-web: color-mix(in srgb, var(--color-cat-web) 82%, var(--journal-ink));
  --challenge-tone-pwn: color-mix(in srgb, var(--color-cat-pwn) 72%, var(--journal-ink));
  --challenge-tone-reverse: color-mix(in srgb, var(--color-cat-reverse) 74%, var(--journal-ink));
  --challenge-tone-crypto: color-mix(in srgb, var(--color-cat-crypto) 76%, var(--journal-ink));
  --challenge-tone-misc: color-mix(in srgb, var(--color-cat-misc) 78%, var(--journal-ink));
  --challenge-tone-forensics: color-mix(
    in srgb,
    var(--color-cat-forensics) 78%,
    var(--journal-ink)
  );
  --challenge-diff-beginner: color-mix(in srgb, var(--color-diff-beginner) 76%, var(--journal-ink));
  --challenge-diff-easy: color-mix(in srgb, var(--color-diff-easy) 78%, var(--journal-ink));
  --challenge-diff-medium: color-mix(in srgb, var(--color-diff-medium) 80%, var(--journal-ink));
  --challenge-diff-hard: color-mix(in srgb, var(--color-diff-hard) 80%, var(--journal-ink));
  --challenge-diff-insane: color-mix(in srgb, var(--color-diff-insane) 84%, var(--journal-ink));
}

.challenge-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.challenge-heading {
  min-width: 0;
}

.challenge-title {
  color: var(--journal-ink);
}

.challenge-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

@media (max-width: 960px) {
  .challenge-topbar {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>

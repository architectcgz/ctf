<script setup lang="ts">
import { LayoutDashboard, Target } from 'lucide-vue-next'

import ChallengeDirectoryPanel from '@/components/challenge/ChallengeDirectoryPanel.vue'
import { useChallengeListPage } from '@/features/challenge-list'

type ChallengeSummaryKey = 'total' | 'visible' | 'solved' | 'unsolved'

const challengeSummaryVisuals: Record<
  ChallengeSummaryKey,
  { badge: string; eyebrow: string; accent: string }
> = {
  total: {
    badge: '01',
    eyebrow: 'Library',
    accent: 'var(--challenge-tone-web)',
  },
  visible: {
    badge: '02',
    eyebrow: 'Visible',
    accent: 'var(--challenge-tone-crypto)',
  },
  solved: {
    badge: '03',
    eyebrow: 'Solved',
    accent: 'var(--color-success)',
  },
  unsolved: {
    badge: '04',
    eyebrow: 'Pending',
    accent: 'var(--color-warning)',
  },
}

function getSummaryVisual(key: string) {
  return (
    challengeSummaryVisuals[(key as ChallengeSummaryKey) || 'visible'] ??
    challengeSummaryVisuals.visible
  )
}

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
            <div class="workspace-overline">Challenges</div>
            <h1 class="workspace-page-title challenge-title">靶场训练</h1>
            <p class="challenge-subtitle">
              统一查看训练题目，按分类、难度和关键词收束范围后直接进入做题。
            </p>
          </div>

          <div class="challenge-hero-rail">
            <div class="challenge-actions">
              <button type="button" class="ui-btn ui-btn--primary" @click="goToDashboard">
                <LayoutDashboard class="h-4 w-4" />
                返回仪表盘
              </button>
              <button type="button" class="ui-btn ui-btn--ghost" @click="openSkillProfile">
                能力画像
              </button>
            </div>
          </div>
        </header>

        <section class="challenge-summary metric-panel-default-surface">
          <div class="challenge-summary-title">
            <span class="challenge-summary-title-mark" />
            <Target class="h-4 w-4" />
            <span>题库概况</span>
          </div>
          <div class="challenge-summary-grid metric-panel-grid">
            <div
              v-for="stat in summaryStats"
              :key="stat.key"
              class="challenge-summary-item metric-panel-card"
              :style="{ '--challenge-summary-accent': getSummaryVisual(stat.key).accent }"
            >
              <div class="challenge-summary-ribbon">
                <span class="challenge-summary-badge">{{ getSummaryVisual(stat.key).badge }}</span>
                <span class="challenge-summary-eyebrow">{{
                  getSummaryVisual(stat.key).eyebrow
                }}</span>
              </div>

              <div class="challenge-summary-main">
                <div class="challenge-summary-label metric-panel-label">
                  {{ stat.label }}
                </div>
                <div class="challenge-summary-value metric-panel-value">
                  {{ stat.value }}
                </div>
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

.challenge-page .challenge-topbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-5);
  align-items: end;
}

.challenge-page .challenge-heading {
  min-width: 0;
  display: grid;
  gap: var(--space-3);
  max-width: min(44rem, 100%);
}

.challenge-title {
  color: var(--journal-ink);
}

.challenge-page .challenge-subtitle {
  max-width: 42rem;
}

.challenge-hero-rail {
  display: grid;
  justify-items: end;
  gap: var(--space-3);
}

.challenge-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

.challenge-page .challenge-summary {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
  --metric-panel-grid-gap: var(--space-3);
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-6);
  padding: var(--space-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  border-radius: var(--radius-2xl);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 9%, transparent),
      transparent 36%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 76%, var(--color-bg-base))
    );
  box-shadow: 0 18px 34px color-mix(in srgb, var(--color-shadow-soft) 34%, transparent);
}

.challenge-page .challenge-summary .challenge-summary-title {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2-5);
  color: var(--journal-ink);
}

.challenge-summary-title-mark {
  width: var(--space-1-5);
  height: var(--space-1-5);
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 70%, transparent);
  box-shadow: 0 0 0 var(--space-1) color-mix(in srgb, var(--color-primary) 14%, transparent);
}

.challenge-page .challenge-summary .challenge-summary-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.challenge-page .challenge-summary-item {
  display: grid;
  gap: var(--space-3);
  min-width: 0;
  padding: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--challenge-summary-accent) 20%, var(--journal-border));
  border-radius: var(--radius-xl);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--challenge-summary-accent) 10%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface) 96%, transparent)
  );
  box-shadow: 0 14px 26px color-mix(in srgb, var(--color-shadow-soft) 18%, transparent);
}

.challenge-summary-ribbon {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.challenge-summary-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: calc(var(--space-6) + var(--space-1));
  min-height: calc(var(--space-6) + var(--space-1));
  padding: 0 var(--space-2);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--challenge-summary-accent) 16%, transparent);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--challenge-summary-accent);
}

.challenge-summary-eyebrow {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--challenge-summary-accent) 70%, var(--journal-muted));
}

.challenge-summary-main {
  display: grid;
  gap: var(--space-1-5);
}

.challenge-page .challenge-summary-item .challenge-summary-label {
  color: var(--journal-muted);
}

.challenge-page .challenge-summary-item .challenge-summary-value {
  color: var(--journal-ink);
}

.challenge-page .challenge-summary-item .challenge-summary-helper {
  max-width: 20rem;
}

@media (max-width: 960px) {
  .challenge-page .challenge-topbar {
    grid-template-columns: minmax(0, 1fr);
    align-items: flex-start;
  }

  .challenge-hero-rail {
    width: 100%;
    justify-items: stretch;
  }

  .challenge-actions {
    justify-content: flex-start;
  }

  .challenge-page .challenge-summary .challenge-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .challenge-page .challenge-summary .challenge-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>

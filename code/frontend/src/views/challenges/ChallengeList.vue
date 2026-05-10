<script setup lang="ts">
import type { Component } from 'vue'

import { BookOpen, Layers3, LayoutDashboard, ShieldCheck, Target } from 'lucide-vue-next'

import ChallengeDirectoryPanel from '@/components/challenge/ChallengeDirectoryPanel.vue'
import { useChallengeListPage } from '@/features/challenge-list'

type ChallengeSummaryKey = 'total' | 'visible' | 'solved' | 'unsolved'

const challengeSummaryVisuals: Record<ChallengeSummaryKey, { icon: Component; accent: string }> = {
  total: {
    icon: BookOpen,
    accent: 'var(--challenge-tone-web)',
  },
  visible: {
    icon: Layers3,
    accent: 'var(--challenge-tone-crypto)',
  },
  solved: {
    icon: ShieldCheck,
    accent: 'var(--color-success)',
  },
  unsolved: {
    icon: Target,
    accent: 'var(--color-warning)',
  },
}

const challengeSummaryWaves: Record<ChallengeSummaryKey, string> = {
  total:
    'M10 80 C35 70, 45 88, 65 72 C88 52, 95 20, 115 38 C140 62, 145 14, 165 22 C185 30, 190 8, 210 18',
  visible: 'M10 78 C35 60, 55 50, 80 54 C105 58, 110 28, 135 34 C165 43, 172 15, 210 22',
  solved: 'M8 76 C32 56, 52 74, 70 55 C92 30, 112 24, 132 52 C150 79, 172 14, 210 20',
  unsolved:
    'M10 80 C35 72, 48 82, 68 68 C92 50, 95 18, 118 26 C142 34, 148 70, 170 58 C192 46, 190 10, 212 18',
}

function getSummaryVisual(key: string) {
  return (
    challengeSummaryVisuals[(key as ChallengeSummaryKey) || 'visible'] ??
    challengeSummaryVisuals.visible
  )
}

function getSummaryWave(key: string) {
  return (
    challengeSummaryWaves[(key as ChallengeSummaryKey) || 'visible'] ??
    challengeSummaryWaves.visible
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
        <header class="workspace-page-header challenge-topbar">
          <div class="challenge-heading">
            <div class="workspace-overline">Challenges</div>
            <h1 class="workspace-page-title challenge-title">靶场训练</h1>
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
              <div class="challenge-summary-icon-shell">
                <component :is="getSummaryVisual(stat.key).icon" class="h-5 w-5" />
              </div>

              <div class="challenge-summary-content">
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

              <div class="challenge-summary-wave" aria-hidden="true">
                <svg
                  viewBox="0 0 220 90"
                  fill="none"
                  preserveAspectRatio="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    :d="getSummaryWave(stat.key)"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                    stroke-linecap="round"
                  />
                </svg>
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
  --workspace-page-header-gap: var(--space-5);
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.challenge-page .challenge-heading {
  min-width: 0;
  max-width: min(44rem, 100%);
}

.challenge-title {
  color: var(--journal-ink);
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
  --metric-panel-grid-gap: var(--space-4);
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
  color: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
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
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--space-4);
  min-width: 0;
  overflow: hidden;
  padding: var(--space-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  border-radius: var(--radius-xl);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 72%, var(--color-bg-base))
    ),
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--challenge-summary-accent) 5%, transparent),
      transparent 62%
    );
  box-shadow: 0 12px 24px color-mix(in srgb, var(--color-shadow-soft) 14%, transparent);
}

.challenge-page .challenge-summary-item::after {
  content: '';
  position: absolute;
  inset: 0;
  opacity: 0.92;
  pointer-events: none;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--challenge-summary-accent) 7%, transparent),
    color-mix(in srgb, var(--challenge-summary-accent) 2%, transparent)
  );
}

.challenge-summary-icon-shell {
  position: relative;
  z-index: 1;
  display: grid;
  flex-shrink: 0;
  width: calc(var(--space-12) + var(--space-4));
  height: calc(var(--space-12) + var(--space-4));
  place-items: center;
  border-radius: 999px;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--challenge-summary-accent) 18%, var(--journal-surface)),
    color-mix(in srgb, var(--challenge-summary-accent) 6%, var(--journal-surface))
  );
  color: var(--challenge-summary-accent);
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--journal-border) 32%, transparent),
    0 8px 18px color-mix(in srgb, var(--challenge-summary-accent) 8%, transparent);
}

.challenge-summary-content {
  position: relative;
  z-index: 1;
  display: grid;
  gap: var(--space-1-5);
  min-width: 0;
}

.challenge-page .challenge-summary-item .challenge-summary-label {
  color: var(--journal-muted);
  margin: 0;
}

.challenge-page .challenge-summary-item .challenge-summary-value {
  color: var(--journal-ink);
  margin: 0;
}

.challenge-page .challenge-summary-item .challenge-summary-helper {
  max-width: 16rem;
  margin: 0;
}

.challenge-summary-wave {
  position: absolute;
  right: -0.25rem;
  bottom: -0.75rem;
  width: 8.875rem;
  height: 4.375rem;
  color: var(--challenge-summary-accent);
  opacity: 0.34;
  pointer-events: none;
}

.challenge-summary-wave svg {
  display: block;
  width: 100%;
  height: 100%;
}

@media (max-width: 960px) {
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

  .challenge-page .challenge-summary-item {
    padding: var(--space-4);
  }
}

@media (max-width: 720px) {
  .challenge-page .challenge-summary .challenge-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .challenge-page .challenge-summary-item {
    align-items: flex-start;
  }
}
</style>

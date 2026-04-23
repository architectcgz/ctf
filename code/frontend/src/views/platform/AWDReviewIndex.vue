<script setup lang="ts">
import { computed } from 'vue'
import { FolderKanban, RefreshCcw, ScanEye, Waypoints } from 'lucide-vue-next'

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
      <header class="admin-awd-review-shell__hero">
        <div class="admin-awd-review-shell__hero-main">
          <div class="workspace-overline">
            Review Workspace
          </div>
          <h1 class="workspace-page-title">
            AWD复盘
          </h1>
          <p class="workspace-page-copy">
            在平台视角统一查看可进入的 AWD 赛事、当前状态和报告就绪度，并直接进入复盘详情。
          </p>
        </div>

        <div class="admin-awd-review-shell__hero-actions">
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="router.push({ name: 'PlatformOverview' })"
          >
            返回平台概览
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--primary"
            @click="loadContests"
          >
            <RefreshCcw class="h-4 w-4" />
            刷新目录
          </button>
        </div>
      </header>

      <div
        class="admin-summary-grid admin-awd-review-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
      >
        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>赛事数量</span>
            <FolderKanban class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ contests.length.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前可进入复盘的 AWD 赛事
          </div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>进行中</span>
            <ScanEye class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ runningCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            仍在持续产出攻防信号的赛事
          </div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>可导出报告</span>
            <Waypoints class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ exportReadyCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            已允许导出教师复盘报告的赛事
          </div>
        </article>
      </div>

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

.admin-awd-review-shell__hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.admin-awd-review-shell__hero-main {
  max-width: 48rem;
}

.admin-awd-review-shell__hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.admin-awd-review-shell__hero-actions > .ui-btn {
  --ui-btn-height: 2.75rem;
  --ui-btn-radius: 1rem;
  --ui-btn-padding: var(--space-2-5) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
}

.admin-awd-review-shell__hero-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--admin-control-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
}

.admin-awd-review-shell__summary {
  --metric-panel-columns: 3;
  --metric-panel-border: color-mix(in srgb, var(--workspace-brand) 16%, var(--workspace-line-soft));
  --metric-panel-background:
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--workspace-brand) 16%, transparent),
      transparent 60%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, transparent),
      color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base))
    );
}

@media (max-width: 900px) {
  .admin-awd-review-shell__hero-actions {
    width: 100%;
  }
}
</style>

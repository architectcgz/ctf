<script setup lang="ts">
import { BarChart2, CheckCircle, RefreshCw, Shield, Trophy, Users } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import { useScoreboardDetailPage } from '@/features/scoreboard'

const {
  contest,
  page,
  rows,
  scoreboard,
  total,
  totalPages,
  loading,
  refreshing,
  supportsRealtime,
  accentStyle,
  emptyTitle,
  emptyDescription,
  topScore,
  solvedCount,
  getStatusLabel,
  formatDateTime,
  formatContestWindow,
  getRowClass,
  getRankPillClass,
  getStatusCopy,
  changePage,
  loadScoreboard,
} = useScoreboardDetailPage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
    :style="accentStyle"
  >
    <main class="content-pane">
      <div class="scoreboard-detail-page">
        <header class="workspace-page-header scoreboard-detail-hero">
          <div>
            <div class="workspace-overline scoreboard-panel-overline">
              Contest Scoreboard
            </div>
            <h1 class="scoreboard-detail-title workspace-page-title">
              {{ contest?.title || '竞赛排行榜' }}
            </h1>
            <p class="scoreboard-detail-subtitle">
              {{ getStatusCopy(contest?.status, scoreboard?.frozen) }}
            </p>
          </div>
          <button
            type="button"
            class="ui-btn ui-btn--secondary"
            :disabled="loading || refreshing"
            @click="loadScoreboard(true)"
          >
            <RefreshCw
              class="h-4 w-4"
              :class="{ 'animate-spin': refreshing }"
            />
            刷新
          </button>
        </header>

        <ScoreboardRealtimeBridge
          v-if="contest && supportsRealtime"
          :contest-id="contest.id"
          @updated="loadScoreboard(true)"
        />

        <section class="scoreboard-summary">
          <div class="scoreboard-summary-title">
            <BarChart2 class="h-4 w-4" />
            <span>排行概况</span>
          </div>
          <div class="scoreboard-summary-grid metric-panel-grid">
            <div class="scoreboard-summary-item progress-card metric-panel-card">
              <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                <span>榜单队伍</span>
                <Users class="h-4 w-4" />
              </div>
              <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                {{ total }}
              </div>
              <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                当前进入排行榜的队伍总数
              </div>
            </div>
            <div class="scoreboard-summary-item progress-card metric-panel-card">
              <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                <span>本页最高分</span>
                <Trophy class="h-4 w-4" />
              </div>
              <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                {{ topScore }}
              </div>
              <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                当前页最高排名队伍分数
              </div>
            </div>
            <div class="scoreboard-summary-item progress-card metric-panel-card">
              <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                <span>本页解题</span>
                <CheckCircle class="h-4 w-4" />
              </div>
              <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                {{ solvedCount }}
              </div>
              <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                当前页队伍累计解题数
              </div>
            </div>
            <div class="scoreboard-summary-item progress-card metric-panel-card">
              <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                <span>榜单状态</span>
                <Shield class="h-4 w-4" />
              </div>
              <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                {{ scoreboard?.frozen ? '封榜' : getStatusLabel(contest?.status ?? 'ended') }}
              </div>
              <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                {{ formatContestWindow(contest) }}
              </div>
            </div>
          </div>
        </section>

        <section
          class="student-directory-section workspace-directory-section scoreboard-detail-directory-section"
          aria-label="排行详情"
        >
          <section
            class="student-directory-shell scoreboard-detail-directory workspace-directory-list"
          >
            <header class="student-directory-shell__head student-directory-list-heading list-heading">
              <div class="student-directory-shell__heading student-directory-list-heading__body">
                <div
                  class="journal-note-label student-directory-shell__eyebrow student-directory-list-heading__eyebrow"
                >
                  Scoreboard Detail
                </div>
                <h2 class="student-directory-shell__title student-directory-list-heading__title">
                  排行详情
                </h2>
              </div>
              <div class="student-directory-shell__meta scoreboard-detail-directory__meta">
                第 {{ page }} 页，当前展示 {{ rows.length }} / {{ total }} 支队伍
                <span
                  v-if="scoreboard?.frozen"
                  class="scoreboard-frozen-inline"
                >
                  <Shield class="h-3 w-3" /> 已冻结
                </span>
              </div>
            </header>

            <div
              v-if="loading && !scoreboard"
              class="scoreboard-loading student-directory-state workspace-directory-loading"
            >
              <div class="student-directory-spinner" />
            </div>

            <AppEmpty
              v-else-if="!rows.length"
              class="scoreboard-empty-state student-directory-state workspace-directory-empty"
              icon="Trophy"
              :title="emptyTitle"
              :description="emptyDescription"
            >
              <template #action>
                <button
                  type="button"
                  class="ui-btn ui-btn--secondary"
                  @click="loadScoreboard(true)"
                >
                  重新加载
                </button>
              </template>
            </AppEmpty>

            <template v-else>
              <div class="workspace-directory-grid-head scoreboard-detail-directory-head">
                <span>排名</span>
                <span>队伍</span>
                <span>得分</span>
                <span>解题数</span>
                <span>最近得分</span>
              </div>

              <div
                v-for="item in rows"
                :key="item.team_id"
                data-testid="scoreboard-detail-row"
                class="workspace-directory-grid-row scoreboard-detail-row"
                :class="getRowClass(item.rank)"
              >
                <div class="scoreboard-detail-rank">
                  <span :class="getRankPillClass(item.rank)">{{ item.rank }}</span>
                </div>
                <div class="workspace-directory-row-title scoreboard-detail-team">
                  {{ item.team_name }}
                </div>
                <div class="workspace-directory-compact-text scoreboard-detail-score">
                  {{ item.score }}
                </div>
                <div class="workspace-directory-compact-text">
                  {{ item.solved_count }}
                </div>
                <div class="workspace-directory-compact-text sb-cell--muted">
                  {{ formatDateTime(item.last_submission_at) }}
                </div>
              </div>

              <div class="scoreboard-pagination workspace-directory-pagination">
                <PagePaginationControls
                  :page="page"
                  :total-pages="totalPages"
                  :total="total"
                  :total-label="`共 ${total} 支队伍`"
                  :disabled="loading || refreshing"
                  show-jump
                  @change-page="changePage"
                />
              </div>
            </template>
          </section>
        </section>
      </div>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
}

.scoreboard-detail-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.scoreboard-detail-hero {
  gap: var(--space-4);
}

.scoreboard-panel-overline {
  margin-bottom: var(--space-2);
}

.scoreboard-detail-title {
  max-width: 760px;
}

.scoreboard-detail-subtitle {
  max-width: 720px;
  margin-top: var(--space-3);
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.scoreboard-summary {
  margin-top: var(--space-6);
}

.scoreboard-summary-title {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--journal-ink);
}

.scoreboard-loading {
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.scoreboard-empty-state) {
  margin-top: 0;
}

.scoreboard-detail-directory-section {
  margin-top: var(--space-6);
}

.scoreboard-detail-directory {
  --workspace-directory-grid-columns: 5.5rem minmax(0, 1.2fr) 7.5rem 7.5rem minmax(13rem, 1fr);
}

.scoreboard-detail-directory__meta {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

.scoreboard-frozen-inline {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--scoreboard-accent, var(--journal-accent));
}

.sb-rank-pill {
  display: inline-flex;
  align-items: center;
  min-height: 1.625rem;
  padding: 0 var(--space-2);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-12);
  font-weight: 600;
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.scoreboard-detail-rank,
.scoreboard-detail-team,
.scoreboard-detail-score {
  min-width: 0;
}

.scoreboard-detail-row.sb-row--top1,
.sb-rank-pill--top1 {
  color: color-mix(in srgb, var(--color-warning) 84%, var(--journal-ink));
}

.scoreboard-detail-row.sb-row--top2,
.sb-rank-pill--top2 {
  color: color-mix(in srgb, var(--color-text-secondary) 80%, var(--journal-ink));
}

.scoreboard-detail-row.sb-row--top3,
.sb-rank-pill--top3 {
  color: color-mix(in srgb, var(--color-danger) 42%, var(--color-warning));
}

.scoreboard-detail-row .scoreboard-detail-team {
  color: inherit;
}

.scoreboard-detail-row:not(.sb-row--top1, .sb-row--top2, .sb-row--top3) .scoreboard-detail-team {
  color: var(--journal-ink);
}

.sb-cell--muted {
  color: var(--journal-muted);
}

@media (max-width: 1180px) {
  .scoreboard-detail-directory-head {
    display: none;
  }

  .scoreboard-detail-row {
    grid-template-columns: 1fr;
  }
}
</style>

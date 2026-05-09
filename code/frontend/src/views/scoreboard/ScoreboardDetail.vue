<script setup lang="ts">
import { ArrowLeft, BarChart2, CheckCircle, RefreshCw, Shield, Trophy, Users } from 'lucide-vue-next'

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
        <router-link
          class="scoreboard-back-link"
          :to="{ name: 'Scoreboard' }"
        >
          <ArrowLeft class="h-4 w-4" />
          返回排行列表
        </router-link>

        <header class="scoreboard-detail-hero">
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

        <div
          v-if="loading && !scoreboard"
          class="scoreboard-loading"
        >
          <div class="scoreboard-loading-spinner" />
        </div>

        <AppEmpty
          v-else-if="!rows.length"
          class="scoreboard-empty-state"
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

        <section
          v-else
          class="scoreboard-directory"
          aria-label="排行详情"
        >
          <div class="scoreboard-directory-top">
            <h2 class="scoreboard-directory-title">
              排行详情
            </h2>
            <div class="scoreboard-directory-meta">
              第 {{ page }} 页，当前展示 {{ rows.length }} / {{ total }} 支队伍
              <span
                v-if="scoreboard?.frozen"
                class="scoreboard-frozen-inline"
              >
                <Shield class="h-3 w-3" /> 已冻结
              </span>
            </div>
          </div>

          <div class="scoreboard-table-shell overflow-x-auto">
            <table class="sb-table">
              <thead>
                <tr>
                  <th>排名</th>
                  <th>队伍</th>
                  <th>得分</th>
                  <th>解题数</th>
                  <th>最近得分</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in rows"
                  :key="item.team_id"
                  data-testid="scoreboard-detail-row"
                  :class="getRowClass(item.rank)"
                >
                  <td class="sb-cell--rank">
                    <span :class="getRankPillClass(item.rank)">{{ item.rank }}</span>
                  </td>
                  <td>{{ item.team_name }}</td>
                  <td class="sb-cell--mono">
                    {{ item.score }}
                  </td>
                  <td>{{ item.solved_count }}</td>
                  <td class="sb-cell--muted">
                    {{ formatDateTime(item.last_submission_at) }}
                  </td>
                </tr>
              </tbody>
            </table>
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

.scoreboard-back-link {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: var(--radius-md);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--journal-muted);
}

.scoreboard-back-link:hover,
.scoreboard-back-link:focus-visible {
  color: var(--journal-accent);
  border-color: color-mix(in srgb, var(--journal-accent) 36%, var(--journal-border));
}

.scoreboard-detail-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-top: var(--space-5);
  padding-bottom: var(--space-5);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
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
  padding: var(--space-16) 0;
}

.scoreboard-loading-spinner {
  width: 2rem;
  height: 2rem;
  border: 0.25rem solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: scoreboardSpin 900ms linear infinite;
}

:deep(.scoreboard-empty-state) {
  margin-top: var(--space-6);
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-directory {
  margin-top: var(--space-6);
}

.scoreboard-frozen-inline {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  margin-left: var(--space-2);
  color: var(--scoreboard-accent, var(--journal-accent));
}

.scoreboard-table-shell {
  overflow-x: auto;
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

.sb-table {
  width: 100%;
  border-collapse: collapse;
}

.sb-table th {
  padding: 0 0 var(--space-3);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  text-align: left;
  color: var(--journal-muted);
}

.sb-row td {
  padding: var(--space-4) 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  font-size: var(--font-size-14);
  color: var(--journal-ink);
}

.sb-row--top1 td,
.sb-rank-pill--top1 {
  color: color-mix(in srgb, var(--color-warning) 84%, var(--journal-ink));
}

.sb-row--top2 td,
.sb-rank-pill--top2 {
  color: color-mix(in srgb, var(--color-text-secondary) 80%, var(--journal-ink));
}

.sb-row--top3 td,
.sb-rank-pill--top3 {
  color: color-mix(in srgb, var(--color-danger) 42%, var(--color-warning));
}

.sb-cell--rank,
.sb-cell--mono {
  font-family: var(--font-family-mono);
}

.sb-cell--muted {
  color: var(--journal-muted);
}

@keyframes scoreboardSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}
</style>

<script setup lang="ts">
import { ArrowRight, BarChart2, Clock3, Flag, Shield, Trophy, Users } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import {
  useScoreboardContestDirectoryPage,
  useScoreboardRoutePage,
  useScoreboardView,
} from '@/features/scoreboard'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

const { panelTabs, activeTab, setTabButtonRef, selectTab, handleTabKeydown } =
  useScoreboardRoutePage()

const {
  contestPage,
  contestPageSize,
  contestSummary,
  contestTotal,
  contestTotalPages,
  changeContestPage,
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
  getCardDescription,
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

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <div class="scoreboard-page">
      <nav
        class="workspace-tabbar top-tabs"
        role="tablist"
        aria-label="排行榜视图切换"
      >
        <button
          v-for="(tab, index) in panelTabs"
          :id="tab.tabId"
          :key="tab.tabId"
          :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
          type="button"
          role="tab"
          class="workspace-tab top-tab"
          :class="{ active: activeTab === tab.key }"
          :aria-selected="activeTab === tab.key ? 'true' : 'false'"
          :aria-controls="tab.panelId"
          :tabindex="activeTab === tab.key ? 0 : -1"
          @click="selectTab(tab.key)"
          @keydown="handleTabKeydown($event, index)"
        >
          {{ tab.label }}
        </button>
      </nav>

      <main class="content-pane">
        <section
          v-show="activeTab === 'contest'"
          id="scoreboard-panel-contest"
          class="tab-panel"
          :class="{ active: activeTab === 'contest' }"
          role="tabpanel"
          aria-labelledby="scoreboard-tab-contest"
          :aria-hidden="activeTab === 'contest' ? 'false' : 'true'"
        >
          <div class="workspace-overline scoreboard-panel-overline">
            Contest Scoreboard
          </div>

          <section class="scoreboard-summary">
            <div class="scoreboard-summary-title">
              <BarChart2 class="h-4 w-4" />
              <span>当前排行概况</span>
            </div>
            <div class="scoreboard-summary-grid metric-panel-grid">
              <div class="scoreboard-summary-item progress-card metric-panel-card">
                <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                  <span>展示竞赛</span>
                  <Trophy class="h-4 w-4" />
                </div>
                <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                  {{ contestCount }}
                </div>
                <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                  当前可查看排行的竞赛总数
                </div>
              </div>
              <div class="scoreboard-summary-item progress-card metric-panel-card">
                <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                  <span>进行中</span>
                  <Clock3 class="h-4 w-4" />
                </div>
                <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                  {{ runningCount }}
                </div>
                <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                  支持进入后实时刷新的竞赛数量
                </div>
              </div>
              <div class="scoreboard-summary-item progress-card metric-panel-card">
                <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                  <span>冻结竞赛</span>
                  <Shield class="h-4 w-4" />
                </div>
                <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                  {{ frozenCount }}
                </div>
                <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                  当前处于封榜阶段的竞赛数量
                </div>
              </div>
              <div class="scoreboard-summary-item progress-card metric-panel-card">
                <div class="scoreboard-summary-label progress-card-label metric-panel-label">
                  <span>已结束</span>
                  <Flag class="h-4 w-4" />
                </div>
                <div class="scoreboard-summary-value progress-card-value metric-panel-value">
                  {{ endedCount }}
                </div>
                <div class="scoreboard-summary-helper progress-card-hint metric-panel-helper">
                  可查看最终成绩的历史竞赛
                </div>
              </div>
            </div>
          </section>

          <div
            v-if="loading && !hasSections"
            class="scoreboard-loading"
          >
            <div class="scoreboard-loading-spinner" />
          </div>

          <AppEmpty
            v-else-if="!hasSections"
            class="scoreboard-empty-state"
            icon="Trophy"
            :title="emptyTitle"
            :description="selectionHint"
          >
            <template #action>
              <button
                type="button"
                class="ui-btn ui-btn--secondary"
                @click="refresh"
              >
                重新加载
              </button>
            </template>
          </AppEmpty>

          <section
            v-else
            class="scoreboard-directory workspace-directory-list"
            aria-label="排行榜列表"
          >
            <div class="scoreboard-directory-top">
              <h2 class="scoreboard-directory-title">
                竞赛排行列表
              </h2>
              <div class="scoreboard-directory-meta">
                按竞赛开始时间倒序展示排行榜
              </div>
            </div>

            <div class="scoreboard-sections">
              <router-link
                v-for="(section, index) in paginatedSections"
                :key="section.contest.id"
                data-testid="scoreboard-card"
                class="scoreboard-card scoreboard-card-link"
                :style="sectionAccentStyle(section.contest.status)"
                :to="{ name: 'ScoreboardDetail', params: { contestId: section.contest.id } }"
              >
                <div class="scoreboard-card-header">
                  <div class="scoreboard-card-main">
                    <div class="scoreboard-card-chips">
                      <span class="sb-index">{{
                        String(contestPageStartIndex + index + 1).padStart(2, '0')
                      }}</span>
                      <span class="sb-status-chip">{{
                        getStatusLabel(section.contest.status)
                      }}</span>
                      <span class="sb-mode-chip">{{ getModeLabel(section.contest.mode) }}</span>
                      <span
                        v-if="section.frozen"
                        class="sb-frozen-chip"
                      >
                        <Shield class="h-3 w-3" /> 已冻结
                      </span>
                    </div>
                    <h3 class="scoreboard-card-title">
                      {{ section.contest.title }}
                    </h3>
                    <p class="scoreboard-card-time">
                      {{ formatContestWindow(section.contest.starts_at, section.contest.ends_at) }}
                    </p>
                    <p class="scoreboard-card-description">
                      {{
                        getCardDescription(
                          section.contest.status,
                          section.frozen
                        )
                      }}
                    </p>
                  </div>
                  <div class="scoreboard-card-meta">
                    <Users class="h-3.5 w-3.5" />
                    <span>点击进入排行详情</span>
                    <ArrowRight class="h-4 w-4" />
                  </div>
                </div>
              </router-link>
            </div>

            <div
              class="scoreboard-pagination workspace-directory-pagination"
            >
              <PagePaginationControls
                :page="contestPage"
                :total-pages="contestTotalPages"
                :total="contestTotal"
                :total-label="`共 ${contestTotal} 个竞赛`"
                :disabled="loading"
                show-jump
                @change-page="changeContestPage"
              />
            </div>
          </section>
        </section>

        <section
          v-show="activeTab === 'points'"
          id="scoreboard-panel-points"
          class="tab-panel"
          :class="{ active: activeTab === 'points' }"
          role="tabpanel"
          aria-labelledby="scoreboard-tab-points"
          :aria-hidden="activeTab === 'points' ? 'false' : 'true'"
        >
          <div class="workspace-overline scoreboard-panel-overline">
            Points Scoreboard
          </div>

          <div
            v-if="rankingLoading"
            class="scoreboard-loading"
          >
            <div class="scoreboard-loading-spinner" />
          </div>

          <AppEmpty
            v-else-if="!hasRankingRows"
            class="scoreboard-empty-state"
            icon="Trophy"
            :title="pointsEmptyTitle"
            :description="rankingHint"
          >
            <template #action>
              <button
                type="button"
                class="ui-btn ui-btn--secondary"
                @click="refreshPracticeRanking"
              >
                重新加载
              </button>
            </template>
          </AppEmpty>

          <div
            v-else
            class="scoreboard-table-shell workspace-directory-list overflow-x-auto"
          >
            <table class="sb-table">
              <thead>
                <tr>
                  <th>排名</th>
                  <th>用户名</th>
                  <th>积分</th>
                  <th>解题数</th>
                  <th>班级</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in rankingRows"
                  :key="item.user_id"
                  :class="getRowClass(item.rank)"
                >
                  <td class="sb-cell--rank">
                    <span :class="getRankPillClass(item.rank)">{{ item.rank }}</span>
                  </td>
                  <td>{{ item.username }}</td>
                  <td class="sb-cell--mono">
                    {{ item.total_score }}
                  </td>
                  <td>{{ item.solved_count }}</td>
                  <td class="sb-cell--muted">
                    {{ item.class_name || '未分配班级' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </main>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
}

.scoreboard-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.scoreboard-panel-overline {
  margin-bottom: 18px;
}

.scoreboard-inline-note {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 8px;
  font-size: var(--font-size-12);
  font-weight: 600;
  color: var(--journal-muted);
}

.scoreboard-inline-note-danger {
  color: var(--color-danger);
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--journal-border));
  background: color-mix(in srgb, var(--color-danger) 6%, transparent);
}

.scoreboard-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
}

.scoreboard-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: scoreboardSpin 900ms linear infinite;
}

:deep(.scoreboard-empty-state) {
  margin-top: 24px;
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-directory {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  margin-top: 24px;
  box-shadow: 0 18px 34px color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.scoreboard-sections {
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-card {
  padding: var(--space-5) var(--space-4-5);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-card-link {
  display: block;
  transition:
    background 160ms ease,
    border-color 160ms ease,
    transform 160ms ease;
}

.scoreboard-card-link:hover,
.scoreboard-card-link:focus-visible {
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 4%, transparent);
  transform: translateY(-0.0625rem);
}

.scoreboard-card-header {
  display: flex;
  flex-wrap: wrap;
  align-items: start;
  justify-content: space-between;
  gap: 16px;
}

.scoreboard-card-main {
  min-width: 0;
}

.scoreboard-card-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.scoreboard-card-title {
  margin-top: 10px;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-18);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
}

.scoreboard-card-time,
.scoreboard-card-meta {
  margin-top: 6px;
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.scoreboard-card-description {
  margin-top: 8px;
  max-width: 700px;
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, var(--journal-ink));
}

.scoreboard-card-meta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.scoreboard-table-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  margin-top: var(--space-4);
  overflow-x: auto;
  box-shadow: 0 18px 34px color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.sb-index,
.sb-status-chip,
.sb-mode-chip,
.sb-frozen-chip,
.sb-rank-pill {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 8px;
  font-size: var(--font-size-12);
  font-weight: 600;
}

.sb-index {
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 12%, transparent);
  color: var(--scoreboard-accent, var(--journal-accent));
}

.sb-status-chip,
.sb-frozen-chip {
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 10%, transparent);
  color: var(--scoreboard-accent, var(--journal-accent));
}

.sb-mode-chip,
.sb-rank-pill {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.sb-table {
  width: 100%;
  border-collapse: collapse;
}

.sb-table th {
  padding: 0 0 12px;
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  text-align: left;
  color: var(--journal-muted);
}

.sb-row td {
  padding: 14px 0;
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

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

function onContestStatusFilterChange(event: Event): void {
  void updateContestStatusFilter(
    (event.target as HTMLSelectElement).value as typeof contestStatusFilter.value
  )
}

function onContestModeFilterChange(event: Event): void {
  void updateContestModeFilter(
    (event.target as HTMLSelectElement).value as typeof contestModeFilter.value
  )
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <div class="scoreboard-page">
      <nav class="workspace-tabbar top-tabs" role="tablist" aria-label="排行榜视图切换">
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

          <section
            class="student-directory-section workspace-directory-section"
            aria-label="排行榜列表"
          >
            <section class="student-directory-shell scoreboard-directory workspace-directory-list">
              <header class="student-directory-shell__head student-directory-list-heading list-heading">
                <div class="student-directory-shell__heading student-directory-list-heading__body">
                  <h2 class="student-directory-shell__title student-directory-list-heading__title">
                    竞赛排行列表
                  </h2>
                </div>
                <div class="student-directory-shell__meta">按竞赛开始时间倒序展示排行榜</div>
              </header>

              <section class="student-directory-filters scoreboard-directory-filters" aria-label="竞赛排行筛选">
                <div class="student-directory-filter-grid scoreboard-directory-filter-grid">
                  <label class="student-directory-filter-field" for="scoreboard-status-filter">
                    <span class="student-directory-filter-label">状态</span>
                    <div class="ui-control-wrap student-directory-filter-control">
                      <select
                        id="scoreboard-status-filter"
                        :value="contestStatusFilter"
                        class="ui-control"
                        @change="onContestStatusFilterChange"
                      >
                        <option value="">全部状态</option>
                        <option value="running">进行中</option>
                        <option value="frozen">已冻结</option>
                        <option value="ended">已结束</option>
                      </select>
                    </div>
                  </label>

                  <label class="student-directory-filter-field" for="scoreboard-mode-filter">
                    <span class="student-directory-filter-label">模式</span>
                    <div class="ui-control-wrap student-directory-filter-control">
                      <select
                        id="scoreboard-mode-filter"
                        :value="contestModeFilter"
                        class="ui-control"
                        @change="onContestModeFilterChange"
                      >
                        <option value="">全部模式</option>
                        <option value="jeopardy">Jeopardy</option>
                        <option value="awd">AWD</option>
                      </select>
                    </div>
                  </label>

                  <div class="student-directory-filter-actions">
                    <span
                      class="student-directory-filter-label student-directory-filter-label--ghost"
                      aria-hidden="true"
                    >
                      操作
                    </span>
                    <div class="student-directory-filter-action-row">
                      <button
                        type="button"
                        class="ui-btn ui-btn--ghost"
                        :disabled="!contestStatusFilter && !contestModeFilter"
                        @click="resetContestFilters"
                      >
                        清空筛选
                      </button>
                    </div>
                  </div>
                </div>
              </section>

              <div
                v-if="loading && !hasSections"
                class="scoreboard-loading student-directory-state workspace-directory-loading"
              >
                <div class="student-directory-spinner" />
              </div>

              <AppEmpty
                v-else-if="!hasSections"
                class="scoreboard-empty-state student-directory-state workspace-directory-empty"
                icon="Trophy"
                :title="emptyTitle"
                :description="selectionHint"
              >
                <template #action>
                  <button type="button" class="ui-btn ui-btn--secondary" @click="refresh">
                    重新加载
                  </button>
                </template>
              </AppEmpty>

              <template v-else>
                <div class="workspace-directory-grid-head scoreboard-directory-head">
                  <span>序号</span>
                  <span>竞赛</span>
                  <span>状态</span>
                  <span>模式</span>
                  <span>时间</span>
                  <span class="scoreboard-directory-head__action">操作</span>
                </div>

                <router-link
                  v-for="(section, index) in paginatedSections"
                  :key="section.contest.id"
                  data-testid="scoreboard-card"
                  class="workspace-directory-grid-row scoreboard-card scoreboard-card-link"
                  :style="sectionAccentStyle(section.contest.status)"
                  :to="{ name: 'ScoreboardDetail', params: { contestId: section.contest.id } }"
                >
                  <div>
                    <span class="workspace-directory-status-pill sb-index">{{
                      String(contestPageStartIndex + index + 1).padStart(2, '0')
                    }}</span>
                  </div>
                  <div class="workspace-directory-cell scoreboard-card-main">
                    <h3 class="workspace-directory-row-title scoreboard-card-title">
                      {{ section.contest.title }}
                    </h3>
                  </div>
                  <div>
                    <span class="workspace-directory-status-pill sb-status-chip">{{
                      getStatusLabel(section.contest.status)
                    }}</span>
                  </div>
                  <div>
                    <span
                      class="workspace-directory-status-pill workspace-directory-status-pill--muted sb-mode-chip"
                      >{{ getModeLabel(section.contest.mode) }}</span
                    >
                    <span
                      v-if="section.frozen"
                      class="workspace-directory-status-pill sb-frozen-chip"
                    >
                      <Shield class="h-3 w-3" /> 已冻结
                    </span>
                  </div>
                  <div class="workspace-directory-compact-text scoreboard-card-time">
                    {{ formatContestWindow(section.contest.starts_at, section.contest.ends_at) }}
                  </div>
                  <div class="workspace-directory-row-btn scoreboard-card-meta">
                    <Users class="h-3.5 w-3.5" />
                    <span>进入详情</span>
                    <ArrowRight class="h-4 w-4" />
                  </div>
                </router-link>

                <div class="scoreboard-pagination workspace-directory-pagination">
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
              </template>
            </section>
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
          <section
            class="student-directory-section workspace-directory-section"
            aria-label="积分排行榜"
          >
            <section
              class="student-directory-shell scoreboard-table-shell workspace-directory-list"
            >
              <header class="student-directory-shell__head student-directory-list-heading list-heading">
                <div class="student-directory-shell__heading student-directory-list-heading__body">
                  <h2 class="student-directory-shell__title student-directory-list-heading__title">
                    积分排行列表
                  </h2>
                </div>
              </header>

              <div
                v-if="rankingLoading"
                class="scoreboard-loading student-directory-state workspace-directory-loading"
              >
                <div class="student-directory-spinner" />
              </div>

              <AppEmpty
                v-else-if="!hasRankingRows"
                class="scoreboard-empty-state student-directory-state workspace-directory-empty"
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

              <div v-else class="scoreboard-table-scroll overflow-x-auto">
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
                      <td>
                        <span
                          class="workspace-directory-status-pill"
                          :class="getRankPillClass(item.rank)"
                          >{{ item.rank }}</span
                        >
                      </td>
                      <td>{{ item.username }}</td>
                      <td>
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
          </section>
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
  margin-bottom: var(--space-3);
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
}

:deep(.scoreboard-empty-state) {
  margin-top: 0;
}

.scoreboard-directory {
  --workspace-directory-grid-columns: 5.5rem minmax(0, 1.25fr) 7.5rem 9.5rem minmax(13rem, 1fr)
    8.5rem;
}

.scoreboard-card {
  --workspace-directory-row-accent: var(--scoreboard-accent, var(--journal-accent));
}

.scoreboard-card-link {
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.scoreboard-card-link:hover,
.scoreboard-card-link:focus-visible {
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 5%, transparent);
}

.scoreboard-card-main {
  min-width: 0;
}

.scoreboard-card-title {
  min-width: 0;
}

.scoreboard-card-meta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

.scoreboard-table-shell {
  overflow-x: auto;
}

.scoreboard-table-scroll {
  overflow-x: auto;
}

.sb-index {
  border-color: color-mix(
    in srgb,
    var(--scoreboard-accent, var(--journal-accent)) 22%,
    transparent
  );
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 12%, transparent);
  color: var(--scoreboard-accent, var(--journal-accent));
}

.sb-status-chip,
.sb-frozen-chip {
  border-color: color-mix(
    in srgb,
    var(--scoreboard-accent, var(--journal-accent)) 22%,
    transparent
  );
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 10%, transparent);
  color: var(--scoreboard-accent, var(--journal-accent));
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

.sb-cell--muted {
  color: var(--journal-muted);
}

@media (max-width: 1180px) {
  .scoreboard-directory-head {
    display: none;
  }

  .scoreboard-card {
    grid-template-columns: minmax(0, 1fr);
  }

  .scoreboard-card-meta {
    justify-content: flex-start;
  }
}
</style>

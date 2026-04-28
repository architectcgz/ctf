<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowRight, BarChart2, Shield, Trophy, Users } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import type { ContestStatus } from '@/api/contracts'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import { useScoreboardView } from '@/composables/useScoreboardView'
import { getContestAccentColor, getModeLabel, getStatusLabel } from '@/utils/contest'

type ScoreboardPanelKey = 'contest' | 'points'

const route = useRoute()
const router = useRouter()
const panelTabs: Array<{ key: ScoreboardPanelKey; label: string; panelId: string; tabId: string }> =
  [
    {
      key: 'contest',
      label: '竞赛排行榜',
      panelId: 'scoreboard-panel-contest',
      tabId: 'scoreboard-tab-contest',
    },
    {
      key: 'points',
      label: '积分排行榜',
      panelId: 'scoreboard-panel-points',
      tabId: 'scoreboard-tab-points',
    },
  ]
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } =
  useRouteQueryTabs<ScoreboardPanelKey>({
    route,
    router,
    orderedTabs: panelTabs.map((tab) => tab.key) as ScoreboardPanelKey[],
    defaultTab: 'contest',
  })

const {
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
const contestCount = computed(() => sections.value.length)
const runningCount = computed(
  () => sections.value.filter((section) => section.contest.status === 'running').length
)
const frozenCount = computed(() => sections.value.filter((section) => section.frozen).length)
const endedCount = computed(
  () => sections.value.filter((section) => section.contest.status === 'ended').length
)
const emptyTitle = computed(() =>
  selectionHint.value.includes('失败') ? '排行榜加载失败' : '暂无可查看的竞赛排行榜'
)
const pointsEmptyTitle = computed(() =>
  rankingError.value ? '积分排行榜加载失败' : '暂无可查看的积分排行榜'
)

function formatDateTime(value?: string): string {
  if (!value) return '未记录'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatContestWindow(startsAt: string, endsAt: string): string {
  return `${formatDateTime(startsAt)} ~ ${formatDateTime(endsAt)}`
}

function sectionAccentStyle(status: ContestStatus): Record<string, string> {
  return { '--scoreboard-accent': getContestAccentColor(status) }
}

function getRowClass(rank: number): string {
  if (rank === 1) return 'sb-row sb-row--top1'
  if (rank === 2) return 'sb-row sb-row--top2'
  if (rank === 3) return 'sb-row sb-row--top3'
  return 'sb-row'
}

function getRankPillClass(rank: number): string[] {
  return [
    'sb-rank-pill',
    rank === 1 ? 'sb-rank-pill--top1' : '',
    rank === 2 ? 'sb-rank-pill--top2' : '',
    rank === 3 ? 'sb-rank-pill--top3' : '',
  ]
}

function getCardDescription(
  status: ContestStatus,
  frozen: boolean
): string {
  if (frozen || status === 'frozen') {
    return '封榜阶段先展示竞赛入口，进入后查看冻结前排名。'
  }

  if (status === 'running') {
    return '进行中竞赛进入详情后支持实时刷新，提交后榜单会自动更新。'
  }

  return '历史竞赛进入详情后展示最终成绩，可用于复盘队伍解题表现。'
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <div class="scoreboard-page">
      <nav
        class="top-tabs"
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
          class="top-tab"
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
              <div class="scoreboard-summary-item metric-panel-card">
                <div class="scoreboard-summary-label metric-panel-label">
                  展示竞赛
                </div>
                <div class="scoreboard-summary-value metric-panel-value">
                  {{ contestCount }}
                </div>
                <div class="scoreboard-summary-helper metric-panel-helper">
                  当前可查看排行的竞赛总数
                </div>
              </div>
              <div class="scoreboard-summary-item metric-panel-card">
                <div class="scoreboard-summary-label metric-panel-label">
                  进行中
                </div>
                <div class="scoreboard-summary-value metric-panel-value">
                  {{ runningCount }}
                </div>
                <div class="scoreboard-summary-helper metric-panel-helper">
                  支持进入后实时刷新的竞赛数量
                </div>
              </div>
              <div class="scoreboard-summary-item metric-panel-card">
                <div class="scoreboard-summary-label metric-panel-label">
                  冻结竞赛
                </div>
                <div class="scoreboard-summary-value metric-panel-value">
                  {{ frozenCount }}
                </div>
                <div class="scoreboard-summary-helper metric-panel-helper">
                  当前处于封榜阶段的竞赛数量
                </div>
              </div>
              <div class="scoreboard-summary-item metric-panel-card">
                <div class="scoreboard-summary-label metric-panel-label">
                  已结束
                </div>
                <div class="scoreboard-summary-value metric-panel-value">
                  {{ endedCount }}
                </div>
                <div class="scoreboard-summary-helper metric-panel-helper">
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
            class="scoreboard-directory"
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
              <article
                v-for="(section, index) in sections"
                :key="section.contest.id"
                data-testid="scoreboard-card"
                class="scoreboard-card"
                :style="sectionAccentStyle(section.contest.status)"
              >
                <div class="scoreboard-card-header">
                  <div class="scoreboard-card-main">
                    <div class="scoreboard-card-chips">
                      <span class="sb-index">{{ String(index + 1).padStart(2, '0') }}</span>
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
                    点击进入排行详情
                  </div>
                </div>

                <div class="scoreboard-card-divider" />

                <router-link
                  class="scoreboard-detail-link"
                  :to="{ name: 'ScoreboardDetail', params: { contestId: section.contest.id } }"
                >
                  <Trophy class="h-4 w-4" />
                  <span>查看完整排行榜</span>
                  <ArrowRight class="h-4 w-4" />
                </router-link>
              </article>
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
            class="scoreboard-table-shell overflow-x-auto"
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
  margin-top: 24px;
}

.scoreboard-sections {
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-card {
  padding: 22px 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
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

.scoreboard-card-divider {
  margin: 16px 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
}

.scoreboard-detail-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 32%, var(--journal-border));
  border-radius: var(--radius-md);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--scoreboard-accent, var(--journal-accent));
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 8%, transparent);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    transform 160ms ease;
}

.scoreboard-detail-link:hover,
.scoreboard-detail-link:focus-visible {
  border-color: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 54%, var(--journal-border));
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 12%, transparent);
  transform: translateY(-0.0625rem);
}

.scoreboard-table-shell {
  overflow-x: auto;
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

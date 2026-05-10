<script setup lang="ts">
import type { Component } from 'vue'
import { CalendarRange, Clock3, Flag, Trophy } from 'lucide-vue-next'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import { useContestListPage } from '@/features/contest-detail'

const {
  loading,
  refresh,
  visibleContests,
  total,
  page,
  totalPages,
  changePage,
  summaryMetrics,
  loadErrorMessage,
  formatTime,
  getTimelineHint,
  openContest,
  contestAccentStyle,
  getStatusLabel,
  getModeLabel,
  getContestActionLabel,
} = useContestListPage()

function summaryMetricIcon(key: string): Component {
  switch (key) {
    case 'running':
      return Clock3
    case 'registering':
      return CalendarRange
    case 'ended':
      return Flag
    default:
      return Trophy
  }
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="contest-page">
        <header class="contest-topbar">
          <div class="contest-heading">
            <div class="workspace-overline">Contests</div>
            <h1 class="contest-title workspace-page-title">竞赛中心</h1>
            <p class="contest-subtitle">查看当前可参加和已结束的竞赛，直接进入竞赛工作区。</p>
          </div>
        </header>

        <section class="contest-summary metric-panel-default-surface">
          <div class="contest-summary-title">
            <Trophy class="h-4 w-4" />
            <span>当前竞赛概况</span>
          </div>
          <div class="contest-summary-grid metric-panel-grid">
            <div
              v-for="stat in summaryMetrics"
              :key="stat.key"
              class="contest-summary-item progress-card metric-panel-card"
            >
              <div class="contest-summary-label progress-card-label metric-panel-label">
                <span>{{ stat.label }}</span>
                <component :is="summaryMetricIcon(stat.key)" class="h-4 w-4" />
              </div>
              <div class="contest-summary-value progress-card-value metric-panel-value">
                {{ stat.value }}
              </div>
              <div class="contest-summary-helper progress-card-hint metric-panel-helper">
                {{ stat.hint }}
              </div>
            </div>
          </div>
        </section>

        <div v-if="loading" class="contest-loading">
          <div class="contest-loading-spinner" />
        </div>

        <AppEmpty
          v-else-if="loadErrorMessage"
          class="contest-empty-state"
          icon="AlertTriangle"
          title="竞赛列表加载失败"
          :description="loadErrorMessage"
        >
          <template #action>
            <button type="button" class="ui-btn ui-btn--secondary" @click="refresh">重试</button>
          </template>
        </AppEmpty>

        <AppEmpty
          v-else-if="visibleContests.length === 0"
          class="contest-empty-state"
          icon="Flag"
          title="暂无竞赛"
          description="当前没有可展示的竞赛，稍后再来查看新的开赛计划。"
        />

        <section
          v-else
          class="contest-directory workspace-directory-list workspace-directory-list--catalog"
          aria-label="竞赛目录"
        >
          <div class="contest-directory-top">
            <h2 class="contest-directory-title">竞赛列表</h2>
            <div class="contest-directory-meta">共 {{ total }} 场</div>
          </div>

          <div class="workspace-directory-grid-head contest-directory-head">
            <span>竞赛</span>
            <span>时间</span>
            <span>状态</span>
            <span>节奏</span>
            <span>操作</span>
          </div>

          <button
            v-for="contest in visibleContests"
            :key="contest.id"
            type="button"
            class="workspace-directory-grid-row contest-row"
            :style="contestAccentStyle(contest.status)"
            :aria-label="`${contest.title}，${getStatusLabel(contest.status)}，${getModeLabel(contest.mode)}`"
            @click="openContest(contest)"
          >
            <div class="workspace-directory-cell contest-row-main">
              <div class="contest-row-status-strip">
                <span
                  class="workspace-directory-status-pill contest-chip"
                  :style="{ '--contest-chip-color': 'var(--contest-row-accent)' }"
                >
                  {{ getStatusLabel(contest.status) }}
                </span>
                <span
                  class="workspace-directory-status-pill workspace-directory-status-pill--muted contest-chip contest-chip-muted"
                  >{{ getModeLabel(contest.mode) }}</span
                >
              </div>
              <h3
                class="contest-row-title workspace-directory-row-title"
                :title="contest.title"
              >
                {{ contest.title }}
              </h3>
            </div>

            <div class="workspace-directory-compact-text contest-row-time">
              <div class="contest-row-time-item">
                <CalendarRange class="h-3.5 w-3.5" />
                <span>{{ formatTime(contest.starts_at) }} - {{ formatTime(contest.ends_at) }}</span>
              </div>
            </div>

            <div class="contest-row-state">
              <span
                class="workspace-directory-status-pill contest-state-chip"
                :style="{ '--contest-state-color': 'var(--contest-row-accent)' }"
              >
                {{ getStatusLabel(contest.status) }}
              </span>
            </div>

            <div class="workspace-directory-compact-text contest-row-timeline">
              <div class="contest-row-time-item contest-row-time-item-strong">
                <Clock3 class="h-3.5 w-3.5" />
                <span>{{ getTimelineHint(contest) }}</span>
              </div>
            </div>

            <div class="workspace-directory-row-btn contest-row-cta">
              <span>{{ getContestActionLabel(contest.status) }}</span>
            </div>
          </button>

          <div v-if="total > 0" class="contest-pagination workspace-directory-pagination">
            <PagePaginationControls
              :page="page"
              :total-pages="totalPages"
              :total="total"
              :total-label="`共 ${total} 场`"
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

.contest-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.contest-subtitle {
  max-width: 680px;
}

.contest-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
}

.contest-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: contestSpin 900ms linear infinite;
}

:deep(.contest-empty-state) {
  margin-top: 24px;
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.contest-directory {
  --workspace-directory-grid-columns: minmax(0, 1.4fr) minmax(13.75rem, 1fr) 7.5rem 11.25rem 7.5rem;
  margin-top: 24px;
}

.contest-pagination {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
}

.contest-row {
  --workspace-directory-row-accent: var(--contest-row-accent, var(--journal-accent));
  cursor: pointer;
}

.contest-row-main {
  min-width: 0;
}

.contest-row-status-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.contest-row-title {
  margin-top: var(--space-2-5);
  font-size: var(--font-size-15);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contest-chip {
  border-color: color-mix(
    in srgb,
    var(--contest-chip-color, var(--journal-accent)) 22%,
    transparent
  );
  background: color-mix(in srgb, var(--contest-chip-color, var(--journal-accent)) 12%, transparent);
  color: var(--contest-chip-color, var(--journal-accent));
}

.contest-state-chip {
  border-color: color-mix(
    in srgb,
    var(--contest-state-color, var(--journal-accent)) 22%,
    transparent
  );
  background: color-mix(
    in srgb,
    var(--contest-state-color, var(--journal-accent)) 12%,
    transparent
  );
  color: var(--contest-state-color, var(--journal-accent));
}

.contest-row-time-item {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
}

.contest-row-time-item-strong {
  color: var(--contest-row-accent, var(--journal-accent));
  font-weight: 600;
}

.contest-row-cta {
  justify-content: flex-start;
  color: var(--contest-row-accent, var(--journal-accent));
}

@keyframes contestSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1180px) {
  .contest-directory-head {
    display: none;
  }

  .contest-row {
    grid-template-columns: 1fr;
  }
}
</style>

<script setup lang="ts">
import type { Component } from 'vue'
import { ArrowRight, CalendarRange, Clock3, Flag, Trophy } from 'lucide-vue-next'
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

        <section
          class="student-directory-section workspace-directory-section"
          aria-label="竞赛目录"
        >
          <section class="student-directory-shell contest-directory workspace-directory-list">
            <header class="student-directory-shell__head">
              <div class="student-directory-shell__heading">
                <div class="journal-note-label student-directory-shell__eyebrow">
                  Contest Directory
                </div>
                <h2 class="student-directory-shell__title">竞赛列表</h2>
              </div>
              <div class="student-directory-shell__meta">共 {{ total }} 场</div>
            </header>

            <div
              v-if="loading"
              class="contest-loading student-directory-state workspace-directory-loading"
            >
              <div class="student-directory-spinner" />
            </div>

            <AppEmpty
              v-else-if="loadErrorMessage"
              class="contest-empty-state student-directory-state workspace-directory-empty"
              icon="AlertTriangle"
              title="竞赛列表加载失败"
              :description="loadErrorMessage"
            >
              <template #action>
                <button type="button" class="ui-btn ui-btn--secondary" @click="refresh">
                  重试
                </button>
              </template>
            </AppEmpty>

            <AppEmpty
              v-else-if="visibleContests.length === 0"
              class="contest-empty-state student-directory-state workspace-directory-empty"
              icon="Flag"
              title="暂无竞赛"
              description="当前没有可展示的竞赛，稍后再来查看新的开赛计划。"
            />

            <template v-else>
              <div class="workspace-directory-grid-head contest-directory-head">
                <span>竞赛</span>
                <span>状态</span>
                <span>模式</span>
                <span>开始时间</span>
                <span>结束时间</span>
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
                  <h3
                    class="contest-row-title workspace-directory-row-title"
                    :title="contest.title"
                  >
                    {{ contest.title }}
                  </h3>
                </div>

                <div class="contest-row-state">
                  <span
                    class="workspace-directory-status-pill contest-state-chip"
                    :style="{ '--contest-state-color': 'var(--contest-row-accent)' }"
                  >
                    {{ getStatusLabel(contest.status) }}
                  </span>
                </div>

                <div class="contest-row-mode">
                  <span
                    class="workspace-directory-status-pill workspace-directory-status-pill--muted contest-chip contest-chip-muted"
                    >{{ getModeLabel(contest.mode) }}</span
                  >
                </div>

                <div class="workspace-directory-compact-text contest-row-start-time">
                  <div class="contest-row-time-item contest-row-time-item-strong">
                    <CalendarRange class="h-3.5 w-3.5" />
                    <span>{{ formatTime(contest.starts_at) }}</span>
                  </div>
                </div>

                <div class="workspace-directory-compact-text contest-row-end-time">
                  <div class="contest-row-time-item">
                    <Clock3 class="h-3.5 w-3.5" />
                    <span>{{ formatTime(contest.ends_at) }}</span>
                  </div>
                </div>

                <div class="workspace-directory-row-btn contest-row-cta">
                  <span>{{ getContestActionLabel(contest.status) }}</span>
                  <ArrowRight class="h-4 w-4" />
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
}

:deep(.contest-empty-state) {
  margin-top: 0;
}

.contest-directory {
  --workspace-directory-grid-columns: minmax(0, 1.15fr) 7rem 7rem minmax(10.5rem, 0.85fr)
    minmax(10.5rem, 0.85fr) 8rem;
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

.contest-row-title {
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

.contest-row-mode {
  min-width: 0;
}

.contest-row-start-time,
.contest-row-end-time {
  min-width: 0;
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

@media (max-width: 1180px) {
  .contest-directory-head {
    display: none;
  }

  .contest-row {
    grid-template-columns: 1fr;
  }
}
</style>

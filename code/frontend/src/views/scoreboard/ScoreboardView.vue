<script setup lang="ts">
import { computed } from 'vue'

import type { ContestStatus } from '@/api/contracts'
import { useScoreboardView } from '@/composables/useScoreboardView'
import { getContestAccentColor, getModeLabel, getStatusLabel } from '@/utils/contest'

const { hasSections, loading, refresh, sections, selectionHint } = useScoreboardView()
const contestCount = computed(() => sections.value.length)
const teamCount = computed(() => sections.value.reduce((sum, section) => sum + section.rows.length, 0))
const hasPartialFailure = computed(() => sections.value.some((section) => section.error))

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

function sectionStyle(status: ContestStatus): Record<string, string> {
  return {
    '--scoreboard-accent': getContestAccentColor(status),
  }
}

function getRowClass(rank: number): string {
  if (rank === 1) return 'scoreboard-row scoreboard-row--top1'
  if (rank === 2) return 'scoreboard-row scoreboard-row--top2'
  if (rank === 3) return 'scoreboard-row scoreboard-row--top3'
  return 'scoreboard-row'
}

function getRankPillClass(rank: number): string[] {
  return [
    'scoreboard-rank-pill',
    rank === 1 ? 'scoreboard-rank-pill--top1' : '',
    rank === 2 ? 'scoreboard-rank-pill--top2' : '',
    rank === 3 ? 'scoreboard-rank-pill--top3' : '',
  ]
}
</script>

<template>
  <div class="scoreboard-view space-y-6">
    <header class="scoreboard-header">
      <div class="scoreboard-header__lead">
        <h1 class="scoreboard-header__title">
          排行榜
        </h1>
        <p class="scoreboard-header__hint">
          {{ selectionHint }}
        </p>
      </div>
      <div class="scoreboard-header__actions">
        <button
          type="button"
          class="scoreboard-refresh"
          @click="refresh"
        >
          刷新
        </button>
      </div>
    </header>

    <section class="scoreboard-summary">
      <div class="scoreboard-metric">
        <div class="scoreboard-metric__label">
          展示竞赛
        </div>
        <div class="scoreboard-metric__value">
          {{ contestCount }}
        </div>
      </div>
      <div class="scoreboard-metric">
        <div class="scoreboard-metric__label">
          排名队伍
        </div>
        <div class="scoreboard-metric__value">
          {{ teamCount }}
        </div>
      </div>
      <div
        v-if="hasPartialFailure"
        class="scoreboard-summary__warn"
      >
        部分竞赛加载失败
      </div>
    </section>

    <div
      v-if="loading && !hasSections"
      class="scoreboard-skeleton"
    >
      <div
        v-for="index in 3"
        :key="index"
        class="scoreboard-skeleton__row"
      />
    </div>

    <div
      v-else-if="!hasSections"
      class="scoreboard-empty"
    >
      暂无可查看的竞赛排行榜
    </div>

    <div
      v-else
      class="scoreboard-sections"
    >
      <article
        v-for="(section, index) in sections"
        :key="section.contest.id"
        data-testid="scoreboard-card"
        class="scoreboard-section"
        :style="sectionStyle(section.contest.status)"
      >
        <div class="scoreboard-section__line" />
        <div class="scoreboard-section__content">
          <header class="scoreboard-section__header">
            <div class="scoreboard-section__prefix">
              {{ String(index + 1).padStart(2, '0') }}
            </div>
            <div class="space-y-2">
              <div class="scoreboard-section__title-wrap">
                <h2 class="scoreboard-section__title">
                  {{ section.contest.title }}
                </h2>
                <span class="scoreboard-status-chip">{{ getStatusLabel(section.contest.status) }}</span>
                <span class="scoreboard-mode-chip">{{ getModeLabel(section.contest.mode) }}</span>
                <span
                  v-if="section.frozen"
                  class="scoreboard-frozen-chip"
                >排行榜已冻结</span>
              </div>
              <p class="scoreboard-section__window">
                {{ formatContestWindow(section.contest.starts_at, section.contest.ends_at) }}
              </p>
            </div>

            <div class="scoreboard-section__meta">
              {{ section.rows.length > 0 ? `展示前 ${section.rows.length} 支队伍` : '暂无排行队伍' }}
            </div>
          </header>

          <div
            v-if="section.error"
            class="scoreboard-note scoreboard-note--danger"
          >
            该竞赛排行榜加载失败，请稍后重试
          </div>

          <div
            v-else-if="section.rows.length === 0"
            class="scoreboard-note"
          >
            暂无排行榜数据
          </div>

          <div
            v-else
            class="scoreboard-table-wrap"
          >
            <table class="scoreboard-table">
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
                  v-for="item in section.rows"
                  :key="`${section.contest.id}-${item.team_id}`"
                  :class="getRowClass(item.rank)"
                >
                  <td class="scoreboard-cell--rank">
                    <span :class="getRankPillClass(item.rank)">
                      {{ item.rank }}
                    </span>
                  </td>
                  <td>{{ item.team_name }}</td>
                  <td class="scoreboard-cell--mono">
                    {{ item.score }}
                  </td>
                  <td>{{ item.solved_count }}</td>
                  <td class="scoreboard-cell--muted">
                    {{ formatDateTime(item.last_submission_at) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.scoreboard-view {
  --scoreboard-accent: var(--color-primary);
}

.scoreboard-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.8rem 1rem;
  padding-bottom: 0.95rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-primary) 24%, var(--color-border-default));
}

.scoreboard-header__lead {
  min-width: 0;
}

.scoreboard-header__actions {
  display: flex;
  justify-content: flex-end;
}

.scoreboard-header__title {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.scoreboard-header__hint {
  margin-top: 0.35rem;
  font-size: 0.86rem;
  line-height: 1.6;
  color: var(--color-text-muted);
}

.scoreboard-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.7rem 1.15rem;
  padding-bottom: 0.95rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 86%, var(--color-primary-soft));
}

.scoreboard-metric {
  display: grid;
  gap: 0.15rem;
  min-width: 7rem;
}

.scoreboard-metric__label {
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.scoreboard-metric__value {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 1.08rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.scoreboard-summary__warn {
  border-left: 2px solid color-mix(in srgb, var(--color-danger) 62%, var(--color-border-default));
  padding-left: 0.6rem;
  font-size: 0.78rem;
  color: color-mix(in srgb, var(--color-danger) 82%, var(--color-text-primary));
}

.scoreboard-refresh {
  border: 1px solid var(--color-border-default);
  border-radius: 6px;
  padding: 0.42rem 0.78rem;
  font-size: 0.82rem;
  color: var(--color-text-primary);
  transition: border-color 180ms ease, color 180ms ease;
}

.scoreboard-refresh:hover {
  border-color: color-mix(in srgb, var(--color-primary) 56%, var(--color-border-default));
  color: color-mix(in srgb, var(--color-primary) 72%, var(--color-text-primary));
}

.scoreboard-skeleton {
  display: grid;
  gap: 0.65rem;
}

.scoreboard-skeleton__row {
  height: 4.9rem;
  border-bottom: 1px solid var(--color-border-default);
  background: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-primary-soft)),
    transparent
  );
  background-size: 220% 100%;
  animation: scoreSkeletonMove 1.25s linear infinite;
}

.scoreboard-empty {
  padding: 1.1rem 0.2rem;
  border-left: 2px solid color-mix(in srgb, var(--color-text-muted) 50%, var(--color-border-default));
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.scoreboard-sections {
  display: grid;
  gap: 1.2rem;
}

.scoreboard-section {
  --scoreboard-accent: var(--color-primary);
  display: grid;
  grid-template-columns: 3px minmax(0, 1fr);
  gap: 0.8rem;
}

.scoreboard-section__line {
  width: 3px;
  background: linear-gradient(
    to bottom,
    color-mix(in srgb, var(--scoreboard-accent) 80%, #ffffff),
    color-mix(in srgb, var(--scoreboard-accent) 36%, transparent)
  );
}

.scoreboard-section__content {
  border-top: 1px solid var(--color-border-default);
  padding-top: 0.72rem;
}

.scoreboard-section__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem 1rem;
}

.scoreboard-section__prefix {
  min-width: 2rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 0.8rem;
  letter-spacing: 0.14em;
  color: color-mix(in srgb, var(--scoreboard-accent) 74%, var(--color-text-secondary));
}

.scoreboard-section__title-wrap {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
}

.scoreboard-section__title {
  font-size: 1.03rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.scoreboard-status-chip {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--scoreboard-accent) 34%, transparent);
  background: color-mix(in srgb, var(--scoreboard-accent) 10%, transparent);
  padding: 0.16rem 0.55rem;
  font-size: 0.72rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--scoreboard-accent) 84%, var(--color-text-primary));
}

.scoreboard-mode-chip {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, var(--scoreboard-accent));
  padding: 0.16rem 0.55rem;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--color-text-secondary);
}

.scoreboard-frozen-chip {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-warning) 36%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  padding: 0.16rem 0.55rem;
  font-size: 0.72rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--color-warning) 84%, var(--color-text-primary));
}

.scoreboard-section__window {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.scoreboard-section__meta {
  font-size: 0.79rem;
  color: var(--color-text-secondary);
}

.scoreboard-note {
  margin-top: 0.72rem;
  padding: 0.5rem 0.72rem;
  border-left: 2px solid color-mix(in srgb, var(--color-text-muted) 46%, var(--color-border-default));
  font-size: 0.83rem;
  color: var(--color-text-secondary);
}

.scoreboard-note--danger {
  border-left-color: color-mix(in srgb, var(--color-danger) 62%, var(--color-border-default));
  color: color-mix(in srgb, var(--color-danger) 86%, var(--color-text-primary));
}

.scoreboard-table-wrap {
  margin-top: 0.62rem;
  overflow-x: auto;
}

.scoreboard-table {
  min-width: 100%;
  border-collapse: collapse;
}

.scoreboard-table th {
  position: sticky;
  top: 0;
  z-index: 1;
  padding: 0.55rem 0.8rem;
  border-bottom: 1px solid var(--color-border-default);
  background: color-mix(in srgb, var(--color-bg-canvas) 92%, transparent);
  text-align: left;
  font-size: 0.71rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.scoreboard-row {
  border-bottom: 1px solid var(--color-border-subtle);
}

.scoreboard-row td {
  padding: 0.6rem 0.8rem;
  font-size: 0.84rem;
  color: var(--color-text-primary);
}

.scoreboard-row:hover {
  background: color-mix(in srgb, var(--scoreboard-accent) 5%, transparent);
}

.scoreboard-row--top1 {
  border-left: 2px solid #d97706;
}

.scoreboard-row--top2 {
  border-left: 2px solid #64748b;
}

.scoreboard-row--top3 {
  border-left: 2px solid #c2410c;
}

.scoreboard-cell--rank,
.scoreboard-cell--mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
}

.scoreboard-cell--rank {
  width: 5.2rem;
}

.scoreboard-cell--muted {
  color: var(--color-text-secondary) !important;
}

.scoreboard-rank-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.9rem;
  padding: 0.14rem 0.4rem;
  border: 1px solid var(--color-border-default);
  border-radius: 0.35rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.scoreboard-rank-pill--top1 {
  border-color: color-mix(in srgb, #d97706 56%, var(--color-border-default));
  color: #d97706;
}

.scoreboard-rank-pill--top2 {
  border-color: color-mix(in srgb, #64748b 56%, var(--color-border-default));
  color: #64748b;
}

.scoreboard-rank-pill--top3 {
  border-color: color-mix(in srgb, #c2410c 56%, var(--color-border-default));
  color: #c2410c;
}

@media (max-width: 720px) {
  .scoreboard-header__title {
    font-size: 1.35rem;
  }

  .scoreboard-section {
    grid-template-columns: 2px minmax(0, 1fr);
    gap: 0.7rem;
  }

  .scoreboard-section__prefix {
    width: 100%;
    min-width: 100%;
  }
}

@keyframes scoreSkeletonMove {
  from {
    background-position-x: 0%;
  }
  to {
    background-position-x: 220%;
  }
}
</style>

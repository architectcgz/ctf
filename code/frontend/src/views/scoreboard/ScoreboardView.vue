<script setup lang="ts">
import { computed } from 'vue'
import { BarChart2, Shield, Users } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import type { ContestStatus } from '@/api/contracts'
import { useScoreboardView } from '@/composables/useScoreboardView'
import { getContestAccentColor, getModeLabel, getStatusLabel } from '@/utils/contest'

const { hasSections, loading, refresh, refreshContestScoreboard, sections, selectionHint } =
  useScoreboardView()
const contestCount = computed(() => sections.value.length)
const teamCount = computed(() =>
  sections.value.reduce((sum, section) => sum + section.rows.length, 0)
)
const frozenCount = computed(() => sections.value.filter((section) => section.frozen).length)
const failureCount = computed(() => sections.value.filter((section) => section.error).length)
const hasPartialFailure = computed(() => sections.value.some((section) => section.error))
const emptyTitle = computed(() =>
  selectionHint.value.includes('失败') ? '排行榜加载失败' : '暂无可查看的竞赛排行榜'
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

function supportsRealtime(status: ContestStatus): boolean {
  return status === 'running' || status === 'frozen'
}
</script>

<template>
  <section
    class="journal-shell space-y-6 journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
      <div>
        <div class="journal-eyebrow">Scoreboard</div>
        <h2
          class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
        >
          排行榜
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          {{ selectionHint }}
        </p>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
          <BarChart2 class="h-5 w-5 text-[var(--journal-accent)]" />
          总览
        </div>
        <div class="mt-4 grid grid-cols-2 gap-3">
          <div class="journal-note">
            <div class="journal-note-label">展示竞赛</div>
            <div class="journal-note-value">{{ contestCount }}</div>
            <div class="journal-note-helper">当前被纳入排行榜视图的竞赛数</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">参赛队伍</div>
            <div class="journal-note-value">{{ teamCount }}</div>
            <div class="journal-note-helper">所有已拉取分区内的队伍总数</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">冻结竞赛</div>
            <div class="journal-note-value">{{ frozenCount }}</div>
            <div class="journal-note-helper">处于冻结排行显示状态的场次</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">异常分区</div>
            <div class="journal-note-value">{{ failureCount }}</div>
            <div class="journal-note-helper">仅统计当前已检测到的加载异常</div>
          </div>
        </div>
        <div
          v-if="hasPartialFailure"
          class="mt-3 rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-medium text-amber-700"
        >
          部分竞赛加载失败
        </div>
      </article>
    </div>
    <div class="scoreboard-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
      <div v-if="loading && !hasSections" class="space-y-3 py-1">
        <div
          v-for="i in 3"
          :key="i"
          class="h-32 rounded-[18px] animate-pulse"
          style="background: rgba(226, 232, 240, 0.5)"
        />
      </div>

      <AppEmpty
        v-else-if="!hasSections"
        class="scoreboard-empty-state"
        icon="Trophy"
        :title="emptyTitle"
        :description="selectionHint"
      >
        <template #action>
          <button type="button" class="sb-refresh-btn" @click="refresh">重新加载</button>
        </template>
      </AppEmpty>

      <div v-else class="scoreboard-sections mt-5">
        <article
          v-for="(section, index) in sections"
          :key="section.contest.id"
          data-testid="scoreboard-card"
          class="journal-log scoreboard-card px-6 py-5"
          :style="{
            ...sectionAccentStyle(section.contest.status),
            borderLeftWidth: '3px',
            borderLeftColor: 'var(--scoreboard-accent)',
          }"
        >
          <ScoreboardRealtimeBridge
            v-if="supportsRealtime(section.contest.status)"
            :contest-id="section.contest.id"
            @updated="refreshContestScoreboard(section.contest.id)"
          />
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="min-w-0">
              <div class="flex flex-wrap gap-2 items-center">
                <span class="sb-index">{{ String(index + 1).padStart(2, '0') }}</span>
                <span class="sb-status-chip">{{ getStatusLabel(section.contest.status) }}</span>
                <span class="sb-mode-chip">{{ getModeLabel(section.contest.mode) }}</span>
                <span v-if="section.frozen" class="sb-frozen-chip">
                  <Shield class="h-3 w-3" /> 已冻结
                </span>
              </div>
              <h3 class="mt-2 font-semibold text-lg text-[var(--journal-ink)] leading-snug">
                {{ section.contest.title }}
              </h3>
              <p class="mt-1 text-xs text-[var(--journal-muted)]">
                {{ formatContestWindow(section.contest.starts_at, section.contest.ends_at) }}
              </p>
            </div>
            <div
              class="flex items-center gap-1.5 text-xs text-[var(--journal-muted)] shrink-0 pt-1"
            >
              <Users class="h-3.5 w-3.5" />
              {{
                section.rows.length > 0 ? `展示前 ${section.rows.length} 支队伍` : '暂无排行队伍'
              }}
            </div>
          </div>

          <div class="scoreboard-card-divider" />

          <div
            v-if="section.error"
            class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
          >
            该竞赛排行榜加载失败，请稍后重试
          </div>

          <div
            v-else-if="section.rows.length === 0"
            class="rounded-[18px] border border-[var(--journal-border)] border-dashed px-4 py-6 text-sm text-[var(--journal-muted)]"
          >
            暂无排行榜数据
          </div>

          <div v-else class="scoreboard-table-shell overflow-x-auto">
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
                  v-for="item in section.rows"
                  :key="`${section.contest.id}-${item.team_id}`"
                  :class="getRowClass(item.rank)"
                >
                  <td class="sb-cell--rank">
                    <span :class="getRankPillClass(item.rank)">{{ item.rank }}</span>
                  </td>
                  <td>{{ item.team_name }}</td>
                  <td class="sb-cell--mono">{{ item.score }}</td>
                  <td>{{ item.solved_count }}</td>
                  <td class="sb-cell--muted">{{ formatDateTime(item.last_submission_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-color: var(--journal-border);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.journal-log {
  background: var(--journal-surface);
  border-color: var(--journal-border);
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease;
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.625rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.scoreboard-board {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.scoreboard-sections {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  overflow: hidden;
}

.sb-refresh-btn {
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.38rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--journal-ink);
  cursor: pointer;
  transition: border-color 150ms ease;
}

.sb-refresh-btn:hover {
  border-color: var(--journal-accent);
}

.scoreboard-card {
  border-bottom: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base)));
}

.scoreboard-card:last-child {
  border-bottom: 0;
}

.scoreboard-card-divider {
  margin: 1rem 0 1.1rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.scoreboard-table-shell {
  border-radius: 1rem;
  border: 1px solid rgba(226, 232, 240, 0.72);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

:deep(.scoreboard-empty-state) {
  border-top-style: dashed;
  border-bottom-style: dashed;
  border-top-color: rgba(148, 163, 184, 0.58);
  border-bottom-color: rgba(148, 163, 184, 0.58);
}

.sb-index {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--journal-muted);
  font-variant-numeric: tabular-nums;
}

.sb-status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid
    color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 30%, transparent);
  background: color-mix(in srgb, var(--scoreboard-accent, var(--journal-accent)) 10%, transparent);
  padding: 0.18rem 0.6rem;
  font-size: 0.72rem;
  font-weight: 700;
  color: color-mix(
    in srgb,
    var(--scoreboard-accent, var(--journal-accent)) 80%,
    var(--journal-ink)
  );
}

.sb-mode-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid var(--journal-border);
  background: rgba(226, 232, 240, 0.4);
  padding: 0.18rem 0.6rem;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.sb-frozen-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  border-radius: 999px;
  border: 1px solid rgba(245, 158, 11, 0.3);
  background: rgba(245, 158, 11, 0.1);
  padding: 0.18rem 0.6rem;
  font-size: 0.72rem;
  font-weight: 700;
  color: #b45309;
}

.sb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.sb-table th {
  padding: 0.55rem 0.75rem;
  text-align: left;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--journal-muted);
  border-bottom: 1px solid var(--journal-border);
}

.sb-row td {
  padding: 0.65rem 0.75rem;
  color: var(--journal-ink);
  border-bottom: 1px solid rgba(226, 232, 240, 0.4);
}

.sb-row--top1 td {
  background: rgba(234, 179, 8, 0.06);
}

.sb-row--top2 td {
  background: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 34%, transparent);
}

.sb-row--top3 td {
  background: rgba(194, 65, 12, 0.06);
}

.sb-cell--rank {
  width: 4rem;
}

.sb-cell--mono {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: color-mix(
    in srgb,
    var(--scoreboard-accent, var(--journal-accent)) 90%,
    var(--journal-ink)
  );
}

.sb-cell--muted {
  color: var(--journal-muted);
  font-size: 0.8rem;
}

.sb-rank-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.8rem;
  border-radius: 999px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface-subtle);
  padding: 0.1rem 0.5rem;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--journal-muted);
}

.sb-rank-pill--top1 {
  border-color: rgba(234, 179, 8, 0.5);
  background: rgba(234, 179, 8, 0.1);
  color: #92400e;
}

.sb-rank-pill--top2 {
  border-color: rgba(148, 163, 184, 0.5);
  background: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 34%, transparent);
  color: #475569;
}

.sb-rank-pill--top3 {
  border-color: rgba(194, 65, 12, 0.4);
  background: rgba(194, 65, 12, 0.08);
  color: #c2410c;
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(99, 102, 241, 0.14), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .sb-frozen-chip {
  border-color: rgba(245, 158, 11, 0.25);
  background: rgba(245, 158, 11, 0.08);
  color: #fbbf24;
}
</style>

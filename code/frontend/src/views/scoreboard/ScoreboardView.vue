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

function getCardDescription(
  status: ContestStatus,
  frozen: boolean,
  hasError: boolean,
  rowCount: number
): string {
  if (hasError) {
    return '该竞赛排行榜暂时不可用，可稍后重新加载。'
  }

  if (rowCount === 0) {
    return '当前还没有可展示的队伍成绩，提交后会自动进入榜单。'
  }

  if (frozen || status === 'frozen') {
    return '封榜阶段仅展示冻结前排名，解封后会同步最终成绩。'
  }

  if (status === 'running') {
    return '进行中竞赛支持实时刷新，提交后榜单会自动更新。'
  }

  return '历史竞赛展示最终成绩，可用于复盘队伍解题表现。'
}
</script>

<template>
  <section
    class="journal-shell journal-shell-user journal-eyebrow-text journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="scoreboard-page">
      <header class="scoreboard-topbar">
        <div class="scoreboard-heading">
          <div class="journal-eyebrow">Scoreboard</div>
          <h1 class="scoreboard-title">排行榜</h1>
          <p class="scoreboard-subtitle">{{ selectionHint }}</p>
        </div>
      </header>

      <section class="scoreboard-summary">
        <div class="scoreboard-summary-title">
          <BarChart2 class="h-4 w-4" />
          <span>当前排行概况</span>
        </div>
        <div class="scoreboard-summary-grid">
          <div class="scoreboard-summary-item">
            <div class="scoreboard-summary-label">展示竞赛</div>
            <div class="scoreboard-summary-value">{{ contestCount }}</div>
            <div class="scoreboard-summary-helper">当前可查看排行的竞赛总数</div>
          </div>
          <div class="scoreboard-summary-item">
            <div class="scoreboard-summary-label">参赛队伍</div>
            <div class="scoreboard-summary-value">{{ teamCount }}</div>
            <div class="scoreboard-summary-helper">已进入榜单统计的队伍规模</div>
          </div>
          <div class="scoreboard-summary-item">
            <div class="scoreboard-summary-label">冻结竞赛</div>
            <div class="scoreboard-summary-value">{{ frozenCount }}</div>
            <div class="scoreboard-summary-helper">当前处于封榜阶段的竞赛数量</div>
          </div>
          <div class="scoreboard-summary-item">
            <div class="scoreboard-summary-label">异常分区</div>
            <div class="scoreboard-summary-value">{{ failureCount }}</div>
            <div class="scoreboard-summary-helper">排行榜加载异常的竞赛分区</div>
          </div>
        </div>
        <div v-if="hasPartialFailure" class="scoreboard-inline-note">部分竞赛加载失败</div>
      </section>

      <div v-if="loading && !hasSections" class="scoreboard-loading">
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
          <button type="button" class="scoreboard-btn" @click="refresh">重新加载</button>
        </template>
      </AppEmpty>

      <section v-else class="scoreboard-directory" aria-label="排行榜列表">
        <div class="scoreboard-directory-top">
          <h2 class="scoreboard-directory-title">竞赛排行列表</h2>
          <div class="scoreboard-directory-meta">按竞赛开始时间倒序展示排行榜</div>
        </div>

        <div class="scoreboard-sections">
          <article
            v-for="(section, index) in sections"
            :key="section.contest.id"
            data-testid="scoreboard-card"
            class="scoreboard-card"
            :style="sectionAccentStyle(section.contest.status)"
          >
            <ScoreboardRealtimeBridge
              v-if="supportsRealtime(section.contest.status)"
              :contest-id="section.contest.id"
              @updated="refreshContestScoreboard(section.contest.id)"
            />
            <div class="scoreboard-card-header">
              <div class="scoreboard-card-main">
                <div class="scoreboard-card-chips">
                  <span class="sb-index">{{ String(index + 1).padStart(2, '0') }}</span>
                  <span class="sb-status-chip">{{ getStatusLabel(section.contest.status) }}</span>
                  <span class="sb-mode-chip">{{ getModeLabel(section.contest.mode) }}</span>
                  <span v-if="section.frozen" class="sb-frozen-chip">
                    <Shield class="h-3 w-3" /> 已冻结
                  </span>
                </div>
                <h3 class="scoreboard-card-title">{{ section.contest.title }}</h3>
                <p class="scoreboard-card-time">
                  {{ formatContestWindow(section.contest.starts_at, section.contest.ends_at) }}
                </p>
                <p class="scoreboard-card-description">
                  {{
                    getCardDescription(
                      section.contest.status,
                      section.frozen,
                      section.error,
                      section.rows.length
                    )
                  }}
                </p>
              </div>
              <div class="scoreboard-card-meta">
                <Users class="h-3.5 w-3.5" />
                {{
                  section.rows.length > 0 ? `展示前 ${section.rows.length} 支队伍` : '暂无排行队伍'
                }}
              </div>
            </div>

            <div class="scoreboard-card-divider" />

            <div v-if="section.error" class="scoreboard-inline-note scoreboard-inline-note-danger">
              该竞赛排行榜加载失败，请稍后重试
            </div>

            <div v-else-if="section.rows.length === 0" class="scoreboard-inline-note">
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
      </section>
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

.scoreboard-subtitle {
  max-width: 760px;
}

.scoreboard-inline-note {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 8px;
  font-size: 12px;
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
  font-size: 18px;
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
}

.scoreboard-card-time,
.scoreboard-card-meta {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--journal-muted);
}

.scoreboard-card-description {
  margin-top: 8px;
  max-width: 700px;
  font-size: 13px;
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
  font-size: 12px;
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
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  text-align: left;
  color: var(--journal-muted);
}

.sb-row td {
  padding: 14px 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  font-size: 14px;
  color: var(--journal-ink);
}

.sb-row--top1 td,
.sb-rank-pill--top1 {
  color: color-mix(in srgb, #b45309 84%, var(--journal-ink));
}

.sb-row--top2 td,
.sb-rank-pill--top2 {
  color: color-mix(in srgb, #475569 80%, var(--journal-ink));
}

.sb-row--top3 td,
.sb-rank-pill--top3 {
  color: color-mix(in srgb, #92400e 80%, var(--journal-ink));
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

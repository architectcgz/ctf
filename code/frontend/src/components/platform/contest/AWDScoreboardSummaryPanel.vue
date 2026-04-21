<script setup lang="ts">
import { Trophy, TrendingUp, Users } from 'lucide-vue-next'
import type { AWDScoreboardSummaryPanelProps } from '@/components/platform/contest/awdInspector.types'

defineProps<AWDScoreboardSummaryPanelProps>()
</script>

<template>
  <div class="studio-scoreboard-stack">
    <!-- 1. Rank Context HUD (Subtle) -->
    <div class="rank-context">
      <div class="context-item">
        <Trophy class="h-4 w-4 text-amber-500" />
        <span>全场总分排名</span>
      </div>
      <div class="context-divider" />
      <div class="context-item">
        <Users class="h-4 w-4 text-slate-400" />
        <span>活跃参赛队伍: {{ scoreboardRows.length }}</span>
      </div>
      <div class="context-divider" />
      <div
        v-if="scoreboardFrozen"
        class="context-item"
      >
        <div class="frozen-dot" />
        <span class="text-orange-500 font-bold">排行榜已冻结</span>
      </div>
    </div>

    <!-- 2. The Grand Leaderboard -->
    <div class="log-table-wrap">
      <table class="studio-table">
        <thead>
          <tr>
            <th class="w-24">
              当前排名
            </th>
            <th>队伍/选手</th>
            <th class="text-right">
              累积总分
            </th>
            <th class="text-right">
              解题进度
            </th>
            <th class="text-right">
              最后命中时间
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in scoreboardRows"
            :key="item.team_id"
            class="studio-row"
          >
            <td>
              <div
                class="rank-badge"
                :class="`rank-${item.rank}`"
              >
                #{{ String(item.rank).padStart(2, '0') }}
              </div>
            </td>
            <td>
              <div class="team-cell">
                <span class="team-name">{{ item.team_name }}</span>
                <span class="team-id">ID: {{ item.team_id }}</span>
              </div>
            </td>
            <td class="text-right font-mono font-black text-blue-600 text-lg">
              {{ formatScore(item.score) }}
            </td>
            <td class="text-right font-mono text-slate-500">
              <span class="font-bold text-slate-900">{{ item.solved_count }}</span> <small>SOLVED</small>
            </td>
            <td class="text-right text-[11px] text-slate-400">
              {{ formatDateTime(item.last_submission_at).split(' ')[1] || '--' }}
            </td>
          </tr>
          <tr v-if="scoreboardRows.length === 0">
            <td
              colspan="5"
              class="py-24 text-center"
            >
              <div class="flex flex-col items-center gap-3 opacity-20">
                <Trophy class="h-12 w-12" />
                <p class="text-sm font-bold">
                  暂无积分记录，比赛尚未产生得分
                </p>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.studio-scoreboard-stack { display: flex; flex-direction: column; gap: var(--space-6); }

/* Rank HUD */
.rank-context { display: flex; align-items: center; gap: var(--space-6); padding: var(--space-2) 0; }
.context-item { display: flex; align-items: center; gap: var(--space-2-5); font-size: var(--font-size-12); font-weight: 700; color: var(--color-text-secondary); }
.context-divider { width: 1px; height: 1rem; background: var(--color-border-default); }

.frozen-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--color-warning); animation: blink 1.5s infinite; }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }

/* Leaderboard Table */
.log-table-wrap { width: 100%; overflow: hidden; background: var(--color-bg-surface); }
.studio-table { width: 100%; border-collapse: collapse; }
.studio-table th { background: var(--color-bg-elevated); padding: 0.75rem 1rem; text-align: left; font-size: var(--font-size-10); font-weight: 800; text-transform: uppercase; color: var(--color-text-muted); border-top: 1px solid var(--color-border-default); border-bottom: 1px solid var(--color-border-default); }
.studio-table td { padding: 1.15rem 1rem; border-bottom: 1px solid var(--color-border-subtle); }

.studio-row:hover { background: var(--color-bg-elevated); }

.rank-badge { font-family: var(--font-family-mono); font-size: var(--font-size-14); font-weight: 900; color: var(--color-text-muted); }
.rank-1 { color: var(--color-warning); font-size: var(--font-size-18); }
.rank-2 { color: var(--color-text-muted); }
.rank-3 { color: color-mix(in srgb, var(--color-warning) 80%, black); }

.team-cell { display: flex; flex-direction: column; gap: 0.15rem; }
.team-name { font-size: var(--font-size-14); font-weight: 800; color: var(--color-text-primary); }
.team-id { font-size: var(--font-size-10); color: var(--color-text-muted); font-weight: 600; }

.text-right { text-align: right; }
.w-24 { width: 6rem; }
</style>

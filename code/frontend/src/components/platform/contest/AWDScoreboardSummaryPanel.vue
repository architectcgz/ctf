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
.studio-scoreboard-stack { display: flex; flex-direction: column; gap: 1.5rem; }

/* Rank HUD */
.rank-context { display: flex; align-items: center; gap: 1.5rem; padding: 0.5rem 0; }
.context-item { display: flex; align-items: center; gap: 0.65rem; font-size: 12px; font-weight: 700; color: #64748b; }
.context-divider { width: 1px; height: 1rem; background: #e2e8f0; }

.frozen-dot { width: 6px; height: 6px; border-radius: 50%; background: #f97316; animation: blink 1.5s infinite; }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }

/* Leaderboard Table */
.log-table-wrap { width: 100%; overflow: hidden; background: white; }
.studio-table { width: 100%; border-collapse: collapse; }
.studio-table th { background: #f8fafc; padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; border-top: 1px solid #e2e8f0; border-bottom: 1px solid #e2e8f0; }
.studio-table td { padding: 1.15rem 1rem; border-bottom: 1px solid #f1f5f9; }

.studio-row:hover { background: #f8fafc; }

.rank-badge { font-family: var(--font-family-mono); font-size: 14px; font-weight: 900; color: #94a3b8; }
.rank-1 { color: #f59e0b; font-size: 18px; }
.rank-2 { color: #94a3b8; }
.rank-3 { color: #b45309; }

.team-cell { display: flex; flex-direction: column; gap: 0.15rem; }
.team-name { font-size: 14px; font-weight: 800; color: #0f172a; }
.team-id { font-size: 10px; color: #94a3b8; font-weight: 600; }

.text-right { text-align: right; }
.w-24 { width: 6rem; }
</style>

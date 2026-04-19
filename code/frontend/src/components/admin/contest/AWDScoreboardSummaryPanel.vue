<script setup lang="ts">
import type { AWDScoreboardSummaryPanelProps } from '@/components/admin/contest/awdInspector.types'

defineProps<AWDScoreboardSummaryPanelProps>()
</script>

<template>
  <div class="space-y-6">
    <div class="overflow-hidden rounded-2xl border border-border">
      <div class="awd-scoreboard-head">
        <div class="awd-section-title">实时排行榜</div>
        <span v-if="scoreboardFrozen" class="awd-scoreboard-frozen-pill">排行榜已冻结</span>
      </div>
      <table class="min-w-full divide-y divide-border">
        <thead class="awd-table-head">
          <tr>
            <th class="px-4 py-3">排名</th>
            <th class="px-4 py-3">队伍</th>
            <th class="px-4 py-3">得分</th>
            <th class="px-4 py-3">解题数</th>
            <th class="px-4 py-3">最近得分</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border bg-surface/70">
          <tr v-for="item in scoreboardRows" :key="item.team_id">
            <td class="awd-table-cell awd-table-cell--primary awd-scoreboard-rank">#{{ item.rank }}</td>
            <td class="awd-table-cell awd-table-cell--primary">{{ item.team_name }}</td>
            <td class="awd-table-cell awd-table-cell--primary">{{ formatScore(item.score) }}</td>
            <td class="awd-table-cell awd-table-cell--muted">{{ item.solved_count }}</td>
            <td class="awd-table-cell awd-table-cell--muted">
              {{ formatDateTime(item.last_submission_at) }}
            </td>
          </tr>
          <tr v-if="scoreboardRows.length === 0">
            <td colspan="5" class="awd-empty-row">当前赛事还没有排行榜数据。</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="overflow-hidden rounded-2xl border border-border">
      <div class="awd-summary-head">本轮汇总</div>
      <table class="min-w-full divide-y divide-border">
        <thead class="awd-table-head">
          <tr>
            <th class="px-4 py-3">队伍</th>
            <th class="px-4 py-3">总分</th>
            <th class="px-4 py-3">SLA / 攻击 / 防守</th>
            <th class="px-4 py-3">服务状态</th>
            <th class="px-4 py-3">被攻击情况</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border bg-surface/70">
          <tr v-for="item in summary?.items || []" :key="item.team_id">
            <td class="awd-table-cell awd-table-cell--primary awd-summary-team">{{ item.team_name }}</td>
            <td class="awd-table-cell awd-table-cell--primary">{{ item.total_score }}</td>
            <td class="awd-table-cell awd-table-cell--secondary">
              SLA {{ item.sla_score ?? 0 }} / 攻击 {{ item.attack_score }} / 防守
              {{ item.defense_score }}
            </td>
            <td class="awd-table-cell awd-table-cell--secondary">
              正常 {{ item.service_up_count }} / 下线 {{ item.service_down_count }} / 失陷
              {{ item.service_compromised_count }}
            </td>
            <td class="awd-table-cell awd-table-cell--secondary">
              攻破 {{ item.successful_breach_count }} 次，攻击方
              {{ item.unique_attackers_against }} 支
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.awd-scoreboard-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border-bottom: 1px solid var(--color-border-default);
  background: color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-bg-base));
  padding: var(--space-3) var(--space-4);
}

.awd-summary-head {
  border-bottom: 1px solid var(--color-border-default);
  background: color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-bg-base));
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-primary);
}

.awd-section-title {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-primary);
}

.awd-scoreboard-frozen-pill {
  display: inline-flex;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-warning) 20%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  padding: 0.25rem 0.75rem;
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-warning);
}

.awd-table-head {
  background: color-mix(in srgb, var(--color-surface-alt) 40%, transparent);
  text-align: left;
  font-size: var(--font-size-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  color: var(--color-text-muted);
}

.awd-table-cell {
  padding: var(--space-4);
  font-size: var(--font-size-sm);
}

.awd-table-cell--primary {
  color: var(--color-text-primary);
}

.awd-table-cell--secondary {
  color: var(--color-text-secondary);
}

.awd-table-cell--muted {
  color: var(--color-text-muted);
}

.awd-scoreboard-rank {
  font-weight: 600;
}

.awd-summary-team {
  font-weight: 500;
}

.awd-empty-row {
  padding: var(--space-8) var(--space-4);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}
</style>

<script setup lang="ts">
import { Trophy } from 'lucide-vue-next'

import type { ScoreboardRow } from '@/api/contracts'
import { formatProjectorScore } from '@/components/platform/contest/projector/contestProjectorFormatters'

defineProps<{
  topThreeRows: ScoreboardRow[]
  leaderboardRows: ScoreboardRow[]
  scoreboardRowsLength: number
}>()
</script>

<template>
  <section class="leaderboard-panel">
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          排行榜
        </div>
        <h3>实时排名</h3>
      </div>
      <Trophy class="panel-icon panel-icon--accent" />
    </header>

    <div class="podium-grid">
      <article
        v-for="row in topThreeRows"
        :key="row.team_id"
        class="podium-card"
        :class="`rank-${row.rank}`"
      >
        <span class="podium-rank">#{{ row.rank }}</span>
        <strong>{{ row.team_name }}</strong>
        <span>{{ formatProjectorScore(row.score) }} pts</span>
      </article>
    </div>

    <div class="leaderboard-list">
      <div
        v-for="row in leaderboardRows"
        :key="row.team_id"
        class="leaderboard-row"
        :class="{ 'leaderboard-row--top': row.rank <= 3 }"
      >
        <span class="leaderboard-rank">#{{ row.rank }}</span>
        <strong>{{ row.team_name }}</strong>
        <span class="leaderboard-score">{{ formatProjectorScore(row.score) }}</span>
      </div>
      <div
        v-if="scoreboardRowsLength === 0"
        class="panel-empty"
      >
        暂无得分记录
      </div>
    </div>
  </section>
</template>

<style scoped>
.leaderboard-panel,
.leaderboard-list {
  display: flex;
  flex-direction: column;
}

.leaderboard-panel {
  gap: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 86%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 56%, transparent);
  padding: var(--space-4);
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.panel-head h3 {
  margin: var(--space-1) 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-00);
  font-weight: 900;
}

.projector-overline {
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.panel-icon {
  width: var(--space-5);
  height: var(--space-5);
}

.panel-icon--accent {
  color: var(--color-warning);
}

.podium-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.podium-card {
  display: flex;
  min-height: 8rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--journal-accent) 9%, transparent);
  color: var(--color-text-secondary);
  text-align: center;
}

.podium-card.rank-1 {
  min-height: 9rem;
  background: color-mix(in srgb, var(--color-warning) 18%, transparent);
}

.podium-card strong {
  max-width: 90%;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-1-00);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.podium-rank {
  color: var(--color-warning);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-35);
  font-weight: 900;
}

.leaderboard-list {
  gap: var(--space-2);
  margin-top: var(--space-4);
}

.leaderboard-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  border: 1px solid transparent;
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 44%, transparent);
  padding: var(--space-2-5) var(--space-3);
  color: var(--color-text-secondary);
}

.leaderboard-row--top {
  border-color: color-mix(in srgb, var(--color-warning) 24%, transparent);
}

.leaderboard-rank,
.leaderboard-score {
  font-family: var(--font-family-mono);
  font-weight: 900;
}

.leaderboard-rank {
  color: var(--color-warning);
}

.leaderboard-row strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-score {
  color: var(--journal-ink);
}

.panel-empty {
  padding: var(--space-4) 0;
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
}

@media (max-width: 900px) {
  .podium-grid {
    grid-template-columns: 1fr;
  }
}
</style>

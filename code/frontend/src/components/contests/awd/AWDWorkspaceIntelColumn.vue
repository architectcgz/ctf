<script setup lang="ts">
import { computed } from 'vue'
import { BarChart3, History } from 'lucide-vue-next'

import type {
  ContestAWDWorkspaceEventDirection,
  ContestAWDWorkspaceRecentEventData,
  ScoreboardRow,
} from '@/api/contracts'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  scoreboardRows: ScoreboardRow[]
  myTeamId?: string
  recentEvents: ContestAWDWorkspaceRecentEventData[]
  getChallengeTitleForEvent: (event: {
    service_id?: string
    awd_challenge_id: string
  }) => string
  eventDirectionLabel: (direction: ContestAWDWorkspaceEventDirection) => string
  eventResultLabel: (success: boolean) => string
  formatServiceRef: (serviceId?: string) => string
}>()

const topScoreboardRows = computed(() => props.scoreboardRows.slice(0, 10))
</script>

<template>
  <aside class="war-room-col column-intel">
    <section class="ops-panel">
      <header class="ops-panel__header">
        <BarChart3 class="ops-panel__icon ops-panel__icon--accent h-4 w-4" />
        <h3 class="ops-panel__title">战场情报</h3>
      </header>
      <div class="ops-panel__content custom-scrollbar">
        <div
          v-for="item in topScoreboardRows"
          :key="item.team_id"
          class="intel-row"
          :class="{ 'is-me': item.team_id === myTeamId }"
        >
          <span class="intel-rank">#{{ item.rank }}</span>
          <span class="intel-name truncate">{{ item.team_name }}</span>
          <span class="intel-score font-mono">{{ item.score }}</span>
        </div>
      </div>
    </section>

    <section class="ops-panel">
      <header class="ops-panel__header">
        <History class="ops-panel__icon ops-panel__icon--history h-4 w-4" />
        <h3 class="ops-panel__title">最近战报</h3>
      </header>
      <div class="ops-panel__content custom-scrollbar">
        <div
          v-for="event in recentEvents"
          :key="event.id"
          class="feedback-item"
          :class="event.direction"
        >
          <div class="feedback-item__meta">
            <span>{{ eventDirectionLabel(event.direction) }}</span>
            <span>{{ formatTime(event.created_at) }}</span>
          </div>
          <div class="feedback-item__title">
            <span>{{ event.peer_team_name }} / </span>
            <span data-testid="awd-feedback-challenge-title">{{
              getChallengeTitleForEvent(event)
            }}</span>
          </div>
          <div class="feedback-ref">
            {{ formatServiceRef(event.service_id) }}
          </div>
          <div class="feedback-item__result">
            <span :class="event.is_success ? 'feedback-result--success' : 'feedback-result--muted'">
              {{ eventResultLabel(event.is_success) }}
            </span>
            <span class="feedback-score-gain">+{{ event.score_gained }}</span>
          </div>
        </div>
        <div v-if="recentEvents.length === 0" class="panel-note">暂无最近战报。</div>
      </div>
    </section>
  </aside>
</template>

<style scoped>
.column-intel {
  grid-area: intel;
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  gap: var(--space-5);
}

.ops-panel {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  box-shadow: var(--color-shadow-soft);
  display: flex;
  flex-direction: column;
  min-height: 18rem;
  overflow: hidden;
}

.ops-panel__header {
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
}

.ops-panel__title {
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 900;
  letter-spacing: 0.15em;
  margin: 0;
}

.ops-panel__icon--accent {
  color: var(--color-primary);
}

.ops-panel__icon--history {
  color: var(--color-text-secondary);
}

.ops-panel__content {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem;
}

.intel-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--color-border-subtle);
  font-size: var(--font-size-13);
}

.intel-row.is-me {
  color: var(--color-primary);
}

.intel-rank {
  color: var(--color-text-muted);
  font-weight: 900;
  width: 2rem;
}

.intel-name {
  color: var(--color-text-primary);
  flex: 1;
  font-weight: 700;
}

.intel-score {
  color: var(--color-text-primary);
  font-weight: 900;
}

.is-me .intel-name,
.is-me .intel-score {
  color: var(--color-primary);
}

.feedback-item {
  background: var(--color-bg-elevated);
  border-left: 2px solid var(--color-border-default);
  border-radius: 0.5rem;
  margin-bottom: 0.75rem;
  padding: 0.75rem 1rem;
}

.feedback-item.attack_out {
  border-left-color: var(--color-danger);
}

.feedback-item.attack_in {
  border-left-color: var(--color-warning);
}

.feedback-item__meta,
.feedback-item__result {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.feedback-item__title {
  font-size: var(--font-size-12);
  margin-top: var(--space-1);
}

.feedback-ref {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.04em;
  margin-top: 0.35rem;
}

.feedback-result--success,
.feedback-score-gain {
  color: var(--color-success);
}

.feedback-result--muted {
  color: var(--color-text-muted);
}

.panel-note {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 700;
  padding: 3rem 0;
  text-align: center;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--color-border-default);
  border-radius: 10px;
}

@media (max-width: 1280px) {
  .column-intel {
    grid-template-columns: 1fr;
  }
}
</style>

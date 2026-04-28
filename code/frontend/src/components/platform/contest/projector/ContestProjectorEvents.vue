<script setup lang="ts">
import { Sword, Zap } from 'lucide-vue-next'

import type { AWDAttackLogData } from '@/api/contracts'
import {
  formatProjectorScore,
  formatProjectorTime,
  getAttackTypeLabel,
} from '@/components/platform/contest/projector/contestProjectorFormatters'
import type { ContestProjectorAttackLeader } from '@/components/platform/contest/projector/contestProjectorTypes'

defineProps<{
  firstBlood: AWDAttackLogData | null
  attackLeaders: ContestProjectorAttackLeader[]
  latestAttackEvents: AWDAttackLogData[]
}>()
</script>

<template>
  <section class="event-panel">
    <article class="first-blood-panel">
      <header class="panel-head">
        <div>
          <div class="projector-overline">
            首血
          </div>
          <h3>First Blood</h3>
        </div>
        <Zap class="panel-icon panel-icon--accent" />
      </header>
      <div
        v-if="firstBlood"
        class="first-blood-body"
      >
        <strong>{{ firstBlood.attacker_team }}</strong>
        <span>攻破 {{ firstBlood.victim_team }}</span>
        <small>{{ formatProjectorScore(firstBlood.score_gained) }} pts · {{ formatProjectorTime(firstBlood.created_at) }}</small>
      </div>
      <div
        v-else
        class="panel-empty"
      >
        暂无首血记录
      </div>
    </article>

    <article class="attack-panel">
      <header class="panel-head">
        <div>
          <div class="projector-overline">
            攻击榜
          </div>
          <h3>命中队伍</h3>
        </div>
        <Sword class="panel-icon panel-icon--attack" />
      </header>
      <div class="attack-list">
        <div
          v-for="leader in attackLeaders"
          :key="leader.team_id"
          class="attack-row"
        >
          <span>{{ leader.team_name }}</span>
          <strong>{{ leader.success }} HIT</strong>
          <small>{{ formatProjectorScore(leader.score) }} pts</small>
        </div>
        <div
          v-if="attackLeaders.length === 0"
          class="panel-empty"
        >
          暂无成功攻击
        </div>
      </div>
    </article>

    <article class="attack-feed-panel">
      <header class="panel-head">
        <div>
          <div class="projector-overline">
            攻击流水
          </div>
          <h3>实时事件</h3>
        </div>
      </header>
      <div class="attack-feed">
        <div
          v-for="event in latestAttackEvents"
          :key="event.id"
          class="attack-event"
          :class="{ 'attack-event--success': event.is_success }"
        >
          <div class="attack-event__line">
            <strong>{{ event.attacker_team }}</strong>
            <span>→</span>
            <strong>{{ event.victim_team }}</strong>
          </div>
          <div class="attack-event__meta">
            <span>{{ getAttackTypeLabel(event.attack_type) }}</span>
            <span>{{ event.is_success ? '成功' : '未命中' }}</span>
            <span>{{ formatProjectorTime(event.created_at) }}</span>
          </div>
        </div>
        <div
          v-if="latestAttackEvents.length === 0"
          class="panel-empty"
        >
          暂无攻击流水
        </div>
      </div>
    </article>
  </section>
</template>

<style scoped>
.event-panel,
.first-blood-panel,
.attack-panel,
.attack-feed-panel,
.attack-list,
.attack-feed {
  display: flex;
  flex-direction: column;
}

.event-panel {
  gap: var(--space-4);
}

.first-blood-panel,
.attack-panel,
.attack-feed-panel {
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

.panel-icon--attack {
  color: var(--color-danger);
}

.first-blood-body {
  display: flex;
  min-height: 6rem;
  flex: 1;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-1);
}

.first-blood-body strong {
  color: var(--journal-ink);
  font-size: var(--font-size-1-25);
  font-weight: 900;
}

.first-blood-body span,
.first-blood-body small,
.attack-row small,
.panel-empty {
  color: var(--color-text-muted);
}

.attack-list,
.attack-feed {
  gap: var(--space-2-5);
}

.attack-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  align-items: center;
  gap: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
}

.attack-row span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-row strong {
  color: var(--journal-ink);
  font-family: var(--font-family-mono);
}

.attack-event {
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 50%, transparent);
  padding: var(--space-2-5);
}

.attack-event--success {
  border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 9%, transparent);
}

.attack-event__line,
.attack-event__meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-2);
}

.attack-event__line strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-12);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-event__line span {
  color: var(--color-danger);
  font-weight: 900;
}

.attack-event__meta {
  flex-wrap: wrap;
  margin-top: var(--space-1);
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
  font-weight: 800;
}

.panel-empty {
  padding: var(--space-4) 0;
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
}
</style>

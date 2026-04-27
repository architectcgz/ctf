<script setup lang="ts">
import { ArrowRight, Crosshair, ShieldAlert } from 'lucide-vue-next'

import {
  formatProjectorScore,
  formatProjectorTime,
} from '@/components/platform/contest/projector/contestProjectorFormatters'
import type { ContestProjectorAttackEdge } from '@/components/platform/contest/projector/contestProjectorTypes'

defineProps<{
  edges: ContestProjectorAttackEdge[]
}>()

function getLaneStyle(edge: ContestProjectorAttackEdge): Record<string, string> {
  return {
    '--attack-success-rate': `${edge.successRate}%`,
  }
}
</script>

<template>
  <section class="attack-map-panel">
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          攻击关系
        </div>
        <h3>队伍互攻态势</h3>
      </div>
      <Crosshair class="panel-icon panel-icon--attack" />
    </header>

    <div class="attack-lanes">
      <article
        v-for="edge in edges"
        :key="edge.id"
        class="attack-lane"
        :class="{
          'attack-lane--success': edge.success > 0,
          'attack-lane--mutual': edge.reciprocalSuccess > 0,
        }"
        :style="getLaneStyle(edge)"
      >
        <div class="attack-team attack-team--source">
          <span>攻击方</span>
          <strong>{{ edge.attacker_team }}</strong>
        </div>

        <div class="attack-vector">
          <div class="attack-vector__track">
            <i />
          </div>
          <div class="attack-vector__arrow">
            <ArrowRight />
          </div>
          <div class="attack-vector__stats">
            <strong>{{ edge.success }} HIT</strong>
            <span>{{ edge.failed }} MISS</span>
          </div>
        </div>

        <div class="attack-team attack-team--target">
          <span>目标方</span>
          <strong>{{ edge.victim_team }}</strong>
        </div>

        <div class="attack-lane__meta">
          <span>{{ formatProjectorScore(edge.score) }} pts</span>
          <span>{{ edge.latest_service_label }}</span>
          <span>{{ formatProjectorTime(edge.latest_at) }}</span>
          <span
            v-if="edge.reciprocalSuccess > 0"
            class="attack-lane__mutual"
          >
            互攻 {{ edge.reciprocalSuccess }}
          </span>
        </div>
      </article>

      <div
        v-if="edges.length === 0"
        class="panel-empty"
      >
        暂无队伍攻击关系
      </div>
    </div>
  </section>
</template>

<style scoped>
.attack-map-panel,
.attack-lanes,
.attack-team,
.attack-vector,
.attack-vector__stats {
  display: flex;
  flex-direction: column;
}

.attack-map-panel {
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

.panel-icon--attack {
  color: var(--color-danger);
}

.attack-lanes {
  gap: var(--space-3);
}

.attack-lane {
  --attack-success-rate: 0%;

  display: grid;
  grid-template-columns: minmax(7rem, 1fr) minmax(10rem, 1.25fr) minmax(7rem, 1fr);
  gap: var(--space-3);
  align-items: center;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-surface) 52%, transparent);
  padding: var(--space-3);
}

.attack-lane--success {
  border-color: color-mix(in srgb, var(--color-danger) 26%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-surface));
}

.attack-lane--mutual {
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--color-warning) 22%, transparent);
}

.attack-team {
  min-width: 0;
  gap: var(--space-1);
}

.attack-team--target {
  text-align: right;
}

.attack-team span,
.attack-lane__meta,
.attack-vector__stats span,
.panel-empty {
  color: var(--color-text-muted);
}

.attack-team span {
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.12em;
}

.attack-team strong {
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-vector {
  position: relative;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
}

.attack-vector__track {
  position: relative;
  width: 100%;
  height: var(--space-2);
  overflow: hidden;
  border-radius: 999rem;
  background: color-mix(in srgb, var(--color-border-subtle) 72%, transparent);
}

.attack-vector__track i {
  display: block;
  width: var(--attack-success-rate);
  height: 100%;
  border-radius: inherit;
  background: color-mix(in srgb, var(--color-danger) 64%, transparent);
}

.attack-vector__arrow {
  display: inline-flex;
  width: var(--space-8);
  height: var(--space-8);
  align-items: center;
  justify-content: center;
  border-radius: 999rem;
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}

.attack-vector__arrow svg {
  width: var(--space-4);
  height: var(--space-4);
}

.attack-vector__stats {
  align-items: center;
  gap: var(--space-0-5);
  font-family: var(--font-family-mono);
}

.attack-vector__stats strong {
  color: var(--journal-ink);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.attack-vector__stats span {
  font-size: var(--font-size-10);
  font-weight: 900;
}

.attack-lane__meta {
  display: flex;
  min-width: 0;
  grid-column: 1 / -1;
  flex-wrap: wrap;
  gap: var(--space-2);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.attack-lane__meta span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attack-lane__mutual {
  display: inline-flex;
  align-items: center;
  border-radius: var(--ui-control-radius-sm);
  background: color-mix(in srgb, var(--color-warning) 14%, transparent);
  padding: 0 var(--space-2);
  color: var(--color-warning);
}

.panel-empty {
  display: flex;
  min-height: 8rem;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
}

@media (max-width: 900px) {
  .attack-lane {
    grid-template-columns: 1fr;
  }

  .attack-team--target {
    text-align: left;
  }
}
</style>

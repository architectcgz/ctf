<script setup lang="ts">
import { Activity, MonitorUp, ShieldAlert, Sword, Trophy, Users } from 'lucide-vue-next'

defineProps<{
  title: string
  contestStatusLabel: string
  roundLabel: string
  teamCount: number
  leaderName: string
  serviceHealthRate: number
  successfulAttackCount: number
  trafficRequestCount: number
  abnormalServiceCount: number
}>()
</script>

<template>
  <header class="projector-hero">
    <div class="projector-hero__title">
      <div class="projector-overline">
        实时态势
      </div>
      <h2 class="scoreboard-projector__title">
        {{ title }}
      </h2>
    </div>
    <div class="projector-hero__status">
      <div class="projector-status">
        <MonitorUp class="status-icon" />
        <span>{{ contestStatusLabel }}</span>
      </div>
      <div class="projector-round-pill">
        {{ roundLabel }}
      </div>
    </div>
  </header>

  <div class="projector-metrics">
    <article class="projector-metric">
      <Users class="metric-icon" />
      <span>队伍</span>
      <strong>{{ teamCount }}</strong>
    </article>
    <article class="projector-metric">
      <Trophy class="metric-icon metric-icon--rank" />
      <span>榜首</span>
      <strong>{{ leaderName }}</strong>
    </article>
    <article class="projector-metric">
      <Activity class="metric-icon metric-icon--live" />
      <span>服务健康</span>
      <strong>{{ serviceHealthRate }}%</strong>
    </article>
    <article class="projector-metric">
      <Sword class="metric-icon metric-icon--attack" />
      <span>成功攻击</span>
      <strong>{{ successfulAttackCount }}</strong>
    </article>
    <article class="projector-metric">
      <Activity class="metric-icon metric-icon--live" />
      <span>代理流量</span>
      <strong>{{ trafficRequestCount }}</strong>
    </article>
    <article class="projector-metric">
      <ShieldAlert class="metric-icon metric-icon--danger" />
      <span>异常服务</span>
      <strong>{{ abnormalServiceCount }}</strong>
    </article>
  </div>
</template>

<style scoped>
.projector-hero,
.projector-hero__status,
.projector-status,
.projector-metric {
  display: flex;
  align-items: center;
}

.projector-hero {
  justify-content: space-between;
  gap: var(--space-4);
}

.projector-overline {
  color: var(--color-text-muted);
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.scoreboard-projector__title {
  margin: var(--space-1) 0 0;
  color: var(--journal-ink);
  font-size: var(--font-size-1-60);
  font-weight: 900;
  line-height: 1.05;
}

.projector-hero__title {
  min-width: 0;
}

.projector-hero__status {
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.projector-status,
.projector-round-pill {
  gap: var(--space-2);
  border: 1px solid var(--color-border-subtle);
  border-radius: 999rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 70%, transparent);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.status-icon,
.metric-icon {
  width: var(--space-4);
  height: var(--space-4);
}

.projector-metrics {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: var(--space-3);
}

.projector-metric {
  justify-content: space-between;
  gap: var(--space-3);
  min-width: 0;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 62%, transparent);
  padding: var(--space-3);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.projector-metric strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-00);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric-icon--rank {
  color: var(--color-warning);
}

.metric-icon--live {
  color: var(--color-primary);
}

.metric-icon--attack {
  color: var(--color-danger);
}

.metric-icon--danger {
  color: var(--color-warning);
}

@media (max-width: 1280px) {
  .projector-metrics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .projector-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .projector-hero__status {
    justify-content: flex-start;
  }

  .projector-metrics {
    grid-template-columns: 1fr;
  }
}
</style>

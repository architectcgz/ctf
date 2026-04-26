<script setup lang="ts">
import { Activity } from 'lucide-vue-next'

import type { AWDTrafficSummaryData, AWDTrafficTopTeamData } from '@/api/contracts'
import { formatProjectorTime } from '@/components/platform/contest/projector/contestProjectorFormatters'
import type { ContestProjectorTrafficTrendBar } from '@/components/platform/contest/projector/contestProjectorTypes'

defineProps<{
  summary: AWDTrafficSummaryData | null
  trendBars: ContestProjectorTrafficTrendBar[]
  hotVictims: AWDTrafficTopTeamData[]
}>()
</script>

<template>
  <section class="traffic-panel">
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          流量态势
        </div>
        <h3>代理流量</h3>
      </div>
      <Activity class="panel-icon panel-icon--live" />
    </header>
    <div class="traffic-strip">
      <span>请求 {{ summary?.total_request_count ?? 0 }}</span>
      <span>攻击方 {{ summary?.active_attacker_team_count ?? 0 }}</span>
      <span>目标 {{ summary?.victim_team_count ?? 0 }}</span>
      <span>错误 {{ summary?.error_request_count ?? 0 }}</span>
    </div>
    <div class="traffic-trend">
      <span
        v-for="bucket in trendBars"
        :key="bucket.bucket_start_at"
        class="traffic-bar"
        :style="{ height: bucket.height }"
        :title="`${formatProjectorTime(bucket.bucket_start_at)} · ${bucket.request_count} req`"
      >
        <i :style="{ height: bucket.errorHeight }" />
      </span>
    </div>
    <div class="traffic-columns">
      <div class="traffic-list">
        <div class="traffic-list__title">
          活跃攻击方
        </div>
        <div
          v-for="item in (summary?.top_attackers ?? []).slice(0, 4)"
          :key="item.team_id"
          class="attack-row"
        >
          <span>{{ item.team_name }}</span>
          <strong>{{ item.request_count }} REQ</strong>
          <small>{{ item.error_count }} ERR</small>
        </div>
      </div>
      <div class="traffic-list">
        <div class="traffic-list__title">
          高压目标
        </div>
        <div
          v-for="item in hotVictims"
          :key="item.team_id"
          class="attack-row"
        >
          <span>{{ item.team_name }}</span>
          <strong>{{ item.request_count }} REQ</strong>
          <small>{{ item.error_count }} ERR</small>
        </div>
      </div>
    </div>
    <div
      v-if="(summary?.top_attackers ?? []).length === 0 && hotVictims.length === 0"
      class="panel-empty"
    >
      暂无代理流量
    </div>
  </section>
</template>

<style scoped>
.traffic-panel,
.traffic-list {
  display: flex;
  flex-direction: column;
}

.traffic-panel {
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

.panel-icon--live {
  color: var(--color-primary);
}

.traffic-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.traffic-strip span {
  display: inline-flex;
  min-height: var(--ui-control-height-sm);
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  padding: 0 var(--space-2-5);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.traffic-trend {
  display: flex;
  height: 8rem;
  align-items: flex-end;
  gap: var(--space-1-5);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-bg-surface) 48%, transparent);
  padding: var(--space-3);
}

.traffic-bar {
  position: relative;
  display: block;
  min-height: var(--space-3);
  flex: 1;
  overflow: hidden;
  border-radius: 999rem 999rem 0 0;
  background: color-mix(in srgb, var(--color-primary) 58%, transparent);
}

.traffic-bar i {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  display: block;
  background: color-mix(in srgb, var(--color-danger) 66%, transparent);
}

.traffic-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.traffic-list {
  gap: var(--space-2-5);
}

.traffic-list__title {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
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

.attack-row small,
.panel-empty {
  color: var(--color-text-muted);
}

.panel-empty {
  padding: var(--space-4) 0;
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
}

@media (max-width: 900px) {
  .traffic-columns {
    grid-template-columns: 1fr;
  }
}
</style>

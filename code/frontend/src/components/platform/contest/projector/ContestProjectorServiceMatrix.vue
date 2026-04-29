<script setup lang="ts">
import type { AWDTeamServiceData } from '@/api/contracts'
import type { ContestProjectorServiceMatrixRow } from '@/components/platform/contest/projector/contestProjectorTypes'
import { getServiceStatusLabel } from '@/components/platform/contest/projector/contestProjectorFormatters'

defineProps<{
  rows: ContestProjectorServiceMatrixRow[]
  upCount: number
  downCount: number
  compromisedCount: number
}>()

function getServiceDisplayName(service: AWDTeamServiceData): string {
  return (
    service.service_name?.trim() ||
    service.awd_challenge_title?.trim() ||
    (service.service_id ? `服务 ${service.service_id}` : `题目 ${service.awd_challenge_id}`)
  )
}
</script>

<template>
  <section class="service-matrix-panel">
    <header class="panel-head">
      <div>
        <div class="projector-overline">
          服务墙
        </div>
        <h3>队伍服务状态</h3>
      </div>
      <div class="service-counts">
        <span class="service-chip service-chip--up">UP {{ upCount }}</span>
        <span class="service-chip service-chip--down">DOWN {{ downCount }}</span>
        <span class="service-chip service-chip--compromised">PWN {{ compromisedCount }}</span>
      </div>
    </header>
    <div class="service-matrix">
      <div
        v-for="row in rows"
        :key="row.team_id"
        class="service-team-row"
      >
        <strong>{{ row.team_name }}</strong>
        <div class="service-cell-list">
          <span
            v-for="service in row.services.slice(0, 10)"
            :key="service.id"
            class="service-cell"
            :class="`service-cell--${service.service_status}`"
            :title="`${getServiceDisplayName(service)} · ${getServiceStatusLabel(service.service_status)}`"
          >
            <span class="service-cell__name">{{ getServiceDisplayName(service) }}</span>
            <span class="service-cell__status">{{ getServiceStatusLabel(service.service_status) }}</span>
          </span>
        </div>
      </div>
      <div
        v-if="rows.length === 0"
        class="panel-empty"
      >
        暂无服务状态
      </div>
    </div>
  </section>
</template>

<style scoped>
.service-matrix-panel,
.service-matrix {
  display: flex;
  flex-direction: column;
}

.service-matrix-panel {
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

.service-matrix {
  gap: var(--space-2);
}

.service-team-row {
  display: grid;
  grid-template-columns: minmax(8rem, 12rem) minmax(0, 1fr);
  align-items: center;
  gap: var(--space-3);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
}

.service-team-row strong {
  min-width: 0;
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-13);
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.service-counts,
.service-cell-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.service-chip,
.service-cell {
  display: inline-flex;
  min-height: var(--ui-control-height-sm);
  align-items: center;
  justify-content: center;
  border-radius: var(--ui-control-radius-sm);
  padding: 0 var(--space-2-5);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.service-cell {
  min-width: var(--ui-service-cell-min-width);
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-0-5);
  padding-block: var(--space-1);
}

.service-cell__name,
.service-cell__status {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.service-cell__name {
  color: var(--journal-ink);
  font-size: var(--font-size-12);
}

.service-cell__status {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-10);
}

.service-chip--up,
.service-cell--up {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: var(--color-success);
}

.service-chip--down,
.service-cell--down {
  background: color-mix(in srgb, var(--color-danger) 14%, transparent);
  color: var(--color-danger);
}

.service-chip--compromised,
.service-cell--compromised {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: var(--color-warning);
}

.panel-empty {
  padding: var(--space-4) 0;
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  font-weight: 800;
  text-align: center;
}

@media (max-width: 900px) {
  .service-team-row {
    grid-template-columns: 1fr;
  }
}
</style>

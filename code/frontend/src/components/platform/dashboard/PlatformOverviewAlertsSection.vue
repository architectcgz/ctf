<script setup lang="ts">
import type { AdminDashboardData } from '@/api/contracts'

defineProps<{
  loading: boolean
  alertCount: number
  alerts: AdminDashboardData['alerts']
}>()
</script>

<template>
  <section id="admin-dashboard-alerts" class="workspace-directory-section overview-directory-section">
    <header class="list-heading">
      <div>
        <div class="section-kicker">Alert Stack</div>
        <h2 class="section-title list-heading__title">当前告警</h2>
      </div>
      <div
        class="status-pill"
        :class="['workspace-directory-status-pill', alertCount > 0 ? 'danger' : 'ready']"
      >
        {{ alertCount }} 条
      </div>
    </header>

    <div v-if="loading" class="workspace-directory-loading overview-state">正在同步告警数据...</div>
    <div v-else-if="alertCount === 0" class="workspace-directory-empty overview-state">
      当前没有资源告警。
    </div>
    <div v-else class="workspace-directory-list overview-list-shell">
      <div class="insight-list">
        <div
          v-for="alert in alerts"
          :key="`${alert.container_id}-${alert.type}`"
          class="insight-item"
        >
          <div>
            <strong>{{ alert.message }}</strong>
            <div class="insight-meta">
              <span class="chip danger" :class="'workspace-directory-status-pill'">
                {{ alert.type.toUpperCase() }}
              </span>
              <span
                class="chip"
                :class="'workspace-directory-status-pill workspace-directory-status-pill--muted'"
              >
                {{ alert.container_id }}
              </span>
            </div>
            <div class="item-copy">
              当前 {{ Math.round(alert.value) }}% / 阈值 {{ Math.round(alert.threshold) }}%，建议优先核查该容器最近任务与资源分配情况。
            </div>
          </div>
          <div class="status-pill danger" :class="'workspace-directory-status-pill'">
            {{ Math.round(alert.value) }}%
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.section-title {
  margin-top: var(--space-2-5);
}

.section-kicker {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.overview-state {
  padding: var(--space-5);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--workspace-faint);
}

.overview-list-shell {
  overflow: hidden;
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
}

.insight-list {
  display: grid;
}

.insight-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4-5);
  padding: var(--space-4-5) var(--space-5);
  border-top: 1px solid var(--workspace-line-soft);
}

.insight-item:first-child {
  border-top: 0;
}

.insight-item strong {
  display: block;
  font-size: var(--font-size-15);
  color: var(--journal-ink);
}

.insight-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.item-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.chip,
.status-pill {
  letter-spacing: 0.01em;
  color: var(--journal-muted);
}

.chip.ready,
.status-pill.ready {
  border-color: color-mix(in srgb, var(--workspace-success) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-success) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-success) 82%, var(--journal-ink));
}

.chip.danger,
.status-pill.danger {
  border-color: color-mix(in srgb, var(--workspace-danger) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-danger) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.status-pill {
  min-width: 4.875rem;
}

@media (max-width: 860px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .insight-item {
    padding-left: var(--space-4);
    padding-right: var(--space-4);
  }
}
</style>

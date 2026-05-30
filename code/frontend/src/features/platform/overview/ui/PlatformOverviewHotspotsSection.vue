<script setup lang="ts">
import type { AdminDashboardData } from '@/api/contracts'

defineProps<{
  loading: boolean
  sortedContainers: AdminDashboardData['container_stats']
  formatPercent: (value: number | undefined) => string
  formatUsageBarWidth: (value: number | undefined) => string
  formatBytes: (value: number | undefined) => string
  usageTone: (value: number | undefined) => string
}>()
</script>

<template>
  <section id="admin-dashboard-hotspots" class="workspace-directory-section overview-directory-section">
    <header class="list-heading">
      <div>
        <div class="section-kicker">Resource Hotspots</div>
        <h2 class="section-title list-heading__title">资源热点</h2>
      </div>
    </header>

    <div v-if="loading" class="workspace-directory-loading overview-state">正在同步容器资源数据...</div>
    <div v-else-if="sortedContainers.length === 0" class="workspace-directory-empty overview-state">
      暂无容器运行数据。
    </div>
    <div v-else class="workspace-directory-list overview-list-shell">
      <div class="hotspot-list">
        <article
          v-for="item in sortedContainers"
          :key="item.container_id"
          class="hotspot-item"
        >
          <div class="hotspot-main">
            <div class="hotspot-title-row">
              <strong>{{ item.container_name || item.container_id }}</strong>
              <span
                class="chip"
                :class="[
                  'workspace-directory-status-pill',
                  Math.max(item.cpu_percent ?? 0, item.memory_percent ?? 0) >= 90
                    ? 'danger'
                    : 'warning',
                ]"
              >
                峰值 {{ formatPercent(Math.max(item.cpu_percent ?? 0, item.memory_percent ?? 0)) }}
              </span>
            </div>
            <div class="item-copy hotspot-copy">
              {{ item.container_id }}
            </div>
            <div class="hotspot-memory">
              {{ formatBytes(item.memory_usage) }} / {{ formatBytes(item.memory_limit) }}
            </div>
          </div>

          <div class="hotspot-stats">
            <div class="hotspot-stat">
              <div class="hotspot-stat-head">
                <span>CPU</span>
                <span>{{ formatPercent(item.cpu_percent) }}</span>
              </div>
              <div class="usage-track">
                <div
                  class="usage-bar"
                  :class="usageTone(item.cpu_percent)"
                  :style="{ width: formatUsageBarWidth(item.cpu_percent) }"
                />
              </div>
            </div>

            <div class="hotspot-stat">
              <div class="hotspot-stat-head">
                <span>内存</span>
                <span>{{ formatPercent(item.memory_percent) }}</span>
              </div>
              <div class="usage-track">
                <div
                  class="usage-bar"
                  :class="usageTone(item.memory_percent)"
                  :style="{ width: formatUsageBarWidth(item.memory_percent) }"
                />
              </div>
            </div>
          </div>
        </article>
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

.hotspot-list {
  display: grid;
}

.hotspot-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(17.5rem, 22.5rem);
  gap: var(--space-4-5);
  padding: var(--space-4-5) var(--space-5);
  border-top: 1px solid var(--workspace-line-soft);
}

.hotspot-item:first-child {
  border-top: 0;
}

.hotspot-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2) var(--space-2-5);
}

.hotspot-title-row strong {
  display: block;
  font-size: var(--font-size-15);
  color: var(--journal-ink);
}

.item-copy,
.hotspot-memory {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.hotspot-copy {
  font-family: var(--workspace-font-mono);
}

.chip {
  letter-spacing: 0.01em;
  color: var(--journal-muted);
}

.chip.warning {
  border-color: color-mix(in srgb, var(--workspace-warning) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-warning) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-warning) 86%, var(--journal-ink));
}

.chip.danger {
  border-color: color-mix(in srgb, var(--workspace-danger) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-danger) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.hotspot-stats {
  display: grid;
  gap: var(--space-3);
}

.hotspot-stat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

.usage-track {
  margin-top: var(--space-2);
  height: 0.5rem;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--workspace-panel-soft) 84%, transparent);
}

.usage-bar {
  height: 100%;
  border-radius: 999px;
}

.usage-bar--danger {
  background: var(--color-danger);
}

.usage-bar--warning {
  background: var(--color-warning);
}

.usage-bar--primary {
  background: var(--color-primary);
}

@media (max-width: 1180px) {
  .hotspot-item {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 860px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .hotspot-item {
    padding-left: var(--space-4);
    padding-right: var(--space-4);
  }
}
</style>

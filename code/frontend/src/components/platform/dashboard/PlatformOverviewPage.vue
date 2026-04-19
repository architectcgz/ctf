<script setup lang="ts">
import { toRef } from 'vue'
import { AlertTriangle } from 'lucide-vue-next'

import type { AdminDashboardData } from '@/api/contracts'
import { usePlatformOverviewWorkspace } from '@/composables/usePlatformOverviewWorkspace'

const props = defineProps<{
  dashboard: AdminDashboardData | null
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openAuditLog: []
  openCheatDetection: []
}>()

const {
  alertCount,
  sortedContainers,
  metaPills,
  overviewMetrics,
  railScore,
  railCopy,
  formatPercent,
  formatBytes,
  usageTone,
} = usePlatformOverviewWorkspace(toRef(props, 'dashboard'))
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <section
          id="admin-dashboard-overview"
          class="workspace-hero"
        >
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Operations Workspace</div>
            <h1 class="hero-title">系统值守台</h1>
            <p class="hero-summary">在这里查看平台状态、异常和当前资源热点。</p>

            <div class="meta-strip">
              <span
                v-for="(pill, index) in metaPills"
                :key="pill"
                class="meta-pill"
                :class="{ brand: index === 0 }"
              >
                {{ pill }}
              </span>
            </div>

            <div class="progress-strip metric-panel-grid">
              <article
                v-for="item in overviewMetrics"
                :key="item.key"
                class="progress-card metric-panel-card"
              >
                <div class="progress-card-label metric-panel-label">
                  {{ item.label }}
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ item.value }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  {{ item.hint }}
                </div>
              </article>
            </div>

            <div class="overview-quick-actions">
              <div class="workspace-overline">Quick Actions</div>
              <div class="quick-actions">
                <button
                  type="button"
                  class="quick-action ui-btn ui-btn--primary"
                  @click="emit('openAuditLog')"
                >
                  <span>审计日志</span><span>→</span>
                </button>
                <button
                  type="button"
                  class="quick-action ui-btn ui-btn--ghost"
                  @click="emit('openCheatDetection')"
                >
                  <span>风险研判</span><span>→</span>
                </button>
                <a class="quick-action" href="#admin-dashboard-alerts">
                  <span>查看当前告警</span><span>→</span>
                </a>
                <a class="quick-action" href="#admin-dashboard-hotspots">
                  <span>查看资源热点</span><span>→</span>
                </a>
              </div>
            </div>

            <div v-if="error" class="workspace-alert" role="alert" aria-live="polite">
              <div class="workspace-alert-title-row">
                <AlertTriangle class="workspace-alert-icon" />
                <div class="workspace-alert-title">管理端概览加载失败</div>
              </div>
              <div class="workspace-alert-copy">
                {{ error }}
              </div>
              <div class="workspace-alert-copy">
                可先重试刷新资源状态，再继续查看当前告警与资源热点；若持续失败，建议优先进入审计日志确认后台任务与容器记录。
              </div>
              <div class="workspace-alert-actions">
                <button
                  type="button"
                  class="quick-action quick-action--compact"
                  @click="emit('retry')"
                >
                  <span>重试加载</span><span>→</span>
                </button>
                <button
                  type="button"
                  class="quick-action quick-action--compact"
                  @click="emit('openAuditLog')"
                >
                  <span>审计日志</span><span>→</span>
                </button>
              </div>
            </div>

            <div v-else-if="loading" class="progress-strip metric-panel-grid">
              <div
                v-for="index in 4"
                :key="index"
                class="progress-card progress-card--skeleton metric-panel-card"
              />
            </div>
          </div>

          <aside class="hero-rail">
            <div class="rail-label">System Pulse</div>
            <div class="rail-score">
              {{ railScore }}
              <small>% peak</small>
            </div>
            <div class="rail-copy">
              {{ railCopy }}
            </div>
          </aside>
        </section>

        <section
          id="admin-dashboard-alerts"
          class="section"
        >
          <div class="section-head list-heading">
            <div>
              <div class="section-kicker">Alert Stack</div>
              <h2 class="section-title list-heading__title">当前告警</h2>
            </div>
            <div class="status-pill" :class="alertCount > 0 ? 'danger' : 'ready'">
              {{ alertCount }} 条
            </div>
          </div>

          <article class="panel panel-pad">
            <div v-if="loading" class="empty-inline">正在同步告警数据...</div>
            <div v-else-if="alertCount === 0" class="empty-inline">当前没有资源告警。</div>
            <div v-else class="insight-list">
              <div
                v-for="alert in dashboard?.alerts"
                :key="`${alert.container_id}-${alert.type}`"
                class="insight-item"
              >
                <div>
                  <strong>{{ alert.message }}</strong>
                  <div class="insight-meta">
                    <span class="chip danger">{{ alert.type.toUpperCase() }}</span>
                    <span class="chip">{{ alert.container_id }}</span>
                  </div>
                  <div class="item-copy">
                    当前 {{ Math.round(alert.value) }}% / 阈值
                    {{ Math.round(alert.threshold) }}%，建议优先核查该容器最近任务与资源分配情况。
                  </div>
                </div>
                <div class="status-pill danger">{{ Math.round(alert.value) }}%</div>
              </div>
            </div>
          </article>
        </section>

        <section
          id="admin-dashboard-hotspots"
          class="section"
        >
          <div class="section-head list-heading">
            <div>
              <div class="section-kicker">Resource Hotspots</div>
              <h2 class="section-title list-heading__title">资源热点</h2>
            </div>
          </div>

          <article class="panel panel-pad">
            <div v-if="loading" class="empty-inline">正在同步容器资源数据...</div>
            <div v-else-if="sortedContainers.length === 0" class="empty-inline">
              暂无容器运行数据。
            </div>
            <div v-else class="hotspot-list">
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
                      :class="
                        Math.max(item.cpu_percent ?? 0, item.memory_percent ?? 0) >= 90
                          ? 'danger'
                          : 'warning'
                      "
                    >
                      峰值
                      {{ formatPercent(Math.max(item.cpu_percent ?? 0, item.memory_percent ?? 0)) }}
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
                        :style="{ width: `${Math.round(item.cpu_percent ?? 0)}%` }"
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
                        :style="{ width: `${Math.round(item.memory_percent ?? 0)}%` }"
                      />
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </article>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-success: var(--color-success);
  --workspace-warning: var(--color-warning);
  --workspace-danger: var(--color-danger);
  --workspace-shadow-shell: 0 24px 84px
    color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-xl: 28px;
  --workspace-radius-lg: 18px;
  --workspace-radius-md: 14px;
  --workspace-font-sans: var(--font-family-sans);
  --workspace-font-mono: var(--font-family-mono);
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 244px;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  max-width: 11ch;
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-4-5);
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 var(--space-2-5);
  border: 1px solid var(--workspace-line-soft);
  border-radius: 8px;
  background: color-mix(in srgb, var(--workspace-panel) 72%, transparent);
  font-size: var(--font-size-12);
  color: var(--journal-muted);
}

.meta-pill.brand {
  border-color: color-mix(in srgb, var(--workspace-brand) 20%, transparent);
  background: var(--workspace-brand-soft);
  color: var(--workspace-brand-ink);
}

.progress-strip {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
  --metric-panel-grid-gap: var(--space-3);
  margin-top: var(--space-5-5);
}

.panel {
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
}

.progress-card {
  --metric-panel-padding: var(--space-3-5) var(--space-4) var(--space-4);
}

.progress-card--skeleton {
  min-height: 110px;
}

.progress-card-label,
.section-kicker {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.progress-card-value {
  --metric-panel-value-margin-top: var(--space-2-5);
  --metric-panel-value-size: 26px;
}

.progress-card-hint,
.item-copy,
.rail-copy,
.workspace-alert-copy,
.hotspot-memory {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.overview-quick-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2-5) var(--space-4);
  margin-top: var(--space-4-5);
}

.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
}

.quick-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2-5);
  min-height: 52px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--workspace-line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--workspace-panel) 82%, transparent);
  color: var(--journal-ink);
  text-decoration: none;
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.quick-action span:last-child {
  color: var(--workspace-faint);
}

.quick-action:hover,
.quick-action:focus-visible {
  border-color: color-mix(in srgb, var(--workspace-brand) 34%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 8%, var(--workspace-panel));
  color: var(--workspace-brand-ink);
  outline: none;
}

.quick-action--compact {
  min-height: 42px;
}

.quick-action.ui-btn {
  min-height: var(--ui-btn-height, 52px);
  border-color: var(--ui-btn-border, var(--workspace-line-soft));
  background: var(--ui-btn-background, color-mix(in srgb, var(--workspace-panel) 82%, transparent));
  padding: var(--ui-btn-padding, 0 var(--space-3-5));
  color: var(--ui-btn-color, var(--journal-ink));
  box-shadow: none;
}

.quick-action.ui-btn:hover:not(:disabled) {
  border-color: var(--ui-btn-hover-border, color-mix(in srgb, var(--workspace-brand) 34%, transparent));
  background: var(--ui-btn-hover-background, color-mix(in srgb, var(--workspace-brand) 8%, var(--workspace-panel)));
  color: var(--ui-btn-hover-color, var(--workspace-brand-ink));
}

.quick-action.ui-btn:focus-visible {
  outline: 2px solid var(--ui-btn-focus-ring, color-mix(in srgb, var(--workspace-brand) 16%, transparent));
  outline-offset: 2px;
}

.quick-actions > .ui-btn {
  --ui-btn-height: 2.75rem;
  --ui-btn-padding: var(--space-2-5) var(--space-4);
  --ui-btn-radius: 1rem;
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.quick-actions > .ui-btn.ui-btn--primary {
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 46%, var(--journal-border));
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: var(--color-primary-hover);
  --ui-btn-primary-hover-shadow: 0 12px 24px color-mix(in srgb, var(--journal-accent) 24%, transparent);
}

.quick-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--journal-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.hero-rail {
  padding-left: var(--space-6);
  border-left: 1px solid var(--workspace-line-soft);
}

.rail-label {
  font-size: var(--font-size-11);
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.rail-score {
  margin-top: var(--space-2-5);
  font: 700 38px/1 var(--workspace-font-mono);
  color: var(--journal-ink);
}

.rail-score small {
  margin-left: var(--space-1);
  font-size: var(--font-size-15);
  color: var(--workspace-faint);
}

.rail-copy {
  padding-top: var(--space-3-5);
  border-top: 1px solid var(--workspace-line-soft);
}

.panel-pad {
  padding: var(--space-5);
}

.panel-title {
  margin: 0;
  font-size: var(--font-size-18);
  line-height: 1.2;
  color: var(--journal-ink);
}

.section {
  padding-top: var(--space-6);
  border-top: 1px solid var(--workspace-line-soft);
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
}

.list-heading__title,
.section-title {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-22);
  line-height: 1.12;
  color: var(--journal-ink);
}

.insight-list {
  display: grid;
  border-top: 1px solid var(--workspace-line-soft);
}

.insight-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4-5);
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--workspace-line-soft);
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

.chip,
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 var(--space-2);
  border-radius: 7px;
  border: 1px solid var(--workspace-line-soft);
  font-size: var(--font-size-11-5);
  font-weight: 600;
  letter-spacing: 0.01em;
  color: var(--journal-muted);
}

.chip.ready,
.status-pill.ready {
  border-color: color-mix(in srgb, var(--workspace-success) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-success) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-success) 82%, var(--journal-ink));
}

.chip.warning,
.status-pill.warning {
  border-color: color-mix(in srgb, var(--workspace-warning) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-warning) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-warning) 86%, var(--journal-ink));
}

.chip.danger,
.status-pill.danger {
  border-color: color-mix(in srgb, var(--workspace-danger) 28%, transparent);
  background: color-mix(in srgb, var(--workspace-danger) 10%, transparent);
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.status-pill {
  min-height: 30px;
  min-width: 78px;
  border-radius: 8px;
}

.workspace-alert {
  margin-top: var(--space-4-5);
  padding: var(--space-4) var(--space-4-5);
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
}

.workspace-alert-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2-5);
}

.workspace-alert-icon {
  width: 18px;
  height: 18px;
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.workspace-alert-title {
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-3-5);
}

.empty-inline {
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--workspace-faint);
}

.hotspot-list {
  display: grid;
  gap: var(--space-3-5);
}

.hotspot-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
  gap: var(--space-4-5);
  padding: var(--space-4-5) 0;
  border-top: 1px solid var(--workspace-line-soft);
}

.hotspot-item:first-child {
  padding-top: 0;
  border-top: 0;
}

.hotspot-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2) var(--space-2-5);
}

.hotspot-title-row strong {
  font-size: var(--font-size-15);
  color: var(--journal-ink);
}

.hotspot-copy {
  font-family: var(--workspace-font-mono);
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
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--workspace-panel-soft) 84%, transparent);
}

.usage-bar {
  height: 100%;
  border-radius: 999px;
}

.bg-\[var\(--color-danger\)\] {
  background: var(--color-danger);
}

.bg-\[var\(--color-warning\)\] {
  background: var(--color-warning);
}

.bg-\[var\(--color-primary\)\] {
  background: var(--color-primary);
}

.admin-action-row {
  --admin-action-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border: 1px solid var(--admin-action-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-4) var(--space-4);
  text-align: left;
  transition:
    border-color 150ms ease,
    background-color 150ms ease;
}

.admin-action-row:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, var(--journal-border));
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--journal-surface));
}

.admin-action-row:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

@media (max-width: 1180px) {
  .workspace-hero,
  .hotspot-item {
    grid-template-columns: 1fr;
  }

  .hero-rail {
    padding-top: var(--space-5);
    padding-left: 0;
    border-top: 1px solid var(--workspace-line-soft);
    border-left: 0;
  }
}

@media (max-width: 860px) {
  .progress-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .list-heading,
  .section-head {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .content-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }

  .progress-strip {
    grid-template-columns: 1fr;
  }
}
</style>

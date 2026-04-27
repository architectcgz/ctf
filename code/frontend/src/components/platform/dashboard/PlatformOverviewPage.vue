<script setup lang="ts">
import { toRef } from 'vue'
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Clock,
  Server,
  ShieldCheck,
  Users,
} from 'lucide-vue-next'

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
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero overview-shell">
    <div class="workspace-grid">
      <main class="content-pane overview-content">
        <section
          id="admin-dashboard-overview"
          class="overview-panel"
        >
          <section class="workspace-hero">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">
                Operations Workspace
              </div>
              <h1 class="hero-title workspace-page-title">
                系统值守台
              </h1>
              <p class="hero-summary workspace-page-copy">
                在这里查看平台状态、异常和当前资源热点。
              </p>

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
            </div>

            <div class="overview-hero-actions">
              <div class="hero-meta-badge">
                <span class="hero-meta-badge__label">System Pulse</span>
                <span class="hero-meta-badge__value">
                  {{ railScore }}
                  <small>% peak</small>
                </span>
                <span class="hero-meta-badge__hint">{{ railCopy }}</span>
              </div>

              <div class="overview-action-grid">
                <button
                  type="button"
                  class="ui-btn ui-btn--primary overview-action-main"
                  @click="emit('openAuditLog')"
                >
                  <Clock class="h-4 w-4" />
                  审计日志
                </button>
                <button
                  type="button"
                  class="ui-btn ui-btn--ghost"
                  @click="emit('openCheatDetection')"
                >
                  <ShieldCheck class="h-4 w-4" />
                  风险研判
                </button>
                <a
                  class="ui-btn ui-btn--ghost overview-anchor-btn"
                  href="#admin-dashboard-alerts"
                >
                  <AlertTriangle class="h-4 w-4" />
                  当前告警
                </a>
                <a
                  class="ui-btn ui-btn--ghost overview-anchor-btn"
                  href="#admin-dashboard-hotspots"
                >
                  <Server class="h-4 w-4" />
                  资源热点
                </a>
              </div>
            </div>
          </section>

          <div
            v-if="loading && !dashboard"
            class="admin-summary-grid overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
          >
            <article
              v-for="index in 4"
              :key="index"
              class="journal-note progress-card metric-panel-card progress-card--skeleton animate-pulse"
            >
              <div class="overview-skeleton-block" />
            </article>
          </div>

          <div
            v-else
            class="admin-summary-grid overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
          >
            <article
              v-for="item in overviewMetrics"
              :key="item.key"
              class="journal-note progress-card metric-panel-card"
            >
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>{{ item.label }}</span>
                <component
                  :is="
                    item.key === 'online_users'
                      ? Users
                      : item.key === 'active_containers'
                        ? Server
                        : item.key === 'cpu_usage'
                          ? Activity
                          : ShieldCheck
                  "
                  class="h-4 w-4"
                />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ item.value.padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                {{ item.hint }}
              </div>
            </article>
          </div>

          <div
            v-if="error"
            class="workspace-alert"
            role="alert"
            aria-live="polite"
          >
            <div class="workspace-alert-title-row">
              <AlertTriangle class="workspace-alert-icon" />
              <div class="workspace-alert-title">
                管理端概览加载失败
              </div>
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
                class="ui-btn ui-btn--ghost"
                @click="emit('retry')"
              >
                <ArrowRight class="h-4 w-4" />
                重试加载
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="emit('openAuditLog')"
              >
                <Clock class="h-4 w-4" />
                审计日志
              </button>
            </div>
          </div>
        </section>

        <section
          id="admin-dashboard-alerts"
          class="workspace-directory-section overview-directory-section"
        >
          <header class="list-heading">
            <div>
              <div class="section-kicker">
                Alert Stack
              </div>
              <h2 class="section-title list-heading__title">当前告警</h2>
            </div>
            <div
              class="status-pill"
              :class="alertCount > 0 ? 'danger' : 'ready'"
            >
              {{ alertCount }} 条
            </div>
          </header>

          <div
            v-if="loading"
            class="workspace-directory-loading overview-state"
          >
            正在同步告警数据...
          </div>
          <div
            v-else-if="alertCount === 0"
            class="workspace-directory-empty overview-state"
          >
            当前没有资源告警。
          </div>
          <div
            v-else
            class="workspace-directory-list overview-list-shell"
          >
            <div class="insight-list">
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
                <div class="status-pill danger">
                  {{ Math.round(alert.value) }}%
                </div>
              </div>
            </div>
          </div>
        </section>

        <section
          id="admin-dashboard-hotspots"
          class="workspace-directory-section overview-directory-section"
        >
          <header class="list-heading">
            <div>
              <div class="section-kicker">
                Resource Hotspots
              </div>
              <h2 class="section-title list-heading__title">资源热点</h2>
            </div>
          </header>

          <div
            v-if="loading"
            class="workspace-directory-loading overview-state"
          >
            正在同步容器资源数据...
          </div>
          <div
            v-else-if="sortedContainers.length === 0"
            class="workspace-directory-empty overview-state"
          >
            暂无容器运行数据。
          </div>
          <div
            v-else
            class="workspace-directory-list overview-list-shell"
          >
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
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.overview-shell {
  --journal-shell-dark-accent: var(--color-primary-hover);
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
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
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-lg: 18px;
  --workspace-font-mono: var(--font-family-mono);
}

.overview-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}

.overview-panel {
  display: grid;
  gap: 0;
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  max-width: 11ch;
}

.hero-summary {
  max-width: 48rem;
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-6);
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 var(--space-2-5);
  border: 1px solid var(--workspace-line-soft);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--workspace-panel) 72%, transparent);
  font-size: var(--font-size-12);
  color: var(--journal-muted);
}

.meta-pill.brand {
  border-color: color-mix(in srgb, var(--workspace-brand) 20%, transparent);
  background: var(--workspace-brand-soft);
  color: var(--workspace-brand-ink);
}

.overview-hero-actions {
  display: grid;
  align-self: start;
  justify-content: flex-end;
  align-content: start;
  gap: var(--space-2-5);
  width: min(19rem, 100%);
  min-width: 16rem;
  padding: var(--space-3);
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--workspace-brand) 12%, transparent),
      transparent 46%
    ),
    color-mix(in srgb, var(--workspace-panel) 90%, transparent);
  box-shadow: var(--workspace-shadow-panel);
}

.hero-meta-badge {
  display: grid;
  gap: var(--space-1);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-meta-badge__label {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.hero-meta-badge__value {
  font: 700 var(--font-size-24, 1.5rem) / 1 var(--workspace-font-mono);
  color: var(--journal-ink);
}

.hero-meta-badge__value small {
  margin-left: var(--space-1);
  font-size: var(--font-size-12);
  color: var(--workspace-faint);
}

.hero-meta-badge__hint {
  font-size: var(--font-size-12);
  line-height: 1.45;
  color: var(--journal-muted);
}

.overview-action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
}

.overview-hero-actions > .ui-btn,
.overview-action-grid > .ui-btn,
.workspace-alert-actions > .ui-btn {
  --ui-btn-height: 2.5rem;
  --ui-btn-padding: var(--space-2) var(--space-3);
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 16%, transparent);
  justify-content: center;
  min-width: 0;
}

.overview-action-main {
  grid-column: 1 / -1;
}

.overview-action-grid > .ui-btn.ui-btn--primary {
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 46%, var(--journal-border));
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: var(--color-primary-hover);
  --ui-btn-primary-hover-shadow: 0 12px 24px color-mix(in srgb, var(--journal-accent) 24%, transparent);
}

.overview-action-grid > .ui-btn.ui-btn--ghost,
.workspace-alert-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--journal-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.overview-anchor-btn {
  text-decoration: none;
}

.overview-summary {
  --metric-panel-columns: 4;
  --metric-panel-grid-gap: var(--space-3);
}

.overview-skeleton-block {
  min-height: 6.875rem;
  border-radius: 1rem;
  background: color-mix(in srgb, var(--workspace-panel-soft) 84%, transparent);
}

.workspace-alert {
  margin-top: var(--space-5);
  padding: var(--space-4) var(--space-4-5);
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 1rem;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
}

.workspace-alert-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2-5);
}

.workspace-alert-icon {
  width: 1.125rem;
  height: 1.125rem;
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.workspace-alert-title {
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-3-5);
}

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

.insight-list,
.hotspot-list {
  display: grid;
}

.insight-item,
.hotspot-item {
  display: grid;
  gap: var(--space-4-5);
  padding: var(--space-4-5) var(--space-5);
  border-top: 1px solid var(--workspace-line-soft);
}

.insight-item:first-child,
.hotspot-item:first-child {
  border-top: 0;
}

.insight-item {
  grid-template-columns: minmax(0, 1fr) auto;
}

.insight-item strong,
.hotspot-title-row strong {
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

.item-copy,
.hotspot-memory {
  margin-top: var(--space-2);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.chip,
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.5rem;
  padding: 0 var(--space-2);
  border-radius: 0.4375rem;
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
  min-height: 1.875rem;
  min-width: 4.875rem;
  border-radius: 0.5rem;
}

.hotspot-item {
  grid-template-columns: minmax(0, 1fr) minmax(17.5rem, 22.5rem);
}

.hotspot-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2) var(--space-2-5);
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
  .workspace-hero,
  .hotspot-item {
    grid-template-columns: 1fr;
  }

  .overview-hero-actions {
    width: 100%;
    min-width: 0;
  }
}

@media (max-width: 720px) {
  .overview-action-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 860px) {
  .overview-summary {
    --metric-panel-columns: 2;
  }

  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 640px) {
  .overview-summary {
    --metric-panel-columns: 1;
  }

  .content-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }

  .insight-item,
  .hotspot-item {
    padding-left: var(--space-4);
    padding-right: var(--space-4);
  }
}
</style>

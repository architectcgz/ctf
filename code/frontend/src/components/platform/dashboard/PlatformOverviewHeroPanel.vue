<script setup lang="ts">
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Clock,
  Server,
  ShieldCheck,
  Users,
} from 'lucide-vue-next'

const props = defineProps<{
  error: string | null
  showSkeleton: boolean
  metaPills: string[]
  railScore: string
  railCopy: string
  overviewMetrics: Array<{
    key: string
    label: string
    value: string
    hint: string
  }>
}>()

const emit = defineEmits<{
  retry: []
  openAuditLog: []
  openCheatDetection: []
}>()
</script>

<template>
  <section id="admin-dashboard-overview" class="overview-panel">
    <header class="workspace-page-header overview-page-header">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">Operations Workspace</div>
        <h1 class="hero-title workspace-page-title">系统值守台</h1>
        <p class="hero-summary workspace-page-copy">在这里查看平台状态、异常和当前资源热点。</p>

        <div class="meta-strip">
          <span
            v-for="(pill, index) in props.metaPills"
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
            {{ props.railScore }}
            <small>% peak</small>
          </span>
          <span class="hero-meta-badge__hint">{{ props.railCopy }}</span>
        </div>

        <div class="header-actions overview-action-grid">
          <button type="button" class="header-btn header-btn--primary" @click="emit('openAuditLog')">
            <Clock class="h-4 w-4" />
            审计日志
          </button>
          <button
            type="button"
            class="header-btn header-btn--ghost"
            @click="emit('openCheatDetection')"
          >
            <ShieldCheck class="h-4 w-4" />
            风险研判
          </button>
          <a class="header-btn header-btn--ghost overview-anchor-btn" href="#admin-dashboard-alerts">
            <AlertTriangle class="h-4 w-4" />
            当前告警
          </a>
          <a class="header-btn header-btn--ghost overview-anchor-btn" href="#admin-dashboard-hotspots">
            <Server class="h-4 w-4" />
            资源热点
          </a>
        </div>
      </div>
    </header>

    <div
      v-if="props.showSkeleton"
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
        v-for="item in props.overviewMetrics"
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

    <div v-if="props.error" class="workspace-alert" role="alert" aria-live="polite">
      <div class="workspace-alert-title-row">
        <AlertTriangle class="workspace-alert-icon" />
        <div class="workspace-alert-title">管理端概览加载失败</div>
      </div>
      <div class="workspace-alert-copy">
        {{ props.error }}
      </div>
      <div class="workspace-alert-copy">
        可先重试刷新资源状态，再继续查看当前告警与资源热点；若持续失败，建议优先进入审计日志确认后台任务与容器记录。
      </div>
      <div class="workspace-alert-actions">
        <button type="button" class="ui-btn ui-btn--ghost" @click="emit('retry')">
          <ArrowRight class="h-4 w-4" />
          重试加载
        </button>
        <button type="button" class="ui-btn ui-btn--ghost" @click="emit('openAuditLog')">
          <Clock class="h-4 w-4" />
          审计日志
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.overview-panel {
  display: grid;
  gap: 0;
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

.overview-page-header {
  align-items: start;
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

.workspace-alert-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--journal-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

@media (max-width: 1180px) {
  .overview-hero-actions {
    width: 100%;
    min-width: 0;
  }
}

@media (max-width: 860px) {
  .overview-summary {
    --metric-panel-columns: 2;
  }
}

@media (max-width: 720px) {
  .overview-action-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .overview-summary {
    --metric-panel-columns: 1;
  }
}
</style>

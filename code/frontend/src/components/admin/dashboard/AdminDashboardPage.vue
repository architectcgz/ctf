<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { AlertTriangle, ArrowRight, ShieldAlert, SquareStack } from 'lucide-vue-next'

import type { AdminDashboardData } from '@/api/contracts'

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

const panelTabs = [
  {
    key: 'overview',
    label: '总览',
    tabId: 'admin-dashboard-tab-overview',
    panelId: 'admin-dashboard-panel-overview',
  },
  {
    key: 'alerts',
    label: '当前告警',
    tabId: 'admin-dashboard-tab-alerts',
    panelId: 'admin-dashboard-panel-alerts',
  },
  {
    key: 'hotspots',
    label: '资源热点',
    tabId: 'admin-dashboard-tab-hotspots',
    panelId: 'admin-dashboard-panel-hotspots',
  },
] as const

type DashboardPanelKey = (typeof panelTabs)[number]['key']

const dashboardPanelSet = new Set<DashboardPanelKey>(panelTabs.map((tab) => tab.key))

function resolvePanelFromLocation(): DashboardPanelKey {
  if (typeof window === 'undefined') return 'overview'
  if (!window.location.pathname || window.location.pathname === '/') return 'overview'
  const panel = new URLSearchParams(window.location.search).get('panel')
  if (panel && dashboardPanelSet.has(panel as DashboardPanelKey)) {
    return panel as DashboardPanelKey
  }
  return 'overview'
}

function syncPanelToLocation(panelKey: DashboardPanelKey): void {
  if (typeof window === 'undefined') return
  const url = new URL(window.location.href)
  url.searchParams.set('panel', panelKey)
  window.history.replaceState(window.history.state, '', `${url.pathname}${url.search}${url.hash}`)
}

const activePanel = ref<DashboardPanelKey>(resolvePanelFromLocation())
const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])

const alertCount = computed(() => props.dashboard?.alerts.length ?? 0)
const healthSummary = computed(() => {
  const cpu = props.dashboard?.cpu_usage ?? 0
  const memory = props.dashboard?.memory_usage ?? 0
  if (alertCount.value > 0 || cpu >= 90 || memory >= 90)
    return { label: '高风险', accent: 'danger' as const }
  if (cpu >= 75 || memory >= 75) return { label: '需要关注', accent: 'warning' as const }
  return { label: '运行稳定', accent: 'success' as const }
})

const quickSignals = computed(() => [
  {
    label: '在线用户',
    value: props.dashboard?.online_users ?? 0,
    helper: '当前在线账号',
    accent: 'primary' as const,
  },
  {
    label: '活跃容器',
    value: props.dashboard?.active_containers ?? 0,
    helper: '正在运行的实例',
    accent: 'success' as const,
  },
  {
    label: '平均 CPU',
    value: formatPercent(props.dashboard?.cpu_usage),
    helper: '当前资源水位',
    accent: healthSummary.value.accent,
  },
  {
    label: '平均内存',
    value: formatPercent(props.dashboard?.memory_usage),
    helper: '结合阈值判断回收',
    accent: healthSummary.value.accent,
  },
])

const sortedContainers = computed(() =>
  [...(props.dashboard?.container_stats ?? [])].sort((left, right) => {
    const leftPeak = Math.max(left.cpu_percent ?? 0, left.memory_percent ?? 0)
    const rightPeak = Math.max(right.cpu_percent ?? 0, right.memory_percent ?? 0)
    return rightPeak - leftPeak
  })
)

const metaPills = computed(() => [
  'Admin Workspace',
  healthSummary.value.label,
  alertCount.value > 0 ? `${alertCount.value} 条资源告警` : '暂无资源告警',
  `活跃容器 ${props.dashboard?.active_containers ?? 0} 个`,
])

const overviewMetrics = computed(() =>
  quickSignals.value.map((item) => ({
    key: item.label,
    label: item.label,
    value: String(item.value),
    hint: item.helper,
  }))
)

const peakContainer = computed(() => sortedContainers.value[0] ?? null)

const railScore = computed(() =>
  String(Math.round(Math.max(props.dashboard?.cpu_usage ?? 0, props.dashboard?.memory_usage ?? 0)))
)

const railCopy = computed(() => {
  if (alertCount.value > 0) {
    return `当前共有 ${alertCount.value} 条资源告警，建议先处理高阈值容器，再结合审计日志确认是否存在持续异常。`
  }

  if (peakContainer.value) {
    return `当前最需要关注的是 ${peakContainer.value.container_name || peakContainer.value.container_id}，可以继续查看资源热点判断是否需要回收或扩容。`
  }

  return '当前没有明显异常，可以继续保持对容器负载和审计记录的例行巡检。'
})

function setTabButtonRef(index: number, element: HTMLButtonElement | null): void {
  tabButtonRefs.value[index] = element
}

function selectPanel(panelKey: DashboardPanelKey): void {
  if (activePanel.value === panelKey) return
  activePanel.value = panelKey
  syncPanelToLocation(panelKey)
}

function focusTabByIndex(index: number): void {
  nextTick(() => {
    tabButtonRefs.value[index]?.focus()
  })
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return

  event.preventDefault()

  if (event.key === 'Home') {
    selectPanel(panelTabs[0].key)
    focusTabByIndex(0)
    return
  }

  if (event.key === 'End') {
    const endIndex = panelTabs.length - 1
    selectPanel(panelTabs[endIndex].key)
    focusTabByIndex(endIndex)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + panelTabs.length) % panelTabs.length
  selectPanel(panelTabs[nextIndex].key)
  focusTabByIndex(nextIndex)
}

function formatPercent(value: number | undefined): string {
  return `${Math.round(value ?? 0)}%`
}

function formatBytes(value: number | undefined): string {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function usageTone(value: number | undefined): string {
  const normalized = Math.round(value ?? 0)
  if (normalized >= 90) return 'bg-[var(--color-danger)]'
  if (normalized >= 75) return 'bg-[var(--color-warning)]'
  return 'bg-[var(--color-primary)]'
}
</script>

<template>
  <div class="workspace-shell">
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Operations Workspace</span>
        <span class="class-chip">系统值守</span>
      </div>
      <div class="top-note">
        <span>资源告警 {{ alertCount }} 条</span>
        <span>热点容器 {{ sortedContainers.length }} 个</span>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="系统值守视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(index, element as HTMLButtonElement | null)"
        class="top-tab"
        type="button"
        role="tab"
        :tabindex="activePanel === tab.key ? 0 : -1"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :class="{ active: activePanel === tab.key }"
        @click="selectPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <div class="workspace-grid">
      <main class="content-pane">
        <section
          v-show="activePanel === 'overview'"
          id="admin-dashboard-panel-overview"
          class="workspace-hero tab-panel"
          :class="{ active: activePanel === 'overview' }"
          role="tabpanel"
          aria-labelledby="admin-dashboard-tab-overview"
          :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
        >
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Operations Workspace</div>
            <h1 class="hero-title workspace-tab-heading__title">系统值守台</h1>
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

            <div class="progress-strip">
              <article v-for="item in overviewMetrics" :key="item.key" class="progress-card">
                <div class="progress-card-label">
                  {{ item.label }}
                </div>
                <div class="progress-card-value">
                  {{ item.value }}
                </div>
                <div class="progress-card-hint">
                  {{ item.hint }}
                </div>
              </article>
            </div>

            <div class="overview-quick-actions">
              <div class="workspace-overline">Quick Actions</div>
              <div class="quick-actions">
                <button
                  type="button"
                  class="quick-action admin-btn admin-btn-primary"
                  @click="emit('openAuditLog')"
                >
                  <span>审计日志</span><span>→</span>
                </button>
                <button
                  type="button"
                  class="quick-action admin-btn admin-btn-ghost"
                  @click="emit('openCheatDetection')"
                >
                  <span>风险研判</span><span>→</span>
                </button>
                <button type="button" class="quick-action" @click="selectPanel('alerts')">
                  <span>查看当前告警</span><span>→</span>
                </button>
                <button type="button" class="quick-action" @click="selectPanel('hotspots')">
                  <span>查看资源热点</span><span>→</span>
                </button>
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

            <div v-else-if="loading" class="progress-strip">
              <div v-for="index in 4" :key="index" class="progress-card progress-card--skeleton" />
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
          v-show="activePanel === 'alerts'"
          id="admin-dashboard-panel-alerts"
          class="section tab-panel"
          :class="{ active: activePanel === 'alerts' }"
          role="tabpanel"
          aria-labelledby="admin-dashboard-tab-alerts"
          :aria-hidden="activePanel === 'alerts' ? 'false' : 'true'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Alert Stack</div>
              <h2 class="section-title workspace-tab-heading__title">当前告警</h2>
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
          v-show="activePanel === 'hotspots'"
          id="admin-dashboard-panel-hotspots"
          class="section tab-panel"
          :class="{ active: activePanel === 'hotspots' }"
          role="tabpanel"
          aria-labelledby="admin-dashboard-tab-hotspots"
          :aria-hidden="activePanel === 'hotspots' ? 'false' : 'true'"
        >
          <div class="section-head workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="section-kicker">Resource Hotspots</div>
              <h2 class="section-title workspace-tab-heading__title">资源热点</h2>
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
  --workspace-font-sans:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
  --workspace-font-mono: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 244px;
  gap: 28px;
  padding-bottom: 26px;
  border-bottom: 1px solid var(--workspace-line-soft);
}

.tab-panel.workspace-hero.active {
  display: grid;
}

.hero-title {
  max-width: 11ch;
}

.hero-summary {
  max-width: 760px;
  margin-top: 14px;
  font-size: 15px;
  line-height: 1.9;
  color: var(--journal-muted);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 18px;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 9px;
  border: 1px solid var(--workspace-line-soft);
  border-radius: 8px;
  background: color-mix(in srgb, var(--workspace-panel) 72%, transparent);
  font-size: 12px;
  color: var(--journal-muted);
}

.meta-pill.brand {
  border-color: color-mix(in srgb, var(--workspace-brand) 20%, transparent);
  background: var(--workspace-brand-soft);
  color: var(--workspace-brand-ink);
}

.progress-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 22px;
}

.progress-card,
.panel {
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
}

.progress-card {
  padding: 14px 16px 15px;
}

.progress-card--skeleton {
  min-height: 110px;
}

.progress-card-label,
.section-kicker {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.progress-card-value {
  margin-top: 10px;
  font-size: 26px;
  line-height: 1;
  letter-spacing: -0.03em;
  color: var(--journal-ink);
}

.progress-card-hint,
.item-copy,
.rail-copy,
.workspace-alert-copy,
.hotspot-memory {
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--journal-muted);
}

.overview-quick-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px 16px;
  margin-top: 18px;
}

.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.quick-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 52px;
  padding: 0 14px;
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

.hero-rail {
  padding-left: 24px;
  border-left: 1px solid var(--workspace-line-soft);
}

.rail-label {
  font-size: 11px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--workspace-faint);
}

.rail-score {
  margin-top: 10px;
  font: 700 38px/1 var(--workspace-font-mono);
  color: var(--journal-ink);
}

.rail-score small {
  margin-left: 4px;
  font-size: 15px;
  color: var(--workspace-faint);
}

.rail-copy {
  padding-top: 14px;
  border-top: 1px solid var(--workspace-line-soft);
}

.panel-pad {
  padding: 20px;
}

.panel-title {
  margin: 0;
  font-size: 18px;
  line-height: 1.2;
  color: var(--journal-ink);
}

.section {
  padding-top: 26px;
  border-top: 1px solid var(--workspace-line-soft);
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.section-title:not(.workspace-tab-heading__title) {
  margin: 10px 0 0;
  font-size: 22px;
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
  gap: 18px;
  padding: 16px 0;
  border-bottom: 1px solid var(--workspace-line-soft);
}

.insight-item strong {
  display: block;
  font-size: 15px;
  color: var(--journal-ink);
}

.insight-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.chip,
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 8px;
  border-radius: 7px;
  border: 1px solid var(--workspace-line-soft);
  font-size: 11.5px;
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
  margin-top: 18px;
  padding: 16px 18px;
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
}

.workspace-alert-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.workspace-alert-icon {
  width: 18px;
  height: 18px;
  color: color-mix(in srgb, var(--workspace-danger) 82%, var(--journal-ink));
}

.workspace-alert-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 14px;
}

.empty-inline {
  font-size: 14px;
  line-height: 1.75;
  color: var(--workspace-faint);
}

.hotspot-list {
  display: grid;
  gap: 14px;
}

.hotspot-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
  gap: 18px;
  padding: 18px 0;
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
  gap: 8px 10px;
}

.hotspot-title-row strong {
  font-size: 15px;
  color: var(--journal-ink);
}

.hotspot-copy {
  font-family: var(--workspace-font-mono);
}

.hotspot-stats {
  display: grid;
  gap: 12px;
}

.hotspot-stat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  color: var(--journal-muted);
}

.usage-track {
  margin-top: 8px;
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

.admin-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  border: 1px solid transparent;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  box-shadow: var(--admin-btn-shadow, none);
  transition: all 150ms ease;
}

.admin-btn:focus-visible {
  outline: none;
  box-shadow:
    var(--admin-btn-shadow, none),
    0 0 0 3px color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.admin-btn-primary {
  --admin-btn-shadow: 0 12px 24px color-mix(in srgb, var(--journal-accent) 24%, transparent);
  border-color: color-mix(in srgb, var(--journal-accent) 46%, var(--journal-border));
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-primary:hover {
  background: #1d4ed8;
}

.admin-btn-ghost {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: rgba(37, 99, 235, 0.28);
  color: var(--journal-accent);
}

.admin-action-row {
  --admin-action-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid var(--admin-action-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: 0.95rem 1rem;
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
    padding-top: 20px;
    padding-left: 0;
    border-top: 1px solid var(--workspace-line-soft);
    border-left: 0;
  }
}

@media (max-width: 860px) {
  .progress-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .section-head {
    display: block;
  }
}

@media (max-width: 640px) {
  .workspace-topbar,
  .top-tabs,
  .content-pane {
    padding-left: 18px;
    padding-right: 18px;
  }

  .workspace-topbar {
    display: block;
  }

  .top-note {
    justify-content: flex-start;
    margin-top: 12px;
  }

  .progress-strip {
    grid-template-columns: 1fr;
  }
}
</style>

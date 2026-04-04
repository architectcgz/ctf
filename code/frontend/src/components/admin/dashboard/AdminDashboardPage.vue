<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, ArrowRight, ShieldAlert, SquareStack } from 'lucide-vue-next'

import type { AdminDashboardData } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'

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

const alertCount = computed(() => props.dashboard?.alerts.length ?? 0)
const healthSummary = computed(() => {
  const cpu = props.dashboard?.cpu_usage ?? 0
  const memory = props.dashboard?.memory_usage ?? 0
  if (alertCount.value > 0 || cpu >= 90 || memory >= 90) return { label: '高风险', accent: 'danger' as const }
  if (cpu >= 75 || memory >= 75) return { label: '需要关注', accent: 'warning' as const }
  return { label: '运行稳定', accent: 'success' as const }
})

const quickSignals = computed(() => [
  { label: '在线用户', value: props.dashboard?.online_users ?? 0, helper: '当前在线账号', accent: 'primary' as const },
  { label: '活跃容器', value: props.dashboard?.active_containers ?? 0, helper: '正在运行的实例', accent: 'success' as const },
  { label: '平均 CPU', value: formatPercent(props.dashboard?.cpu_usage), helper: '当前资源水位', accent: healthSummary.value.accent },
  { label: '平均内存', value: formatPercent(props.dashboard?.memory_usage), helper: '结合阈值判断回收', accent: healthSummary.value.accent },
])

const sortedContainers = computed(() =>
  [...(props.dashboard?.container_stats ?? [])].sort((left, right) => {
    const leftPeak = Math.max(left.cpu_percent ?? 0, left.memory_percent ?? 0)
    const rightPeak = Math.max(right.cpu_percent ?? 0, right.memory_percent ?? 0)
    return rightPeak - leftPeak
  }),
)

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
  <section class="journal-shell journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div class="journal-eyebrow">Admin Console</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            系统值守台
          </h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里查看平台状态、异常和当前资源热点。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="admin-btn admin-btn-primary" @click="emit('openAuditLog')">
              审计日志
            </button>
            <button type="button" class="admin-btn admin-btn-ghost" @click="emit('openCheatDetection')">
              风险研判
            </button>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center justify-between gap-3">
            <div>
              <div class="journal-note-label">当前状态</div>
              <div class="mt-2 text-2xl font-semibold text-[var(--journal-ink)]">
                {{ healthSummary.label }}
              </div>
              <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                当前共有 {{ alertCount }} 条需要处理的资源告警。
              </p>
            </div>
            <div class="journal-brief-icon">
              <ShieldAlert class="h-5 w-5" />
            </div>
          </div>

          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div v-for="item in quickSignals" :key="item.label" class="journal-note">
              <div class="journal-note-label">{{ item.label }}</div>
              <div class="journal-note-value">{{ item.value }}</div>
              <div class="journal-note-helper">{{ item.helper }}</div>
            </div>
          </div>
        </article>
      </div>
      <div class="journal-divider mt-6" />

      <div v-if="error" class="admin-feedback admin-feedback-danger">
        {{ error }}
        <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
      </div>

      <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div v-for="index in 4" :key="index" class="h-28 animate-pulse rounded-[18px] bg-[var(--journal-surface)]" />
      </div>

      <template v-else-if="dashboard">
        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <AppCard
            v-for="item in quickSignals"
            :key="item.label"
            variant="metric"
            :accent="item.accent"
            :eyebrow="item.label"
            :title="String(item.value)"
            :subtitle="item.helper"
          />
        </div>

        <div class="journal-divider mt-6" />

        <section class="space-y-4">
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Alert Stack</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">当前告警</h2>
            </div>
            <div class="admin-pill">
              <AlertTriangle class="h-4 w-4" />
              {{ alertCount }} 条
            </div>
          </div>

          <div v-if="alertCount === 0" class="admin-empty">
            当前没有资源告警。
          </div>

          <div v-else class="space-y-3">
            <AppCard
              v-for="alert in dashboard.alerts"
              :key="`${alert.container_id}-${alert.type}`"
              variant="action"
              accent="danger"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                    <AlertTriangle class="h-4 w-4 text-[var(--color-danger)]" />
                    {{ alert.container_id }}
                  </div>
                  <p class="mt-2 text-sm leading-6 text-text-secondary">{{ alert.message }}</p>
                </div>
                <span class="admin-tag admin-tag-danger">
                  {{ alert.type.toUpperCase() }}
                </span>
              </div>
              <div class="mt-4 text-xs text-text-secondary">
                当前 {{ Math.round(alert.value) }}% / 阈值 {{ Math.round(alert.threshold) }}%
              </div>
            </AppCard>
          </div>

          <div class="grid gap-3 lg:grid-cols-2">
            <button type="button" class="admin-action-row" @click="emit('openCheatDetection')">
              <div>
                <div class="text-sm font-medium text-text-primary">进入风险研判</div>
                <div class="mt-1 text-sm text-text-secondary">先看异常行为，再判断是否需要深挖容器与账号。</div>
              </div>
              <ArrowRight class="h-4 w-4 text-primary" />
            </button>
            <button type="button" class="admin-action-row" @click="emit('openAuditLog')">
              <div>
                <div class="text-sm font-medium text-text-primary">查看审计日志</div>
                <div class="mt-1 text-sm text-text-secondary">结合操作记录，快速定位异常来源。</div>
              </div>
              <ArrowRight class="h-4 w-4 text-primary" />
            </button>
          </div>
        </section>

        <div class="journal-divider mt-6" />

        <section class="space-y-4">
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Resource Hotspots</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">资源热点</h2>
            </div>
          </div>

          <div v-if="sortedContainers.length === 0" class="admin-empty">
            暂无容器运行数据。
          </div>

          <div v-else class="grid gap-4">
            <AppCard
              v-for="item in sortedContainers"
              :key="item.container_id"
              variant="action"
              accent="neutral"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div class="flex items-center gap-2 text-base font-semibold text-text-primary">
                    <SquareStack class="h-4 w-4 text-primary" />
                    {{ item.container_name || item.container_id }}
                  </div>
                  <div class="mt-2 font-mono text-xs text-text-secondary">{{ item.container_id }}</div>
                </div>
                <div class="text-right text-sm text-text-secondary">
                  <div>{{ formatBytes(item.memory_usage) }} / {{ formatBytes(item.memory_limit) }}</div>
                  <div class="mt-1 text-xs uppercase tracking-[0.16em] text-text-muted">内存用量</div>
                </div>
              </div>

              <div class="mt-5 grid gap-4 md:grid-cols-2">
                <div>
                  <div class="flex items-center justify-between gap-3 text-sm">
                    <span class="text-text-secondary">CPU</span>
                    <span class="font-medium text-text-primary">{{ formatPercent(item.cpu_percent) }}</span>
                  </div>
                  <div class="mt-2 h-2.5 overflow-hidden rounded-full bg-[var(--color-bg-base)]">
                    <div class="h-full rounded-full" :class="usageTone(item.cpu_percent)" :style="{ width: `${Math.round(item.cpu_percent ?? 0)}%` }" />
                  </div>
                </div>
                <div>
                  <div class="flex items-center justify-between gap-3 text-sm">
                    <span class="text-text-secondary">内存</span>
                    <span class="font-medium text-text-primary">{{ formatPercent(item.memory_percent) }}</span>
                  </div>
                  <div class="mt-2 h-2.5 overflow-hidden rounded-full bg-[var(--color-bg-base)]">
                    <div class="h-full rounded-full" :class="usageTone(item.memory_percent)" :style="{ width: `${Math.round(item.memory_percent ?? 0)}%` }" />
                  </div>
                </div>
              </div>
            </AppCard>
          </div>
        </section>
      </template>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #2563eb;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.journal-brief-icon {
  display: inline-flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.1);
  color: var(--journal-accent);
}

.journal-divider {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
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
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease,
    transform 150ms ease;
}

.admin-btn:focus-visible {
  outline: none;
  border-color: color-mix(in srgb, var(--journal-accent) 52%, var(--journal-border));
  box-shadow:
    0 0 0 3px color-mix(in srgb, var(--journal-accent) 16%, transparent),
    var(--admin-btn-shadow, none);
}

.admin-btn-primary {
  --admin-btn-shadow: 0 12px 24px color-mix(in srgb, var(--journal-accent) 24%, transparent);
  border-color: color-mix(in srgb, var(--journal-accent) 46%, var(--journal-border));
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 88%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-accent) 76%, var(--journal-surface-subtle))
  );
  color: #eff6ff;
}

.admin-btn-primary:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 58%, var(--journal-border));
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 94%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-accent) 82%, var(--journal-surface-subtle))
  );
  transform: translateY(-1px);
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: rgba(37, 99, 235, 0.28);
  color: var(--journal-accent);
  transform: translateY(-1px);
}

.admin-feedback {
  margin-bottom: 1rem;
  border-radius: 1rem;
  padding: 0.9rem 1rem;
  font-size: 0.875rem;
}

.admin-feedback-danger {
  border: 1px solid rgba(239, 68, 68, 0.2);
  background: rgba(254, 242, 242, 0.9);
  color: #b91c1c;
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 999px;
  border: 1px solid rgba(37, 99, 235, 0.16);
  background: rgba(37, 99, 235, 0.06);
  padding: 0.48rem 0.9rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.admin-tag {
  border-radius: 999px;
  padding: 0.3rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
}

.admin-tag-danger {
  border: 1px solid rgba(239, 68, 68, 0.16);
  background: rgba(239, 68, 68, 0.08);
  color: #dc2626;
}

.admin-empty {
  border: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  border-radius: 16px;
  padding: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

.admin-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid rgba(226, 232, 240, 0.86);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 0.95rem 1rem;
  text-align: left;
  transition: border-color 150ms ease, background-color 150ms ease;
}

.admin-action-row:hover {
  border-color: rgba(37, 99, 235, 0.24);
  background: var(--journal-surface);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #60a5fa;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}

@media (max-width: 767px) {
  .journal-hero {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>

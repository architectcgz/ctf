<script setup lang="ts">
import { computed } from 'vue'
import { Activity, AlertTriangle, ArrowRight, ShieldAlert, Siren, SquareStack } from 'lucide-vue-next'

import type { AdminDashboardData } from '@/api/contracts'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

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
  {
    label: '系统健康',
    value: healthSummary.value.label,
    description: `当前有 ${alertCount.value} 条资源告警`,
    icon: ShieldAlert,
  },
  {
    label: 'CPU 水位',
    value: formatPercent(props.dashboard?.cpu_usage),
    description: '用于判断容器资源是否接近瓶颈',
    icon: Activity,
  },
  {
    label: '告警态势',
    value: `${alertCount.value} 条`,
    description: '优先处理持续高于阈值的容器',
    icon: AlertTriangle,
  },
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
  <div class="space-y-6">
    <PageHeader
      eyebrow="Control Plane"
      title="系统值守台"
      description="这里不再是通用后台概览，而是围绕平台风险、资源热点和立即动作单独设计的管理员值守页面。"
    >
      <ElButton plain @click="emit('openAuditLog')">审计日志</ElButton>
      <ElButton type="primary" @click="emit('openCheatDetection')">风险研判</ElButton>
    </PageHeader>

    <section class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <div class="overflow-hidden rounded-[30px] border border-emerald-500/20 bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.16),transparent_42%),linear-gradient(145deg,rgba(2,6,23,0.98),rgba(15,23,42,0.92))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.24em] text-emerald-100/80">
          <span>Operations Pulse</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">状态：{{ healthSummary.label }}</span>
        </div>
        <h2 class="mt-4 text-3xl font-semibold tracking-tight text-white">当前平台运行{{ healthSummary.label }}</h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-slate-200/78">
          这个页面只围绕“值守”展开。先判断平台是否稳定，再沿着资源热点和告警卡片往下看，而不是继续使用普通表格式后台总览。
        </p>
        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-emerald-100/60">在线用户</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ dashboard?.online_users ?? 0 }}</div>
            <div class="mt-2 text-sm text-slate-200/70">当前仍在平台活动的用户数</div>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-emerald-100/60">活跃容器</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ dashboard?.active_containers ?? 0 }}</div>
            <div class="mt-2 text-sm text-slate-200/70">正在运行的靶场与竞赛容器</div>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-emerald-100/60">资源告警</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ alertCount }}</div>
            <div class="mt-2 text-sm text-slate-200/70">需要管理员优先处理的异常数量</div>
          </div>
        </div>
      </div>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <article
          v-for="item in quickSignals"
          :key="item.label"
          class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]"
        >
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">{{ item.label }}</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ item.value }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <component :is="item.icon" class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">{{ item.description }}</div>
        </article>
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div v-for="index in 4" :key="index" class="h-32 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]" />
    </div>

    <template v-else-if="dashboard">
      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="在线用户" :value="dashboard.online_users" hint="当前仍在平台内活动的用户数" accent="primary" />
        <MetricCard label="活跃容器" :value="dashboard.active_containers" hint="处于运行状态的靶场与竞赛容器总数" accent="success" />
        <MetricCard label="平均 CPU" :value="formatPercent(dashboard.cpu_usage)" hint="超过 75% 时建议重点关注" :accent="healthSummary.accent" />
        <MetricCard label="平均内存" :value="formatPercent(dashboard.memory_usage)" hint="结合容器上限判断是否需要回收资源" :accent="healthSummary.accent" />
      </section>

      <section class="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
        <div class="space-y-6">
          <SectionCard title="告警栈" subtitle="把当前超过阈值的异常按卡片堆叠，方便值守时逐条处理。">
            <template #header>
              <span class="rounded-full bg-red-500/12 px-3 py-1 text-xs font-semibold text-red-600">{{ alertCount }} 条</span>
            </template>

            <div v-if="alertCount === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
              当前没有资源告警。
            </div>

            <div v-else class="space-y-3">
              <article
                v-for="alert in dashboard.alerts"
                :key="`${alert.container_id}-${alert.type}`"
                class="rounded-[24px] border border-red-500/20 bg-red-500/8 px-5 py-5"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                      <Siren class="h-4 w-4 text-red-400" />
                      {{ alert.container_id }}
                    </div>
                    <p class="mt-2 text-sm leading-6 text-text-secondary">{{ alert.message }}</p>
                  </div>
                  <span class="rounded-full border border-red-500/20 bg-red-500/10 px-3 py-1 text-xs font-semibold text-red-300">
                    {{ alert.type.toUpperCase() }}
                  </span>
                </div>
                <div class="mt-4 text-xs uppercase tracking-[0.16em] text-red-200/80">
                  当前 {{ Math.round(alert.value) }}% / 阈值 {{ Math.round(alert.threshold) }}%
                </div>
              </article>
            </div>
          </SectionCard>

          <SectionCard title="立即动作" subtitle="值守时最常见的两条下一步。">
            <div class="grid gap-3">
              <button
                type="button"
                class="flex items-center justify-between rounded-[24px] border border-border bg-base/70 px-4 py-4 text-left transition hover:border-primary/60"
                @click="emit('openCheatDetection')"
              >
                <div>
                  <div class="text-sm font-medium text-text-primary">进入风险研判</div>
                  <div class="mt-1 text-sm text-text-secondary">当资源和异常都开始上升时，先确认是否伴随异常操作模式。</div>
                </div>
                <ArrowRight class="h-4 w-4 text-primary" />
              </button>
              <button
                type="button"
                class="flex items-center justify-between rounded-[24px] border border-border bg-base/70 px-4 py-4 text-left transition hover:border-primary/60"
                @click="emit('openAuditLog')"
              >
                <div>
                  <div class="text-sm font-medium text-text-primary">查看审计日志</div>
                  <div class="mt-1 text-sm text-text-secondary">用于追踪高负载容器背后的管理动作和访问行为。</div>
                </div>
                <ArrowRight class="h-4 w-4 text-primary" />
              </button>
            </div>
          </SectionCard>
        </div>

        <SectionCard title="资源热点" subtitle="把容器按热点程度重排，不再继续使用普通资源表格。">
          <div v-if="sortedContainers.length === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
            暂无容器运行数据。
          </div>

          <div v-else class="grid gap-4">
            <article
              v-for="item in sortedContainers"
              :key="item.container_id"
              class="rounded-[24px] border border-border bg-[linear-gradient(180deg,rgba(15,23,42,0.9),rgba(8,15,32,0.78))] px-5 py-5"
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
            </article>
          </div>
        </SectionCard>
      </section>
    </template>
  </div>
</template>

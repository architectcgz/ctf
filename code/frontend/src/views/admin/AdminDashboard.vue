<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { getDashboard } from '@/api/admin'
import type { AdminDashboardData } from '@/api/contracts'

const loading = ref(false)
const error = ref<string | null>(null)
const dashboard = ref<AdminDashboardData | null>(null)

const alertCount = computed(() => dashboard.value?.alerts.length ?? 0)

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

async function loadDashboard(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    dashboard.value = await getDashboard()
  } catch (err) {
    console.error('加载系统概览失败:', err)
    error.value = '加载系统概览失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<template>
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(15,23,42,0.08),rgba(8,145,178,0.15))] p-7 shadow-sm">
      <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--color-primary)]/85">Admin Console</p>
      <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">系统概览</h1>
      <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
        聚合在线用户、容器资源和告警状态，帮助管理员快速判断平台当前健康度。
      </p>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="loadDashboard">重试</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div v-for="index in 4" :key="index" class="h-32 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"></div>
    </div>

    <template v-else-if="dashboard">
      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">在线用户</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ dashboard.online_users }}</p>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">活跃容器</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ dashboard.active_containers }}</p>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">平均 CPU</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-primary)]">{{ formatPercent(dashboard.cpu_usage) }}</p>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">平均内存</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ formatPercent(dashboard.memory_usage) }}</p>
        </div>
      </section>

      <section class="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">资源告警</h2>
              <p class="mt-1 text-sm text-[var(--color-text-secondary)]">当前超过阈值的容器资源异常。</p>
            </div>
            <span class="rounded-full bg-red-500/12 px-3 py-1 text-xs font-semibold text-red-600">{{ alertCount }} 条</span>
          </div>

          <div v-if="alertCount === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
            当前没有资源告警。
          </div>

          <div v-else class="mt-5 space-y-3">
            <div
              v-for="alert in dashboard.alerts"
              :key="`${alert.container_id}-${alert.type}`"
              class="rounded-xl border border-red-200 bg-red-50 px-4 py-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="font-medium text-red-700">{{ alert.container_id }}</p>
                  <p class="mt-1 text-sm text-red-600">{{ alert.message }}</p>
                </div>
                <span class="rounded-full bg-white px-3 py-1 text-xs font-semibold text-red-600">
                  {{ alert.type.toUpperCase() }}
                </span>
              </div>
              <p class="mt-2 text-xs text-red-500">
                当前 {{ Math.round(alert.value) }}%，阈值 {{ Math.round(alert.threshold) }}%
              </p>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">容器资源明细</h2>
              <p class="mt-1 text-sm text-[var(--color-text-secondary)]">观察热点容器的 CPU 和内存占用。</p>
            </div>
          </div>

          <div v-if="dashboard.container_stats.length === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
            暂无容器运行数据。
          </div>

          <div v-else class="mt-5 overflow-hidden rounded-xl border border-[var(--color-border-default)]">
            <table class="min-w-full divide-y divide-[var(--color-border-default)] text-sm">
              <thead class="bg-[var(--color-bg-base)]">
                <tr>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">容器</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">CPU</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">内存</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">内存用量</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-[var(--color-border-default)] bg-[var(--color-bg-surface)]">
                <tr v-for="item in dashboard.container_stats" :key="item.container_id">
                  <td class="px-4 py-3">
                    <p class="font-medium text-[var(--color-text-primary)]">{{ item.container_name || item.container_id }}</p>
                    <p class="mt-1 font-mono text-xs text-[var(--color-text-secondary)]">{{ item.container_id }}</p>
                  </td>
                  <td class="px-4 py-3 text-[var(--color-text-primary)]">{{ formatPercent(item.cpu_percent) }}</td>
                  <td class="px-4 py-3 text-[var(--color-text-primary)]">{{ formatPercent(item.memory_percent) }}</td>
                  <td class="px-4 py-3 text-[var(--color-text-secondary)]">
                    {{ formatBytes(item.memory_usage) }} / {{ formatBytes(item.memory_limit) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

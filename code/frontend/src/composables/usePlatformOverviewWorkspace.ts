import { computed, type Ref } from 'vue'

import type { AdminDashboardData } from '@/api/contracts'

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

export function usePlatformOverviewWorkspace(dashboard: Ref<AdminDashboardData | null>) {
  const alertCount = computed(() => dashboard.value?.alerts.length ?? 0)

  const healthSummary = computed(() => {
    const cpu = dashboard.value?.cpu_usage ?? 0
    const memory = dashboard.value?.memory_usage ?? 0
    if (alertCount.value > 0 || cpu >= 90 || memory >= 90) {
      return { label: '高风险', accent: 'danger' as const }
    }
    if (cpu >= 75 || memory >= 75) {
      return { label: '需要关注', accent: 'warning' as const }
    }
    return { label: '运行稳定', accent: 'success' as const }
  })

  const quickSignals = computed(() => [
    {
      label: '在线用户',
      value: dashboard.value?.online_users ?? 0,
      helper: '当前在线账号',
      accent: 'primary' as const,
    },
    {
      label: '活跃容器',
      value: dashboard.value?.active_containers ?? 0,
      helper: '正在运行的实例',
      accent: 'success' as const,
    },
    {
      label: '平均 CPU',
      value: formatPercent(dashboard.value?.cpu_usage),
      helper: '当前资源水位',
      accent: healthSummary.value.accent,
    },
    {
      label: '平均内存',
      value: formatPercent(dashboard.value?.memory_usage),
      helper: '结合阈值判断回收',
      accent: healthSummary.value.accent,
    },
  ])

  const sortedContainers = computed(() =>
    [...(dashboard.value?.container_stats ?? [])].sort((left, right) => {
      const leftPeak = Math.max(left.cpu_percent ?? 0, left.memory_percent ?? 0)
      const rightPeak = Math.max(right.cpu_percent ?? 0, right.memory_percent ?? 0)
      return rightPeak - leftPeak
    })
  )

  const metaPills = computed(() => [
    'Admin Workspace',
    healthSummary.value.label,
    alertCount.value > 0 ? `${alertCount.value} 条资源告警` : '暂无资源告警',
    `活跃容器 ${dashboard.value?.active_containers ?? 0} 个`,
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
    String(
      Math.round(Math.max(dashboard.value?.cpu_usage ?? 0, dashboard.value?.memory_usage ?? 0))
    )
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

  return {
    alertCount,
    healthSummary,
    sortedContainers,
    metaPills,
    overviewMetrics,
    railScore,
    railCopy,
    formatPercent,
    formatBytes,
    usageTone,
  }
}

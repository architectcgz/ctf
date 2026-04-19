import { computed, ref, watch, type Ref } from 'vue'

import type { AWDTrafficStatusGroup, AWDTrafficSummaryData } from '@/api/contracts'

export interface AWDTrafficFilters {
  attacker_team_id: string
  victim_team_id: string
  service_id: string
  challenge_id: string
  status_group: 'all' | AWDTrafficStatusGroup
  path_keyword: string
  page: number
  page_size: number
}

interface UseAwdTrafficPanelOptions {
  trafficSummary: Ref<AWDTrafficSummaryData | null>
  trafficEventsTotal: Ref<number>
  trafficFilters: Ref<AWDTrafficFilters>
  loadingTrafficEvents: Ref<boolean>
  trafficPathKeyword: Readonly<Ref<string>>
  formatDateTime: (value?: string) => string
  formatPercent: (value: number) => string
  applyTrafficFilters: (patch: Partial<AWDTrafficFilters>) => void
  changeTrafficPage: (page: number) => void
}

export function useAwdTrafficPanel({
  trafficSummary,
  trafficEventsTotal,
  trafficFilters,
  loadingTrafficEvents,
  trafficPathKeyword,
  formatDateTime,
  formatPercent,
  applyTrafficFilters,
  changeTrafficPage,
}: UseAwdTrafficPanelOptions) {
  const trafficPathKeywordInput = ref('')

  watch(
    trafficPathKeyword,
    (keyword) => {
      trafficPathKeywordInput.value = keyword
    },
    { immediate: true }
  )

  const trafficErrorRate = computed(() => {
    if (!trafficSummary.value || trafficSummary.value.total_request_count <= 0) {
      return 0
    }
    return (
      (trafficSummary.value.error_request_count / trafficSummary.value.total_request_count) * 100
    )
  })

  const trafficTotalPages = computed(() =>
    Math.max(1, Math.ceil(trafficEventsTotal.value / Math.max(trafficFilters.value.page_size, 1)))
  )

  const trafficTrendRows = computed(() => {
    const buckets = trafficSummary.value?.trend_buckets || []
    const peak = Math.max(1, ...buckets.map((item) => item.request_count))
    return buckets.map((item) => ({
      ...item,
      ratio: Math.max(6, Math.round((item.request_count / peak) * 100)),
      label: new Date(item.bucket_start_at).toLocaleTimeString('zh-CN', {
        hour: '2-digit',
        minute: '2-digit',
      }),
    }))
  })

  const trafficSummaryStats = computed(() => {
    const summary = trafficSummary.value
    if (!summary) {
      return []
    }

    return [
      {
        key: 'requests',
        label: '代理请求总量',
        value: String(summary.total_request_count),
        hint: summary.latest_event_at
          ? `最近请求 ${formatDateTime(summary.latest_event_at)}`
          : '当前轮尚未记录最近请求时间',
      },
      {
        key: 'attackers',
        label: '活跃攻击队',
        value: String(summary.active_attacker_team_count),
        hint: `热点攻击队 ${summary.top_attackers[0]?.team_name || '暂无'}`,
      },
      {
        key: 'victims',
        label: '被攻击目标队',
        value: String(summary.victim_team_count),
        hint: `热点目标队 ${summary.top_victims[0]?.team_name || '暂无'}`,
      },
      {
        key: 'paths',
        label: '唯一路径数',
        value: String(summary.unique_path_count),
        hint: `异常路径 ${summary.top_error_paths[0]?.path || '暂无'}`,
      },
      {
        key: 'errors',
        label: '错误请求率',
        value: formatPercent(trafficErrorRate.value),
        hint: `错误请求 ${summary.error_request_count}`,
      },
    ]
  })

  const trafficTrendNarrative = computed(() => {
    if (trafficTrendRows.value.length === 0) {
      return '当前轮尚未形成趋势桶。'
    }
    const peakBucket = [...trafficTrendRows.value].sort((left, right) => {
      if (left.request_count !== right.request_count) {
        return right.request_count - left.request_count
      }
      return right.error_count - left.error_count
    })[0]
    return `${peakBucket.label} 请求最高，共 ${peakBucket.request_count}，错误 ${peakBucket.error_count}。`
  })

  const trafficStatusGroupOptions: Array<{ value: 'all' | AWDTrafficStatusGroup; label: string }> =
    [
      { value: 'all', label: '全部状态' },
      { value: 'success', label: '成功' },
      { value: 'redirect', label: '重定向' },
      { value: 'client_error', label: '客户端错误' },
      { value: 'server_error', label: '服务端错误' },
    ]

  function applyTrafficFilterPatch(patch: Partial<AWDTrafficFilters>): void {
    applyTrafficFilters(patch)
  }

  function applyTrafficKeywordFilter(): void {
    applyTrafficFilterPatch({ path_keyword: trafficPathKeywordInput.value.trim() })
  }

  function onTrafficStatusGroupChange(value: string): void {
    if (
      value !== 'all' &&
      value !== 'success' &&
      value !== 'redirect' &&
      value !== 'client_error' &&
      value !== 'server_error'
    ) {
      return
    }
    applyTrafficFilterPatch({ status_group: value })
  }

  function clearTrafficKeywordFilter(): void {
    trafficPathKeywordInput.value = ''
    applyTrafficFilterPatch({ path_keyword: '' })
  }

  function handleTrafficPageChange(targetPage: number): void {
    if (
      loadingTrafficEvents.value ||
      targetPage < 1 ||
      targetPage > trafficTotalPages.value ||
      targetPage === trafficFilters.value.page
    ) {
      return
    }
    changeTrafficPage(targetPage)
  }

  return {
    trafficPathKeywordInput,
    trafficErrorRate,
    trafficTotalPages,
    trafficTrendRows,
    trafficSummaryStats,
    trafficTrendNarrative,
    trafficStatusGroupOptions,
    applyTrafficKeywordFilter,
    onTrafficStatusGroupChange,
    clearTrafficKeywordFilter,
    applyTrafficFilterPatch,
    handleTrafficPageChange,
  }
}

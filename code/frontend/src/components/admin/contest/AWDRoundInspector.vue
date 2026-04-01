<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RefreshCw, Search, ShieldAlert, ShieldCheck, Sword, TimerReset } from 'lucide-vue-next'

import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficStatusGroup,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { downloadCSVFile, downloadJSONFile } from '@/utils/csv'

interface AWDRoundInspectorFilterState {
  service_team_id: string
  service_status: 'all' | AWDTeamServiceData['service_status']
  service_check_source: string
  service_alert_reason: string
  attack_team_id: string
  attack_result: 'all' | 'success' | 'failed'
  attack_source: 'all' | AWDAttackLogData['source']
}

interface AWDProbeAttemptView {
  probe: string
  healthy: boolean
  latency_ms?: number
  error_code?: string
  error?: string
}

interface AWDProbeTargetView {
  access_url?: string
  healthy: boolean
  probe?: string
  latency_ms?: number
  error_code?: string
  error?: string
  attempts: AWDProbeAttemptView[]
}

interface AWDServiceAlertView {
  key: string
  label: string
  count: number
  affected_teams: string[]
  samples: Array<{
    service_id: string
    team_name: string
    challenge_title: string
  }>
}

interface AWDTrafficFilters {
  attacker_team_id: string
  victim_team_id: string
  challenge_id: string
  status_group: 'all' | AWDTrafficStatusGroup
  path_keyword: string
  page: number
  page_size: number
}

function getInspectorFilterStorageKey(contestId: string, roundId: string): string {
  return `ctf_admin_awd_filters:${contestId}:${roundId}`
}

function loadInspectorFilterState(contestId: string, roundId: string): AWDRoundInspectorFilterState | null {
  if (typeof window === 'undefined') {
    return null
  }
  const raw = window.sessionStorage.getItem(getInspectorFilterStorageKey(contestId, roundId))
  if (!raw) {
    return null
  }
  try {
    const parsed = JSON.parse(raw) as Partial<AWDRoundInspectorFilterState>
    return {
      service_team_id: typeof parsed.service_team_id === 'string' ? parsed.service_team_id : '',
      service_status:
        parsed.service_status === 'up' || parsed.service_status === 'down' || parsed.service_status === 'compromised'
          ? parsed.service_status
          : 'all',
      service_check_source:
        typeof parsed.service_check_source === 'string' ? parsed.service_check_source : '',
      service_alert_reason:
        typeof parsed.service_alert_reason === 'string' ? parsed.service_alert_reason : '',
      attack_team_id: typeof parsed.attack_team_id === 'string' ? parsed.attack_team_id : '',
      attack_result:
        parsed.attack_result === 'success' || parsed.attack_result === 'failed' ? parsed.attack_result : 'all',
      attack_source:
        parsed.attack_source === 'manual_attack_log' ||
        parsed.attack_source === 'submission' ||
        parsed.attack_source === 'legacy'
          ? parsed.attack_source
          : 'all',
    }
  } catch {
    return null
  }
}

function persistInspectorFilterState(
  contestId: string,
  roundId: string,
  state: AWDRoundInspectorFilterState
): void {
  if (typeof window === 'undefined') {
    return
  }
  window.sessionStorage.setItem(getInspectorFilterStorageKey(contestId, roundId), JSON.stringify(state))
}

const props = defineProps<{
  contest: ContestDetailData
  rounds: AWDRoundData[]
  selectedRoundId: string | null
  services: AWDTeamServiceData[]
  attacks: AWDAttackLogData[]
  challengeLinks: AdminContestChallengeData[]
  summary: AWDRoundSummaryData | null
  trafficSummary: AWDTrafficSummaryData | null
  trafficEvents: AWDTrafficEventData[]
  trafficEventsTotal: number
  trafficFilters: AWDTrafficFilters
  scoreboardRows: ScoreboardRow[]
  scoreboardFrozen: boolean
  loadingRounds: boolean
  loadingRoundDetail: boolean
  loadingTrafficSummary: boolean
  loadingTrafficEvents: boolean
  checking: boolean
  shouldAutoRefresh: boolean
  canRecordServiceChecks: boolean
  canRecordAttackLogs: boolean
  serviceCheckHint?: string
  attackLogHint?: string
}>()

const emit = defineEmits<{
  refresh: []
  openCreateRoundDialog: []
  openServiceCheckDialog: []
  openAttackLogDialog: []
  runSelectedRoundCheck: []
  applyTrafficFilters: [filters: Partial<AWDTrafficFilters>]
  changeTrafficPage: [page: number]
  resetTrafficFilters: []
  'update:selectedRoundId': [roundId: string]
}>()

const selectedRound = computed(
  () => props.rounds.find((item) => item.id === props.selectedRoundId) || null
)
const summaryMetrics = computed(() => props.summary?.metrics || null)
const totalServiceCount = computed(
  () => summaryMetrics.value?.total_service_count ?? props.services.length
)
const totalAttackCount = computed(
  () => summaryMetrics.value?.total_attack_count ?? props.attacks.length
)
const upCount = computed(
  () =>
    summaryMetrics.value?.service_up_count ??
    props.services.filter((item) => item.service_status === 'up').length
)
const compromisedCount = computed(
  () =>
    summaryMetrics.value?.service_compromised_count ??
    props.services.filter((item) => item.service_status === 'compromised').length
)
const downCount = computed(
  () =>
    summaryMetrics.value?.service_down_count ??
    props.services.filter((item) => item.service_status === 'down').length
)
const successfulAttackCount = computed(
  () =>
    summaryMetrics.value?.successful_attack_count ??
    props.attacks.filter((item) => item.is_success).length
)
const failedAttackCount = computed(
  () =>
    summaryMetrics.value?.failed_attack_count ??
    props.attacks.filter((item) => !item.is_success).length
)
const attackedServiceCount = computed(
  () =>
    summaryMetrics.value?.attacked_service_count ??
    props.services.filter((item) => item.attack_received > 0).length
)
const defenseSuccessCount = computed(
  () =>
    summaryMetrics.value?.defense_success_count ??
    props.services.filter((item) => item.attack_received > 0 && item.service_status === 'up').length
)
const manualCheckCount = computed(() => {
  if (summaryMetrics.value) {
    return (
      summaryMetrics.value.manual_current_round_check_count +
      summaryMetrics.value.manual_selected_round_check_count +
      summaryMetrics.value.manual_service_check_count
    )
  }
  return props.services.filter((item) => {
    const source = getServiceCheckSourceValue(item.check_result)
    return (
      source === 'manual_current_round' ||
      source === 'manual_selected_round' ||
      source === 'manual_service_check'
    )
  }).length
})
const serviceTeamFilter = ref('')
const serviceStatusFilter = ref<'all' | AWDTeamServiceData['service_status']>('all')
const serviceCheckSourceFilter = ref('')
const serviceAlertReasonFilter = ref('')
const attackTeamFilter = ref('')
const attackResultFilter = ref<'all' | 'success' | 'failed'>('all')
const attackSourceFilter = ref<'all' | AWDAttackLogData['source']>('all')
const trafficPathKeywordInput = ref('')

const trafficErrorRate = computed(() => {
  if (!props.trafficSummary || props.trafficSummary.total_request_count <= 0) {
    return 0
  }
  return (props.trafficSummary.error_request_count / props.trafficSummary.total_request_count) * 100
})

const trafficTotalPages = computed(() =>
  Math.max(1, Math.ceil(props.trafficEventsTotal / Math.max(props.trafficFilters.page_size, 1)))
)
const canGotoTrafficPrevPage = computed(() => props.trafficFilters.page > 1)
const canGotoTrafficNextPage = computed(() => props.trafficFilters.page < trafficTotalPages.value)

function formatDateTime(value?: string): string {
  if (!value) {
    return '未记录'
  }
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function getRoundStatusLabel(status: AWDRoundData['status']): string {
  const labels: Record<AWDRoundData['status'], string> = {
    pending: '待开始',
    running: '进行中',
    finished: '已结束',
  }
  return labels[status]
}

function getRoundStatusClass(status: AWDRoundData['status']): string {
  const classes: Record<AWDRoundData['status'], string> = {
    pending: 'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20',
    running: 'bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20',
    finished: 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-secondary)] border border-[var(--color-text-muted)]/20',
  }
  return classes[status]
}

function getServiceStatusLabel(status: AWDTeamServiceData['service_status']): string {
  const labels: Record<AWDTeamServiceData['service_status'], string> = {
    up: '正常',
    down: '下线',
    compromised: '已失陷',
  }
  return labels[status]
}

function getServiceStatusClass(status: AWDTeamServiceData['service_status']): string {
  const classes: Record<AWDTeamServiceData['service_status'], string> = {
    up: 'bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20',
    down: 'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20',
    compromised: 'bg-[var(--color-danger)]/10 text-[var(--color-danger)] border border-[var(--color-danger)]/20',
  }
  return classes[status]
}

function getAttackTypeLabel(type: AWDAttackLogData['attack_type']): string {
  return type === 'service_exploit' ? '服务利用' : 'Flag 获取'
}

function getAttackSourceLabel(source: AWDAttackLogData['source']): string {
  const labels: Record<AWDAttackLogData['source'], string> = {
    legacy: '历史记录',
    manual_attack_log: '人工补录',
    submission: '学员提交',
  }
  return labels[source]
}

function formatPercent(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '0%'
  }
  return `${value.toFixed(1)}%`
}

function getTrafficStatusGroupLabel(statusGroup: AWDTrafficStatusGroup): string {
  const labels: Record<AWDTrafficStatusGroup, string> = {
    success: '成功',
    redirect: '重定向',
    client_error: '客户端错误',
    server_error: '服务端错误',
  }
  return labels[statusGroup]
}

function getTrafficStatusGroupClass(statusGroup: AWDTrafficStatusGroup): string {
  switch (statusGroup) {
    case 'success':
      return 'bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20'
    case 'redirect':
      return 'bg-[var(--color-primary)]/10 text-[var(--color-primary)] border border-[var(--color-primary)]/20'
    case 'client_error':
      return 'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20'
    case 'server_error':
      return 'bg-[var(--color-danger)]/10 text-[var(--color-danger)] border border-[var(--color-danger)]/20'
  }
  return 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-secondary)] border border-[var(--color-text-muted)]/20'
}

function getCheckSourceLabel(value: unknown): string {
  switch (value) {
    case 'manual_current_round':
      return '当前轮手动巡检'
    case 'manual_selected_round':
      return '指定轮次重跑'
    case 'manual_service_check':
      return '人工补录'
    case 'scheduler':
      return '调度巡检'
    default:
      return ''
  }
}

function getCheckStatusLabel(value: unknown): string {
  if (typeof value !== 'string' || value.trim() === '') {
    return ''
  }
  const labels: Record<string, string> = {
    healthy: '全部正常',
    partial_available: '部分可用',
    no_running_instances: '无运行实例',
    unexpected_http_status: 'HTTP 状态异常',
    http_request_failed: 'HTTP 请求失败',
    invalid_access_url: '访问地址无效',
    all_probes_failed: '巡检失败',
  }
  return labels[value] || value
}

function summarizeCheckResult(result: Record<string, unknown>): string {
  const sourceLabel = getCheckSourceLabel(result.check_source)
  const statusLabel = getCheckStatusLabel(result.status_reason)
  const checkedAt =
    typeof result.checked_at === 'string' && result.checked_at.trim() !== ''
      ? formatDateTime(result.checked_at)
      : ''

  const entries = [
    sourceLabel ? `来源: ${sourceLabel}` : '',
    statusLabel ? `状态: ${statusLabel}` : '',
    checkedAt ? `时间: ${checkedAt}` : '',
  ].filter(Boolean)

  if (entries.length > 0) {
    return entries.join(' | ')
  }

  const fallbackEntries = Object.entries(result)
    .filter(([, value]) => value !== null && value !== undefined && value !== '')
    .slice(0, 3)
    .map(([key, value]) => `${key}: ${String(value)}`)

  return fallbackEntries.length > 0 ? fallbackEntries.join(' | ') : '未返回检查明细'
}

function readProbeAttempts(value: unknown): AWDProbeAttemptView[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
    .map((item) => ({
      probe: typeof item.probe === 'string' ? item.probe : '',
      healthy: item.healthy === true,
      latency_ms: typeof item.latency_ms === 'number' ? item.latency_ms : undefined,
      error_code: typeof item.error_code === 'string' ? item.error_code : undefined,
      error: typeof item.error === 'string' ? item.error : undefined,
    }))
}

function getCheckTargets(result: Record<string, unknown>): AWDProbeTargetView[] {
  if (!Array.isArray(result.targets)) {
    return []
  }
  return result.targets
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
    .map((item) => ({
      access_url: typeof item.access_url === 'string' ? item.access_url : undefined,
      healthy: item.healthy === true,
      probe: typeof item.probe === 'string' ? item.probe : undefined,
      latency_ms: typeof item.latency_ms === 'number' ? item.latency_ms : undefined,
      error_code: typeof item.error_code === 'string' ? item.error_code : undefined,
      error: typeof item.error === 'string' ? item.error : undefined,
      attempts: readProbeAttempts(item.attempts),
    }))
}

function getTargetProbeSummary(result: Record<string, unknown>): string {
  const targets = getCheckTargets(result)
  if (targets.length === 0) {
    return ''
  }
  const healthyTargets = targets.filter((item) => item.healthy).length
  return `探测目标 ${healthyTargets}/${targets.length} 正常`
}

function getProbeStatusText(healthy: boolean, errorCode?: string, error?: string): string {
  if (healthy) {
    return '探测成功'
  }
  return getCheckStatusLabel(errorCode) || error || '探测失败'
}

function formatLatency(value?: number): string {
  if (typeof value !== 'number' || Number.isNaN(value) || value <= 0) {
    return ''
  }
  return `${Math.round(value)} ms`
}

function getChallengeTitle(challengeId: string): string {
  const matched = props.challengeLinks.find((item) => item.challenge_id === challengeId)
  return matched?.title?.trim() || `Challenge #${challengeId}`
}

function getTrafficTeamName(teamId: string, teamName?: string): string {
  if (teamName && teamName.trim() !== '') {
    return teamName
  }
  return trafficTeamOptions.value.find((item) => item.id === teamId)?.name || `Team #${teamId}`
}

function getTrafficChallengeTitle(challengeId: string, fallbackTitle?: string): string {
  if (fallbackTitle && fallbackTitle.trim() !== '') {
    return fallbackTitle
  }
  return getChallengeTitle(challengeId)
}

function getTrafficSourceLabel(source: string): string {
  if (source === 'proxy_audit' || source === 'runtime_proxy') {
    return '平台代理'
  }
  return source || '未标记'
}

function buildExportFilename(suffix: string): string {
  const title = props.contest.title
    .trim()
    .replace(/[^a-zA-Z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
  const contestPart = title || `contest-${props.contest.id}`
  const roundPart = selectedRound.value?.round_number || 'unknown'
  return `${contestPart}-round-${roundPart}-${suffix}.csv`
}

function getSelectedRoundLabel(): string {
  if (!selectedRound.value) {
    return '未知轮次'
  }
  return `第 ${selectedRound.value.round_number} 轮`
}

function formatScore(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

function getSourceOverviewLabel(): string {
  const metrics = summaryMetrics.value
  if (!metrics) {
    return '等待轮次汇总'
  }
  return `巡检 调度 ${metrics.scheduler_check_count} / 手动 ${manualCheckCount.value}`
}

function getSourceOverviewDescription(): string {
  const metrics = summaryMetrics.value
  if (!metrics) {
    return '攻击日志来源将在轮次汇总返回后展示。'
  }
  return `日志 提交 ${metrics.submission_attack_count} / 人工 ${metrics.manual_attack_log_count} / 历史 ${metrics.legacy_attack_log_count}`
}

function getServiceCheckSourceValue(result: Record<string, unknown>): string {
  return typeof result.check_source === 'string' ? result.check_source : ''
}

const serviceTeamOptions = computed(() => {
  const seen = new Set<string>()
  return props.services.filter((item) => {
    if (seen.has(item.team_id)) {
      return false
    }
    seen.add(item.team_id)
    return true
  })
})

const serviceCheckSourceOptions = computed(() => {
  const seen = new Set<string>()
  return props.services
    .map((item) => getServiceCheckSourceValue(item.check_result))
    .filter((item) => {
      if (!item || seen.has(item)) {
        return false
      }
      seen.add(item)
      return true
    })
})

const baseFilteredServices = computed(() =>
  props.services.filter((item) => {
    if (serviceTeamFilter.value && item.team_id !== serviceTeamFilter.value) {
      return false
    }
    if (serviceStatusFilter.value !== 'all' && item.service_status !== serviceStatusFilter.value) {
      return false
    }
    if (
      serviceCheckSourceFilter.value &&
      getServiceCheckSourceValue(item.check_result) !== serviceCheckSourceFilter.value
    ) {
      return false
    }
    return true
  })
)

const serviceAlerts = computed<AWDServiceAlertView[]>(() => {
  const grouped = new Map<string, AWDServiceAlertView>()
  for (const service of baseFilteredServices.value) {
    const reason = getServiceAlertReason(service)
    if (!reason) {
      continue
    }
    const existing = grouped.get(reason) || {
      key: reason,
      label: getCheckStatusLabel(reason) || reason,
      count: 0,
      affected_teams: [],
      samples: [],
    }
    existing.count += 1
    if (!existing.affected_teams.includes(service.team_name)) {
      existing.affected_teams.push(service.team_name)
    }
    if (existing.samples.length < 3) {
      existing.samples.push({
        service_id: service.id,
        team_name: service.team_name,
        challenge_title: getChallengeTitle(service.challenge_id),
      })
    }
    grouped.set(reason, existing)
  }

  return [...grouped.values()].sort((left, right) => {
    if (left.count !== right.count) {
      return right.count - left.count
    }
    return left.label.localeCompare(right.label, 'zh-CN')
  })
})

const filteredServices = computed(() =>
  baseFilteredServices.value.filter((item) => {
    if (!serviceAlertReasonFilter.value) {
      return true
    }
    return getServiceAlertReason(item) === serviceAlertReasonFilter.value
  })
)

const attackTeamOptions = computed(() => {
  const seen = new Set<string>()
  return props.attacks.flatMap((item) => {
    const entries = [
      { id: item.attacker_team_id, name: item.attacker_team },
      { id: item.victim_team_id, name: item.victim_team },
    ]
    return entries.filter((entry) => {
      if (seen.has(entry.id)) {
        return false
      }
      seen.add(entry.id)
      return true
    })
  })
})

const attackSourceOptions = computed(() => {
  const seen = new Set<AWDAttackLogData['source']>()
  return props.attacks
    .map((item) => item.source)
    .filter((item) => {
      if (seen.has(item)) {
        return false
      }
      seen.add(item)
      return true
    })
})

const trafficTeamOptions = computed(() => {
  const entries = new Map<string, string>()

  for (const service of props.services) {
    entries.set(service.team_id, service.team_name)
  }
  for (const attackTeam of attackTeamOptions.value) {
    entries.set(attackTeam.id, attackTeam.name)
  }
  for (const item of props.trafficSummary?.top_attackers || []) {
    entries.set(item.team_id, item.team_name)
  }
  for (const item of props.trafficSummary?.top_victims || []) {
    entries.set(item.team_id, item.team_name)
  }
  for (const item of props.trafficEvents) {
    if (item.attacker_team_name?.trim()) {
      entries.set(item.attacker_team_id, item.attacker_team_name)
    }
    if (item.victim_team_name?.trim()) {
      entries.set(item.victim_team_id, item.victim_team_name)
    }
  }

  return [...entries.entries()]
    .map(([id, name]) => ({ id, name }))
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
})

const trafficTrendRows = computed(() => {
  const buckets = props.trafficSummary?.trend_buckets || []
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
  const summary = props.trafficSummary
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

const trafficStatusGroupOptions: Array<{ value: 'all' | AWDTrafficStatusGroup; label: string }> = [
  { value: 'all', label: '全部状态' },
  { value: 'success', label: '成功' },
  { value: 'redirect', label: '重定向' },
  { value: 'client_error', label: '客户端错误' },
  { value: 'server_error', label: '服务端错误' },
]

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

function applyTrafficFilterPatch(patch: Partial<AWDTrafficFilters>): void {
  emit('applyTrafficFilters', patch)
}

function moveTrafficPage(delta: number): void {
  const targetPage = props.trafficFilters.page + delta
  if (targetPage < 1 || targetPage > trafficTotalPages.value) {
    return
  }
  emit('changeTrafficPage', targetPage)
}

const filteredAttacks = computed(() =>
  props.attacks.filter((item) => {
    if (
      attackTeamFilter.value &&
      item.attacker_team_id !== attackTeamFilter.value &&
      item.victim_team_id !== attackTeamFilter.value
    ) {
      return false
    }
    if (attackResultFilter.value === 'success' && !item.is_success) {
      return false
    }
    if (attackResultFilter.value === 'failed' && item.is_success) {
      return false
    }
    if (attackSourceFilter.value !== 'all' && item.source !== attackSourceFilter.value) {
      return false
    }
    return true
  })
)
let syncingPersistedFilters = false

function resetFilters() {
  serviceTeamFilter.value = ''
  serviceStatusFilter.value = 'all'
  serviceCheckSourceFilter.value = ''
  serviceAlertReasonFilter.value = ''
  attackTeamFilter.value = ''
  attackResultFilter.value = 'all'
  attackSourceFilter.value = 'all'
}

function getServiceAlertReason(service: AWDTeamServiceData): string {
  if (service.service_status === 'up') {
    return ''
  }
  const errorCode =
    typeof service.check_result.error_code === 'string' ? service.check_result.error_code.trim() : ''
  if (errorCode) {
    return errorCode
  }
  const statusReason =
    typeof service.check_result.status_reason === 'string' ? service.check_result.status_reason.trim() : ''
  if (statusReason && statusReason !== 'healthy') {
    return statusReason
  }
  if (service.service_status === 'compromised') {
    return 'service_compromised'
  }
  if (service.service_status === 'down') {
    return 'service_down'
  }
  return ''
}

function getServiceAlertSubtitle(alert: AWDServiceAlertView): string {
  const teamLabel =
    alert.affected_teams.length === 0 ? '无受影响队伍' : `影响队伍 ${alert.affected_teams.slice(0, 3).join(' / ')}`
  return `${teamLabel}${alert.affected_teams.length > 3 ? ' 等' : ''}`
}

function getServiceAlertClass(alertKey: string): string {
  switch (alertKey) {
    case 'invalid_access_url':
    case 'service_compromised':
      return 'border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 text-[var(--color-danger)]'
    case 'unexpected_http_status':
    case 'http_request_failed':
    case 'all_probes_failed':
      return 'border-[var(--color-warning)]/20 bg-[var(--color-warning)]/10 text-[var(--color-warning)]'
    default:
      return 'border-[var(--color-text-muted)]/20 bg-[var(--color-text-muted)]/10 text-[var(--color-text-primary)]'
  }
}

function getServiceAlertLabel(alertKey: string): string {
  return getCheckStatusLabel(alertKey) || alertKey
}

function applyServiceAlertFilter(alertKey: string): void {
  serviceAlertReasonFilter.value = serviceAlertReasonFilter.value === alertKey ? '' : alertKey
}

function exportFilteredServices() {
  if (filteredServices.value.length === 0) {
    return
  }
  downloadCSVFile(
    buildExportFilename('services'),
    filteredServices.value.map((item) => ({
      赛事: props.contest.title,
      轮次: getSelectedRoundLabel(),
      筛选队伍: serviceTeamFilter.value
        ? serviceTeamOptions.value.find((team) => team.team_id === serviceTeamFilter.value)?.team_name || serviceTeamFilter.value
        : '全部队伍',
      筛选状态:
        serviceStatusFilter.value === 'all' ? '全部状态' : getServiceStatusLabel(serviceStatusFilter.value),
      筛选来源:
        serviceCheckSourceFilter.value
          ? getCheckSourceLabel(serviceCheckSourceFilter.value) || serviceCheckSourceFilter.value
          : '全部来源',
      筛选告警:
        serviceAlertReasonFilter.value ? getServiceAlertLabel(serviceAlertReasonFilter.value) : '全部告警',
      队伍: item.team_name,
      靶题: getChallengeTitle(item.challenge_id),
      服务状态: getServiceStatusLabel(item.service_status),
      巡检来源: getCheckSourceLabel(item.check_result.check_source) || '',
      检查摘要: summarizeCheckResult(item.check_result),
      防守得分: item.defense_score,
      受攻击次数: item.attack_received,
      更新时间: formatDateTime(item.updated_at),
    }))
  )
}

function exportFilteredAttacks() {
  if (filteredAttacks.value.length === 0) {
    return
  }
  downloadCSVFile(
    buildExportFilename('attacks'),
    filteredAttacks.value.map((item) => ({
      赛事: props.contest.title,
      轮次: getSelectedRoundLabel(),
      筛选队伍: attackTeamFilter.value
        ? attackTeamOptions.value.find((team) => team.id === attackTeamFilter.value)?.name || attackTeamFilter.value
        : '全部队伍',
      筛选结果:
        attackResultFilter.value === 'all'
          ? '全部结果'
          : attackResultFilter.value === 'success'
            ? '仅成功'
            : '仅失败',
      筛选来源:
        attackSourceFilter.value === 'all' ? '全部来源' : getAttackSourceLabel(attackSourceFilter.value),
      时间: formatDateTime(item.created_at),
      攻击方: item.attacker_team,
      受害方: item.victim_team,
      靶题: getChallengeTitle(item.challenge_id),
      攻击类型: getAttackTypeLabel(item.attack_type),
      记录来源: getAttackSourceLabel(item.source),
      攻击结果: item.is_success ? '成功' : '失败',
      得分: item.score_gained,
      提交Flag: item.submitted_flag || '',
    }))
  )
}

function exportReviewPackage() {
  if (!selectedRound.value) {
    return
  }

  downloadJSONFile(buildExportFilename('review-package').replace(/\.csv$/, '.json'), {
    exported_at: new Date().toISOString(),
    contest: {
      id: props.contest.id,
      title: props.contest.title,
      status: props.contest.status,
      mode: props.contest.mode,
    },
    round: {
      id: selectedRound.value.id,
      round_number: selectedRound.value.round_number,
      status: selectedRound.value.status,
      attack_score: selectedRound.value.attack_score,
      defense_score: selectedRound.value.defense_score,
      started_at: selectedRound.value.started_at || null,
      ended_at: selectedRound.value.ended_at || null,
      updated_at: selectedRound.value.updated_at,
    },
    filters: {
      service: {
        team_id: serviceTeamFilter.value || null,
        team_name:
          serviceTeamOptions.value.find((team) => team.team_id === serviceTeamFilter.value)?.team_name || null,
        status: serviceStatusFilter.value,
        check_source: serviceCheckSourceFilter.value || null,
        check_source_label: serviceCheckSourceFilter.value
          ? getCheckSourceLabel(serviceCheckSourceFilter.value) || serviceCheckSourceFilter.value
          : '全部来源',
        alert_reason: serviceAlertReasonFilter.value || null,
        alert_reason_label: serviceAlertReasonFilter.value ? getServiceAlertLabel(serviceAlertReasonFilter.value) : '全部告警',
      },
      attack: {
        team_id: attackTeamFilter.value || null,
        team_name: attackTeamOptions.value.find((team) => team.id === attackTeamFilter.value)?.name || null,
        result: attackResultFilter.value,
        source: attackSourceFilter.value,
        source_label: attackSourceFilter.value === 'all' ? '全部来源' : getAttackSourceLabel(attackSourceFilter.value),
      },
    },
    summary: {
      round: props.summary?.round || null,
      metrics: props.summary?.metrics || null,
      items: props.summary?.items || [],
      service_alerts: serviceAlerts.value.map((alert) => ({
        key: alert.key,
        label: alert.label,
        count: alert.count,
        affected_teams: alert.affected_teams,
        samples: alert.samples,
      })),
    },
    scoreboard: {
      frozen: props.scoreboardFrozen,
      rows: props.scoreboardRows,
    },
    services: filteredServices.value.map((item) => ({
      id: item.id,
      team_id: item.team_id,
      team_name: item.team_name,
      challenge_id: item.challenge_id,
      challenge_title: getChallengeTitle(item.challenge_id),
      service_status: item.service_status,
      service_status_label: getServiceStatusLabel(item.service_status),
      check_source: getServiceCheckSourceValue(item.check_result) || null,
      check_source_label: getCheckSourceLabel(item.check_result.check_source) || '',
      check_result: item.check_result,
      defense_score: item.defense_score,
      attack_received: item.attack_received,
      attack_score: item.attack_score,
      updated_at: item.updated_at,
    })),
    attacks: filteredAttacks.value.map((item) => ({
      id: item.id,
      attacker_team_id: item.attacker_team_id,
      attacker_team: item.attacker_team,
      victim_team_id: item.victim_team_id,
      victim_team: item.victim_team,
      challenge_id: item.challenge_id,
      challenge_title: getChallengeTitle(item.challenge_id),
      attack_type: item.attack_type,
      attack_type_label: getAttackTypeLabel(item.attack_type),
      source: item.source,
      source_label: getAttackSourceLabel(item.source),
      is_success: item.is_success,
      score_gained: item.score_gained,
      submitted_flag: item.submitted_flag || null,
      created_at: item.created_at,
    })),
  })
}

watch(
  () => props.trafficFilters.path_keyword,
  (keyword) => {
    trafficPathKeywordInput.value = keyword
  },
  { immediate: true }
)

watch(
  () => [props.contest.id, props.selectedRoundId] as const,
  ([contestId, roundId]) => {
    syncingPersistedFilters = true
    if (!roundId) {
      resetFilters()
      syncingPersistedFilters = false
      return
    }
    const storedState = loadInspectorFilterState(contestId, roundId)
    if (!storedState) {
      resetFilters()
      syncingPersistedFilters = false
      return
    }
    serviceTeamFilter.value = storedState.service_team_id
    serviceStatusFilter.value = storedState.service_status
    serviceCheckSourceFilter.value = storedState.service_check_source
    serviceAlertReasonFilter.value = storedState.service_alert_reason
    attackTeamFilter.value = storedState.attack_team_id
    attackResultFilter.value = storedState.attack_result
    attackSourceFilter.value = storedState.attack_source
    syncingPersistedFilters = false
  },
  { immediate: true }
)

watch(
  [
    () => props.contest.id,
    () => props.selectedRoundId,
    serviceTeamFilter,
    serviceStatusFilter,
    serviceCheckSourceFilter,
    serviceAlertReasonFilter,
    attackTeamFilter,
    attackResultFilter,
    attackSourceFilter,
  ],
  ([contestId, roundId]) => {
    if (syncingPersistedFilters || !roundId) {
      return
    }
    persistInspectorFilterState(contestId, roundId, {
      service_team_id: serviceTeamFilter.value,
      service_status: serviceStatusFilter.value,
      service_check_source: serviceCheckSourceFilter.value,
      service_alert_reason: serviceAlertReasonFilter.value,
      attack_team_id: attackTeamFilter.value,
      attack_result: attackResultFilter.value,
      attack_source: attackSourceFilter.value,
    })
  },
  { immediate: true }
)

const checkButtonLabel = computed(() => {
  if (props.checking) {
    return '执行巡检中...'
  }
  if (!selectedRound.value) {
    return '立即巡检所选轮次'
  }
  if (selectedRound.value.status === 'running') {
    return '立即巡检当前轮'
  }
  return `重跑第 ${selectedRound.value.round_number} 轮巡检`
})
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <div class="rounded-[28px] border border-[var(--color-primary)]/20 bg-[linear-gradient(145deg,rgba(8,145,178,0.15),rgba(15,23,42,0.94))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-primary-hover)]/75">
          <span>AWD Operations</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">真实接口</span>
        </div>
        <div class="mt-3 flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-3xl font-semibold tracking-tight text-white">{{ contest.title }}</h2>
            <p class="mt-3 text-sm leading-7 text-[var(--color-text-secondary)]/90">
              针对当前 AWD 赛事查看轮次态势、服务健康、攻击记录，并支持立即触发当前轮巡检。
            </p>
          </div>
          <span
            v-if="selectedRound"
            class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
            :class="getRoundStatusClass(selectedRound.status)"
          >
            第 {{ selectedRound.round_number }} 轮 · {{ getRoundStatusLabel(selectedRound.status) }}
          </span>
        </div>

        <div class="mt-6 flex flex-wrap items-center gap-3">
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary"
            :disabled="loadingRounds || loadingRoundDetail"
            @click="emit('refresh')"
          >
            <RefreshCw class="h-4 w-4" />
            刷新 AWD 数据
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary"
            @click="emit('openCreateRoundDialog')"
          >
            <TimerReset class="h-4 w-4" />
            创建轮次
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!selectedRoundId || !canRecordServiceChecks"
            @click="emit('openServiceCheckDialog')"
          >
            <ShieldCheck class="h-4 w-4" />
            录入服务检查
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!selectedRoundId || !canRecordAttackLogs"
            @click="emit('openAttackLogDialog')"
          >
            <Sword class="h-4 w-4" />
            补录攻击日志
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="checking || !selectedRoundId"
            @click="emit('runSelectedRoundCheck')"
          >
            <TimerReset class="h-4 w-4" />
            {{ checkButtonLabel }}
          </button>
        </div>
        <p
          v-if="shouldAutoRefresh"
          class="mt-3 text-xs text-[var(--color-primary-hover)]/70"
        >
          当前正在跟随 live 轮次，面板会每 15 秒自动刷新一次。
        </p>
        <p
          v-if="selectedRoundId && !canRecordServiceChecks && serviceCheckHint"
          class="mt-1 text-xs text-[var(--color-primary-hover)]/70"
        >
          {{ serviceCheckHint }}
        </p>
        <p
          v-if="selectedRoundId && !canRecordAttackLogs && attackLogHint"
          class="mt-1 text-xs text-[var(--color-primary-hover)]/70"
        >
          {{ attackLogHint }}
        </p>
      </div>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <AppCard variant="metric" accent="primary" eyebrow="轮次数量" :title="String(rounds.length)" subtitle="当前赛事已创建的 AWD 轮次。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary">
              <TimerReset class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard variant="metric" accent="warning" eyebrow="失陷服务" :title="String(compromisedCount)" subtitle="当前所选轮次中已被攻破的服务数。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 text-[var(--color-danger)]">
              <ShieldAlert class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="metric"
          accent="success"
          eyebrow="攻击流量"
          :title="String(totalAttackCount)"
          :subtitle="`成功 ${successfulAttackCount} / 失败 ${failedAttackCount}`"
        >
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-[var(--color-success)]/20 bg-[var(--color-success)]/10 text-[var(--color-success)]">
              <Sword class="h-5 w-5" />
            </div>
          </template>
        </AppCard>
      </div>
    </section>

    <section class="grid gap-6 xl:grid-cols-[0.85fr_1.15fr]">
      <SectionCard title="轮次切换" subtitle="查看当前轮的基础参数与状态。">
        <div class="space-y-4">
          <label class="space-y-2">
            <span class="text-sm text-[var(--color-text-secondary)]">选择轮次</span>
            <select
              id="awd-round-selector"
              :value="selectedRoundId || ''"
              class="w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
              :disabled="loadingRounds || rounds.length === 0"
              @change="emit('update:selectedRoundId', ($event.target as HTMLSelectElement).value)"
            >
              <option v-for="round in rounds" :key="round.id" :value="round.id">
                第 {{ round.round_number }} 轮 · {{ getRoundStatusLabel(round.status) }}
              </option>
            </select>
          </label>

          <div v-if="loadingRounds" class="flex justify-center py-8">
            <AppLoading>正在同步轮次...</AppLoading>
          </div>

          <AppEmpty
            v-else-if="rounds.length === 0"
            title="当前赛事还没有 AWD 轮次"
            description="先让后台调度创建轮次，随后这里会展示服务状态和攻击数据。"
            icon="Flag"
          />

          <div v-else-if="selectedRound" class="grid gap-3">
            <AppCard variant="action" accent="neutral" eyebrow="轮次状态" :subtitle="getRoundStatusLabel(selectedRound.status)">
              <template #default>
                <div class="text-sm text-[var(--color-text-secondary)]">
                  <p>攻击分值：{{ selectedRound.attack_score }}</p>
                  <p class="mt-1">防守分值：{{ selectedRound.defense_score }}</p>
                </div>
              </template>
            </AppCard>

            <AppCard variant="action" accent="neutral" eyebrow="时间窗口" :subtitle="formatDateTime(selectedRound.started_at)">
              <template #default>
                <div class="text-sm text-[var(--color-text-secondary)]">
                  <p>开始：{{ formatDateTime(selectedRound.started_at) }}</p>
                  <p class="mt-1">结束：{{ formatDateTime(selectedRound.ended_at) }}</p>
                </div>
              </template>
            </AppCard>

            <AppCard variant="action" accent="warning" eyebrow="异常速览" :subtitle="`下线 ${downCount} · 失陷 ${compromisedCount}`">
              <template #default>
                <div class="text-sm text-[var(--color-text-secondary)]">
                  <p>服务总数：{{ totalServiceCount }}</p>
                  <p class="mt-1">最后巡检：{{ formatDateTime(selectedRound.updated_at) }}</p>
                </div>
              </template>
            </AppCard>
          </div>
        </div>
      </SectionCard>

      <SectionCard title="回合态势" subtitle="排行榜、服务明细与攻击流水共用当前选中轮次。">
        <div v-if="loadingRoundDetail" class="flex justify-center py-12">
          <AppLoading>正在同步当前轮次数据...</AppLoading>
        </div>

        <div v-else-if="!selectedRound" class="py-4">
          <AppEmpty
            title="请选择一个 AWD 轮次"
            description="轮次选定后即可查看服务健康、攻击记录和本轮汇总。"
            icon="FileChartColumnIncreasing"
          />
        </div>

        <div v-else class="space-y-6">
          <div class="flex justify-end">
            <button
              id="awd-export-review-package"
              type="button"
              class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="!selectedRound"
              @click="exportReviewPackage"
            >
              导出当前轮复盘包
            </button>
          </div>

          <div class="grid gap-3 md:grid-cols-2 2xl:grid-cols-4">
            <AppCard
              variant="metric"
              accent="primary"
              eyebrow="服务总览"
              :title="String(totalServiceCount)"
              :subtitle="`正常 ${upCount} / 下线 ${downCount} / 失陷 ${compromisedCount}`"
            />
            <AppCard
              variant="metric"
              accent="success"
              eyebrow="攻击流量"
              :title="String(totalAttackCount)"
              :subtitle="`成功 ${successfulAttackCount} / 失败 ${failedAttackCount}`"
            />
            <AppCard
              variant="metric"
              accent="warning"
              eyebrow="防守成功"
              :title="String(defenseSuccessCount)"
              :subtitle="`受攻击服务 ${attackedServiceCount}`"
            />
            <AppCard
              variant="metric"
              accent="neutral"
              eyebrow="来源构成"
              :title="getSourceOverviewLabel()"
              :subtitle="getSourceOverviewDescription()"
            />
          </div>

          <section class="space-y-4 border-t border-border pt-6">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h3 class="text-base font-semibold text-[var(--color-text-primary)]">攻击流量态势</h3>
                <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                  代理请求摘要，不等同于已确认攻破结果。
                </p>
              </div>
              <span class="text-xs text-[var(--color-text-muted)]">最近更新时间：{{ formatDateTime(selectedRound.updated_at) }}</span>
            </div>

            <div
              v-if="loadingTrafficSummary"
              class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
            >
              正在加载攻击流量摘要...
            </div>
            <div
              v-else-if="!trafficSummary"
              class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
            >
              当前轮次暂未返回攻击流量摘要。
            </div>
            <template v-else>
              <div class="border-y border-border/80 bg-[linear-gradient(180deg,rgba(8,145,178,0.06),rgba(8,145,178,0))]">
                <div class="grid gap-0 md:grid-cols-2 xl:grid-cols-5">
                  <div
                    v-for="item in trafficSummaryStats"
                    :key="item.key"
                    class="border-b border-border/70 px-4 py-4 last:border-b-0 md:[&:nth-last-child(-n+2)]:border-b-0 xl:border-b-0 xl:border-r xl:last:border-r-0"
                  >
                    <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]">
                      {{ item.label }}
                    </p>
                    <p class="mt-3 text-2xl font-semibold tracking-tight text-[var(--color-text-primary)]">
                      {{ item.value }}
                    </p>
                    <p class="mt-2 text-xs leading-6 text-[var(--color-text-muted)]">
                      {{ item.hint }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="grid gap-4 xl:grid-cols-[1.3fr_0.7fr]">
                <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                  <div class="rounded-xl border border-border/80">
                    <div class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]">热点攻击队</div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_attackers.slice(0, 5)"
                        :key="`traffic-attacker-${item.team_id}`"
                        class="flex items-center justify-between px-3 py-2 text-sm"
                      >
                        <span class="text-[var(--color-text-primary)]">{{ item.team_name }}</span>
                        <span class="font-medium text-[var(--color-text-secondary)]">{{ item.request_count }}</span>
                      </li>
                      <li v-if="trafficSummary.top_attackers.length === 0" class="px-3 py-3 text-xs text-[var(--color-text-muted)]">
                        暂无攻击队热点数据
                      </li>
                    </ol>
                  </div>
                  <div class="rounded-xl border border-border/80">
                    <div class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]">热点受害队</div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_victims.slice(0, 5)"
                        :key="`traffic-victim-${item.team_id}`"
                        class="flex items-center justify-between px-3 py-2 text-sm"
                      >
                        <span class="text-[var(--color-text-primary)]">{{ item.team_name }}</span>
                        <span class="font-medium text-[var(--color-text-secondary)]">{{ item.request_count }}</span>
                      </li>
                      <li v-if="trafficSummary.top_victims.length === 0" class="px-3 py-3 text-xs text-[var(--color-text-muted)]">
                        暂无目标热点数据
                      </li>
                    </ol>
                  </div>
                  <div class="rounded-xl border border-border/80">
                    <div class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]">热点题目</div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_challenges.slice(0, 5)"
                        :key="`traffic-challenge-${item.challenge_id}`"
                        class="flex items-center justify-between gap-3 px-3 py-2 text-sm"
                      >
                        <span class="truncate text-[var(--color-text-primary)]">{{ getTrafficChallengeTitle(item.challenge_id, item.challenge_title) }}</span>
                        <span class="shrink-0 font-medium text-[var(--color-text-secondary)]">{{ item.request_count }}</span>
                      </li>
                      <li v-if="trafficSummary.top_challenges.length === 0" class="px-3 py-3 text-xs text-[var(--color-text-muted)]">
                        暂无题目热点数据
                      </li>
                    </ol>
                  </div>
                  <div class="rounded-xl border border-border/80">
                    <div class="border-b border-border bg-surface-alt/40 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]">异常路径</div>
                    <ol class="divide-y divide-border/70">
                      <li
                        v-for="item in trafficSummary.top_error_paths.slice(0, 5)"
                        :key="`traffic-path-${item.path}`"
                        class="px-3 py-2 text-sm"
                      >
                        <p class="truncate font-mono text-[var(--color-text-primary)]">{{ item.path }}</p>
                        <p class="mt-1 text-xs text-[var(--color-text-muted)]">请求 {{ item.request_count }} / 错误 {{ item.error_count }}</p>
                      </li>
                      <li v-if="trafficSummary.top_error_paths.length === 0" class="px-3 py-3 text-xs text-[var(--color-text-muted)]">
                        暂无异常路径数据
                      </li>
                    </ol>
                  </div>
                </div>

                <div class="rounded-xl border border-border/80">
                  <div class="border-b border-border bg-surface-alt/40 px-3 py-2">
                    <p class="text-xs font-semibold tracking-[0.12em] text-[var(--color-text-muted)]">趋势摘要（最近 12 桶）</p>
                    <p class="mt-2 text-xs leading-6 text-[var(--color-text-muted)]">{{ trafficTrendNarrative }}</p>
                  </div>
                  <div class="space-y-2 px-3 py-3">
                    <div
                      v-for="bucket in trafficTrendRows"
                      :key="bucket.bucket_start_at"
                      class="space-y-1"
                    >
                      <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
                        <span>{{ bucket.label }}</span>
                        <span>请求 {{ bucket.request_count }} / 错误 {{ bucket.error_count }}</span>
                      </div>
                      <div class="h-1.5 overflow-hidden rounded-full bg-[var(--color-text-muted)]/20">
                        <div
                          class="h-full rounded-full bg-[var(--color-primary)]/70"
                          :style="{ width: `${bucket.ratio}%` }"
                        />
                      </div>
                    </div>
                    <p v-if="trafficTrendRows.length === 0" class="text-xs text-[var(--color-text-muted)]">
                      当前没有趋势桶数据。
                    </p>
                  </div>
                </div>
              </div>
            </template>

            <div class="overflow-hidden rounded-xl border border-border">
              <div class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/60 px-4 py-3">
                <div>
                  <p class="text-sm font-semibold text-[var(--color-text-primary)]">流量明细表</p>
                  <p class="mt-1 text-xs text-[var(--color-text-muted)]">按攻击方、受害方、题目、状态分桶和路径关键字筛选。</p>
                </div>
                <button
                  id="awd-traffic-reset-filters"
                  type="button"
                  class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-[var(--color-text-primary)] transition hover:border-primary"
                  @click="emit('resetTrafficFilters')"
                >
                  重置筛选
                </button>
              </div>

              <div class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-5">
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">攻击方</span>
                  <select
                    id="awd-traffic-filter-attacker"
                    :value="trafficFilters.attacker_team_id"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="applyTrafficFilterPatch({ attacker_team_id: ($event.target as HTMLSelectElement).value })"
                  >
                    <option value="">全部攻击方</option>
                    <option v-for="team in trafficTeamOptions" :key="`traffic-attacker-option-${team.id}`" :value="team.id">
                      {{ team.name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">受害方</span>
                  <select
                    id="awd-traffic-filter-victim"
                    :value="trafficFilters.victim_team_id"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="applyTrafficFilterPatch({ victim_team_id: ($event.target as HTMLSelectElement).value })"
                  >
                    <option value="">全部受害方</option>
                    <option v-for="team in trafficTeamOptions" :key="`traffic-victim-option-${team.id}`" :value="team.id">
                      {{ team.name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">题目</span>
                  <select
                    id="awd-traffic-filter-challenge"
                    :value="trafficFilters.challenge_id"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="applyTrafficFilterPatch({ challenge_id: ($event.target as HTMLSelectElement).value })"
                  >
                    <option value="">全部题目</option>
                    <option v-for="challenge in challengeLinks" :key="challenge.id" :value="challenge.challenge_id">
                      {{ challenge.title || `Challenge #${challenge.challenge_id}` }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">状态分桶</span>
                  <select
                    id="awd-traffic-filter-status-group"
                    :value="trafficFilters.status_group"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                    @change="onTrafficStatusGroupChange(($event.target as HTMLSelectElement).value)"
                  >
                    <option v-for="item in trafficStatusGroupOptions" :key="item.value" :value="item.value">
                      {{ item.label }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">路径关键字</span>
                  <div class="flex items-center gap-2">
                    <input
                      id="awd-traffic-filter-path"
                      v-model="trafficPathKeywordInput"
                      type="text"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                      placeholder="/api/..."
                      @keydown.enter.prevent="applyTrafficKeywordFilter"
                    />
                    <button
                      id="awd-traffic-filter-search"
                      type="button"
                      class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border text-[var(--color-text-secondary)] transition hover:border-primary hover:text-[var(--color-text-primary)]"
                      @click="applyTrafficKeywordFilter"
                    >
                      <Search class="h-4 w-4" />
                    </button>
                  </div>
                  <button
                    v-if="trafficFilters.path_keyword"
                    type="button"
                    class="text-xs text-[var(--color-text-muted)] underline-offset-2 transition hover:text-[var(--color-text-primary)] hover:underline"
                    @click="clearTrafficKeywordFilter"
                  >
                    清除路径关键字
                  </button>
                </label>
              </div>

              <table class="min-w-full divide-y divide-border">
                <thead class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]">
                  <tr>
                    <th class="px-4 py-3">时间</th>
                    <th class="px-4 py-3">攻击方 / 受害方</th>
                    <th class="px-4 py-3">靶题</th>
                    <th class="px-4 py-3">请求</th>
                    <th class="px-4 py-3">状态</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border bg-surface/70">
                  <tr v-if="loadingTrafficEvents">
                    <td colspan="5" class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">正在加载流量明细...</td>
                  </tr>
                  <tr
                    v-for="event in trafficEvents"
                    :key="`${event.occurred_at}-${event.attacker_team_id}-${event.victim_team_id}-${event.challenge_id}-${event.method}-${event.path}`"
                  >
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">{{ formatDateTime(event.occurred_at) }}</td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">
                      <p>{{ getTrafficTeamName(event.attacker_team_id, event.attacker_team_name) }}</p>
                      <p class="mt-1 text-xs text-[var(--color-text-muted)]">→ {{ getTrafficTeamName(event.victim_team_id, event.victim_team_name) }}</p>
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">{{ getTrafficChallengeTitle(event.challenge_id, event.challenge_title) }}</td>
                    <td class="px-4 py-4 text-sm">
                      <p class="font-mono text-[var(--color-text-primary)]">{{ event.method.toUpperCase() }} {{ event.path }}</p>
                      <p class="mt-1 text-xs text-[var(--color-text-muted)]">HTTP {{ event.status_code }}</p>
                    </td>
                    <td class="px-4 py-4 text-sm">
                      <span
                        class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                        :class="getTrafficStatusGroupClass(event.status_group)"
                      >
                        {{ getTrafficStatusGroupLabel(event.status_group) }}
                      </span>
                      <p class="mt-1 text-xs text-[var(--color-text-muted)]">{{ getTrafficSourceLabel(event.source) }}</p>
                    </td>
                  </tr>
                  <tr v-if="!loadingTrafficEvents && trafficEvents.length === 0">
                    <td colspan="5" class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
                      当前筛选条件下没有流量事件。
                    </td>
                  </tr>
                </tbody>
              </table>

              <div class="flex items-center justify-between gap-3 border-t border-border bg-surface-alt/20 px-4 py-3 text-xs text-[var(--color-text-muted)]">
                <span>共 {{ trafficEventsTotal }} 条 · 第 {{ trafficFilters.page }} / {{ trafficTotalPages }} 页</span>
                <div class="flex items-center gap-2">
                  <button
                    id="awd-traffic-page-prev"
                    type="button"
                    class="rounded-lg border border-border px-3 py-1.5 text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="!canGotoTrafficPrevPage || loadingTrafficEvents"
                    @click="moveTrafficPage(-1)"
                  >
                    上一页
                  </button>
                  <button
                    id="awd-traffic-page-next"
                    type="button"
                    class="rounded-lg border border-border px-3 py-1.5 text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="!canGotoTrafficNextPage || loadingTrafficEvents"
                    @click="moveTrafficPage(1)"
                  >
                    下一页
                  </button>
                </div>
              </div>
            </div>
          </section>

          <div
            v-if="serviceAlerts.length > 0"
            class="grid gap-3 md:grid-cols-2 2xl:grid-cols-3"
          >
            <button
              v-for="alert in serviceAlerts"
              :key="alert.key"
              type="button"
              class="text-left"
              @click="applyServiceAlertFilter(alert.key)"
            >
              <AppCard
                variant="action"
                accent="warning"
                eyebrow="重点告警"
                :title="alert.label"
                :subtitle="getServiceAlertSubtitle(alert)"
              >
                <template #default>
                  <div
                    class="rounded-2xl border px-3 py-3 text-sm transition"
                    :class="[
                      getServiceAlertClass(alert.key),
                      serviceAlertReasonFilter === alert.key ? 'ring-2 ring-primary/60' : '',
                    ]"
                  >
                    <div class="font-medium">受影响服务 {{ alert.count }} 个</div>
                    <div class="mt-2 space-y-1 text-xs">
                      <div
                        v-for="sample in alert.samples"
                        :key="`${alert.key}-${sample.service_id}`"
                      >
                        {{ sample.team_name }} · {{ sample.challenge_title }}
                      </div>
                    </div>
                    <div class="mt-2 text-xs text-[var(--color-text-secondary)]/80">
                      {{ serviceAlertReasonFilter === alert.key ? '再次点击可取消筛选' : '点击筛选同类异常' }}
                    </div>
                  </div>
                </template>
              </AppCard>
            </button>
          </div>

          <div class="overflow-hidden rounded-2xl border border-border">
            <div class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3">
              <div class="text-sm font-semibold text-[var(--color-text-primary)]">实时排行榜</div>
              <span
                v-if="scoreboardFrozen"
                class="inline-flex rounded-full border border-[var(--color-warning)]/20 bg-[var(--color-warning)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-warning)]"
              >
                排行榜已冻结
              </span>
            </div>
            <table class="min-w-full divide-y divide-border">
              <thead class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]">
                <tr>
                  <th class="px-4 py-3">排名</th>
                  <th class="px-4 py-3">队伍</th>
                  <th class="px-4 py-3">得分</th>
                  <th class="px-4 py-3">解题数</th>
                  <th class="px-4 py-3">最近得分</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border bg-surface/70">
                <tr v-for="item in scoreboardRows" :key="item.team_id">
                  <td class="px-4 py-4 text-sm font-semibold text-[var(--color-text-primary)]">#{{ item.rank }}</td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">{{ item.team_name }}</td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">{{ formatScore(item.score) }}</td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">{{ item.solved_count }}</td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">{{ formatDateTime(item.last_submission_at) }}</td>
                </tr>
                <tr v-if="scoreboardRows.length === 0">
                  <td colspan="5" class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
                    当前赛事还没有排行榜数据。
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="overflow-hidden rounded-2xl border border-border">
            <div class="border-b border-border bg-surface-alt/70 px-4 py-3 text-sm font-semibold text-[var(--color-text-primary)]">
              本轮汇总
            </div>
            <table class="min-w-full divide-y divide-border">
              <thead class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]">
                <tr>
                  <th class="px-4 py-3">队伍</th>
                  <th class="px-4 py-3">总分</th>
                  <th class="px-4 py-3">攻击 / 防守</th>
                  <th class="px-4 py-3">服务状态</th>
                  <th class="px-4 py-3">被攻击情况</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border bg-surface/70">
                <tr v-for="item in summary?.items || []" :key="item.team_id">
                  <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">{{ item.team_name }}</td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-primary)]">{{ item.total_score }}</td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                    {{ item.attack_score }} / {{ item.defense_score }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                    正常 {{ item.service_up_count }} / 下线 {{ item.service_down_count }} / 失陷 {{ item.service_compromised_count }}
                  </td>
                  <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                    攻破 {{ item.successful_breach_count }} 次，攻击方 {{ item.unique_attackers_against }} 支
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="grid gap-6 xl:grid-cols-2">
            <div class="overflow-hidden rounded-2xl border border-border">
              <div class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3">
                <div class="text-sm font-semibold text-[var(--color-text-primary)]">服务状态表</div>
                <button
                  id="awd-export-services"
                  type="button"
                  class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="filteredServices.length === 0"
                  @click="exportFilteredServices"
                >
                  导出当前筛选
                </button>
              </div>
              <div class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-4">
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">队伍</span>
                  <select
                    id="awd-service-filter-team"
                    v-model="serviceTeamFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部队伍</option>
                    <option v-for="team in serviceTeamOptions" :key="team.team_id" :value="team.team_id">
                      {{ team.team_name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">状态</span>
                  <select
                    id="awd-service-filter-status"
                    v-model="serviceStatusFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="all">全部状态</option>
                    <option value="up">正常</option>
                    <option value="down">下线</option>
                    <option value="compromised">已失陷</option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">巡检来源</span>
                  <select
                    id="awd-service-filter-source"
                    v-model="serviceCheckSourceFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部来源</option>
                    <option v-for="source in serviceCheckSourceOptions" :key="source" :value="source">
                      {{ getCheckSourceLabel(source) || source }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">告警类型</span>
                  <select
                    id="awd-service-filter-alert"
                    v-model="serviceAlertReasonFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部告警</option>
                    <option v-for="alert in serviceAlerts" :key="alert.key" :value="alert.key">
                      {{ alert.label }}
                    </option>
                  </select>
                </label>
              </div>
              <table class="min-w-full divide-y divide-border">
                <thead class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]">
                  <tr>
                    <th class="px-4 py-3">队伍</th>
                    <th class="px-4 py-3">靶题</th>
                    <th class="px-4 py-3">服务状态</th>
                    <th class="px-4 py-3">得分</th>
                    <th class="px-4 py-3">检查结果</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border bg-surface/70">
                  <tr v-for="service in filteredServices" :key="service.id">
                    <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">{{ service.team_name }}</td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">{{ getChallengeTitle(service.challenge_id) }}</td>
                    <td class="px-4 py-4">
                      <span
                        class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                        :class="getServiceStatusClass(service.service_status)"
                      >
                        {{ getServiceStatusLabel(service.service_status) }}
                      </span>
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      防守 {{ service.defense_score }} / 受攻击 {{ service.attack_received }}
                    </td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-muted)]">
                      <div>{{ summarizeCheckResult(service.check_result) }}</div>
                      <div
                        v-if="getTargetProbeSummary(service.check_result)"
                        class="mt-2 text-xs text-[var(--color-text-muted)]"
                      >
                        {{ getTargetProbeSummary(service.check_result) }}
                      </div>
                      <details
                        v-if="getCheckTargets(service.check_result).length > 0"
                        class="mt-2 rounded-xl border border-border/80 bg-surface-alt/40 p-3 text-xs text-[var(--color-text-secondary)]"
                      >
                        <summary class="cursor-pointer select-none text-[var(--color-text-primary)]">
                          查看探测明细
                        </summary>
                        <div class="mt-3 space-y-3">
                          <div
                            v-for="(target, targetIndex) in getCheckTargets(service.check_result)"
                            :key="`${service.id}-target-${targetIndex}`"
                            class="rounded-xl border border-border/70 bg-surface/70 p-3"
                          >
                            <div class="font-medium text-[var(--color-text-primary)]">
                              {{ target.access_url || `Target #${targetIndex + 1}` }}
                            </div>
                            <div class="mt-1 text-[var(--color-text-muted)]">
                              {{ getProbeStatusText(target.healthy, target.error_code, target.error) }}
                              <span v-if="target.probe"> · {{ target.probe.toUpperCase() }}</span>
                              <span v-if="formatLatency(target.latency_ms)"> · {{ formatLatency(target.latency_ms) }}</span>
                            </div>
                            <div
                              v-if="target.attempts.length > 0"
                              class="mt-2 space-y-1 border-t border-border/60 pt-2"
                            >
                              <div
                                v-for="(attempt, attemptIndex) in target.attempts"
                                :key="`${service.id}-target-${targetIndex}-attempt-${attemptIndex}`"
                              >
                                Attempt {{ attemptIndex + 1 }}:
                                {{ attempt.probe.toUpperCase() || 'UNKNOWN' }}
                                · {{ getProbeStatusText(attempt.healthy, attempt.error_code, attempt.error) }}
                                <span v-if="formatLatency(attempt.latency_ms)"> · {{ formatLatency(attempt.latency_ms) }}</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </details>
                    </td>
                  </tr>
                  <tr v-if="filteredServices.length === 0">
                    <td colspan="5" class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
                      {{ services.length === 0 ? '当前轮次还没有服务巡检记录。' : '当前筛选条件下没有服务记录。' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="overflow-hidden rounded-2xl border border-border">
              <div class="flex items-center justify-between gap-3 border-b border-border bg-surface-alt/70 px-4 py-3">
                <div class="text-sm font-semibold text-[var(--color-text-primary)]">攻击日志</div>
                <button
                  id="awd-export-attacks"
                  type="button"
                  class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="filteredAttacks.length === 0"
                  @click="exportFilteredAttacks"
                >
                  导出当前筛选
                </button>
              </div>
              <div class="grid gap-3 border-b border-border bg-surface-alt/30 px-4 py-3 md:grid-cols-3">
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">队伍</span>
                  <select
                    id="awd-attack-filter-team"
                    v-model="attackTeamFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="">全部队伍</option>
                    <option v-for="team in attackTeamOptions" :key="team.id" :value="team.id">
                      {{ team.name }}
                    </option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">结果</span>
                  <select
                    id="awd-attack-filter-result"
                    v-model="attackResultFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="all">全部结果</option>
                    <option value="success">仅成功</option>
                    <option value="failed">仅失败</option>
                  </select>
                </label>
                <label class="space-y-1">
                  <span class="text-[11px] uppercase tracking-[0.18em] text-[var(--color-text-muted)]">记录来源</span>
                  <select
                    id="awd-attack-filter-source"
                    v-model="attackSourceFilter"
                    class="w-full rounded-xl border border-border bg-surface px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
                  >
                    <option value="all">全部来源</option>
                    <option v-for="source in attackSourceOptions" :key="source" :value="source">
                      {{ getAttackSourceLabel(source) }}
                    </option>
                  </select>
                </label>
              </div>
              <table class="min-w-full divide-y divide-border">
                <thead class="bg-surface-alt/40 text-left text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-text-muted)]">
                  <tr>
                    <th class="px-4 py-3">时间</th>
                    <th class="px-4 py-3">攻击方</th>
                    <th class="px-4 py-3">受害方</th>
                    <th class="px-4 py-3">类型</th>
                    <th class="px-4 py-3">结果</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border bg-surface/70">
                  <tr v-for="attack in filteredAttacks" :key="attack.id">
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">{{ formatDateTime(attack.created_at) }}</td>
                    <td class="px-4 py-4 text-sm font-medium text-[var(--color-text-primary)]">{{ attack.attacker_team }}</td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">{{ attack.victim_team }}</td>
                    <td class="px-4 py-4 text-sm text-[var(--color-text-secondary)]">
                      <div>{{ getAttackTypeLabel(attack.attack_type) }}</div>
                      <div class="mt-1 text-xs text-[var(--color-text-muted)]">{{ getChallengeTitle(attack.challenge_id) }}</div>
                      <div class="mt-1 text-xs text-[var(--color-text-muted)]">{{ getAttackSourceLabel(attack.source) }}</div>
                    </td>
                    <td class="px-4 py-4 text-sm">
                      <span
                        class="inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold"
                        :class="attack.is_success ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]' : 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-secondary)]'"
                      >
                        <ShieldCheck v-if="attack.is_success" class="h-3.5 w-3.5" />
                        {{ attack.is_success ? `成功 +${attack.score_gained}` : '失败' }}
                      </span>
                    </td>
                  </tr>
                  <tr v-if="filteredAttacks.length === 0">
                    <td colspan="5" class="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
                      {{ attacks.length === 0 ? '当前轮次还没有攻击记录。' : '当前筛选条件下没有攻击记录。' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </SectionCard>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { AdminContestChallengeViewData } from '@/api/contracts'
import { useAwdCheckResultPresentation } from '@/features/awd-inspector'
import type { ContestAwdConfigRouteTarget } from '../model'
import AWDChallengeConfigDirectorySection from './AWDChallengeConfigDirectorySection.vue'
import AWDChallengeConfigHeader from './AWDChallengeConfigHeader.vue'
import './awdChallengeConfigPanel.css'
import type {
  AWDChallengeConfigDirectoryItemView,
  AWDChallengeConfigSummaryItem,
} from './awdChallengeConfigPanel.types'

const props = defineProps<{
  challengeLinks: AdminContestChallengeViewData[]
  buildEditRoute?: (challenge: AdminContestChallengeViewData) => ContestAwdConfigRouteTarget | null
}>()

const sortedChallengeLinks = computed(() =>
  [...props.challengeLinks].sort(
    (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
  )
)

const summaryItems = computed<AWDChallengeConfigSummaryItem[]>(() => [
  {
    key: 'total',
    label: '已关联题目',
    value: String(sortedChallengeLinks.value.length),
    hint: '当前 AWD 赛事中可参与攻防的服务题目数量',
  },
  {
    key: 'configured',
    label: '已配 Checker',
    value: String(
      sortedChallengeLinks.value.filter(
        (item) =>
          Boolean(item.awd_checker_type) || Object.keys(item.awd_checker_config || {}).length > 0
      ).length
    ),
    hint: '已写入 checker 类型或 checker 配置的题目数',
  },
  {
    key: 'standard-checker',
    label: '标准 Checker',
    value: String(
      sortedChallengeLinks.value.filter(
        (item) =>
          item.awd_checker_type === 'http_standard' || item.awd_checker_type === 'tcp_standard'
      ).length
    ),
    hint: '已切到 HTTP / TCP 标准 Checker 的题目数',
  },
  {
    key: 'hidden',
    label: '隐藏题目',
    value: String(sortedChallengeLinks.value.filter((item) => !item.is_visible).length),
    hint: '当前不会直接对选手展示的赛事题目数',
  },
  {
    key: 'service-linked',
    label: '已建服务关联',
    value: String(sortedChallengeLinks.value.filter((item) => Boolean(item.awd_service_id)).length),
    hint: '已落入赛事级服务关联表的题目数',
  },
])

function formatValidationDateTime(value?: string): string {
  if (!value) {
    return '未记录'
  }
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const { getPrimaryAccessURL, getValidationStateLabel } = useAwdCheckResultPresentation({
  formatDateTime: formatValidationDateTime,
})

function getCheckerTypeLabel(value?: string): string {
  switch (value) {
    case 'legacy_probe':
      return '基础探活'
    case 'http_standard':
      return 'HTTP 标准 Checker'
    case 'tcp_standard':
      return 'TCP 标准 Checker'
    case 'script_checker':
      return '脚本 Checker'
    default:
      return '未配置'
  }
}

function getConfigSummary(item: AdminContestChallengeViewData): string {
  const config = item.awd_checker_config || {}
  if (item.awd_checker_type === 'tcp_standard') {
    const steps = Array.isArray(config.steps) ? config.steps.length : 0
    const timeout = typeof config.timeout_ms === 'number' ? `${config.timeout_ms}ms` : ''
    return [`TCP ${steps} steps`, timeout].filter(Boolean).join(' · ') || '未配置 TCP 步骤'
  }
  const putFlag = readActionSummary(config.put_flag, 'PUT')
  const getFlag = readActionSummary(config.get_flag, 'GET')
  const havoc = readActionSummary(config.havoc, 'Havoc')
  const healthPath =
    typeof config.health_path === 'string' && config.health_path.trim() !== ''
      ? `Health ${config.health_path.trim()}`
      : ''

  return [putFlag, getFlag, havoc, healthPath].filter(Boolean).join(' · ') || '未配置动作摘要'
}

function readActionSummary(value: unknown, label: string): string {
  if (!value || typeof value !== 'object') {
    return ''
  }
  const item = value as Record<string, unknown>
  const path = typeof item.path === 'string' ? item.path : ''
  if (!path) {
    return label
  }
  return `${label} ${path}`
}

function getChallengeTitle(item: AdminContestChallengeViewData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}

function getChallengePreviewRoute(item: AdminContestChallengeViewData) {
  return {
    name: 'PlatformChallengeDetail',
    params: { id: item.challenge_id },
  }
}

function buildPresentationResult(item: AdminContestChallengeViewData): Record<string, unknown> {
  const preview = item.awd_checker_last_preview_result
  if (!preview) {
    return {}
  }
  return {
    ...preview.check_result,
    preview_context: preview.preview_context,
  }
}

function getValidationStateText(item: AdminContestChallengeViewData): string {
  return getValidationStateLabel(item.awd_checker_validation_state) || '未验证'
}

function getValidationHint(item: AdminContestChallengeViewData): string {
  const previewAccessURL = getPrimaryAccessURL(buildPresentationResult(item))
  const entries = [
    item.awd_checker_last_preview_at
      ? `最近校验 ${formatValidationDateTime(item.awd_checker_last_preview_at)}`
      : '',
    previewAccessURL ? `目标 ${previewAccessURL}` : '',
  ].filter(Boolean)

  if (entries.length > 0) {
    return entries.join(' · ')
  }

  switch (item.awd_checker_validation_state) {
    case 'stale':
      return 'Checker 草稿已变化，需要重新试跑。'
    case 'failed':
      return '最近一次保存的试跑结果未通过。'
    case 'passed':
      return '最近一次保存的试跑结果已通过。'
    case 'pending':
    default:
      return '保存后可通过试跑绑定最近一次校验结果。'
  }
}

const directoryItems = computed<AWDChallengeConfigDirectoryItemView[]>(() =>
  sortedChallengeLinks.value.map((item) => ({
    source: item,
    challengeId: item.challenge_id,
    title: getChallengeTitle(item),
    category: item.category,
    order: item.order,
    checkerTypeLabel: getCheckerTypeLabel(item.awd_checker_type),
    slaScore: item.awd_sla_score,
    defenseScore: item.awd_defense_score,
    configSummary: getConfigSummary(item),
    validationState: item.awd_checker_validation_state,
    validationStateText: getValidationStateText(item),
    validationPrimaryHint: getValidationHint(item).split(' · ')[0],
    previewRoute: getChallengePreviewRoute(item),
  }))
)

</script>

<template>
  <div class="studio-awd-config">
    <AWDChallengeConfigHeader :summary-items="summaryItems" />
    <AWDChallengeConfigDirectorySection
      :items="directoryItems"
      :build-edit-route="buildEditRoute"
    />
  </div>
</template>

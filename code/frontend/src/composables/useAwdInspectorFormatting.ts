import type { Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficStatusGroup,
  AWDTeamServiceData,
  AdminContestChallengeData,
  ContestDetailData,
} from '@/api/contracts'

interface UseAwdInspectorFormattingOptions {
  contest: Ref<ContestDetailData>
  challengeLinks: Ref<AdminContestChallengeData[]>
  selectedRound: Ref<AWDRoundData | null>
  summaryMetrics: Ref<AWDRoundSummaryData['metrics'] | null>
  manualCheckCount: Ref<number>
}

export function useAwdInspectorFormatting({
  contest,
  challengeLinks,
  selectedRound,
  summaryMetrics,
  manualCheckCount,
}: UseAwdInspectorFormattingOptions) {
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
      pending:
        'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20',
      running:
        'bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20',
      finished:
        'bg-[var(--color-text-muted)]/10 text-[var(--color-text-secondary)] border border-[var(--color-text-muted)]/20',
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
      down:
        'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20',
      compromised:
        'bg-[var(--color-danger)]/10 text-[var(--color-danger)] border border-[var(--color-danger)]/20',
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

  function getChallengeTitle(challengeId: string): string {
    const matched = challengeLinks.value.find((item) => item.challenge_id === challengeId)
    return matched?.title?.trim() || `Challenge #${challengeId}`
  }

  function buildExportFilename(suffix: string): string {
    const title = contest.value.title
      .trim()
      .replace(/[^a-zA-Z0-9_-]+/g, '-')
      .replace(/^-+|-+$/g, '')
    const contestPart = title || `contest-${contest.value.id}`
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

  return {
    formatDateTime,
    getRoundStatusLabel,
    getRoundStatusClass,
    getServiceStatusLabel,
    getServiceStatusClass,
    getAttackTypeLabel,
    getAttackSourceLabel,
    formatPercent,
    getTrafficStatusGroupLabel,
    getTrafficStatusGroupClass,
    getChallengeTitle,
    buildExportFilename,
    getSelectedRoundLabel,
    formatScore,
    getSourceOverviewLabel,
    getSourceOverviewDescription,
  }
}

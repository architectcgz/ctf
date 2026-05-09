import type { Component } from 'vue'
import { Activity, CheckCircle, MousePointerClick, Waypoints } from 'lucide-vue-next'

import type {
  TeacherAttackEventData,
  TeacherAttackSessionData,
  TeacherEvidenceData,
  TeacherEvidenceSummaryData,
} from '@/api/contracts'

export const TEACHER_STUDENT_REVIEW_WORKSPACE_COPY = {
  emptyTitle: '暂无攻击会话',
  emptyDescription: '当前学员还没有可用于复盘的攻击过程记录。',
} as const

export interface ReviewWorkspaceSummaryItem {
  key: string
  label: string
  value: number
  hint: string
  icon: Component
}

export interface ReviewWorkspaceFilterOption {
  value: string
  label: string
}

export interface ReviewWorkspaceObservationItem {
  key: string
  label: string
  level: 'good' | 'attention'
  summary: string
}

export function buildTeacherStudentReviewSummaryItems(input: {
  sessionSummary?: {
    total_sessions: number
    event_count: number
    success_count: number
  } | null
  evidenceSummary?: Pick<TeacherEvidenceSummaryData, 'proxy_request_count'> | null
}): ReviewWorkspaceSummaryItem[] {
  return [
    {
      key: 'total_sessions',
      label: '会话数',
      value: input.sessionSummary?.total_sessions ?? 0,
      hint: '已聚合的攻击或解题过程',
      icon: Waypoints,
    },
    {
      key: 'event_count',
      label: '事件数',
      value: input.sessionSummary?.event_count ?? 0,
      hint: '纳入会话的关键动作',
      icon: Activity,
    },
    {
      key: 'success_count',
      label: '成功会话',
      value: input.sessionSummary?.success_count ?? 0,
      hint: '命中提交或 AWD 攻击成功',
      icon: CheckCircle,
    },
    {
      key: 'proxy_request_count',
      label: '实操请求',
      value: input.evidenceSummary?.proxy_request_count ?? 0,
      hint: '经平台代理记录的请求',
      icon: MousePointerClick,
    },
  ]
}

export function sessionModeLabel(mode: string): string {
  switch (mode) {
    case 'practice':
      return '训练'
    case 'jeopardy':
      return 'Jeopardy'
    case 'awd':
      return 'AWD'
    default:
      return mode || '未知'
  }
}

export function sessionResultLabel(result: string): string {
  switch (result) {
    case 'success':
      return '成功'
    case 'failed':
      return '失败'
    case 'in_progress':
      return '进行中'
    default:
      return '未知'
  }
}

export function sessionResultClass(result: string): string {
  switch (result) {
    case 'success':
      return 'writeup-chip writeup-chip--success'
    case 'failed':
      return 'writeup-chip writeup-chip--warning'
    case 'in_progress':
      return 'writeup-chip writeup-chip--primary'
    default:
      return 'writeup-chip writeup-chip--muted'
  }
}

export function eventTypeLabel(type: string): string {
  switch (type) {
    case 'instance_access':
      return '访问'
    case 'instance_proxy_request':
      return '请求'
    case 'challenge_submission':
      return '提交'
    case 'awd_attack_submission':
      return 'AWD 攻击'
    case 'awd_traffic':
      return 'AWD 流量'
    case 'writeup':
      return 'Writeup'
    case 'manual_review':
      return '人工评审'
    default:
      return type || '事件'
  }
}

export function buildChallengeFilterOptions(input: {
  evidence?: TeacherEvidenceData | null
  attackSessions?: { sessions: TeacherAttackSessionData[] } | null
}): ReviewWorkspaceFilterOption[] {
  const options = new Map<string, string>()

  input.evidence?.events.forEach((event) => {
    if (!event.challenge_id) return
    options.set(event.challenge_id, event.title || `题目 ${event.challenge_id}`)
  })

  input.attackSessions?.sessions.forEach((session) => {
    if (!session.challenge_id) return
    options.set(session.challenge_id, session.title || `题目 ${session.challenge_id}`)
  })

  return Array.from(options.entries())
    .sort((left, right) => left[0].localeCompare(right[0], 'zh-CN', { numeric: true }))
    .map(([value, label]) => ({
      value,
      label,
    }))
}

export function buildReviewWorkspaceObservations(input: {
  evidence?: TeacherEvidenceData | null
  attackSessions?: { sessions: TeacherAttackSessionData[] } | null
}): ReviewWorkspaceObservationItem[] {
  const evidence = input.evidence
  const sessions = input.attackSessions?.sessions ?? []
  const items: ReviewWorkspaceObservationItem[] = []

  const hasProxyRequest = (evidence?.summary.proxy_request_count ?? 0) > 0
  if (hasProxyRequest) {
    items.push({
      key: 'hands_on_activity',
      label: '实操参与',
      level: 'good',
      summary: '已出现真实请求交互，当前复盘可以直接围绕利用过程展开。',
    })
  } else {
    items.push({
      key: 'hands_on_activity',
      label: '实操参与',
      level: 'attention',
      summary: '暂未看到平台代理请求记录，当前更像浏览或阅读阶段。',
    })
  }

  if (hasRepeatedWrongSubmissions(evidence)) {
    items.push({
      key: 'submission_stability',
      label: '提交稳定性',
      level: 'attention',
      summary: '存在连续错误提交，适合回看试错顺序和验证方式。',
    })
  }

  if (sessions.some((session) => session.result === 'in_progress')) {
    items.push({
      key: 'access_without_submit',
      label: '过程中断',
      level: 'attention',
      summary: '至少有一个会话停留在访问或利用阶段，没有形成最终提交。',
    })
  }

  const hasWriteup = hasEvidenceType(evidence, 'writeup')
  const hasManualReview = hasEvidenceType(evidence, 'manual_review')
  const hasSuccess = sessions.some((session) => session.result === 'success')
  if (hasSuccess && (hasWriteup || hasManualReview)) {
    items.push({
      key: 'training_closure',
      label: '训练闭环',
      level: 'good',
      summary: '成功会话后已经补充复盘材料，实操与输出形成闭环。',
    })
  } else if (hasSuccess) {
    items.push({
      key: 'training_closure',
      label: '训练闭环',
      level: 'attention',
      summary: '已经完成命中，但还缺少题解或人工评审材料。',
    })
  }

  const hasAWDSuccess = sessions.some((session) => session.mode === 'awd' && session.result === 'success')
  if (hasAWDSuccess && !hasWriteup && !hasManualReview) {
    items.push({
      key: 'awd_reflection_gap',
      label: 'AWD 复盘缺口',
      level: 'attention',
      summary: '已有有效攻击命中，但还没有对应复盘材料可供课堂回看。',
    })
  }

  return items
}

export function formatReviewWorkspaceDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

export function evidenceReviewStatusLabel(status: string): string {
  switch (status) {
    case 'approved':
      return '已通过'
    case 'rejected':
      return '已驳回'
    case 'pending':
      return '待审核'
    default:
      return status
  }
}

export function eventMetaItems(
  event: TeacherAttackEventData
): Array<{ key: string; label: string }> {
  const meta = event.meta ?? {}
  const items: Array<{ key: string; label: string }> = []
  const requestMethod = stringMeta(meta, 'request_method') || stringMeta(meta, 'method')
  const targetPath = stringMeta(meta, 'target_path')
  const statusCode = numberMeta(meta, 'status_code')
  const payloadPreview = stringMeta(meta, 'payload_preview')
  const scoreGained = numberMeta(meta, 'score_gained')
  const points = numberMeta(meta, 'points')
  const victimTeam = stringMeta(meta, 'victim_team_name')
  const serviceID = numberMeta(meta, 'service_id')
  const roundID = numberMeta(meta, 'round_id')
  const writeupTitle = stringMeta(meta, 'writeup_title')
  const reviewStatus = stringMeta(meta, 'review_status')

  if (requestMethod) items.push({ key: 'request_method', label: requestMethod })
  if (targetPath) items.push({ key: 'target_path', label: targetPath })
  if (typeof statusCode === 'number') items.push({ key: 'status_code', label: String(statusCode) })
  if (payloadPreview) items.push({ key: 'payload_preview', label: payloadPreview })
  if (typeof points === 'number') items.push({ key: 'points', label: `得分 ${points}` })
  if (typeof scoreGained === 'number') items.push({ key: 'score_gained', label: `得分 ${scoreGained}` })
  if (victimTeam) items.push({ key: 'victim_team', label: `目标 ${victimTeam}` })
  if (typeof serviceID === 'number') items.push({ key: 'service_id', label: `服务 ${serviceID}` })
  if (typeof roundID === 'number') items.push({ key: 'round_id', label: `轮次 ${roundID}` })
  if (writeupTitle) items.push({ key: 'writeup_title', label: writeupTitle })
  if (reviewStatus) items.push({ key: 'review_status', label: evidenceReviewStatusLabel(reviewStatus) })
  return items
}

export function sessionPathSummary(session: TeacherAttackSessionData): string {
  const events = session.events ?? []
  if (events.length === 0) return '暂无事件明细'
  return events
    .slice(0, 4)
    .map((event) => eventTypeLabel(event.type))
    .join(' -> ')
}

function stringMeta(meta: Record<string, unknown>, key: string): string {
  const value = meta[key]
  return typeof value === 'string' ? value : ''
}

function numberMeta(meta: Record<string, unknown>, key: string): number | undefined {
  const value = meta[key]
  return typeof value === 'number' ? value : undefined
}

function hasEvidenceType(evidence: TeacherEvidenceData | null | undefined, type: string): boolean {
  return evidence?.events.some((event) => event.type === type) ?? false
}

function hasRepeatedWrongSubmissions(evidence: TeacherEvidenceData | null | undefined): boolean {
  const events = evidence?.events ?? []
  let streak = 0

  for (const event of events) {
    const tracked = submissionResult(event)
    if (tracked === undefined) continue
    if (tracked) {
      streak = 0
      continue
    }
    streak += 1
    if (streak >= 2) {
      return true
    }
  }

  return false
}

function submissionResult(event: TeacherEvidenceData['events'][number]): boolean | undefined {
  if (!event.meta) return undefined
  if (event.type === 'challenge_submission') {
    return typeof event.meta.is_correct === 'boolean' ? event.meta.is_correct : undefined
  }
  if (event.type === 'awd_attack_submission') {
    return typeof event.meta.is_success === 'boolean' ? event.meta.is_success : undefined
  }
  return undefined
}

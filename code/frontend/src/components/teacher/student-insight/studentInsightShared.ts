import type {
  TeacherAttackEventData,
  TeacherAttackSessionData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
} from '@/api/contracts'

export type StudentInsightSection =
  | 'all'
  | 'overview'
  | 'recommendations'
  | 'writeups'
  | 'manual-review'
  | 'evidence'
  | 'timeline'

export interface InsightMetaItem {
  key: string
  label: string
}

export function visibilityStatusLabel(
  status: TeacherSubmissionWriteupItemData['visibility_status']
): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

export function visibilityStatusClass(
  status: TeacherSubmissionWriteupItemData['visibility_status']
): string {
  return status === 'hidden'
    ? 'writeup-chip writeup-chip--warning'
    : 'writeup-chip writeup-chip--success'
}

export function manualReviewStatusLabel(
  status: TeacherManualReviewSubmissionItemData['review_status']
): string {
  switch (status) {
    case 'approved':
      return '已通过'
    case 'rejected':
      return '已驳回'
    default:
      return '待审核'
  }
}

export function manualReviewStatusClass(
  status: TeacherManualReviewSubmissionItemData['review_status']
): string {
  switch (status) {
    case 'approved':
      return 'writeup-chip writeup-chip--success'
    case 'rejected':
      return 'writeup-chip writeup-chip--warning'
    default:
      return 'writeup-chip writeup-chip--muted'
  }
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

export function formatDateTime(value: string): string {
  if (!value) return '-'
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

export function eventMetaItems(event: TeacherAttackEventData): InsightMetaItem[] {
  const meta = event.meta ?? {}
  const items: InsightMetaItem[] = []
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
  if (typeof scoreGained === 'number')
    items.push({ key: 'score_gained', label: `得分 ${scoreGained}` })
  if (victimTeam) items.push({ key: 'victim_team', label: `目标 ${victimTeam}` })
  if (typeof serviceID === 'number') items.push({ key: 'service_id', label: `服务 ${serviceID}` })
  if (typeof roundID === 'number') items.push({ key: 'round_id', label: `轮次 ${roundID}` })
  if (writeupTitle) items.push({ key: 'writeup_title', label: writeupTitle })
  if (reviewStatus)
    items.push({ key: 'review_status', label: evidenceReviewStatusLabel(reviewStatus) })
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

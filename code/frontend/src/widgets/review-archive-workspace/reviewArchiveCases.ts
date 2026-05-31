import type {
  ReviewArchiveEvidenceItemData,
  ReviewArchiveManualReviewItemData,
  ReviewArchiveWriteupItemData,
  TimelineEvent,
} from '@/api/contracts'

type CaseTone = 'neutral' | 'success' | 'warning'
type PracticeStageKey = 'access' | 'exploit' | 'submit' | 'reflection'
type AwdStageKey = 'attack' | 'result' | 'score'

interface CaseEvent {
  id: string
  label: string
  detail: string
  timestamp: string
  stageLabel: string
  tone: CaseTone
  meta?: string
}

export interface ReviewArchiveStageSummary {
  key: string
  label: string
  count: number
}

export interface ReviewArchiveCase {
  id: string
  title: string
  subtitle: string
  statusLabel: string
  tone: CaseTone
  eventCount: number
  lastActivityAt: string
  metrics: Array<{ label: string; value: string }>
  stages: ReviewArchiveStageSummary[]
  events: CaseEvent[]
}

interface PracticeBucket {
  id: string
  challengeId: string
  title: string
  accessCount: number
  exploitCount: number
  submitCount: number
  successCount: number
  reflectionCount: number
  writeupCount: number
  manualReviewCount: number
  lastActivityAt: string
  events: CaseEvent[]
}

interface AwdBucket {
  id: string
  challengeId: string
  title: string
  victimTeamName: string
  attemptCount: number
  successCount: number
  scoreEvents: number
  scoreTotal: number
  lastActivityAt: string
  events: CaseEvent[]
}

export function buildReviewArchiveCaseGroups(input: {
  timeline: TimelineEvent[]
  evidence: ReviewArchiveEvidenceItemData[]
  writeups: ReviewArchiveWriteupItemData[]
  manualReviews: ReviewArchiveManualReviewItemData[]
}): {
  practiceCases: ReviewArchiveCase[]
  awdCases: ReviewArchiveCase[]
} {
  const practiceBuckets = new Map<string, PracticeBucket>()
  const awdBuckets = new Map<string, AwdBucket>()

  input.timeline.forEach((item) => {
    const challengeId = item.challenge_id ? String(item.challenge_id) : ''
    if (!challengeId) return

    if (isAWDTimeline(item)) {
      const victimTeamName = extractVictimTeamName(item.detail)
      const bucket = ensureAWDBucket(awdBuckets, {
        challengeId,
        title: item.title,
        victimTeamName,
      })
      const isSuccess = Boolean(item.is_correct)
      bucket.attemptCount++
      if (isSuccess) bucket.successCount++
      if ((item.points ?? 0) > 0) {
        bucket.scoreEvents++
        bucket.scoreTotal += item.points ?? 0
      }
      pushCaseEvent(bucket.events, {
        id: item.id,
        label: isSuccess ? '攻击命中' : '攻击尝试',
        detail: item.detail || item.title,
        timestamp: item.created_at,
        stageLabel: isSuccess ? '命中结果' : '攻击尝试',
        tone: isSuccess ? 'success' : 'warning',
        meta: item.points ? `得分 ${item.points}` : undefined,
      })
      bucket.lastActivityAt = latestTime(bucket.lastActivityAt, item.created_at)
      return
    }

    if (!isPracticeTimeline(item)) return
    const bucket = ensurePracticeBucket(practiceBuckets, challengeId, item.title)
    const rawType = rawTimelineType(item)
    if (rawType === 'instance_start' || rawType === 'instance_destroy') {
      bucket.accessCount++
      pushCaseEvent(bucket.events, {
        id: item.id,
        label: rawType === 'instance_start' ? '接入目标' : '结束实例',
        detail: item.detail || item.title,
        timestamp: item.created_at,
        stageLabel: '接入',
        tone: 'neutral',
      })
    } else {
      bucket.submitCount++
      if (item.is_correct) bucket.successCount++
      pushCaseEvent(bucket.events, {
        id: item.id,
        label: item.is_correct ? '命中提交' : '提交尝试',
        detail: item.detail || item.title,
        timestamp: item.created_at,
        stageLabel: '提交',
        tone: item.is_correct ? 'success' : 'warning',
        meta: item.points ? `得分 ${item.points}` : undefined,
      })
    }
    bucket.lastActivityAt = latestTime(bucket.lastActivityAt, item.created_at)
  })

  input.evidence.forEach((item, index) => {
    const challengeId = String(item.challenge_id)
    if (!challengeId) return

    if (item.type === 'awd_attack_submission') {
      const victimTeamName =
        asString(item.meta?.victim_team_name) || extractVictimTeamName(item.detail)
      const bucket = ensureAWDBucket(awdBuckets, {
        challengeId,
        title: item.title,
        victimTeamName,
      })
      const isSuccess = Boolean(item.meta?.is_success)
      const score = numberFromUnknown(item.meta?.score_gained)
      bucket.attemptCount++
      if (isSuccess) bucket.successCount++
      if (score > 0) {
        bucket.scoreEvents++
        bucket.scoreTotal += score
      }
      pushCaseEvent(bucket.events, {
        id: `${item.type}-${challengeId}-${item.timestamp}-${index}`,
        label: isSuccess ? '攻击命中' : '攻击未命中',
        detail: item.detail || item.title,
        timestamp: item.timestamp,
        stageLabel: isSuccess ? '命中结果' : '攻击尝试',
        tone: isSuccess ? 'success' : 'warning',
        meta: score > 0 ? `得分 ${score}` : undefined,
      })
      bucket.lastActivityAt = latestTime(bucket.lastActivityAt, item.timestamp)
      return
    }

    if (!isPracticeEvidence(item.type)) return
    const bucket = ensurePracticeBucket(practiceBuckets, challengeId, item.title)
    const practiceEvent = practiceEventFromEvidence(item, index)
    if (practiceEvent.stage === 'access') bucket.accessCount++
    if (practiceEvent.stage === 'exploit') bucket.exploitCount++
    if (practiceEvent.stage === 'submit') {
      bucket.submitCount++
      if (practiceEvent.event.tone === 'success') bucket.successCount++
    }
    pushCaseEvent(bucket.events, practiceEvent.event)
    bucket.lastActivityAt = latestTime(bucket.lastActivityAt, item.timestamp)
  })

  input.writeups.forEach((item) => {
    const challengeId = String(item.challenge_id)
    const bucket = ensurePracticeBucket(practiceBuckets, challengeId, item.challenge_title)
    bucket.reflectionCount++
    bucket.writeupCount++
    const timestamp = item.published_at || item.updated_at
    pushCaseEvent(bucket.events, {
      id: `writeup-${item.id}`,
      label: '复盘输出',
      detail: item.title,
      timestamp,
      stageLabel: '复盘',
      tone: 'success',
      meta: item.is_recommended ? '推荐题解' : 'Writeup',
    })
    bucket.lastActivityAt = latestTime(bucket.lastActivityAt, timestamp)
  })

  input.manualReviews.forEach((item) => {
    const challengeId = String(item.challenge_id)
    const bucket = ensurePracticeBucket(practiceBuckets, challengeId, item.challenge_title)
    bucket.reflectionCount++
    bucket.manualReviewCount++
    pushCaseEvent(bucket.events, {
      id: `manual-review-${item.id}`,
      label: '人工审核',
      detail: item.review_comment || item.answer,
      timestamp: item.submitted_at,
      stageLabel: '复盘',
      tone: item.review_status === 'approved' ? 'success' : 'warning',
      meta: item.reviewer_name || '待审核',
    })
    bucket.lastActivityAt = latestTime(bucket.lastActivityAt, item.submitted_at)
  })

  return {
    practiceCases: Array.from(practiceBuckets.values())
      .map(toPracticeCase)
      .sort(sortCasesByLastActivity),
    awdCases: Array.from(awdBuckets.values()).map(toAWDCase).sort(sortCasesByLastActivity),
  }
}

function ensurePracticeBucket(
  buckets: Map<string, PracticeBucket>,
  challengeId: string,
  title: string
): PracticeBucket {
  const existing = buckets.get(challengeId)
  if (existing) return existing
  const bucket: PracticeBucket = {
    id: `practice-${challengeId}`,
    challengeId,
    title,
    accessCount: 0,
    exploitCount: 0,
    submitCount: 0,
    successCount: 0,
    reflectionCount: 0,
    writeupCount: 0,
    manualReviewCount: 0,
    lastActivityAt: '',
    events: [],
  }
  buckets.set(challengeId, bucket)
  return bucket
}

function ensureAWDBucket(
  buckets: Map<string, AwdBucket>,
  input: { challengeId: string; title: string; victimTeamName: string }
): AwdBucket {
  const key = `${input.challengeId}:${input.victimTeamName || 'unknown'}`
  const existing = buckets.get(key)
  if (existing) return existing
  const bucket: AwdBucket = {
    id: `awd-${key}`,
    challengeId: input.challengeId,
    title: input.title,
    victimTeamName: input.victimTeamName || '目标队伍',
    attemptCount: 0,
    successCount: 0,
    scoreEvents: 0,
    scoreTotal: 0,
    lastActivityAt: '',
    events: [],
  }
  buckets.set(key, bucket)
  return bucket
}

function pushCaseEvent(events: CaseEvent[], event: CaseEvent): void {
  events.push(event)
  events.sort((a, b) => compareTimeDesc(a.timestamp, b.timestamp))
}

function latestTime(current: string, incoming: string): string {
  if (!current) return incoming
  return compareTimeDesc(current, incoming) <= 0 ? incoming : current
}

function compareTimeDesc(left: string, right: string): number {
  return new Date(right).getTime() - new Date(left).getTime()
}

function rawTimelineType(item: TimelineEvent): string {
  return asString(item.meta?.raw_type) || item.type || ''
}

function isAWDTimeline(item: TimelineEvent): boolean {
  const rawType = rawTimelineType(item)
  return rawType === 'awd_attack_submit' || rawType === 'awd_attack_submission'
}

function isPracticeTimeline(item: TimelineEvent): boolean {
  const rawType = rawTimelineType(item)
  return (
    rawType === 'solve' ||
    rawType === 'submission' ||
    rawType === 'instance_start' ||
    rawType === 'instance_destroy' ||
    item.type === 'solve'
  )
}

function isPracticeEvidence(type: string): boolean {
  return (
    type === 'instance_access' ||
    type === 'instance_proxy_request' ||
    type === 'submission_correct' ||
    type === 'submission_attempt'
  )
}

function practiceEventFromEvidence(
  item: ReviewArchiveEvidenceItemData,
  index: number
): {
  stage: PracticeStageKey
  event: CaseEvent
} {
  const baseId = `${item.type}-${item.challenge_id}-${item.timestamp}-${index}`
  if (item.type === 'instance_access') {
    return {
      stage: 'access',
      event: {
        id: baseId,
        label: '接入目标',
        detail: item.detail || item.title,
        timestamp: item.timestamp,
        stageLabel: '接入',
        tone: 'neutral',
      },
    }
  }
  if (item.type === 'instance_proxy_request') {
    return {
      stage: 'exploit',
      event: {
        id: baseId,
        label: '利用动作',
        detail: item.detail || item.title,
        timestamp: item.timestamp,
        stageLabel: '利用',
        tone: 'neutral',
        meta: asString(item.meta?.method),
      },
    }
  }
  const isSuccess = item.type === 'submission_correct'
  return {
    stage: 'submit',
    event: {
      id: baseId,
      label: isSuccess ? '命中提交' : '提交尝试',
      detail: item.detail || item.title,
      timestamp: item.timestamp,
      stageLabel: '提交',
      tone: isSuccess ? 'success' : 'warning',
      meta: numberFromUnknown(item.meta?.points) > 0 ? `得分 ${item.meta?.points}` : undefined,
    },
  }
}

function extractVictimTeamName(detail?: string): string {
  if (!detail) return ''
  const match = detail.match(/命中\s+(.+?)(?:，|,|$)/)
  return match?.[1]?.trim() || ''
}

function toPracticeCase(bucket: PracticeBucket): ReviewArchiveCase {
  const stages: Record<PracticeStageKey, ReviewArchiveStageSummary> = {
    access: { key: 'access', label: '接入', count: bucket.accessCount },
    exploit: { key: 'exploit', label: '利用', count: bucket.exploitCount },
    submit: { key: 'submit', label: '提交', count: bucket.submitCount },
    reflection: { key: 'reflection', label: '复盘', count: bucket.reflectionCount },
  }
  return {
    id: bucket.id,
    title: bucket.title,
    subtitle: `${bucket.events.length} 条训练事件`,
    statusLabel: bucket.successCount > 0 ? '已形成命中' : '持续练习中',
    tone: bucket.successCount > 0 ? 'success' : 'neutral',
    eventCount: bucket.events.length,
    lastActivityAt: bucket.lastActivityAt,
    metrics: [
      { label: '有效提交', value: String(bucket.successCount) },
      { label: 'Writeup', value: String(bucket.writeupCount) },
      { label: '人工审核', value: String(bucket.manualReviewCount) },
    ],
    stages: Object.values(stages),
    events: bucket.events,
  }
}

function toAWDCase(bucket: AwdBucket): ReviewArchiveCase {
  const stages: Record<AwdStageKey, ReviewArchiveStageSummary> = {
    attack: { key: 'attack', label: '攻击', count: bucket.attemptCount },
    result: { key: 'result', label: '命中', count: bucket.successCount },
    score: { key: 'score', label: '得分', count: bucket.scoreEvents },
  }
  return {
    id: bucket.id,
    title: bucket.title,
    subtitle: bucket.victimTeamName,
    statusLabel: bucket.successCount > 0 ? '有效命中' : '持续对抗中',
    tone: bucket.successCount > 0 ? 'success' : 'warning',
    eventCount: bucket.events.length,
    lastActivityAt: bucket.lastActivityAt,
    metrics: [
      { label: '攻击次数', value: String(bucket.attemptCount) },
      { label: '成功命中', value: String(bucket.successCount) },
      { label: '累计得分', value: String(bucket.scoreTotal) },
    ],
    stages: Object.values(stages),
    events: bucket.events,
  }
}

function sortCasesByLastActivity(left: ReviewArchiveCase, right: ReviewArchiveCase): number {
  return compareTimeDesc(left.lastActivityAt, right.lastActivityAt)
}

function asString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function numberFromUnknown(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

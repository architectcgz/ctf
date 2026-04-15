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
  const victimTeamName = input.victimTeamName || '未标记目标队伍'
  const key = `${input.challengeId}::${victimTeamName}`
  const existing = buckets.get(key)
  if (existing) return existing
  const bucket: AwdBucket = {
    id: `awd-${input.challengeId}-${victimTeamName}`,
    challengeId: input.challengeId,
    title: input.title,
    victimTeamName,
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

function toPracticeCase(bucket: PracticeBucket): ReviewArchiveCase {
  const statusLabel =
    bucket.successCount > 0 && bucket.reflectionCount > 0
      ? '已形成闭环'
      : bucket.successCount > 0
        ? '已完成'
        : bucket.submitCount > 0 || bucket.exploitCount > 0 || bucket.accessCount > 0
          ? '进行中'
          : '已记录'

  const tone: CaseTone =
    bucket.successCount > 0
      ? 'success'
      : bucket.submitCount > 0 || bucket.exploitCount > 0
        ? 'warning'
        : 'neutral'

  return {
    id: bucket.id,
    title: bucket.title,
    subtitle: '练习轨迹',
    statusLabel,
    tone,
    eventCount: bucket.events.length,
    lastActivityAt: bucket.lastActivityAt,
    metrics: [
      { label: '有效提交', value: String(bucket.successCount) },
      { label: '复盘材料', value: String(bucket.writeupCount + bucket.manualReviewCount) },
      { label: '最近活动', value: bucket.lastActivityAt || '--' },
    ],
    stages: [
      { key: 'access', label: '接入', count: bucket.accessCount },
      { key: 'exploit', label: '利用', count: bucket.exploitCount },
      { key: 'submit', label: '提交', count: bucket.submitCount },
      { key: 'reflection', label: '复盘', count: bucket.reflectionCount },
    ],
    events: bucket.events.sort(sortEventsByTime),
  }
}

function toAWDCase(bucket: AwdBucket): ReviewArchiveCase {
  const hit = bucket.successCount > 0
  return {
    id: bucket.id,
    title: bucket.title,
    subtitle: bucket.victimTeamName,
    statusLabel: hit ? '攻击命中' : '攻击未命中',
    tone: hit ? 'success' : 'warning',
    eventCount: bucket.events.length,
    lastActivityAt: bucket.lastActivityAt,
    metrics: [
      { label: '攻击次数', value: String(bucket.attemptCount) },
      { label: '累计得分', value: String(bucket.scoreTotal) },
      { label: '最近活动', value: bucket.lastActivityAt || '--' },
    ],
    stages: [
      { key: 'attack', label: '攻击尝试', count: bucket.attemptCount },
      { key: 'result', label: '命中结果', count: bucket.events.length },
      { key: 'score', label: '得分变化', count: bucket.scoreEvents },
    ],
    events: bucket.events.sort(sortEventsByTime),
  }
}

function practiceEventFromEvidence(
  item: ReviewArchiveEvidenceItemData,
  index: number
): { stage: PracticeStageKey; event: CaseEvent } {
  if (item.type === 'instance_access') {
    return {
      stage: 'access',
      event: {
        id: `${item.type}-${item.challenge_id}-${item.timestamp}-${index}`,
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
        id: `${item.type}-${item.challenge_id}-${item.timestamp}-${index}`,
        label: '利用操作',
        detail: item.detail || item.title,
        timestamp: item.timestamp,
        stageLabel: '利用',
        tone: 'neutral',
      },
    }
  }

  const isCorrect = Boolean(item.meta?.is_correct)
  const points = numberFromUnknown(item.meta?.points)
  return {
    stage: 'submit',
    event: {
      id: `${item.type}-${item.challenge_id}-${item.timestamp}-${index}`,
      label: isCorrect ? '命中提交' : '提交尝试',
      detail: item.detail || item.title,
      timestamp: item.timestamp,
      stageLabel: '提交',
      tone: isCorrect ? 'success' : 'warning',
      meta: points > 0 ? `得分 ${points}` : undefined,
    },
  }
}

function isPracticeTimeline(item: TimelineEvent): boolean {
  const rawType = rawTimelineType(item)
  return rawType === 'instance_start' || rawType === 'instance_destroy' || rawType === 'flag_submit'
}

function isAWDTimeline(item: TimelineEvent): boolean {
  return rawTimelineType(item) === 'awd_attack_submit'
}

function rawTimelineType(item: TimelineEvent): string {
  const raw = asString(item.meta?.raw_type)
  return raw || item.type
}

function isPracticeEvidence(type: string): boolean {
  return (
    type === 'instance_access' ||
    type === 'instance_proxy_request' ||
    type === 'challenge_submission'
  )
}

function extractVictimTeamName(detail?: string): string {
  if (!detail) return ''
  const hit = detail.match(/^AWD 攻击命中\s+(.+?)(?:，得分\s+\d+)?$/)
  if (hit?.[1]) return hit[1]
  const fail = detail.match(/^AWD 攻击未命中\s+(.+)$/)
  if (fail?.[1]) return fail[1]
  return ''
}

function latestTime(current: string, next: string): string {
  if (!current) return next
  return new Date(next).getTime() > new Date(current).getTime() ? next : current
}

function numberFromUnknown(value: unknown): number {
  return typeof value === 'number' ? value : 0
}

function asString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function pushCaseEvent(events: CaseEvent[], event: CaseEvent): void {
  events.push(event)
}

function sortEventsByTime(left: CaseEvent, right: CaseEvent): number {
  return new Date(left.timestamp).getTime() - new Date(right.timestamp).getTime()
}

function sortCasesByLastActivity(left: ReviewArchiveCase, right: ReviewArchiveCase): number {
  return new Date(right.lastActivityAt).getTime() - new Date(left.lastActivityAt).getTime()
}

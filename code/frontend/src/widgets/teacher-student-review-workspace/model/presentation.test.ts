import { describe, expect, it } from 'vitest'

import {
  buildChallengeFilterOptions,
  buildReviewWorkspaceObservations,
  buildTeacherStudentReviewSummaryItems,
  eventMetaItems,
  sessionPathSummary,
  sessionResultLabel,
  TEACHER_STUDENT_REVIEW_WORKSPACE_COPY,
} from './presentation'

describe('teacher student review workspace presentation', () => {
  it('应提供稳定的空状态文案和摘要项顺序', () => {
    const items = buildTeacherStudentReviewSummaryItems({
      sessionSummary: {
        total_sessions: 3,
        event_count: 8,
        success_count: 2,
      },
      evidenceSummary: {
        proxy_request_count: 5,
      },
    })

    expect(TEACHER_STUDENT_REVIEW_WORKSPACE_COPY.emptyTitle).toBe('暂无攻击会话')
    expect(items.map((item) => item.key)).toEqual([
      'total_sessions',
      'event_count',
      'success_count',
      'proxy_request_count',
    ])
    expect(items[2].value).toBe(2)
    expect(items[3].value).toBe(5)
  })

  it('应生成事件元数据标签并兼容评审状态', () => {
    const items = eventMetaItems({
      id: 'evt-1',
      type: 'manual_review',
      stage: 'review',
      source: 'submissions',
      occurred_at: '2026-05-03T08:00:00Z',
      actor: { user_id: 'stu-1' },
      target: {},
      summary: 'summary',
      capture_available: false,
      meta: {
        request_method: 'POST',
        target_path: '/submit',
        status_code: 200,
        review_status: 'approved',
        writeup_title: '从回显到 flag',
      },
    })

    expect(items.map((item) => item.label)).toContain('POST')
    expect(items.map((item) => item.label)).toContain('/submit')
    expect(items.map((item) => item.label)).toContain('200')
    expect(items.map((item) => item.label)).toContain('已通过')
    expect(items.map((item) => item.label)).toContain('从回显到 flag')
  })

  it('应生成挑战筛选项和复盘观察点', () => {
    const options = buildChallengeFilterOptions({
      evidence: {
        summary: {
          total_events: 4,
          proxy_request_count: 1,
          submit_count: 2,
          success_count: 1,
          challenge_id: '11',
        },
        events: [
          {
            type: 'challenge_submission',
            challenge_id: '11',
            title: 'web-1',
            detail: '提交未命中 Flag',
            timestamp: '2026-05-03T08:00:00Z',
            meta: {
              is_correct: false,
            },
          },
          {
            type: 'challenge_submission',
            challenge_id: '11',
            title: 'web-1',
            detail: '提交未命中 Flag',
            timestamp: '2026-05-03T08:02:00Z',
            meta: {
              is_correct: false,
            },
          },
          {
            type: 'writeup',
            challenge_id: '11',
            title: 'web-1',
            detail: '提交题解',
            timestamp: '2026-05-03T08:12:00Z',
          },
        ],
      },
      attackSessions: {
        sessions: [
          {
            id: 'sess-1',
            mode: 'practice',
            student_id: 'stu-1',
            challenge_id: '11',
            title: 'web-1',
            started_at: '2026-05-03T08:00:00Z',
            ended_at: '2026-05-03T08:10:00Z',
            result: 'success',
            event_count: 2,
            capture_count: 0,
          },
        ],
      },
    })

    const observations = buildReviewWorkspaceObservations({
      evidence: {
        summary: {
          total_events: 4,
          proxy_request_count: 1,
          submit_count: 2,
          success_count: 1,
          challenge_id: '11',
        },
        events: [
          {
            type: 'challenge_submission',
            challenge_id: '11',
            title: 'web-1',
            detail: '提交未命中 Flag',
            timestamp: '2026-05-03T08:00:00Z',
            meta: {
              is_correct: false,
            },
          },
          {
            type: 'challenge_submission',
            challenge_id: '11',
            title: 'web-1',
            detail: '提交未命中 Flag',
            timestamp: '2026-05-03T08:02:00Z',
            meta: {
              is_correct: false,
            },
          },
          {
            type: 'writeup',
            challenge_id: '11',
            title: 'web-1',
            detail: '提交题解',
            timestamp: '2026-05-03T08:12:00Z',
          },
        ],
      },
      attackSessions: {
        sessions: [
          {
            id: 'sess-1',
            mode: 'practice',
            student_id: 'stu-1',
            challenge_id: '11',
            title: 'web-1',
            started_at: '2026-05-03T08:00:00Z',
            ended_at: '2026-05-03T08:10:00Z',
            result: 'success',
            event_count: 2,
            capture_count: 0,
          },
        ],
      },
    })

    expect(options).toEqual([{ value: '11', label: 'web-1' }])
    expect(observations.map((item) => item.key)).toContain('hands_on_activity')
    expect(observations.map((item) => item.key)).toContain('submission_stability')
    expect(observations.map((item) => item.key)).toContain('training_closure')
  })

  it('应生成会话路径摘要和结果标签', () => {
    expect(
      sessionPathSummary({
        id: 'sess-1',
        mode: 'practice',
        student_id: 'stu-1',
        title: 'web-1',
        started_at: '2026-05-03T08:00:00Z',
        ended_at: '2026-05-03T08:30:00Z',
        result: 'success',
        event_count: 2,
        capture_count: 0,
        events: [
          {
            id: 'evt-1',
            type: 'instance_access',
            stage: 'access',
            source: 'audit_logs',
            occurred_at: '2026-05-03T08:00:00Z',
            actor: { user_id: 'stu-1' },
            target: {},
            summary: '访问',
            capture_available: false,
          },
          {
            id: 'evt-2',
            type: 'challenge_submission',
            stage: 'submit',
            source: 'submissions',
            occurred_at: '2026-05-03T08:10:00Z',
            actor: { user_id: 'stu-1' },
            target: {},
            summary: '提交',
            capture_available: false,
          },
        ],
      })
    ).toBe('访问 -> 提交')
    expect(sessionResultLabel('success')).toBe('成功')
  })
})

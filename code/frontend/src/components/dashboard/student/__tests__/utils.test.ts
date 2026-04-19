import { describe, expect, it } from 'vitest'

import type { TimelineEvent } from '@/api/contracts'

import { timelineTypeTone } from '../utils'

describe('student dashboard utils', () => {
  const buildTimelineEvent = (type: TimelineEvent['type']): TimelineEvent => ({
    id: '1',
    type,
    created_at: '2026-04-19T10:00:00Z',
    title: 'timeline event',
  })

  it('timelineTypeTone 应返回语义类，而不是直接拼主题 utility', () => {
    expect(timelineTypeTone(buildTimelineEvent('challenge_detail_view'))).toBe(
      'timeline-type-pill timeline-type-pill--primary'
    )
    expect(timelineTypeTone(buildTimelineEvent('instance_proxy_request'))).toBe(
      'timeline-type-pill timeline-type-pill--danger'
    )
    expect(timelineTypeTone(buildTimelineEvent('solve'))).toBe(
      'timeline-type-pill timeline-type-pill--success'
    )
    expect(timelineTypeTone(buildTimelineEvent('submit'))).toBe(
      'timeline-type-pill timeline-type-pill--warning'
    )
    expect(timelineTypeTone(buildTimelineEvent('hint'))).toBe(
      'timeline-type-pill timeline-type-pill--reverse'
    )
  })
})

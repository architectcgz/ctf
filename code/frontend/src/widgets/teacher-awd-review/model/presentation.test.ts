import { describe, expect, it } from 'vitest'

import {
  buildTeacherAwdReviewIndexSummaryItems,
  buildTeacherAwdReviewSummaryItems,
  TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY,
  TEACHER_AWD_REVIEW_WORKSPACE_COPY,
} from './presentation'

describe('teacher awd review widget presentation', () => {
  it('应生成固定顺序的摘要项', () => {
    const items = buildTeacherAwdReviewSummaryItems(
      {
        roundCount: 4,
        teamCount: 6,
        serviceCount: 12,
        attackCount: 8,
        trafficCount: 20,
      },
      false
    )

    expect(items).toHaveLength(4)
    expect(items[0].label).toBe('轮次范围')
    expect(items[1].label).toBe('参与队伍')
    expect(items[2].value).toBe('12 / 8 / 20')
    expect(items[3].value).toBe('链路就绪')
    expect(items[3].valueClass).toBe('awd-review-status-text')
  })

  it('轮询中应显示后台处理中状态', () => {
    const items = buildTeacherAwdReviewSummaryItems(
      {
        roundCount: 1,
        teamCount: 1,
        serviceCount: 1,
        attackCount: 1,
        trafficCount: 1,
      },
      true
    )

    expect(items[3].value).toBe('后台处理中...')
  })

  it('应提供工作区文案出口', () => {
    expect(TEACHER_AWD_REVIEW_WORKSPACE_COPY.overline).toBe('AWD Review')
    expect(TEACHER_AWD_REVIEW_WORKSPACE_COPY.title).toBe('AWD复盘')
    expect(TEACHER_AWD_REVIEW_WORKSPACE_COPY.descriptionSuffix).toContain('多维复盘攻防实战过程')
  })

  it('应生成目录页摘要项与文案出口', () => {
    const items = buildTeacherAwdReviewIndexSummaryItems({
      totalCount: 10,
      runningCount: 4,
      exportReadyCount: 6,
    })

    expect(items).toHaveLength(3)
    expect(items[0].label).toBe('赛事数量')
    expect(items[1].value).toBe(4)
    expect(items[2].value).toBe(6)
    expect(TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.summaryTitle).toBe('Review Snapshot')
    expect(TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.openDashboardAction).toBe('教学概览')
  })
})

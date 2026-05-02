import { describe, expect, it } from 'vitest'

import {
  buildReviewArchiveSummaryCards,
  rankReviewArchiveSkillDimensions,
  REVIEW_ARCHIVE_STATE_COPY,
  REVIEW_ARCHIVE_SUMMARY_COPY,
} from './presentation'

describe('teacher review archive widget presentation', () => {
  it('应保持状态文案出口稳定', () => {
    expect(REVIEW_ARCHIVE_STATE_COPY.errorTitle).toBe('复盘归档加载失败')
    expect(REVIEW_ARCHIVE_STATE_COPY.emptyTitle).toBe('暂无复盘归档')
    expect(REVIEW_ARCHIVE_STATE_COPY.reload).toBe('重新加载')
  })

  it('应保持摘要文案出口稳定', () => {
    expect(REVIEW_ARCHIVE_SUMMARY_COPY.summaryTitle).toBe('训练摘要')
    expect(REVIEW_ARCHIVE_SUMMARY_COPY.skillTitle).toBe('能力画像')
    expect(REVIEW_ARCHIVE_SUMMARY_COPY.correctSubmissionLabel).toBe('有效提交')
  })

  it('应生成固定顺序的摘要卡片 schema', () => {
    const cards = buildReviewArchiveSummaryCards({
      solvedRate: 50,
      totalSolved: 10,
      totalChallenges: 20,
      correctSubmissionCount: 4,
      formattedLastActivity: '2026-04-01 09:20',
    })

    expect(cards).toHaveLength(3)
    expect(cards[0].key).toBe('solved_rate')
    expect(cards[0].value).toBe('50%')
    expect(cards[1].key).toBe('correct_submission')
    expect(cards[2].key).toBe('latest_activity')
    expect(cards[2].valueClass).toBe('summary-card__value--time')
  })

  it('应按得分降序排列能力维度', () => {
    const ranked = rankReviewArchiveSkillDimensions([
      { key: 'crypto', name: '密码', value: 48 },
      { key: 'web', name: 'Web', value: 82 },
      { key: 'misc', name: '杂项', value: 70 },
    ])

    expect(ranked.map((item) => item.key)).toEqual(['web', 'misc', 'crypto'])
  })
})

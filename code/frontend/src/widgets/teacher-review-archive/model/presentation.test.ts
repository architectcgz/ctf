import { describe, expect, it } from 'vitest'

import {
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
})

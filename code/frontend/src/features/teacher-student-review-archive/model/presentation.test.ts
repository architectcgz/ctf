import { describe, expect, it } from 'vitest'

import {
  resolveTeacherStudentReviewArchiveErrorMessage,
  TEACHER_STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES,
} from './presentation'

describe('teacher student review archive presentation', () => {
  it('应保持导出文案出口稳定', () => {
    expect(TEACHER_STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.success).toBe(
      '复盘归档已生成并开始下载'
    )
    expect(TEACHER_STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.pending).toBe(
      '复盘归档开始生成，完成后会自动下载'
    )
  })

  it('应优先返回 error message，其次回退 fallback', () => {
    expect(
      resolveTeacherStudentReviewArchiveErrorMessage(new Error('导出失败'), 'fallback')
    ).toBe('导出失败')
    expect(resolveTeacherStudentReviewArchiveErrorMessage(new Error('   '), 'fallback')).toBe(
      'fallback'
    )
    expect(resolveTeacherStudentReviewArchiveErrorMessage('oops', 'fallback')).toBe('fallback')
  })
})

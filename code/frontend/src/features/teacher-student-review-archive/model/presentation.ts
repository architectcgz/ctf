export const TEACHER_STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES = {
  success: '复盘归档已生成并开始下载',
  pending: '复盘归档开始生成，完成后会自动下载',
  downloadFailed: '复盘归档下载失败，请稍后重试',
  generationFailed: '复盘归档生成失败',
  pollingFailed: '复盘归档生成状态同步失败，请稍后重试',
  exportFailed: '复盘归档导出失败，请稍后重试',
} as const

export function resolveTeacherStudentReviewArchiveErrorMessage(
  error: unknown,
  fallback: string
): string {
  if (!(error instanceof Error)) return fallback
  const message = error.message.trim()
  return message || fallback
}

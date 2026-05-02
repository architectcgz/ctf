import { ApiError, type ApiValidationIssue } from '@/api/request'

export interface UploadErrorDetail {
  message: string
  code?: number
  requestId?: string
}

export function formatChallengeImportValidationIssue(issue: ApiValidationIssue): string {
  const field = issue.field?.trim()
  const message = issue.message?.trim()
  if (field && message) {
    return `${field}: ${message}`
  }
  return message || field || ''
}

export function buildFriendlyChallengeImportMessage(message: string, code: number | undefined): string {
  const normalizedMessage = message.trim()
  const isGenericParameterError =
    code === 10001 ||
    normalizedMessage === '请求参数错误' ||
    normalizedMessage === '参数校验失败，请检查输入'

  if (isGenericParameterError) {
    return '题目包格式校验失败，请确认 Zip 根目录包含 challenge.yml，并对照“查看题目包示例”检查字段。'
  }

  return normalizedMessage || '题目包解析失败'
}

export function normalizeChallengeImportError(error: unknown): UploadErrorDetail {
  if (error instanceof ApiError) {
    const fieldErrors = (error.errors ?? [])
      .map(formatChallengeImportValidationIssue)
      .filter((item) => item.length > 0)

    if (fieldErrors.length > 0) {
      return {
        message: `参数校验失败：${fieldErrors.join('；')}`,
        code: error.code,
        requestId: error.requestId,
      }
    }

    const fallbackMessage = buildFriendlyChallengeImportMessage(error.message, error.code)
    return {
      message: fallbackMessage,
      code: error.code,
      requestId: error.requestId,
    }
  }

  if (error instanceof Error) {
    return {
      message: error.message || '题目包解析失败',
    }
  }

  return {
    message: '题目包解析失败',
  }
}

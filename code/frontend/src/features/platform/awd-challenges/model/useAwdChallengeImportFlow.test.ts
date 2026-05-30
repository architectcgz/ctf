import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '@/api/request'
import { useAwdChallengeImportFlow } from './useAwdChallengeImportFlow'

const awdAuthoringApiMocks = vi.hoisted(() => ({
  commitAdminAwdChallengeImport: vi.fn(),
  listAdminAwdChallengeImports: vi.fn(),
  previewAdminAwdChallengeImport: vi.fn(),
}))

vi.mock('@/api/admin/awd-authoring', () => awdAuthoringApiMocks)

describe('useAwdChallengeImportFlow', () => {
  const refreshChallenges = vi.fn()
  const notifySuccess = vi.fn()
  const notifyError = vi.fn()

  beforeEach(() => {
    awdAuthoringApiMocks.commitAdminAwdChallengeImport.mockReset()
    awdAuthoringApiMocks.listAdminAwdChallengeImports.mockReset()
    awdAuthoringApiMocks.previewAdminAwdChallengeImport.mockReset()
    refreshChallenges.mockReset()
    notifySuccess.mockReset()
    notifyError.mockReset()
    refreshChallenges.mockResolvedValue(undefined)
    awdAuthoringApiMocks.listAdminAwdChallengeImports.mockResolvedValue([])
  })

  it('commit 冲突时沿用 humanizeRequestError 的后端提示', async () => {
    const humanizeRequestError = vi.fn((error: unknown, fallback: string) => {
      if (error instanceof Error && error.message.trim()) {
        return error.message
      }
      return fallback
    })

    awdAuthoringApiMocks.commitAdminAwdChallengeImport.mockRejectedValue(
      new ApiError('AWD 题目 slug awd-bank-portal-01 已被已有题目占用，请改用题目编辑入口更新', {
        code: 10007,
        status: 409,
      })
    )

    const flow = useAwdChallengeImportFlow({
      refreshChallenges,
      humanizeRequestError,
      notifySuccess,
      notifyError,
    })

    const result = await flow.commitImportPreview({
      id: 'imp-1',
      file_name: 'awd-bank-portal-01.zip',
      slug: 'awd-bank-portal-01',
      title: 'Bank Portal AWD',
      category: 'web',
      difficulty: 'hard',
      description: 'desc',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      version: 'v2026.04',
      checker_type: 'http_standard',
      created_at: '2026-05-03T22:24:00.000Z',
    })

    expect(result).toBeNull()
    expect(humanizeRequestError).toHaveBeenCalled()
    expect(notifyError).toHaveBeenCalledWith(
      'AWD 题目 slug awd-bank-portal-01 已被已有题目占用，请改用题目编辑入口更新'
    )
    expect(notifySuccess).not.toHaveBeenCalled()
  })
})

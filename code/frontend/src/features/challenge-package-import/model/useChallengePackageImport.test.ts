import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { ApiError } from '@/api/request'
import { useChallengePackageImport } from './useChallengePackageImport'

const authoringApiMocks = vi.hoisted(() => ({
  commitChallengeImport: vi.fn(),
  getChallengeImport: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}))

const uploadFlowMocks = vi.hoisted(() => ({
  refreshQueue: vi.fn(),
  selectPackage: vi.fn(),
  selectPackages: vi.fn(),
}))

vi.mock('@/api/admin/authoring', () => authoringApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('./challengeImportUploadFlow', () => ({
  useChallengeImportUploadFlow: () => uploadFlowMocks,
}))

describe('useChallengePackageImport', () => {
  beforeEach(() => {
    authoringApiMocks.commitChallengeImport.mockReset()
    authoringApiMocks.getChallengeImport.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    uploadFlowMocks.refreshQueue.mockReset()
    uploadFlowMocks.selectPackage.mockReset()
    uploadFlowMocks.selectPackages.mockReset()
    uploadFlowMocks.refreshQueue.mockResolvedValue(undefined)
  })

  it('commit 冲突时透传后端错误，而不是退化成通用失败文案', async () => {
    authoringApiMocks.commitChallengeImport.mockRejectedValue(
      new ApiError('题目 slug web-sqli-101 已被已有题目占用，请改用题目编辑入口更新', {
        code: 10007,
        status: 409,
      })
    )

    let composable!: ReturnType<typeof useChallengePackageImport>
    const Harness = defineComponent({
      setup() {
        composable = useChallengePackageImport()
        return () => null
      },
    })

    mount(Harness)
    composable.preview.value = {
      id: 'imp-1',
      file_name: 'web-sqli-101.zip',
      slug: 'web-sqli-101',
      title: 'SQL Injection 101',
      description: 'desc',
      category: 'web',
      difficulty: 'easy',
      points: 100,
      attachments: [],
      hints: [],
      flag: { type: 'static', prefix: 'flag' },
      runtime: { type: 'container', image_ref: 'ctf/web-sqli-101:latest' },
      extensions: { topology: { source: '', enabled: false } },
      package_files: [],
      warnings: [],
      created_at: '2026-05-03T22:24:00.000Z',
    }

    const result = await composable.commitPreview()
    await flushPromises()

    expect(result).toBeNull()
    expect(toastMocks.error).toHaveBeenCalledWith(
      '题目 slug web-sqli-101 已被已有题目占用，请改用题目编辑入口更新'
    )
    expect(toastMocks.error).not.toHaveBeenCalledWith('题目导入失败')
  })
})

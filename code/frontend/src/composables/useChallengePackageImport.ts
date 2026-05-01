import { computed, shallowRef } from 'vue'

import {
  commitChallengeImport,
  getChallengeImport,
  listChallengeImports,
  previewChallengeImport,
} from '@/api/admin/authoring'
import { ApiError, type ApiValidationIssue } from '@/api/request'
import type { AdminChallengeImportCommitData, AdminChallengeImportPreview } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseChallengePackageImportOptions {
  onCommitted?: (result: AdminChallengeImportCommitData) => Promise<void> | void
}

interface UploadResultBase {
  fileName: string
  message: string
  code?: number
  requestId?: string
}

interface UploadErrorDetail {
  message: string
  code?: number
  requestId?: string
}

export interface ChallengePackageUploadResult extends UploadResultBase {
  id: string
  status: 'success' | 'error'
  createdAt: string
}

const MAX_UPLOAD_RESULTS = 12

interface SelectPackagesOptions {
  parallel?: boolean
}

export function useChallengePackageImport(options: UseChallengePackageImportOptions = {}) {
  const toast = useToast()
  const preview = shallowRef<AdminChallengeImportPreview | null>(null)
  const uploading = shallowRef(false)
  const committing = shallowRef(false)
  const queueLoading = shallowRef(false)
  const selectedFileName = shallowRef('')
  const queue = shallowRef<AdminChallengeImportPreview[]>([])
  const uploadResults = shallowRef<ChallengePackageUploadResult[]>([])

  const hasPreview = computed(() => preview.value !== null)
  const primaryAttachment = computed(() => preview.value?.attachments?.[0])

  function appendUploadResult(
    status: ChallengePackageUploadResult['status'],
    payload: UploadResultBase
  ): void {
    const item: ChallengePackageUploadResult = {
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      status,
      createdAt: new Date().toISOString(),
      ...payload,
    }

    uploadResults.value = [item, ...uploadResults.value].slice(0, MAX_UPLOAD_RESULTS)
  }

  function normalizeUploadError(error: unknown): UploadErrorDetail {
    if (error instanceof ApiError) {
      const fieldErrors = (error.errors ?? [])
        .map(formatValidationIssue)
        .filter((item) => item.length > 0)

      if (fieldErrors.length > 0) {
        return {
          message: `参数校验失败：${fieldErrors.join('；')}`,
          code: error.code,
          requestId: error.requestId,
        }
      }

      const fallbackMessage = buildFriendlyUploadMessage(error.message, error.code)
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

  function formatValidationIssue(issue: ApiValidationIssue): string {
    const field = issue.field?.trim()
    const message = issue.message?.trim()
    if (field && message) {
      return `${field}: ${message}`
    }
    return message || field || ''
  }

  function buildFriendlyUploadMessage(message: string, code: number | undefined): string {
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

  async function refreshQueue() {
    queueLoading.value = true
    try {
      queue.value = await listChallengeImports()
    } catch {
      toast.error('加载导入任务失败')
    } finally {
      queueLoading.value = false
    }
  }

  async function selectPackage(file: File): Promise<AdminChallengeImportPreview | null> {
    selectedFileName.value = file.name
    uploading.value = true
    try {
      preview.value = await previewChallengeImport(file)
      appendUploadResult('success', {
        fileName: file.name,
        message: '题目包解析完成，已生成导入预览',
      })
      toast.success('题目包解析完成')
      await refreshQueue()
      return preview.value
    } catch (error) {
      const normalizedError = normalizeUploadError(error)
      appendUploadResult('error', {
        fileName: file.name,
        message: normalizedError.message,
        code: normalizedError.code,
        requestId: normalizedError.requestId,
      })
      toast.error(`题目包解析失败：${normalizedError.message}`)
      return null
    } finally {
      uploading.value = false
    }
  }

  async function selectPackages(
    files: File[],
    options: SelectPackagesOptions = {}
  ): Promise<AdminChallengeImportPreview | null> {
    if (files.length === 0) {
      return null
    }

    if (!options.parallel || files.length === 1) {
      let latestSuccess: AdminChallengeImportPreview | null = null
      for (const file of files) {
        const result = await selectPackage(file)
        if (result) {
          latestSuccess = result
        }
      }
      return latestSuccess
    }

    selectedFileName.value = `${files[0]?.name || ''} +${Math.max(0, files.length - 1)}`
    uploading.value = true
    try {
      const tasks = files.map(async (file) => {
        try {
          const result = await previewChallengeImport(file)
          return { ok: true as const, file, result }
        } catch (error) {
          return { ok: false as const, file, error }
        }
      })

      const settled = await Promise.all(tasks)
      let latestSuccess: AdminChallengeImportPreview | null = null
      let successCount = 0

      for (const item of settled) {
        if (item.ok) {
          successCount += 1
          latestSuccess = item.result
          appendUploadResult('success', {
            fileName: item.file.name,
            message: '题目包解析完成，已生成导入预览',
          })
          continue
        }

        const normalizedError = normalizeUploadError(item.error)
        appendUploadResult('error', {
          fileName: item.file.name,
          message: normalizedError.message,
          code: normalizedError.code,
          requestId: normalizedError.requestId,
        })
      }

      if (latestSuccess) {
        preview.value = latestSuccess
        selectedFileName.value = latestSuccess.file_name
      }

      if (successCount === files.length) {
        toast.success(`批量解析完成（${successCount}/${files.length}）`)
      } else if (successCount > 0) {
        toast.warning(`批量解析部分成功（${successCount}/${files.length}）`)
      } else {
        toast.error(`批量解析失败（0/${files.length}）`)
      }

      await refreshQueue()
      return latestSuccess
    } finally {
      uploading.value = false
    }
  }

  async function loadPreview(id: string) {
    uploading.value = true
    try {
      preview.value = await getChallengeImport(id)
      selectedFileName.value = preview.value.file_name
    } catch {
      toast.error('加载导入预览失败')
    } finally {
      uploading.value = false
    }
  }

  function resetPreview() {
    preview.value = null
    selectedFileName.value = ''
  }

  async function commitPreview() {
    if (!preview.value) {
      return
    }

    committing.value = true
    try {
      const result = await commitChallengeImport(preview.value.id)
      toast.success('题目导入成功')
      preview.value = null
      selectedFileName.value = ''
      await refreshQueue()
      await options.onCommitted?.(result)
      return result
    } catch {
      toast.error('题目导入失败')
      return null
    } finally {
      committing.value = false
    }
  }

  return {
    preview,
    uploading,
    committing,
    queueLoading,
    selectedFileName,
    queue,
    uploadResults,
    hasPreview,
    primaryAttachment,
    refreshQueue,
    selectPackage,
    selectPackages,
    loadPreview,
    resetPreview,
    commitPreview,
  }
}

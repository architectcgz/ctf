import { ref, type Ref } from 'vue'

import {
  commitAdminAwdChallengeImport,
  listAdminAwdChallengeImports,
  previewAdminAwdChallengeImport,
} from '@/api/admin/awd-authoring'
import type { AdminAwdChallengeImportPreview } from '@/api/contracts'

import type { PlatformAwdChallengeImportUploadResult } from './usePlatformAwdChallenges'

interface UseAwdChallengeImportFlowOptions {
  refreshChallenges: () => Promise<void>
  humanizeRequestError: (error: unknown, fallback: string) => string
  notifySuccess: (message: string) => void
  notifyError: (message: string) => void
}

export function useAwdChallengeImportFlow(options: UseAwdChallengeImportFlowOptions) {
  const { refreshChallenges, humanizeRequestError, notifySuccess, notifyError } = options

  const uploading = ref(false)
  const queueLoading = ref(false)
  const selectedImportFileName = ref('')
  const importQueue = ref<AdminAwdChallengeImportPreview[]>([])
  const uploadResults = ref<PlatformAwdChallengeImportUploadResult[]>([])

  function appendUploadResult(
    result: Omit<PlatformAwdChallengeImportUploadResult, 'id' | 'createdAt'>
  ) {
    uploadResults.value = [
      {
        id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
        createdAt: new Date().toISOString(),
        ...result,
      },
      ...uploadResults.value,
    ].slice(0, 8)
  }

  async function refreshImportQueue() {
    queueLoading.value = true
    try {
      importQueue.value = await listAdminAwdChallengeImports()
    } catch (error) {
      notifyError(humanizeRequestError(error, '加载 AWD 导入队列失败'))
    } finally {
      queueLoading.value = false
    }
  }

  async function selectImportPackages(files: File[]) {
    if (files.length === 0) {
      return null
    }

    uploading.value = true
    let latestSuccess: AdminAwdChallengeImportPreview | null = null

    try {
      for (const file of files) {
        selectedImportFileName.value = file.name
        try {
          const preview = await previewAdminAwdChallengeImport(file)
          latestSuccess = preview
          appendUploadResult({
            status: 'success',
            fileName: file.name,
            message: 'AWD 题目包解析完成，已进入待确认导入队列。',
          })
        } catch (error) {
          appendUploadResult({
            status: 'error',
            fileName: file.name,
            message: humanizeRequestError(error, 'AWD 题目包解析失败'),
          })
        }
      }

      if (latestSuccess) {
        notifySuccess('AWD 题目包解析完成')
      } else {
        notifyError('AWD 题目包解析失败')
      }
      await refreshImportQueue()
      return latestSuccess
    } finally {
      uploading.value = false
    }
  }

  async function commitImportPreview(preview: AdminAwdChallengeImportPreview) {
    try {
      const result = await commitAdminAwdChallengeImport(preview.id)
      notifySuccess(`已导入题目 ${result.challenge.name}`)
      await Promise.all([refreshChallenges(), refreshImportQueue()])
      return result
    } catch (error) {
      notifyError(humanizeRequestError(error, '导入 AWD 题目失败'))
      return null
    }
  }

  return {
    uploading,
    queueLoading,
    selectedImportFileName,
    importQueue,
    uploadResults,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
  }
}

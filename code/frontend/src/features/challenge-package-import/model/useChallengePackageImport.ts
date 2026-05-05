import { computed, shallowRef } from 'vue'

import {
  commitChallengeImport,
  getChallengeImport,
} from '@/api/admin/authoring'
import type { AdminChallengeImportCommitData, AdminChallengeImportPreview } from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import {
  useChallengeImportUploadFlow,
  type ChallengePackageUploadResult,
} from './challengeImportUploadFlow'

interface UseChallengePackageImportOptions {
  onCommitted?: (result: AdminChallengeImportCommitData) => Promise<void> | void
}

export type { ChallengePackageUploadResult }

function humanizeCommitError(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
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
  const { refreshQueue, selectPackage, selectPackages } = useChallengeImportUploadFlow({
    toast,
    preview,
    uploading,
    queueLoading,
    selectedFileName,
    queue,
    uploadResults,
  })

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
    } catch (error) {
      toast.error(humanizeCommitError(error, '题目导入失败'))
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

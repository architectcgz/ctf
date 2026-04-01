import { computed, shallowRef } from 'vue'

import {
  commitChallengeImport,
  previewChallengeImport,
} from '@/api/admin'
import type { AdminChallengeImportCommitData, AdminChallengeImportPreview } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseChallengePackageImportOptions {
  onCommitted?: (result: AdminChallengeImportCommitData) => Promise<void> | void
}

export function useChallengePackageImport(options: UseChallengePackageImportOptions = {}) {
  const toast = useToast()
  const preview = shallowRef<AdminChallengeImportPreview | null>(null)
  const uploading = shallowRef(false)
  const committing = shallowRef(false)
  const selectedFileName = shallowRef('')

  const hasPreview = computed(() => preview.value !== null)
  const primaryAttachment = computed(() => preview.value?.attachments?.[0])

  async function selectPackage(file: File) {
    selectedFileName.value = file.name
    uploading.value = true
    try {
      preview.value = await previewChallengeImport(file)
      toast.success('题目包解析完成')
    } catch {
      toast.error('题目包解析失败')
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
    selectedFileName,
    hasPreview,
    primaryAttachment,
    selectPackage,
    resetPreview,
    commitPreview,
  }
}

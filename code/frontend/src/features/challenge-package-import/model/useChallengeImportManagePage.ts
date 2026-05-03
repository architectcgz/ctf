import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useChallengePackageImport } from './useChallengePackageImport'

export function useChallengeImportManagePage() {
  const router = useRouter()
  const {
    uploading,
    queueLoading,
    selectedFileName,
    queue,
    uploadResults,
    refreshQueue,
    selectPackages,
  } = useChallengePackageImport()

  const queueCount = computed(() => queue.value.length)

  onMounted(() => {
    void refreshQueue()
  })

  async function handleSelectPackage(files: File[]) {
    const selectedPreview = await selectPackages(files, { parallel: files.length > 1 })
    if (!selectedPreview?.id) {
      return
    }

    await router.push({
      name: 'PlatformChallengeImportPreview',
      params: { importId: selectedPreview.id },
    })
  }

  async function openPackageFormatGuide(): Promise<void> {
    await router.push({ name: 'PlatformChallengePackageFormat' })
  }

  async function backToChallenges(): Promise<void> {
    await router.push({ name: 'ChallengeManage' })
  }

  async function inspectImportTask(importId: string): Promise<void> {
    await router.push({
      name: 'PlatformChallengeImportPreview',
      params: { importId },
    })
  }

  function formatDateTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN')
  }

  return {
    uploading,
    queueLoading,
    selectedFileName,
    queue,
    uploadResults,
    refreshQueue,
    queueCount,
    handleSelectPackage,
    openPackageFormatGuide,
    backToChallenges,
    inspectImportTask,
    formatDateTime,
  }
}

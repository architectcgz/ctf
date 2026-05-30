import { computed, onMounted, ref } from 'vue'

import {
  buildChallengeImportPreviewRoute,
  buildChallengeManageRoute,
  buildChallengePackageFormatRoute,
} from './challengeImportRoutes'
import { useChallengePackageImport } from './useChallengePackageImport'

export function useChallengeImportManagePage() {
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
  const previewRedirectRoute = ref<ReturnType<typeof buildChallengeImportPreviewRoute> | null>(null)
  const backToChallengesRoute = buildChallengeManageRoute()
  const packageFormatGuideRoute = buildChallengePackageFormatRoute()

  onMounted(() => {
    void refreshQueue()
  })

  async function handleSelectPackage(files: File[]) {
    const selectedPreview = await selectPackages(files, { parallel: files.length > 1 })
    if (!selectedPreview?.id) {
      return
    }

    previewRedirectRoute.value = buildChallengeImportPreviewRoute(selectedPreview.id)
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
    previewRedirectRoute,
    backToChallengesRoute,
    packageFormatGuideRoute,
    buildImportPreviewRoute: buildChallengeImportPreviewRoute,
    handleSelectPackage,
    formatDateTime,
  }
}

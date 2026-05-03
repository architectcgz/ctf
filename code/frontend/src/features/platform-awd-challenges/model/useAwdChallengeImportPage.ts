import { onMounted } from 'vue'

import { usePlatformAwdChallenges } from './usePlatformAwdChallenges'

export function useAwdChallengeImportPage() {
  const {
    uploading,
    queueLoading,
    importQueue,
    uploadResults,
    selectedImportFileName,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
  } = usePlatformAwdChallenges()

  onMounted(() => {
    void refreshImportQueue()
  })

  return {
    uploading,
    queueLoading,
    importQueue,
    uploadResults,
    selectedImportFileName,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
  }
}

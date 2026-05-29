import { onMounted } from 'vue'

import { usePlatformAwdChallenges } from './usePlatformAwdChallenges'

interface PlatformAwdChallengeImportRouteTarget {
  name: 'PlatformAwdChallengeImport'
}

export function useAwdChallengeLibraryPage() {
  const {
    list,
    total,
    page,
    pageSize,
    loading,
    refresh,
    changePage,
    keyword,
    serviceTypeFilter,
    statusFilter,
    dialogOpen,
    dialogMode,
    saving,
    uploading,
    queueLoading,
    importQueue,
    uploadResults,
    selectedImportFileName,
    formDraft,
    openEditDialog,
    closeDialog,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
    saveChallenge,
    removeChallenge,
  } = usePlatformAwdChallenges()

  onMounted(() => {
    void refresh()
  })

  const importRoute: PlatformAwdChallengeImportRouteTarget = { name: 'PlatformAwdChallengeImport' }

  function updateKeyword(value: string) {
    keyword.value = value
  }

  function updateServiceTypeFilter(value: typeof serviceTypeFilter.value) {
    serviceTypeFilter.value = value
  }

  function updateStatusFilter(value: typeof statusFilter.value) {
    statusFilter.value = value
  }

  function handleDialogOpenChange(value: boolean) {
    if (!value) {
      closeDialog()
    }
  }

  return {
    list,
    total,
    page,
    pageSize,
    loading,
    refresh,
    changePage,
    keyword,
    serviceTypeFilter,
    statusFilter,
    dialogOpen,
    dialogMode,
    saving,
    uploading,
    queueLoading,
    importQueue,
    uploadResults,
    selectedImportFileName,
    formDraft,
    openEditDialog,
    refreshImportQueue,
    selectImportPackages,
    commitImportPreview,
    saveChallenge,
    removeChallenge,
    updateKeyword,
    updateServiceTypeFilter,
    updateStatusFilter,
    handleDialogOpenChange,
    importRoute,
  }
}

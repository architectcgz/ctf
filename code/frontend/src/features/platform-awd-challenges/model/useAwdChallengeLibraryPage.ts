import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { usePlatformAwdChallenges } from './usePlatformAwdChallenges'

export function useAwdChallengeLibraryPage() {
  const router = useRouter()
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

  function openImportPage(): void {
    void router.push({ name: 'PlatformAwdChallengeImport' })
  }

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
    openImportPage,
  }
}

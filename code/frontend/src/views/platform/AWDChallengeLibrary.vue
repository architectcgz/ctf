<script setup lang="ts">
import { onMounted } from 'vue'

import AWDChallengeEditorDialog from '@/components/platform/awd-service/AWDChallengeEditorDialog.vue'
import AWDChallengeLibraryPage from '@/components/platform/awd-service/AWDChallengeLibraryPage.vue'
import { useAwdChallengeLibraryPage, usePlatformAwdChallenges } from '@/features/platform-awd-challenges'
const { openImportPage } = useAwdChallengeLibraryPage()

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

</script>

<template>
  <div>
    <AWDChallengeLibraryPage
      mode="library"
      :list="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :keyword="keyword"
      :service-type-filter="serviceTypeFilter"
      :status-filter="statusFilter"
      :uploading="uploading"
      :queue-loading="queueLoading"
      :import-queue="importQueue"
      :upload-results="uploadResults"
      :selected-file-name="selectedImportFileName"
      @refresh="refresh"
      @update-keyword="updateKeyword"
      @update-service-type-filter="updateServiceTypeFilter"
      @update-status-filter="updateStatusFilter"
      @refresh-import-queue="refreshImportQueue"
      @select-import-packages="selectImportPackages"
      @commit-import="commitImportPreview"
      @open-import-page="openImportPage"
      @open-edit-dialog="openEditDialog"
      @delete-challenge="removeChallenge"
      @change-page="changePage"
    />

    <AWDChallengeEditorDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :draft="formDraft"
      :saving="saving"
      @update:open="handleDialogOpenChange"
      @save="saveChallenge"
    />
  </div>
</template>

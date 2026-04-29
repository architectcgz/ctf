<script setup lang="ts">
import { onMounted } from 'vue'

import AWDChallengeLibraryPage from '@/components/platform/awd-service/AWDChallengeLibraryPage.vue'
import { usePlatformAwdChallenges } from '@/composables/usePlatformAwdChallenges'

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
</script>

<template>
  <AWDChallengeLibraryPage
    mode="import"
    :list="[]"
    :total="0"
    :page="1"
    :page-size="20"
    :loading="false"
    keyword=""
    service-type-filter=""
    status-filter=""
    :uploading="uploading"
    :queue-loading="queueLoading"
    :import-queue="importQueue"
    :upload-results="uploadResults"
    :selected-file-name="selectedImportFileName"
    @refresh-import-queue="refreshImportQueue"
    @select-import-packages="selectImportPackages"
    @commit-import="commitImportPreview"
  />
</template>

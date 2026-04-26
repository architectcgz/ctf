<script setup lang="ts">
import { onMounted } from 'vue'

import AWDServiceTemplateLibraryPage from '@/components/platform/awd-service/AWDServiceTemplateLibraryPage.vue'
import { usePlatformAwdServiceTemplates } from '@/composables/usePlatformAwdServiceTemplates'

const {
  uploading,
  queueLoading,
  importQueue,
  uploadResults,
  selectedImportFileName,
  refreshImportQueue,
  selectImportPackages,
  commitImportPreview,
} = usePlatformAwdServiceTemplates()

onMounted(() => {
  void refreshImportQueue()
})
</script>

<template>
  <AWDServiceTemplateLibraryPage
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

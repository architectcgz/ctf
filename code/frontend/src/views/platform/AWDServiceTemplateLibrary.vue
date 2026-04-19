<script setup lang="ts">
import { onMounted } from 'vue'

import AWDServiceTemplateEditorDialog from '@/components/platform/awd-service/AWDServiceTemplateEditorDialog.vue'
import AWDServiceTemplateLibraryPage from '@/components/platform/awd-service/AWDServiceTemplateLibraryPage.vue'
import { usePlatformAwdServiceTemplates } from '@/composables/usePlatformAwdServiceTemplates'

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
  formDraft,
  openCreateDialog,
  openEditDialog,
  closeDialog,
  saveTemplate,
  removeTemplate,
} = usePlatformAwdServiceTemplates()

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
  <div class="space-y-6">
    <AWDServiceTemplateLibraryPage
      :list="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :keyword="keyword"
      :service-type-filter="serviceTypeFilter"
      :status-filter="statusFilter"
      @refresh="refresh"
      @update-keyword="updateKeyword"
      @update-service-type-filter="updateServiceTypeFilter"
      @update-status-filter="updateStatusFilter"
      @open-create-dialog="openCreateDialog"
      @open-edit-dialog="openEditDialog"
      @delete-template="removeTemplate"
      @change-page="changePage"
    />

    <AWDServiceTemplateEditorDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :draft="formDraft"
      :saving="saving"
      @update:open="handleDialogOpenChange"
      @save="saveTemplate"
    />
  </div>
</template>

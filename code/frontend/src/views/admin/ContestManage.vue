<script setup lang="ts">
import { onMounted } from 'vue'

import AdminContestFormDialog from '@/components/admin/contest/AdminContestFormDialog.vue'
import ContestOrchestrationPage from '@/components/admin/contest/ContestOrchestrationPage.vue'
import { useAdminContests } from '@/composables/useAdminContests'

const {
  list,
  total,
  page,
  pageSize,
  loading,
  refresh,
  changePage,
  statusFilter,
  dialogOpen,
  mode,
  saving,
  formDraft,
  fieldLocks,
  statusOptions,
  openCreateDialog,
  openEditDialog,
  closeDialog,
  saveContest,
} = useAdminContests()

onMounted(() => {
  void refresh()
})

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
    <ContestOrchestrationPage
      :list="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :status-filter="statusFilter"
      @refresh="refresh"
      @open-create-dialog="openCreateDialog"
      @update-status-filter="updateStatusFilter"
      @open-edit-dialog="openEditDialog"
      @change-page="changePage"
    />

    <AdminContestFormDialog
      :open="dialogOpen"
      :mode="mode"
      :draft="formDraft"
      :saving="saving"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @update:open="handleDialogOpenChange"
      @save="saveContest"
    />
  </div>
</template>

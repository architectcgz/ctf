<script setup lang="ts">
import PlatformUserFormDialog from '@/components/platform/user/PlatformUserFormDialog.vue'
import UserGovernancePage from '@/components/platform/user/UserGovernancePage.vue'
import { usePlatformUserManagePage } from '@/features/platform-users'

const {
  list,
  total,
  page,
  pageSize,
  loading,
  refresh,
  changePage,
  keyword,
  studentNo,
  teacherNo,
  roleFilter,
  statusFilter,
  dialogOpen,
  dialogMode,
  saving,
  formDraft,
  importResult,
  openCreateDialog,
  openEditDialog,
  saveUser,
  importUserFile,
  updateKeyword,
  updateRoleFilter,
  updateStudentNo,
  updateTeacherNo,
  updateStatusFilter,
  handleDelete,
  handleDialogOpenChange,
} = usePlatformUserManagePage()
</script>

<template>
  <div class="space-y-6">
    <UserGovernancePage
      :list="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :keyword="keyword"
      :student-no="studentNo"
      :teacher-no="teacherNo"
      :role-filter="roleFilter"
      :status-filter="statusFilter"
      :import-result="importResult"
      @refresh="refresh"
      @update-keyword="updateKeyword"
      @update-student-no="updateStudentNo"
      @update-teacher-no="updateTeacherNo"
      @update-role-filter="updateRoleFilter"
      @update-status-filter="updateStatusFilter"
      @open-create-dialog="openCreateDialog"
      @open-edit-dialog="openEditDialog"
      @delete-user="handleDelete"
      @change-page="changePage"
      @import-file="importUserFile"
    />

    <PlatformUserFormDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :draft="formDraft"
      :saving="saving"
      @update:open="handleDialogOpenChange"
      @save="saveUser"
    />
  </div>
</template>

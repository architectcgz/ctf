<script setup lang="ts">
import { onMounted } from 'vue'

import PlatformUserFormDialog from '@/components/platform/user/PlatformUserFormDialog.vue'
import UserGovernancePage from '@/components/platform/user/UserGovernancePage.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePlatformUsers } from '@/features/platform-users'

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
  closeDialog,
  saveUser,
  removeUser,
  importUserFile,
} = usePlatformUsers()

onMounted(() => {
  void refresh()
})

function updateKeyword(value: string) {
  keyword.value = value
}

function updateRoleFilter(value: typeof roleFilter.value) {
  roleFilter.value = value
}

function updateStudentNo(value: string) {
  studentNo.value = value
}

function updateTeacherNo(value: string) {
  teacherNo.value = value
}

function updateStatusFilter(value: typeof statusFilter.value) {
  statusFilter.value = value
}

async function handleDelete(userId: string) {
  const user = list.value.find((item) => item.id === userId)
  if (!user) {
    return
  }

  const confirmed = await confirmDestructiveAction({
    title: '删除用户',
    message: `确定删除用户 ${user.username} 吗？`,
    confirmButtonText: '确认删除',
  })
  if (!confirmed) {
    return
  }
  await removeUser(user)
}

function handleDialogOpenChange(value: boolean) {
  if (!value) {
    closeDialog()
  }
}
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

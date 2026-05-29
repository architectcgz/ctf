import { computed, onMounted } from 'vue'
import { type LocationQueryRaw, useRoute, useRouter } from 'vue-router'

import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import {
  buildUserGovernancePanelQuery,
  resolveUserGovernancePanel,
  type UserPanelKey,
} from './useUserGovernancePanelRoute'
import { usePlatformUsers } from './usePlatformUsers'

export function usePlatformUserManagePage() {
  const route = useRoute()
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
  const activePanel = computed<UserPanelKey>(() => resolveUserGovernancePanel(route.query.panel))

  onMounted(() => {
    void refresh()
  })

  async function switchPanel(panel: UserPanelKey): Promise<void> {
    if (activePanel.value === panel) {
      return
    }

    await router.replace({
      name: 'UserManage',
      query: buildUserGovernancePanelQuery(route.query, panel) as LocationQueryRaw,
    })
  }

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

  return {
    list,
    total,
    page,
    pageSize,
    loading,
    activePanel,
    refresh,
    changePage,
    switchPanel,
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
  }
}

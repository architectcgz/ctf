import { computed, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import {
  createUser,
  deleteUser,
  getUsers,
  importUsers,
  updateUser,
  type AdminUserCreatePayload,
  type AdminUserUpdatePayload,
} from '@/api/admin'
import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'
import { ApiError } from '@/api/request'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import type { UserRole } from '@/utils/constants'

type UserFilterRole = UserRole | 'all'
type UserFilterStatus = UserStatus | 'all'

export interface PlatformUserFormDraft {
  username: string
  name: string
  password: string
  email: string
  student_no: string
  teacher_no: string
  class_name: string
  role: UserRole
  status: UserStatus
}

function createEmptyDraft(): PlatformUserFormDraft {
  return {
    username: '',
    name: '',
    password: '',
    email: '',
    student_no: '',
    teacher_no: '',
    class_name: '',
    role: 'student',
    status: 'active',
  }
}

export function usePlatformUsers() {
  const toast = useToast()
  const keyword = ref('')
  const studentNo = ref('')
  const teacherNo = ref('')
  const roleFilter = ref<UserFilterRole>('all')
  const statusFilter = ref<UserFilterStatus>('all')
  const dialogOpen = ref(false)
  const saving = ref(false)
  const editingUserId = ref<string | null>(null)
  const formDraft = ref<PlatformUserFormDraft>(createEmptyDraft())
  const importResult = ref<AdminUserImportData | null>(null)

  const pagination = usePagination<AdminUserListItem>(({ page, page_size, signal }) =>
    getUsers({
      page,
      page_size,
      keyword: keyword.value.trim() || undefined,
      student_no: studentNo.value.trim() || undefined,
      teacher_no: teacherNo.value.trim() || undefined,
      role: roleFilter.value === 'all' ? undefined : roleFilter.value,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
    }, {
      signal,
    })
  )

  const dialogMode = computed<'create' | 'edit'>(() => (editingUserId.value ? 'edit' : 'create'))
  type DebouncedRefresh = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const scheduleTextFilterRefresh = useDebounceFn(() => {
    void pagination.changePage(1)
  }, 250) as DebouncedRefresh

  function notifyUserActionError(error: unknown, fallback: string): void {
    console.error(fallback, error)
    if (error instanceof ApiError) {
      return
    }
    const message = error instanceof Error && error.message.trim() ? error.message : fallback
    toast.error(message)
  }

  watch([keyword, studentNo, teacherNo], () => {
    scheduleTextFilterRefresh()
  })

  watch([roleFilter, statusFilter], async () => {
    scheduleTextFilterRefresh.cancel?.()
    await pagination.changePage(1)
  })

  function openCreateDialog() {
    editingUserId.value = null
    formDraft.value = createEmptyDraft()
    dialogOpen.value = true
  }

  function openEditDialog(user: AdminUserListItem) {
    editingUserId.value = user.id
    formDraft.value = {
      username: user.username,
      name: user.name || '',
      password: '',
      email: user.email || '',
      student_no: user.student_no || '',
      teacher_no: user.teacher_no || '',
      class_name: user.class_name || '',
      role: user.roles[0] || 'student',
      status: user.status,
    }
    dialogOpen.value = true
  }

  function closeDialog() {
    dialogOpen.value = false
  }

  async function saveUser(draft: PlatformUserFormDraft) {
    saving.value = true
    try {
      if (editingUserId.value) {
        const payload: AdminUserUpdatePayload = {
          email: draft.email.trim() || undefined,
          name: draft.name.trim() || undefined,
          student_no: draft.student_no.trim() || undefined,
          teacher_no: draft.teacher_no.trim() || undefined,
          class_name: draft.class_name.trim() || undefined,
          role: draft.role,
          status: draft.status,
          password: draft.password.trim() || undefined,
        }
        await updateUser(editingUserId.value, payload)
        toast.success('用户已更新')
      } else {
        const payload: AdminUserCreatePayload = {
          username: draft.username.trim(),
          name: draft.name.trim() || undefined,
          password: draft.password,
          email: draft.email.trim() || undefined,
          student_no: draft.student_no.trim() || undefined,
          teacher_no: draft.teacher_no.trim() || undefined,
          class_name: draft.class_name.trim() || undefined,
          role: draft.role,
          status: draft.status,
        }
        await createUser(payload)
        toast.success('用户已创建')
      }

      dialogOpen.value = false
      await pagination.refresh()
    } catch (error) {
      notifyUserActionError(
        error,
        editingUserId.value ? '用户更新失败，请稍后重试' : '用户创建失败，请稍后重试'
      )
    } finally {
      saving.value = false
    }
  }

  async function removeUser(user: AdminUserListItem) {
    try {
      await deleteUser(user.id)
      toast.success(`已删除用户 ${user.username}`)
      await pagination.refresh()
    } catch (error) {
      notifyUserActionError(error, `删除用户失败：${user.username}`)
    }
  }

  async function importUserFile(file: File) {
    try {
      importResult.value = await importUsers(file)
      toast.success('批量导入已完成')
      await pagination.refresh()
    } catch (error) {
      notifyUserActionError(error, '批量导入失败，请稍后重试')
    }
  }

  return {
    ...pagination,
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
  }
}
